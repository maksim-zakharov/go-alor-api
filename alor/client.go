package alor

import (
	"context"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Client — HTTP + WebSocket клиент Alor OpenAPI V2.
type Client struct {
	cfg  Config
	http *http.Client

	limiter *rate.Limiter

	tokenMu    sync.RWMutex
	refreshMu  sync.Mutex
	access     string
	accessExp  time.Time
	portfolios []string

	ws        *wsConn
	cws       *cwsConn
	closed    chan struct{}
	closeOnce sync.Once
}

// Option — функциональная опция конструктора.
type Option func(*Client)

// WithHTTPClient подменяет HTTP-клиент.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.http = h
		}
	}
}

/**
 * New создаёт клиент.
 *
 * Для авторизованных запросов нужен Config.RefreshToken или заранее SetAccessToken.
 */
func New(cfg Config, opts ...Option) *Client {
	cfg = cfg.Resolve()
	c := &Client{
		cfg:     cfg,
		http:    &http.Client{Timeout: cfg.HTTPTimeout},
		limiter: rate.NewLimiter(rate.Limit(cfg.RestRPS), cfg.RestBurst),
		closed:  make(chan struct{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	if cfg.AccessToken != "" {
		exp := time.Now().Add(25 * time.Minute)
		if cfg.AccessTokenTTL > 0 {
			exp = time.Now().Add(cfg.AccessTokenTTL)
		}
		c.SetAccessToken(cfg.AccessToken, exp)
	}
	c.ws = newWSConn(c)
	c.cws = newCWSConn(c)
	return c
}

/**
 * NewProduction создаёт клиент боевого контура с refresh-токеном.
 */
func NewProduction(refreshToken string, opts ...Option) *Client {
	return New(Config{Environment: EnvProduction, RefreshToken: refreshToken}, opts...)
}

/**
 * NewDev создаёт клиент тестового контура с refresh-токеном.
 */
func NewDev(refreshToken string, opts ...Option) *Client {
	return New(Config{Environment: EnvDev, RefreshToken: refreshToken}, opts...)
}

// Config возвращает копию текущей конфигурации.
func (c *Client) Config() Config {
	return c.cfg
}

// Close останавливает WebSocket-соединения и помечает клиент закрытым.
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)
		if c.ws != nil {
			c.ws.close()
		}
		if c.cws != nil {
			c.cws.close()
		}
	})
}

func (c *Client) isClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

// WS возвращает менеджер подписок /ws.
func (c *Client) WS() *WS {
	return &WS{c: c}
}

// CWS возвращает менеджер торговых команд /cws.
func (c *Client) CWS() *CWS {
	return &CWS{c: c}
}

/**
 * Run запускает фоновый refresh JWT и удержание WS/CWS (если уже были подписки/команды).
 * Блокирует до отмены ctx или Close.
 */
func (c *Client) Run(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		if _, err := c.accessToken(ctx); err != nil && err != ErrNoToken {
			// не фатально: повторим
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.closed:
			return ErrClosed
		case <-ticker.C:
		}
	}
}
