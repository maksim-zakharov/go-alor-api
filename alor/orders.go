package alor

import (
	"context"
	"fmt"
)

// LimitOrderRequest — тело создания/изменения лимитной заявки (Command API).
type LimitOrderRequest struct {
	Side            Side        `json:"side"`
	Quantity        int32       `json:"quantity"`
	Price           float64     `json:"price"`
	Instrument      Instrument  `json:"instrument"`
	User            UserRef     `json:"user"`
	Comment         string      `json:"comment,omitempty"`
	TimeInForce     TimeInForce `json:"timeInForce,omitempty"`
	AllowMargin     bool        `json:"allowMargin,omitempty"`
	IcebergFixed    int32       `json:"icebergFixed,omitempty"`
	IcebergVariance int32       `json:"icebergVariance,omitempty"`
	CheckDuplicates bool        `json:"checkDuplicates,omitempty"`
}

// MarketOrderRequest — тело рыночной заявки.
type MarketOrderRequest struct {
	Side            Side        `json:"side"`
	Quantity        int32       `json:"quantity"`
	Instrument      Instrument  `json:"instrument"`
	User            UserRef     `json:"user"`
	Comment         string      `json:"comment,omitempty"`
	TimeInForce     TimeInForce `json:"timeInForce,omitempty"`
	AllowMargin     bool        `json:"allowMargin,omitempty"`
	CheckDuplicates bool        `json:"checkDuplicates,omitempty"`
}

// StopOrderRequest — тело стоп-заявки (Command API v2).
type StopOrderRequest struct {
	Side              Side       `json:"side"`
	Quantity          int32      `json:"quantity"`
	Condition         string     `json:"condition"`
	TriggerPrice      float64    `json:"triggerPrice"`
	StopEndUnixTime   int64      `json:"stopEndUnixTime,omitempty"`
	Instrument        Instrument `json:"instrument"`
	User              UserRef    `json:"user"`
	AllowMargin       bool       `json:"allowMargin,omitempty"`
	CheckDuplicates   bool       `json:"checkDuplicates,omitempty"`
	ProtectingSeconds int32      `json:"protectingSeconds,omitempty"`
	Comment           string     `json:"comment,omitempty"`
	Activate          *bool      `json:"activate,omitempty"`
}

// StopLimitOrderRequest — тело стоп-лимитной заявки.
type StopLimitOrderRequest struct {
	Side              Side        `json:"side"`
	Quantity          int32       `json:"quantity"`
	Price             float64     `json:"price"`
	Condition         string      `json:"condition"`
	TriggerPrice      float64     `json:"triggerPrice"`
	StopEndUnixTime   int64       `json:"stopEndUnixTime,omitempty"`
	Instrument        Instrument  `json:"instrument"`
	User              UserRef     `json:"user"`
	TimeInForce       TimeInForce `json:"timeInForce,omitempty"`
	AllowMargin       bool        `json:"allowMargin,omitempty"`
	IcebergFixed      int32       `json:"icebergFixed,omitempty"`
	IcebergVariance   int32       `json:"icebergVariance,omitempty"`
	CheckDuplicates   bool        `json:"checkDuplicates,omitempty"`
	ProtectingSeconds int32       `json:"protectingSeconds,omitempty"`
	Comment           string      `json:"comment,omitempty"`
	Activate          *bool       `json:"activate,omitempty"`
}

// Instrument — символ + биржа (+ группа) для торговых запросов.
type Instrument struct {
	Symbol          string   `json:"symbol"`
	Exchange        Exchange `json:"exchange"`
	InstrumentGroup string   `json:"instrumentGroup,omitempty"`
}

// UserRef — портфель пользователя.
type UserRef struct {
	Portfolio string `json:"portfolio"`
}

// OrderActionResult — ответ на create/update заявки.
type OrderActionResult struct {
	Message     string `json:"message,omitempty"`
	OrderNumber string `json:"orderNumber,omitempty"`
	OrderID     string `json:"orderId,omitempty"`
}

