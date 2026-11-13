package alor

import "time"

// Environment — контур Alor OpenAPI.
type Environment string

const (
	// EnvProduction — боевой контур.
	EnvProduction Environment = "production"
	// EnvDev — тестовый (игровой) контур.
	EnvDev Environment = "dev"
)

// Адреса контуров по умолчанию.
const (
	APIURLProduction   = "https://api.alor.ru"
	APIURLDev          = "https://apidev.alor.ru"
	OAuthURLProduction = "https://oauth.alor.ru"
	OAuthURLDev        = "https://oauthdev.alor.ru"
	WSURLProduction    = "wss://api.alor.ru/ws"
	WSURLDev           = "wss://apidev.alor.ru/ws"
	CWSURLProduction   = "wss://api.alor.ru/cws"
	CWSURLDev          = "wss://apidev.alor.ru/cws"
)

// Format — формат сериализации ответов HTTP/WS (Simple | Slim | Heavy).
type Format string

const (
	FormatSimple Format = "Simple"
	FormatSlim   Format = "Slim"
	FormatHeavy  Format = "Heavy"
)

// Exchange — код биржи.
type Exchange string

const (
	ExchangeMOEX Exchange = "MOEX"
	ExchangeSPBX Exchange = "SPBX"
)

// Side — сторона заявки/сделки.
type Side string

const (
	SideBuy  Side = "buy"
	SideSell Side = "sell"
)

// OrderType — тип заявки.
type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

// TimeInForce — период действия заявки.
type TimeInForce string

const (
	TimeInForceOneDay            TimeInForce = "oneday"
	TimeInForceImmediateOrCancel TimeInForce = "immediateorcancel"
	TimeInForceFillOrKill        TimeInForce = "fillorkill"
	TimeInForceGoodTillCancelled TimeInForce = "goodtillcancelled"
)

// Config — настройки клиента.
type Config struct {
	// Environment контур; по умолчанию EnvProduction.
	Environment Environment
	// RefreshToken refresh-токен из кабинета разработчика (обязателен для авторизованных запросов).
	RefreshToken string
	// AccessToken готовый JWT; если задан вместе с AccessTokenTTL — refresh не обязателен.
	AccessToken string
	// AccessTokenTTL срок жизни AccessToken при ручной установке.
	AccessTokenTTL time.Duration

	// APIURL переопределение базового HTTP URL.
	APIURL string
	// OAuthURL переопределение OAuth URL.
	OAuthURL string
	// WSURL переопределение URL WebSocket подписок (/ws).
	WSURL string
	// CWSURL переопределение URL WebSocket команд (/cws).
	CWSURL string

	// DefaultFormat формат ответов по умолчанию (Simple).
	DefaultFormat Format
	// HTTPTimeout таймаут HTTP-клиента.
	HTTPTimeout time.Duration
	/** RestRPS средняя скорость REST (запросов/сек). 0 → 25. */
	RestRPS float64
	/** RestBurst размер burst token-bucket. 0 → 25. */
	RestBurst int

	// RefreshSkew обновлять JWT за это время до истечения. 0 → 60s.
	RefreshSkew time.Duration
	// WSHandshakeTimeout таймаут dial WebSocket. 0 → 15s.
	WSHandshakeTimeout time.Duration
	// WSReconnectMin минимальный backoff reconnect. 0 → 1s.
	WSReconnectMin time.Duration
	// WSReconnectMax максимальный backoff reconnect. 0 → 30s.
	WSReconnectMax time.Duration
	// EnableCompression включать permessage-deflate на /ws (нужно для depth>20).
	EnableCompression *bool
}

// Resolve заполняет URL по Environment и значения по умолчанию.
func (c Config) Resolve() Config {
	out := c
	if out.Environment == "" {
		out.Environment = EnvProduction
	}
	if out.APIURL == "" {
		if out.Environment == EnvDev {
			out.APIURL = APIURLDev
		} else {
			out.APIURL = APIURLProduction
		}
	}
	if out.OAuthURL == "" {
		if out.Environment == EnvDev {
			out.OAuthURL = OAuthURLDev
		} else {
			out.OAuthURL = OAuthURLProduction
		}
	}
	if out.WSURL == "" {
		if out.Environment == EnvDev {
			out.WSURL = WSURLDev
		} else {
			out.WSURL = WSURLProduction
		}
	}
	if out.CWSURL == "" {
		if out.Environment == EnvDev {
			out.CWSURL = CWSURLDev
		} else {
			out.CWSURL = CWSURLProduction
		}
	}
	if out.DefaultFormat == "" {
		out.DefaultFormat = FormatSimple
	}
	if out.HTTPTimeout == 0 {
		out.HTTPTimeout = 30 * time.Second
	}
	if out.RestRPS == 0 {
		out.RestRPS = 25
	}
	if out.RestBurst == 0 {
		out.RestBurst = 25
	}
	if out.RefreshSkew == 0 {
		out.RefreshSkew = 60 * time.Second
	}
	if out.WSHandshakeTimeout == 0 {
		out.WSHandshakeTimeout = 15 * time.Second
	}
	if out.WSReconnectMin == 0 {
		out.WSReconnectMin = time.Second
	}
	if out.WSReconnectMax == 0 {
		out.WSReconnectMax = 30 * time.Second
	}
	if out.EnableCompression == nil {
		v := true
		out.EnableCompression = &v
	}
	return out
}
