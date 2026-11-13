package alor

import (
	"context"
	"fmt"
	"net/url"
)

/**
 * GetTime возвращает текущее UTC-время сервера (Unix seconds).
 */
func (c *Client) GetTime(ctx context.Context) (int64, error) {
	var t int64
	if err := c.GetJSON(ctx, "/md/v2/time", nil, &t); err != nil {
		return 0, err
	}
	return t, nil
}

/**
 * GetSecurities возвращает список инструментов (опционально с поиском).
 */
func (c *Client) GetSecurities(ctx context.Context, q Query) (Securities, error) {
	var out Securities
	if err := c.GetJSON(ctx, "/md/v2/Securities", c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetSecuritiesByExchange возвращает инструменты выбранной биржи.
 */
func (c *Client) GetSecuritiesByExchange(ctx context.Context, exchange Exchange, q Query) (Securities, error) {
	var out Securities
	path := fmt.Sprintf("/md/v2/Securities/%s", exchange)
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetSecurity возвращает информацию об инструменте.
 */
func (c *Client) GetSecurity(ctx context.Context, exchange Exchange, symbol string, q Query) (*Security, error) {
	var out Security
	path := fmt.Sprintf("/md/v2/Securities/%s/%s", exchange, url.PathEscape(symbol))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetAvailableBoards возвращает список бордов инструмента.
 */
func (c *Client) GetAvailableBoards(ctx context.Context, exchange Exchange, symbol string) ([]string, error) {
	var out []string
	path := fmt.Sprintf("/md/v2/Securities/%s/%s/availableBoards", exchange, url.PathEscape(symbol))
	if err := c.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetAllTrades возвращает сделки по инструменту за текущую сессию.
 */
func (c *Client) GetAllTrades(ctx context.Context, exchange Exchange, symbol string, q Query) (AllTrades, error) {
	var out AllTrades
	path := fmt.Sprintf("/md/v2/Securities/%s/%s/alltrades", exchange, url.PathEscape(symbol))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetAllTradesHistory возвращает исторические сделки по инструменту.
 */
func (c *Client) GetAllTradesHistory(ctx context.Context, exchange Exchange, symbol string, q Query) (*AllTradesHistory, error) {
	var out AllTradesHistory
	path := fmt.Sprintf("/md/v2/Securities/%s/%s/alltrades/history", exchange, url.PathEscape(symbol))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetActualFuturesQuote возвращает котировки ближайшего фьючерса по коду.
 */
func (c *Client) GetActualFuturesQuote(ctx context.Context, exchange Exchange, symbol string, q Query) (*Futures, error) {
	var out Futures
	path := fmt.Sprintf("/md/v2/Securities/%s/%s/actualFuturesQuote", exchange, url.PathEscape(symbol))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetQuotes возвращает котировки для списка символов (через запятую: MOEX:SBER,MOEX:GAZP).
 * Символы подставляются в path as-is (как в доке Alor); не экранируйте вручную.
 */
func (c *Client) GetQuotes(ctx context.Context, symbols string, q Query) (Symbols, error) {
	var out Symbols
	path := "/md/v2/Securities/" + symbols + "/quotes"
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetCurrencyPairs возвращает список валютных пар.
 */
func (c *Client) GetCurrencyPairs(ctx context.Context, q Query) (CurrencyPairs, error) {
	var out CurrencyPairs
	if err := c.GetJSON(ctx, "/md/v2/Securities/currencyPairs", c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetOrderBook возвращает биржевой стакан.
 */
func (c *Client) GetOrderBook(ctx context.Context, exchange Exchange, symbol string, q Query) (*OrderBook, error) {
	var out OrderBook
	path := fmt.Sprintf("/md/v2/orderbooks/%s/%s", exchange, url.PathEscape(symbol))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetHistory возвращает свечи (историю) по инструменту.
 * Обязательные параметры в q: symbol, exchange, tf, from, to (или часть из них по доке).
 */
func (c *Client) GetHistory(ctx context.Context, q Query) (*History, error) {
	var out History
	if err := c.GetJSON(ctx, "/md/v2/history", c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetRiskRates возвращает ставки риска.
 */
func (c *Client) GetRiskRates(ctx context.Context, q Query) (*RiskRates, error) {
	var out RiskRates
	if err := c.GetJSON(ctx, "/md/v2/risk/rates", c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetMarketStatus возвращает статус торгов по бирже и рынку.
 */
func (c *Client) GetMarketStatus(ctx context.Context, exchange Exchange, market string, q Query) (*TradingSession, error) {
	var out TradingSession
	path := fmt.Sprintf("/md/status/%s/%s", exchange, url.PathEscape(market))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
