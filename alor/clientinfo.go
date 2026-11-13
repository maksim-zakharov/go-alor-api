package alor

import (
	"context"
	"fmt"
	"net/url"
)

/**
 * GetOrders возвращает все заявки портфеля.
 */
func (c *Client) GetOrders(ctx context.Context, exchange Exchange, portfolio string, q Query) (Orders, error) {
	var out Orders
	path := fmt.Sprintf("/md/v2/Clients/%s/%s/orders", exchange, url.PathEscape(portfolio))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetOrder возвращает заявку по ID.
 */
func (c *Client) GetOrder(ctx context.Context, exchange Exchange, portfolio string, orderID int64, q Query) (*Order, error) {
	var out Order
	path := fmt.Sprintf("/md/v2/Clients/%s/%s/orders/%d", exchange, url.PathEscape(portfolio), orderID)
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetStopOrders возвращает стоп-заявки портфеля.
 */
func (c *Client) GetStopOrders(ctx context.Context, exchange Exchange, portfolio string, q Query) (StopOrders, error) {
	var out StopOrders
	path := fmt.Sprintf("/md/v2/Clients/%s/%s/stoporders", exchange, url.PathEscape(portfolio))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetStopOrder возвращает стоп-заявку по ID.
 */
func (c *Client) GetStopOrder(ctx context.Context, exchange Exchange, portfolio string, orderID int64, q Query) (*StopOrder, error) {
	var out StopOrder
	path := fmt.Sprintf("/md/v2/Clients/%s/%s/stoporders/%d", exchange, url.PathEscape(portfolio), orderID)
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetMoney возвращает информацию о деньгах портфеля (legacy).
 */
func (c *Client) GetMoney(ctx context.Context, exchange Exchange, portfolio string, q Query) (*Money, error) {
	var out Money
	path := fmt.Sprintf("/md/v2/Clients/legacy/%s/%s/money", exchange, url.PathEscape(portfolio))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetSummary возвращает сводку по портфелю.
 */
func (c *Client) GetSummary(ctx context.Context, exchange Exchange, portfolio string, q Query) (*Summary, error) {
	var out Summary
	path := fmt.Sprintf("/md/v2/Clients/%s/%s/summary", exchange, url.PathEscape(portfolio))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetPositions возвращает все позиции портфеля.
 */
func (c *Client) GetPositions(ctx context.Context, exchange Exchange, portfolio string, q Query) (Positions, error) {
	var out Positions
	path := fmt.Sprintf("/md/v2/Clients/%s/%s/positions", exchange, url.PathEscape(portfolio))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetPosition возвращает позицию по инструменту.
 */
func (c *Client) GetPosition(ctx context.Context, exchange Exchange, portfolio, symbol string, q Query) (*Position, error) {
	var out Position
	path := fmt.Sprintf("/md/v2/Clients/%s/%s/positions/%s", exchange, url.PathEscape(portfolio), url.PathEscape(symbol))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetTrades возвращает сделки портфеля за текущую сессию.
 */
func (c *Client) GetTrades(ctx context.Context, exchange Exchange, portfolio string, q Query) (TradesV2, error) {
	var out TradesV2
	path := fmt.Sprintf("/md/v2/Clients/%s/%s/trades", exchange, url.PathEscape(portfolio))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetSymbolTrades возвращает сделки портфеля по инструменту за текущую сессию.
 */
func (c *Client) GetSymbolTrades(ctx context.Context, exchange Exchange, portfolio, symbol string, q Query) (TradesV2, error) {
	var out TradesV2
	path := fmt.Sprintf("/md/v2/Clients/%s/%s/%s/trades", exchange, url.PathEscape(portfolio), url.PathEscape(symbol))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetFortsRisk возвращает риски срочного рынка.
 */
func (c *Client) GetFortsRisk(ctx context.Context, exchange Exchange, portfolio string, q Query) (*FortsRisk, error) {
	var out FortsRisk
	path := fmt.Sprintf("/md/v2/Clients/%s/%s/fortsrisk", exchange, url.PathEscape(portfolio))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetRisk возвращает портфельные риски.
 */
func (c *Client) GetRisk(ctx context.Context, exchange Exchange, portfolio string, q Query) (*Risk, error) {
	var out Risk
	path := fmt.Sprintf("/md/v2/Clients/%s/%s/risk", exchange, url.PathEscape(portfolio))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * GetLoginPositions возвращает позиции по логину.
 */
func (c *Client) GetLoginPositions(ctx context.Context, login string, q Query) (Positions, error) {
	var out Positions
	path := fmt.Sprintf("/md/v2/Clients/%s/positions", url.PathEscape(login))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetHistoryTrades возвращает историю всех сделок (md/stats).
 */
func (c *Client) GetHistoryTrades(ctx context.Context, exchange Exchange, portfolio string, q Query) (Trades, error) {
	var out Trades
	path := fmt.Sprintf("/md/stats/%s/%s/history/trades", exchange, url.PathEscape(portfolio))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetHistorySymbolTrades возвращает историю сделок по инструменту (md/stats).
 */
func (c *Client) GetHistorySymbolTrades(ctx context.Context, exchange Exchange, portfolio, symbol string, q Query) (Trades, error) {
	var out Trades
	path := fmt.Sprintf("/md/stats/%s/%s/history/trades/%s", exchange, url.PathEscape(portfolio), url.PathEscape(symbol))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetStatsHistoryTrades возвращает сделки прошлых сессий (md/v2/Stats).
 */
func (c *Client) GetStatsHistoryTrades(ctx context.Context, exchange Exchange, portfolio string, q Query) (TradesV2, error) {
	var out TradesV2
	path := fmt.Sprintf("/md/v2/Stats/%s/%s/history/trades", exchange, url.PathEscape(portfolio))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}

/**
 * GetStatsHistorySymbolTrades возвращает сделки прошлых сессий по инструменту.
 */
func (c *Client) GetStatsHistorySymbolTrades(ctx context.Context, exchange Exchange, portfolio, symbol string, q Query) (TradesV2, error) {
	var out TradesV2
	path := fmt.Sprintf("/md/v2/Stats/%s/%s/history/trades/%s", exchange, url.PathEscape(portfolio), url.PathEscape(symbol))
	if err := c.GetJSON(ctx, path, c.withFormat(q), &out); err != nil {
		return nil, err
	}
	return out, nil
}
