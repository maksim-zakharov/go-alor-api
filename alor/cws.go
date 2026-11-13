package alor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// CWSOpcode — код операции торговых команд /cws.
type CWSOpcode string

const (
	CWSAuthorize       CWSOpcode = "authorize"
	CWSCreateLimit     CWSOpcode = "create:limit"
	CWSCreateMarket    CWSOpcode = "create:market"
	CWSCreateStop      CWSOpcode = "create:stop"
	CWSCreateStopLimit CWSOpcode = "create:stopLimit"
	CWSDeleteLimit     CWSOpcode = "delete:limit"
	CWSDeleteMarket    CWSOpcode = "delete:market"
	CWSDeleteStop      CWSOpcode = "delete:stop"
	CWSDeleteStopLimit CWSOpcode = "delete:stopLimit"
	CWSUpdateLimit     CWSOpcode = "update:limit"
	CWSUpdateMarket    CWSOpcode = "update:market"
	CWSUpdateStop      CWSOpcode = "update:stop"
	CWSUpdateStopLimit CWSOpcode = "update:stopLimit"
)

// CWS — фасад торговых команд WebSocket (/cws).
type CWS struct {
	c *Client
}

/**
 * Connected сообщает, авторизовано ли /cws соединение.
 */
func (x *CWS) Connected() bool {
	return x.c.cws.isAuthorized()
}

/**
 * Ensure подключает и авторизует /cws.
 */
func (x *CWS) Ensure(ctx context.Context) error {
	return x.c.cws.ensure(ctx)
}

/**
 * Send отправляет произвольную команду /cws и ждёт ответ с тем же guid.
 */
func (x *CWS) Send(ctx context.Context, opcode CWSOpcode, body map[string]any) (*CWSResponse, error) {
	return x.c.cws.send(ctx, opcode, body)
}

/**
 * CreateLimit создаёт лимитную заявку через /cws.
 */
func (x *CWS) CreateLimit(ctx context.Context, req LimitOrderRequest) (*CWSResponse, error) {
	return x.sendOrder(ctx, CWSCreateLimit, req)
}

/**
 * CreateMarket создаёт рыночную заявку через /cws.
 */
func (x *CWS) CreateMarket(ctx context.Context, req MarketOrderRequest) (*CWSResponse, error) {
	return x.sendOrder(ctx, CWSCreateMarket, req)
}

/**
 * CreateStop создаёт стоп-заявку через /cws.
 */
func (x *CWS) CreateStop(ctx context.Context, req StopOrderRequest) (*CWSResponse, error) {
	return x.sendOrder(ctx, CWSCreateStop, req)
}

/**
 * CreateStopLimit создаёт стоп-лимит через /cws.
 */
func (x *CWS) CreateStopLimit(ctx context.Context, req StopLimitOrderRequest) (*CWSResponse, error) {
	return x.sendOrder(ctx, CWSCreateStopLimit, req)
}

/**
 * UpdateLimit изменяет лимитную заявку (в body нужен orderId по доке).
 */
func (x *CWS) UpdateLimit(ctx context.Context, body map[string]any) (*CWSResponse, error) {
	return x.c.cws.send(ctx, CWSUpdateLimit, body)
}

/**
 * UpdateMarket изменяет рыночную заявку.
 */
func (x *CWS) UpdateMarket(ctx context.Context, body map[string]any) (*CWSResponse, error) {
	return x.c.cws.send(ctx, CWSUpdateMarket, body)
}

/**
 * UpdateStop изменяет стоп-заявку.
 */
func (x *CWS) UpdateStop(ctx context.Context, body map[string]any) (*CWSResponse, error) {
	return x.c.cws.send(ctx, CWSUpdateStop, body)
}

/**
 * UpdateStopLimit изменяет стоп-лимит.
 */
func (x *CWS) UpdateStopLimit(ctx context.Context, body map[string]any) (*CWSResponse, error) {
	return x.c.cws.send(ctx, CWSUpdateStopLimit, body)
}

/**
 * DeleteLimit снимает лимитную заявку.
 */
func (x *CWS) DeleteLimit(ctx context.Context, body map[string]any) (*CWSResponse, error) {
	return x.c.cws.send(ctx, CWSDeleteLimit, body)
}

/**
 * DeleteMarket снимает рыночную заявку.
 */
func (x *CWS) DeleteMarket(ctx context.Context, body map[string]any) (*CWSResponse, error) {
	return x.c.cws.send(ctx, CWSDeleteMarket, body)
}

/**
 * DeleteStop снимает стоп-заявку.
 */
func (x *CWS) DeleteStop(ctx context.Context, body map[string]any) (*CWSResponse, error) {
	return x.c.cws.send(ctx, CWSDeleteStop, body)
}

/**
 * DeleteStopLimit снимает стоп-лимит.
 */
func (x *CWS) DeleteStopLimit(ctx context.Context, body map[string]any) (*CWSResponse, error) {
	return x.c.cws.send(ctx, CWSDeleteStopLimit, body)
}

