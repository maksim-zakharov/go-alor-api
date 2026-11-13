package alor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// refreshResponse — тело ответа oauth /refresh.
type refreshResponse struct {
	AccessToken string `json:"AccessToken"`
}

// jwtClaims — минимальный разбор payload JWT (без проверки подписи).
type jwtClaims struct {
	Exp        int64    `json:"exp"`
	Portfolios []string `json:"portfolios"`
	Agreements string   `json:"agreements"`
	Name       string   `json:"name"`
}

// TokenInfo — сведения о текущем access-токене.
type TokenInfo struct {
	/** AccessToken JWT. */
	AccessToken string
	/** ExpiresAt время истечения (из claim exp). */
	ExpiresAt time.Time
	/** Portfolios портфели из JWT. */
	Portfolios []string
}

/**
 * EnsureAccessToken возвращает валидный JWT, при необходимости обновляя через refresh.
 */
func (c *Client) EnsureAccessToken(ctx context.Context) (string, error) {
	return c.accessToken(ctx)
}

/**
 * Refresh принудительно обновляет JWT через refresh-токен.
 */
func (c *Client) Refresh(ctx context.Context) error {
	c.tokenMu.Lock()
	c.access = ""
	c.accessExp = time.Time{}
	c.tokenMu.Unlock()
	_, err := c.refreshAccessToken(ctx)
	return err
}

/**
 * SetWSURL переопределяет URL WebSocket подписок (/ws).
 */
func (c *Client) SetWSURL(wsURL string) {
	if c == nil {
		return
	}
	wsURL = strings.TrimSpace(wsURL)
	if wsURL == "" {
		return
	}
	c.cfg.WSURL = wsURL
}

/**
 * TokenInfo возвращает текущий access-токен и метаданные (без принудительного refresh).
 */
func (c *Client) TokenInfo() TokenInfo {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return TokenInfo{
		AccessToken: c.access,
		ExpiresAt:   c.accessExp,
		Portfolios:  append([]string(nil), c.portfolios...),
	}
}

/**
 * Portfolios возвращает список портфелей из текущего JWT.
 */
func (c *Client) Portfolios(ctx context.Context) ([]string, error) {
	if _, err := c.accessToken(ctx); err != nil {
		return nil, err
	}
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return append([]string(nil), c.portfolios...), nil
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	if c.cfg.AccessTokenFunc != nil {
		return c.cfg.AccessTokenFunc(ctx)
	}
	c.tokenMu.RLock()
	token := c.access
	exp := c.accessExp
	c.tokenMu.RUnlock()
	if token != "" && time.Until(exp) > c.cfg.RefreshSkew {
		return token, nil
	}
	return c.refreshAccessToken(ctx)
}

/**
 * SetAccessTokenFunc задаёт внешний источник JWT (вызывать до активных запросов/WS).
 */
func (c *Client) SetAccessTokenFunc(fn func(ctx context.Context) (string, error)) {
	if c == nil {
		return
	}
	c.cfg.AccessTokenFunc = fn
}

func (c *Client) refreshAccessToken(ctx context.Context) (string, error) {
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()

	// double-check после захвата refreshMu
	c.tokenMu.RLock()
	token := c.access
	exp := c.accessExp
	c.tokenMu.RUnlock()
	if token != "" && time.Until(exp) > c.cfg.RefreshSkew {
		return token, nil
	}

	if c.cfg.RefreshToken == "" {
		if token != "" {
			return token, nil
		}
		return "", ErrNoToken
	}

	endpoint := strings.TrimRight(c.cfg.OAuthURL, "/") + "/refresh?token=" + url.QueryEscape(c.cfg.RefreshToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("alor: refresh token: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", &APIError{
			StatusCode: resp.StatusCode,
			Method:     http.MethodPost,
			Path:       "/refresh",
			Body:       strings.TrimSpace(string(body)),
			Message:    "не удалось обновить JWT",
		}
	}

	access := ""
	var parsed refreshResponse
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.AccessToken != "" {
		access = parsed.AccessToken
	} else {
		// иногда тело — «сырой» JSON-string
		var asString string
		if err := json.Unmarshal(body, &asString); err == nil && asString != "" {
			access = asString
		} else {
			trimmed := strings.TrimSpace(string(body))
			trimmed = strings.Trim(trimmed, `"`)
			if trimmed != "" {
				access = trimmed
			}
		}
	}
	if access == "" {
		return "", fmt.Errorf("alor: пустой AccessToken в ответе /refresh")
	}

	claims, _ := parseJWTClaims(access)
	expAt := time.Now().Add(25 * time.Minute)
	var portfolios []string
	if claims != nil {
		if claims.Exp > 0 {
			expAt = time.Unix(claims.Exp, 0)
		}
		portfolios = claims.Portfolios
	}

	c.tokenMu.Lock()
	c.access = access
	c.accessExp = expAt
	c.portfolios = portfolios
	c.tokenMu.Unlock()
	return access, nil
}

func parseJWTClaims(token string) (*jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("некорректный JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// иногда padding
		payload, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, err
		}
	}
	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}

// SetAccessToken устанавливает JWT вручную (без refresh).
func (c *Client) SetAccessToken(token string, expiresAt time.Time) {
	claims, _ := parseJWTClaims(token)
	var portfolios []string
	if claims != nil {
		portfolios = claims.Portfolios
		if expiresAt.IsZero() && claims.Exp > 0 {
			expiresAt = time.Unix(claims.Exp, 0)
		}
	}
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(25 * time.Minute)
	}
	c.tokenMu.Lock()
	c.access = token
	c.accessExp = expiresAt
	c.portfolios = portfolios
	c.tokenMu.Unlock()
}