/**
 * CreateLimitOrder создаёт лимитную заявку (Command API).
 */
func (c *Client) CreateLimitOrder(ctx context.Context, req LimitOrderRequest, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	if err := c.PostJSON(ctx, "/commandapi/warptrans/TRADE/v2/client/orders/actions/limit", q, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * CreateMarketOrder создаёт рыночную заявку.
 */
func (c *Client) CreateMarketOrder(ctx context.Context, req MarketOrderRequest, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	if err := c.PostJSON(ctx, "/commandapi/warptrans/TRADE/v2/client/orders/actions/market", q, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * UpdateLimitOrder изменяет лимитную заявку.
 */
func (c *Client) UpdateLimitOrder(ctx context.Context, orderID int64, req LimitOrderRequest, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	path := fmt.Sprintf("/commandapi/warptrans/TRADE/v2/client/orders/actions/limit/%d", orderID)
	if err := c.PutJSON(ctx, path, q, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * UpdateMarketOrder изменяет рыночную заявку.
 */
func (c *Client) UpdateMarketOrder(ctx context.Context, orderID int64, req MarketOrderRequest, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	path := fmt.Sprintf("/commandapi/warptrans/TRADE/v2/client/orders/actions/market/%d", orderID)
	if err := c.PutJSON(ctx, path, q, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * CreateStopOrder создаёт стоп-заявку (Command API).
 */
func (c *Client) CreateStopOrder(ctx context.Context, req StopOrderRequest, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	if err := c.PostJSON(ctx, "/commandapi/warptrans/TRADE/v2/client/orders/actions/stop", q, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * CreateStopLimitOrder создаёт стоп-лимитную заявку (Command API).
 */
func (c *Client) CreateStopLimitOrder(ctx context.Context, req StopLimitOrderRequest, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	if err := c.PostJSON(ctx, "/commandapi/warptrans/TRADE/v2/client/orders/actions/stopLimit", q, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * UpdateStopOrder изменяет стоп-заявку.
 */
func (c *Client) UpdateStopOrder(ctx context.Context, stopOrderID int64, req StopOrderRequest, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	path := fmt.Sprintf("/commandapi/warptrans/TRADE/v2/client/orders/actions/stop/%d", stopOrderID)
	if err := c.PutJSON(ctx, path, q, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * UpdateStopLimitOrder изменяет стоп-лимитную заявку.
 */
func (c *Client) UpdateStopLimitOrder(ctx context.Context, stopOrderID int64, req StopLimitOrderRequest, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	path := fmt.Sprintf("/commandapi/warptrans/TRADE/v2/client/orders/actions/stopLimit/%d", stopOrderID)
	if err := c.PutJSON(ctx, path, q, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * DeleteOrder снимает одну заявку (limit/market).
 * query: portfolio, exchange, stop (bool — стоп-заявка).
 */
func (c *Client) DeleteOrder(ctx context.Context, orderID int64, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	path := fmt.Sprintf("/commandapi/warptrans/TRADE/v2/client/orders/%d", orderID)
	if err := c.DeleteJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * DeleteAllOrders снимает все заявки портфеля.
 * query: portfolio, exchange, stop.
 */
func (c *Client) DeleteAllOrders(ctx context.Context, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	if err := c.DeleteJSON(ctx, "/commandapi/warptrans/TRADE/v2/client/orders/all", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * EstimateOrder оценивает одну заявку.
 */
func (c *Client) EstimateOrder(ctx context.Context, req ObjEstimateOrder, q Query) (*EstimateOrder, error) {
	var out EstimateOrder
	if err := c.PostJSON(ctx, "/commandapi/warptrans/TRADE/v2/client/orders/estimate", q, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/**
 * EstimateOrders оценивает несколько заявок.
 */
func (c *Client) EstimateOrders(ctx context.Context, req []ObjEstimateOrder, q Query) (EstimateOrders, error) {
	var out EstimateOrders
	if err := c.PostJSON(ctx, "/commandapi/warptrans/TRADE/v2/client/orders/estimate/all", q, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// --- Warptrans stop/take (tradeServerCode) ---

/**
 * CreateStopLoss создаёт стоп-лосс через warptrans/{tradeServerCode}.
 */
func (c *Client) CreateStopLoss(ctx context.Context, tradeServerCode string, body any, q Query) (*OrderActionResult, error) {
	return c.postWarpAction(ctx, tradeServerCode, "stopLoss", body, q)
}

/**
 * CreateTakeProfit создаёт take-profit через warptrans.
 */
func (c *Client) CreateTakeProfit(ctx context.Context, tradeServerCode string, body any, q Query) (*OrderActionResult, error) {
	return c.postWarpAction(ctx, tradeServerCode, "takeProfit", body, q)
}

/**
 * CreateTakeProfitLimit создаёт take-profit limit через warptrans.
 */
func (c *Client) CreateTakeProfitLimit(ctx context.Context, tradeServerCode string, body any, q Query) (*OrderActionResult, error) {
	return c.postWarpAction(ctx, tradeServerCode, "takeProfitLimit", body, q)
}

/**
 * CreateStopLossLimit создаёт stop-loss limit через warptrans.
 */
func (c *Client) CreateStopLossLimit(ctx context.Context, tradeServerCode string, body any, q Query) (*OrderActionResult, error) {
	return c.postWarpAction(ctx, tradeServerCode, "stopLossLimit", body, q)
}

/**
 * UpdateStopLoss изменяет стоп-лосс.
 */
func (c *Client) UpdateStopLoss(ctx context.Context, tradeServerCode string, orderID int64, body any, q Query) (*OrderActionResult, error) {
	return c.putWarpAction(ctx, tradeServerCode, "stopLoss", orderID, body, q)
}

/**
 * UpdateTakeProfit изменяет take-profit.
 */
func (c *Client) UpdateTakeProfit(ctx context.Context, tradeServerCode string, orderID int64, body any, q Query) (*OrderActionResult, error) {
	return c.putWarpAction(ctx, tradeServerCode, "takeProfit", orderID, body, q)
}

/**
 * UpdateTakeProfitLimit изменяет take-profit limit.
 */
func (c *Client) UpdateTakeProfitLimit(ctx context.Context, tradeServerCode string, orderID int64, body any, q Query) (*OrderActionResult, error) {
	return c.putWarpAction(ctx, tradeServerCode, "takeProfitLimit", orderID, body, q)
}

/**
 * UpdateStopLossLimit изменяет stop-loss limit.
 */
func (c *Client) UpdateStopLossLimit(ctx context.Context, tradeServerCode string, orderID int64, body any, q Query) (*OrderActionResult, error) {
	return c.putWarpAction(ctx, tradeServerCode, "stopLossLimit", orderID, body, q)
}

/**
 * DeleteWarpOrder снимает стоп-заявку через warptrans/{tradeServerCode}.
 */
func (c *Client) DeleteWarpOrder(ctx context.Context, tradeServerCode string, orderID int64, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	path := fmt.Sprintf("/warptrans/%s/v2/client/orders/%d", tradeServerCode, orderID)
	if err := c.DeleteJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) postWarpAction(ctx context.Context, tradeServerCode, action string, body any, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	path := fmt.Sprintf("/warptrans/%s/v2/client/orders/actions/%s", tradeServerCode, action)
	if err := c.PostJSON(ctx, path, q, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) putWarpAction(ctx context.Context, tradeServerCode, action string, orderID int64, body any, q Query) (*OrderActionResult, error) {
	var out OrderActionResult
	path := fmt.Sprintf("/warptrans/%s/v2/client/orders/actions/%s/%d", tradeServerCode, action, orderID)
	if err := c.PutJSON(ctx, path, q, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
