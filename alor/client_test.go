package alor_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maksim-zakharov/go-alor-api/alor"
)

func TestGetTimeWithoutAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/md/v2/time" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte("1710000000"))
	}))
	t.Cleanup(srv.Close)

	c := alor.New(alor.Config{APIURL: srv.URL, OAuthURL: srv.URL})
	ts, err := c.GetTime(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ts != 1710000000 {
		t.Fatalf("got %d", ts)
	}
}

func TestRefreshAndAuthorizedGet(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"AccessToken": makeTestJWT(t, time.Now().Add(time.Hour).Unix(), []string{"D12345"}),
		})
	})
	mux.HandleFunc("/md/v2/Securities/MOEX:SBER/quotes", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Fatalf("missing bearer: %q", auth)
		}
		_, _ = w.Write([]byte(`[{"symbol":"SBER","last_price":250.5}]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := alor.New(alor.Config{
		APIURL:       srv.URL,
		OAuthURL:     srv.URL,
		RefreshToken: "refresh-test",
	})
	quotes, err := c.GetQuotes(context.Background(), "MOEX:SBER", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].Symbol != "SBER" {
		t.Fatalf("unexpected quotes: %+v", quotes)
	}
	ports, err := c.Portfolios(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ports) != 1 || ports[0] != "D12345" {
		t.Fatalf("portfolios=%v", ports)
	}
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"bad request"}`))
	}))
	t.Cleanup(srv.Close)

	c := alor.New(alor.Config{APIURL: srv.URL, AccessToken: "x", AccessTokenTTL: time.Hour})
	_, err := c.GetSecurity(context.Background(), alor.ExchangeMOEX, "SBER", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *alor.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want APIError, got %T %v", err, err)
	}
	if apiErr.StatusCode != 400 || apiErr.Message != "bad request" {
		t.Fatalf("%+v", apiErr)
	}
}

func TestSubscribeOnClosedClient(t *testing.T) {
	sub := alor.Subscription{
		Opcode:   alor.OpcodeOrderBook,
		Exchange: alor.ExchangeMOEX,
		Code:     "SBER",
		Depth:    10,
		Format:   alor.FormatSimple,
	}
	c := alor.New(alor.Config{
		RefreshToken:   "x",
		AccessToken:    makeTestJWT(t, time.Now().Add(time.Hour).Unix(), nil),
		AccessTokenTTL: time.Hour,
	})
	c.Close()
	_, err := c.WS().Subscribe(context.Background(), sub, func(string, json.RawMessage) {})
	if !errors.Is(err, alor.ErrClosed) {
		t.Fatalf("want ErrClosed, got %v", err)
	}
}

func makeTestJWT(t *testing.T, exp int64, portfolios []string) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payloadMap := map[string]any{"exp": exp}
	if portfolios != nil {
		payloadMap["portfolios"] = portfolios
	}
	raw, err := json.Marshal(payloadMap)
	if err != nil {
		t.Fatal(err)
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	return header + "." + payload + ".sig"
}
