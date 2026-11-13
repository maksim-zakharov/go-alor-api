package alor

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WSOpcode — код операции подписки /ws.
type WSOpcode string

const (
	OpcodeAllTrades            WSOpcode = "AllTradesGetAndSubscribe"
	OpcodeBars                 WSOpcode = "BarsGetAndSubscribe"
	OpcodeInstruments          WSOpcode = "InstrumentsGetAndSubscribeV2"
	OpcodeOrderBook            WSOpcode = "OrderBookGetAndSubscribe"
	OpcodeOrders               WSOpcode = "OrdersGetAndSubscribeV2"
	OpcodePositions            WSOpcode = "PositionsGetAndSubscribeV2"
	OpcodeQuotes               WSOpcode = "QuotesSubscribe"
	OpcodeRisks                WSOpcode = "RisksGetAndSubscribe"
	OpcodeSpectraRisks         WSOpcode = "SpectraRisksGetAndSubscribe"
	OpcodeStopOrders           WSOpcode = "StopOrdersGetAndSubscribeV2"
	OpcodeStopOrdersDeprecated WSOpcode = "StopOrdersGetAndSubscribe"
	OpcodeSummaries            WSOpcode = "SummariesGetAndSubscribeV2"
	OpcodeTrades               WSOpcode = "TradesGetAndSubscribeV2"
	OpcodeUnsubscribe          WSOpcode = "unsubscribe"
)

// DataHandler — обработчик data-сообщений подписки (data уже без обёртки guid/data).
type DataHandler func(guid string, data json.RawMessage)

// WSMessage — сырое сообщение /ws.
type WSMessage struct {
	GUID        string          `json:"guid"`
	Data        json.RawMessage `json:"data,omitempty"`
	HTTPCode    int             `json:"httpCode,omitempty"`
	Message     string          `json:"message,omitempty"`
	RequestGUID string          `json:"requestGuid,omitempty"`
}

// Subscription — параметры подписки /ws.
type Subscription struct {
	Opcode               WSOpcode
	Exchange             Exchange
	Code                 string
	InstrumentGroup      string
	Portfolio            string
	Depth                int
	Frequency            int
	Format               Format
	From                 int64
	TF                   string
	IncludeVirtualTrades *bool
	SkipHistory          *bool
	// GUID если пуст — генерируется UUID.
	GUID string
	// Extra дополнительные поля тела запроса.
	Extra map[string]any
}

func (s Subscription) command() map[string]any {
	cmd := map[string]any{
		"opcode": string(s.Opcode),
	}
	if s.Exchange != "" {
		cmd["exchange"] = string(s.Exchange)
	}
	if s.Code != "" {
		cmd["code"] = s.Code
	}
	if s.InstrumentGroup != "" {
		cmd["instrumentGroup"] = s.InstrumentGroup
	}
	if s.Portfolio != "" {
		cmd["portfolio"] = s.Portfolio
	}
	if s.Depth > 0 {
		cmd["depth"] = s.Depth
	}
	if s.Frequency > 0 {
		cmd["frequency"] = s.Frequency
	}
	if s.Format != "" {
		cmd["format"] = string(s.Format)
	}
	if s.From > 0 {
		cmd["from"] = s.From
	}
	if s.TF != "" {
		cmd["tf"] = s.TF
	}
	if s.IncludeVirtualTrades != nil {
		cmd["includeVirtualTrades"] = *s.IncludeVirtualTrades
	}
	if s.SkipHistory != nil {
		cmd["skipHistory"] = *s.SkipHistory
	}
	for k, v := range s.Extra {
		cmd[k] = v
	}
	return cmd
}

// WS — публичный фасад подписок.
type WS struct {
	c *Client
}

/**
 * Subscribe регистрирует подписку и возвращает guid.
 * Соединение поднимается лениво; при reconnect подписки восстанавливаются.
 */
func (w *WS) Subscribe(ctx context.Context, sub Subscription, handler DataHandler) (string, error) {
	return w.c.ws.subscribe(ctx, sub, handler)
}

/**
 * Unsubscribe отменяет подписку по guid.
 */
func (w *WS) Unsubscribe(ctx context.Context, guid string) error {
	return w.c.ws.unsubscribe(ctx, guid)
}

/**
 * Connected сообщает, установлено ли /ws соединение.
 */
func (w *WS) Connected() bool {
	return w.c.ws.connected()
}

/**
 * SubscriptionCount — число активных локальных подписок.
 */
func (w *WS) SubscriptionCount() int {
	return w.c.ws.subscriptionCount()
}

type wsSubscription struct {
	command map[string]any
	handler DataHandler
}

type wsConn struct {
	client *Client

	mu            sync.Mutex
	conn          *websocket.Conn
	writeMu       sync.Mutex
	subscriptions map[string]*wsSubscription
	dialing       bool
	closed        bool
}

func newWSConn(c *Client) *wsConn {
	return &wsConn{
		client:        c,
		subscriptions: make(map[string]*wsSubscription),
	}
}

func (w *wsConn) connected() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn != nil
}

func (w *wsConn) subscriptionCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.subscriptions)
}

