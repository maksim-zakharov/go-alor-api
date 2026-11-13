# go-alor-api

Go-клиент [Alor OpenAPI V2](https://alor.dev/docs/): HTTP REST, WebSocket подписки (`/ws`) и торговые команды (`/cws`).

Спецификация: [`api/openapi/WarpOpenAPIv2.yml`](api/openapi/WarpOpenAPIv2.yml) (с [alor.dev/rawdocs2](https://alor.dev/rawdocs2/WarpOpenAPIv2.yml)).

## Установка

```bash
go get github.com/maksim-zakharov/go-alor-api
```

## Быстрый старт

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/maksim-zakharov/go-alor-api/alor"
)

func main() {
	client := alor.NewProduction(os.Getenv("ALOR_REFRESH_TOKEN"))
	defer client.Close()

	ctx := context.Background()
	quotes, err := client.GetQuotes(ctx, "MOEX:SBER", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", quotes)

	guid, err := client.WS().SubscribeOrderBook(ctx, alor.ExchangeMOEX, "SBER", 10,
		func(guid string, data json.RawMessage) {
			fmt.Println(guid, string(data))
		},
		alor.WithInstrumentGroup("TQBR"),
		alor.WithFrequency(250),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.WS().Unsubscribe(ctx, guid)

	select {}
}
```

Тестовый контур: `alor.NewDev(refreshToken)`.

## Возможности

### Авторизация

- Refresh → JWT (`oauth.alor.ru` / `oauthdev.alor.ru`)
- Автообновление access-токена
- Портфели из JWT: `client.Portfolios(ctx)`
- Ручная установка: `SetAccessToken`

### HTTP (все ручки OpenAPI V2)

| Группа | Методы |
|--------|--------|
| Market data | `GetTime`, `GetSecurities*`, `GetSecurity`, `GetQuotes`, `GetOrderBook`, `GetHistory`, `GetAllTrades*`, `GetCurrencyPairs`, `GetRiskRates`, `GetMarketStatus`, … |
| Client info | `GetOrders`, `GetStopOrders`, `GetPositions`, `GetSummary`, `GetTrades`, `GetRisk`, `GetFortsRisk`, история сделок, … |
| Orders | `Create/Update/Delete` limit/market/stop/stopLimit, `EstimateOrder(s)`, warptrans stopLoss/takeProfit* |
| Order groups | `Get/Create/Update/DeleteOrderGroup(s)` |

Формат ответов по умолчанию — `Simple` (`Config.DefaultFormat`).

### WebSocket `/ws`

Подписки: Quotes, OrderBook, Bars, AllTrades, Instruments, Orders, StopOrders, Positions, Summaries, Risks, SpectraRisks, Trades + `unsubscribe`.

- reconnect с backoff + resubscribe
- `permessage-deflate` (нужен для `depth > 20`)
- хелперы: `SubscribeQuotes`, `SubscribeOrderBook`, …

### WebSocket `/cws`

Команды: `authorize`, `create|update|delete` × `limit|market|stop|stopLimit`.

```go
resp, err := client.CWS().CreateLimit(ctx, alor.LimitOrderRequest{ /* … */ })
```

## Конфигурация

```go
client := alor.New(alor.Config{
	Environment:  alor.EnvProduction, // или EnvDev
	RefreshToken: "...",
	RestRPS:      25,
	DefaultFormat: alor.FormatSimple,
})
```

## Примеры

```bash
export ALOR_REFRESH_TOKEN=...
go run ./examples/rest_time
go run ./examples/ws_quotes
```

## Обновление типов из OpenAPI

```bash
curl.exe --ssl-no-revoke -L -o api/openapi/WarpOpenAPIv2.yml https://alor.dev/rawdocs2/WarpOpenAPIv2.yml
python _gen_types.py
gofmt -w alor/types_simple_gen.go
```

## Лицензия

Apache-2.0 (как у спецификации Alor OpenAPI).