func (x *CWS) sendOrder(ctx context.Context, opcode CWSOpcode, req any) (*CWSResponse, error) {
	raw, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	return x.c.cws.send(ctx, opcode, body)
}

// CWSResponse — ответ на команду /cws.
type CWSResponse struct {
	GUID        string          `json:"guid"`
	HTTPCode    int             `json:"httpCode,omitempty"`
	Message     string          `json:"message,omitempty"`
	OrderNumber string          `json:"orderNumber,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
	Raw         json.RawMessage `json:"-"`
}

// OK сообщает об успешном HTTP-коде ответа.
func (r *CWSResponse) OK() bool {
	return r != nil && (r.HTTPCode == 0 || (r.HTTPCode >= 200 && r.HTTPCode < 300))
}

type cwsPending struct {
	ch chan *CWSResponse
}

type cwsConn struct {
	client *Client

	mu         sync.Mutex
	conn       *websocket.Conn
	writeMu    sync.Mutex
	authorized bool
	closed     bool
	pending    map[string]*cwsPending
	generation atomic.Uint64
}

func newCWSConn(c *Client) *cwsConn {
	return &cwsConn{
		client:  c,
		pending: make(map[string]*cwsPending),
	}
}

func (w *cwsConn) isAuthorized() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.authorized && w.conn != nil
}

func (w *cwsConn) close() {
	w.mu.Lock()
	w.closed = true
	conn := w.conn
	w.conn = nil
	w.authorized = false
	for guid, p := range w.pending {
		close(p.ch)
		delete(w.pending, guid)
	}
	w.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

func (w *cwsConn) ensure(ctx context.Context) error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return ErrClosed
	}
	if w.conn != nil && w.authorized {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()
	return w.connect(ctx)
}

func (w *cwsConn) connect(ctx context.Context) error {
	token, err := w.client.accessToken(ctx)
	if err != nil {
		return err
	}
	dialer := websocket.Dialer{HandshakeTimeout: w.client.cfg.WSHandshakeTimeout}
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	conn, _, err := dialer.DialContext(ctx, w.client.cfg.CWSURL, header)
	if err != nil {
		return fmt.Errorf("alor: cws dial: %w", err)
	}

	gen := w.generation.Add(1)
	w.mu.Lock()
	old := w.conn
	w.conn = conn
	w.authorized = false
	w.mu.Unlock()
	if old != nil {
		_ = old.Close()
	}
	go w.readLoop(conn, gen)

	authGUID := uuid.NewString()
	resp, err := w.sendRaw(ctx, map[string]any{
		"opcode": string(CWSAuthorize),
		"guid":   authGUID,
		"token":  token,
	}, authGUID)
	if err != nil {
		return err
	}
	if !resp.OK() {
		return &WSError{GUID: authGUID, Code: resp.HTTPCode, Message: firstNonEmpty(resp.Message, "authorize failed")}
	}
	w.mu.Lock()
	w.authorized = true
	w.mu.Unlock()
	return nil
}

func (w *cwsConn) send(ctx context.Context, opcode CWSOpcode, body map[string]any) (*CWSResponse, error) {
	if err := w.ensure(ctx); err != nil {
		return nil, err
	}
	guid := uuid.NewString()
	cmd := cloneMap(body)
	cmd["opcode"] = string(opcode)
	cmd["guid"] = guid
	token, err := w.client.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	cmd["token"] = token
	return w.sendRaw(ctx, cmd, guid)
}

func (w *cwsConn) sendRaw(ctx context.Context, cmd map[string]any, guid string) (*CWSResponse, error) {
	ch := make(chan *CWSResponse, 1)
	w.mu.Lock()
	conn := w.conn
	if conn == nil {
		w.mu.Unlock()
		return nil, ErrNotConnected
	}
	w.pending[guid] = &cwsPending{ch: ch}
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		delete(w.pending, guid)
		w.mu.Unlock()
	}()

	w.writeMu.Lock()
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	err := conn.WriteJSON(cmd)
	w.writeMu.Unlock()
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-w.client.closed:
		return nil, ErrClosed
	case resp, ok := <-ch:
		if !ok || resp == nil {
			return nil, ErrClosed
		}
		return resp, nil
	}
}

func (w *cwsConn) readLoop(conn *websocket.Conn, gen uint64) {
	defer func() {
		w.mu.Lock()
		if w.conn == conn {
			w.conn = nil
			w.authorized = false
		}
		w.mu.Unlock()
		_ = conn.Close()
	}()

	for {
		if w.generation.Load() != gen {
			return
		}
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var resp CWSResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			continue
		}
		resp.Raw = json.RawMessage(append([]byte(nil), data...))
		w.mu.Lock()
		p := w.pending[resp.GUID]
		w.mu.Unlock()
		if p != nil {
			select {
			case p.ch <- &resp:
			default:
			}
		}
	}
}