func (w *wsConn) subscribe(ctx context.Context, sub Subscription, handler DataHandler) (string, error) {
	if w.client.isClosed() {
		return "", ErrClosed
	}
	if handler == nil {
		return "", fmt.Errorf("alor: handler обязателен")
	}
	if sub.Format == "" {
		sub.Format = w.client.cfg.DefaultFormat
	}
	guid := sub.GUID
	if guid == "" {
		guid = uuid.NewString()
	}
	cmd := sub.command()
	cmd["guid"] = guid

	w.mu.Lock()
	w.subscriptions[guid] = &wsSubscription{command: cmd, handler: handler}
	needDial := w.conn == nil && !w.dialing && !w.closed
	conn := w.conn
	w.mu.Unlock()

	if needDial {
		go w.ensureLoop()
	}
	if conn != nil {
		if err := w.sendCommand(ctx, cmd); err != nil {
			// подписка сохранена — уйдёт при reconnect
			return guid, nil
		}
	}
	return guid, nil
}

func (w *wsConn) unsubscribe(ctx context.Context, guid string) error {
	w.mu.Lock()
	_, ok := w.subscriptions[guid]
	delete(w.subscriptions, guid)
	conn := w.conn
	w.mu.Unlock()
	if !ok {
		return nil
	}
	if conn == nil {
		return nil
	}
	return w.sendCommand(ctx, map[string]any{
		"opcode": string(OpcodeUnsubscribe),
		"guid":   guid,
	})
}

func (w *wsConn) close() {
	w.mu.Lock()
	w.closed = true
	conn := w.conn
	w.conn = nil
	w.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

func (w *wsConn) ensureLoop() {
	w.mu.Lock()
	if w.dialing || w.closed {
		w.mu.Unlock()
		return
	}
	w.dialing = true
	w.mu.Unlock()

	backoff := w.client.cfg.WSReconnectMin
	for {
		if w.client.isClosed() {
			w.mu.Lock()
			w.dialing = false
			w.mu.Unlock()
			return
		}
		w.mu.Lock()
		empty := len(w.subscriptions) == 0
		closed := w.closed
		w.mu.Unlock()
		if closed || empty {
			w.mu.Lock()
			w.dialing = false
			w.mu.Unlock()
			return
		}

		err := w.connectAndRead()
		if w.client.isClosed() {
			w.mu.Lock()
			w.dialing = false
			w.mu.Unlock()
			return
		}
		_ = err
		jitter := time.Duration(rand.Int63n(int64(backoff/2) + 1))
		time.Sleep(backoff + jitter)
		backoff *= 2
		if backoff > w.client.cfg.WSReconnectMax {
			backoff = w.client.cfg.WSReconnectMax
		}
	}
}

func (w *wsConn) connectAndRead() error {
	ctx, cancel := context.WithTimeout(context.Background(), w.client.cfg.WSHandshakeTimeout)
	token, err := w.client.accessToken(ctx)
	cancel()
	if err != nil {
		return err
	}

	dialer := websocket.Dialer{
		HandshakeTimeout:  w.client.cfg.WSHandshakeTimeout,
		EnableCompression: w.client.cfg.EnableCompression != nil && *w.client.cfg.EnableCompression,
	}
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	conn, _, err := dialer.Dial(w.client.cfg.WSURL, header)
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.conn = conn
	subs := make([]*wsSubscription, 0, len(w.subscriptions))
	for _, s := range w.subscriptions {
		subs = append(subs, s)
	}
	w.mu.Unlock()

	// resubscribe
	for _, s := range subs {
		cmd := cloneMap(s.command)
		_ = w.sendCommand(context.Background(), cmd)
	}

	defer func() {
		w.mu.Lock()
		if w.conn == conn {
			w.conn = nil
		}
		w.mu.Unlock()
		_ = conn.Close()
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		w.dispatch(data)
	}
}

func (w *wsConn) dispatch(raw []byte) {
	var msg WSMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return
	}
	if msg.HTTPCode != 0 && msg.HTTPCode >= 400 {
		// ошибка подписки — можно пробросить через handler с пустым data; пока игнор
		return
	}
	guid := msg.GUID
	if guid == "" {
		guid = msg.RequestGUID
	}
	w.mu.Lock()
	sub := w.subscriptions[guid]
	w.mu.Unlock()
	if sub == nil || sub.handler == nil {
		return
	}
	payload := msg.Data
	if payload == nil {
		// иногда data на верхнем уровне без обёртки — отдаём весь raw
		payload = json.RawMessage(raw)
	}
	sub.handler(guid, payload)
}

func (w *wsConn) sendCommand(ctx context.Context, cmd map[string]any) error {
	token, err := w.client.accessToken(ctx)
	if err != nil {
		return err
	}
	cmd = cloneMap(cmd)
	cmd["token"] = token

	w.mu.Lock()
	conn := w.conn
	w.mu.Unlock()
	if conn == nil {
		return ErrNotConnected
	}
	w.writeMu.Lock()
	defer w.writeMu.Unlock()
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteJSON(cmd)
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in)+1)
	for k, v := range in {
		out[k] = v
	}
	return out
}
