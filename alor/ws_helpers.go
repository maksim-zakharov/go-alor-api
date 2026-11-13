package alor

import (
	"context"
	"encoding/json"
)

/**
 * SubscribeQuotes подписка на котировки инструмента.
 */
func (w *WS) SubscribeQuotes(ctx context.Context, exchange Exchange, code string, handler DataHandler, opts ...SubscriptionOption) (string, error) {
	sub := Subscription{Opcode: OpcodeQuotes, Exchange: exchange, Code: code}
	applySubOpts(&sub, opts)
	return w.Subscribe(ctx, sub, handler)
}

/**
 * SubscribeOrderBook подписка на стакан.
 */
func (w *WS) SubscribeOrderBook(ctx context.Context, exchange Exchange, code string, depth int, handler DataHandler, opts ...SubscriptionOption) (string, error) {
	sub := Subscription{Opcode: OpcodeOrderBook, Exchange: exchange, Code: code, Depth: depth}
	applySubOpts(&sub, opts)
	return w.Subscribe(ctx, sub, handler)
}

/**
 * SubscribeBars подписка на свечи.
 */
func (w *WS) SubscribeBars(ctx context.Context, exchange Exchange, code, tf string, from int64, handler DataHandler, opts ...SubscriptionOption) (string, error) {
	sub := Subscription{Opcode: OpcodeBars, Exchange: exchange, Code: code, TF: tf, From: from}
	applySubOpts(&sub, opts)
	return w.Subscribe(ctx, sub, handler)
}

/**
 * SubscribeAllTrades подписка на ленту всех сделок по инструменту.
 */
func (w *WS) SubscribeAllTrades(ctx context.Context, exchange Exchange, code string, handler DataHandler, opts ...SubscriptionOption) (string, error) {
	sub := Subscription{Opcode: OpcodeAllTrades, Exchange: exchange, Code: code}
	applySubOpts(&sub, opts)
	return w.Subscribe(ctx, sub, handler)
}

/**
 * SubscribeInstruments подписка на изменения инструмента.
 */
func (w *WS) SubscribeInstruments(ctx context.Context, exchange Exchange, code string, handler DataHandler, opts ...SubscriptionOption) (string, error) {
	sub := Subscription{Opcode: OpcodeInstruments, Exchange: exchange, Code: code}
	applySubOpts(&sub, opts)
	return w.Subscribe(ctx, sub, handler)
}

/**
 * SubscribeOrders подписка на заявки портфеля.
 */
func (w *WS) SubscribeOrders(ctx context.Context, exchange Exchange, portfolio string, handler DataHandler, opts ...SubscriptionOption) (string, error) {
	sub := Subscription{Opcode: OpcodeOrders, Exchange: exchange, Portfolio: portfolio}
	applySubOpts(&sub, opts)
	return w.Subscribe(ctx, sub, handler)
}

/**
 * SubscribeStopOrders подписка на стоп-заявки портфеля.
 */
func (w *WS) SubscribeStopOrders(ctx context.Context, exchange Exchange, portfolio string, handler DataHandler, opts ...SubscriptionOption) (string, error) {
	sub := Subscription{Opcode: OpcodeStopOrders, Exchange: exchange, Portfolio: portfolio}
	applySubOpts(&sub, opts)
	return w.Subscribe(ctx, sub, handler)
}

/**
 * SubscribePositions подписка на позиции портфеля.
 */
func (w *WS) SubscribePositions(ctx context.Context, exchange Exchange, portfolio string, handler DataHandler, opts ...SubscriptionOption) (string, error) {
	sub := Subscription{Opcode: OpcodePositions, Exchange: exchange, Portfolio: portfolio}
	applySubOpts(&sub, opts)
	return w.Subscribe(ctx, sub, handler)
}

/**
 * SubscribeSummaries подписка на сводку портфеля.
 */
func (w *WS) SubscribeSummaries(ctx context.Context, exchange Exchange, portfolio string, handler DataHandler, opts ...SubscriptionOption) (string, error) {
	sub := Subscription{Opcode: OpcodeSummaries, Exchange: exchange, Portfolio: portfolio}
	applySubOpts(&sub, opts)
	return w.Subscribe(ctx, sub, handler)
}

/**
 * SubscribeRisks подписка на портфельные риски.
 */
func (w *WS) SubscribeRisks(ctx context.Context, exchange Exchange, portfolio string, handler DataHandler, opts ...SubscriptionOption) (string, error) {
	sub := Subscription{Opcode: OpcodeRisks, Exchange: exchange, Portfolio: portfolio}
	applySubOpts(&sub, opts)
	return w.Subscribe(ctx, sub, handler)
}

/**
 * SubscribeSpectraRisks подписка на риски FORTS.
 */
func (w *WS) SubscribeSpectraRisks(ctx context.Context, exchange Exchange, portfolio string, handler DataHandler, opts ...SubscriptionOption) (string, error) {
	sub := Subscription{Opcode: OpcodeSpectraRisks, Exchange: exchange, Portfolio: portfolio}
	applySubOpts(&sub, opts)
	return w.Subscribe(ctx, sub, handler)
}

/**
 * SubscribeTrades подписка на сделки портфеля.
 */
func (w *WS) SubscribeTrades(ctx context.Context, exchange Exchange, portfolio string, handler DataHandler, opts ...SubscriptionOption) (string, error) {
	sub := Subscription{Opcode: OpcodeTrades, Exchange: exchange, Portfolio: portfolio}
	applySubOpts(&sub, opts)
	return w.Subscribe(ctx, sub, handler)
}

// SubscriptionOption — опция подписки.
type SubscriptionOption func(*Subscription)

// WithInstrumentGroup задаёт режим торгов (борд).
func WithInstrumentGroup(group string) SubscriptionOption {
	return func(s *Subscription) { s.InstrumentGroup = group }
}

// WithFrequency задаёт частоту обновлений (мс).
func WithFrequency(ms int) SubscriptionOption {
	return func(s *Subscription) { s.Frequency = ms }
}

// WithFormat задаёт формат данных подписки.
func WithFormat(f Format) SubscriptionOption {
	return func(s *Subscription) { s.Format = f }
}

// WithGUID задаёт фиксированный guid подписки.
func WithGUID(guid string) SubscriptionOption {
	return func(s *Subscription) { s.GUID = guid }
}

// WithDepth задаёт глубину (стакан / alltrades).
func WithDepth(depth int) SubscriptionOption {
	return func(s *Subscription) { s.Depth = depth }
}

// WithCode задаёт код инструмента (для портфельных подписок с фильтром).
func WithCode(code string) SubscriptionOption {
	return func(s *Subscription) { s.Code = code }
}

func applySubOpts(sub *Subscription, opts []SubscriptionOption) {
	for _, opt := range opts {
		opt(sub)
	}
}

/**
 * DecodeData разбирает data-сообщение подписки в типизированную структуру.
 */
func DecodeData[T any](data json.RawMessage) (T, error) {
	var out T
	err := json.Unmarshal(data, &out)
	return out, err
}
