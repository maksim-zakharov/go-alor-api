package alor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Query — удобный builder query-параметров.
type Query map[string]any

func (q Query) values() url.Values {
	out := url.Values{}
	for k, v := range q {
		if v == nil {
			continue
		}
		switch t := v.(type) {
		case string:
			if t != "" {
				out.Set(k, t)
			}
		case Format:
			if t != "" {
				out.Set(k, string(t))
			}
		case Exchange:
			if t != "" {
				out.Set(k, string(t))
			}
		case bool:
			if t {
				out.Set(k, "true")
			} else {
				out.Set(k, "false")
			}
		case *bool:
			if t != nil {
				if *t {
					out.Set(k, "true")
				} else {
					out.Set(k, "false")
				}
			}
		default:
			out.Set(k, fmt.Sprint(v))
		}
	}
	return out
}

func (c *Client) withFormat(q Query) Query {
	if q == nil {
		q = Query{}
	}
	if _, ok := q["format"]; !ok && c.cfg.DefaultFormat != "" {
		q["format"] = c.cfg.DefaultFormat
	}
	return q
}

/**
 * Do выполняет HTTP-запрос к API с Bearer-токеном и rate-limit.
 * body — любой JSON-сериализуемый объект или []byte / io.Reader.
 */
func (c *Client) Do(ctx context.Context, method, path string, query Query, body any, out any) error {
	if c.isClosed() {
		return ErrClosed
	}
	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}

	var reader io.Reader
	if body != nil {
		switch b := body.(type) {
		case []byte:
			reader = bytes.NewReader(b)
		case io.Reader:
			reader = b
		default:
			raw, err := json.Marshal(body)
			if err != nil {
				return err
			}
			reader = bytes.NewReader(raw)
		}
	}

	endpoint := strings.TrimRight(c.cfg.APIURL, "/") + path
	if vals := query.values(); len(vals) > 0 {
		endpoint += "?" + vals.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// X-ALOR-REQID для идемпотентности торговых операций
	if strings.Contains(path, "/orders") || strings.Contains(path, "/orderGroups") {
		req.Header.Set("X-ALOR-REQID", uuid.NewString())
	}

	needsAuth := true
	// публичные ручки без токена допустимы (с задержкой 15 мин)
	if path == "/md/v2/time" {
		needsAuth = false
	}
	if needsAuth || c.cfg.RefreshToken != "" || c.access != "" {
		token, err := c.accessToken(ctx)
		if err != nil && needsAuth {
			return err
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		// один повтор после принудительного refresh
		c.tokenMu.Lock()
		c.access = ""
		c.accessExp = time.Time{}
		c.tokenMu.Unlock()
		if c.cfg.RefreshToken != "" {
			token, rerr := c.refreshAccessToken(ctx)
			if rerr == nil {
				req2, _ := http.NewRequestWithContext(ctx, method, endpoint, cloneBody(body))
				if req2 != nil {
					req2.Header = req.Header.Clone()
					req2.Header.Set("Authorization", "Bearer "+token)
					if strings.Contains(path, "/orders") || strings.Contains(path, "/orderGroups") {
						req2.Header.Set("X-ALOR-REQID", uuid.NewString())
					}
					resp2, err2 := c.http.Do(req2)
					if err2 == nil {
						defer resp2.Body.Close()
						raw, _ = io.ReadAll(resp2.Body)
						resp = resp2
					}
				}
			}
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := ""
		var m struct {
			Message string `json:"message"`
			Detail  string `json:"detail"`
			Title   string `json:"title"`
		}
		if json.Unmarshal(raw, &m) == nil {
			msg = firstNonEmpty(m.Message, m.Detail, m.Title)
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Body:       strings.TrimSpace(string(raw)),
			Message:    msg,
		}
	}
	if out == nil || len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("alor: decode %s %s: %w; body=%s", method, path, err, truncate(string(raw), 200))
	}
	return nil
}

// GetJSON — GET с разбором JSON.
func (c *Client) GetJSON(ctx context.Context, path string, query Query, out any) error {
	return c.Do(ctx, http.MethodGet, path, query, nil, out)
}

// PostJSON — POST с JSON-телом.
func (c *Client) PostJSON(ctx context.Context, path string, query Query, body, out any) error {
	return c.Do(ctx, http.MethodPost, path, query, body, out)
}

// PutJSON — PUT с JSON-телом.
func (c *Client) PutJSON(ctx context.Context, path string, query Query, body, out any) error {
	return c.Do(ctx, http.MethodPut, path, query, body, out)
}

// DeleteJSON — DELETE.
func (c *Client) DeleteJSON(ctx context.Context, path string, query Query, out any) error {
	return c.Do(ctx, http.MethodDelete, path, query, nil, out)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func cloneBody(body any) io.Reader {
	if body == nil {
		return nil
	}
	switch b := body.(type) {
	case []byte:
		return bytes.NewReader(b)
	default:
		raw, err := json.Marshal(body)
		if err != nil {
			return nil
		}
		return bytes.NewReader(raw)
	}
}
