package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/maksim-zakharov/go-alor-api/alor"
)

func main() {
	refresh := os.Getenv("ALOR_REFRESH_TOKEN")
	if refresh == "" {
		log.Fatal("задайте ALOR_REFRESH_TOKEN")
	}
	cfg := alor.Config{Environment: alor.EnvProduction, RefreshToken: refresh}
	if os.Getenv("ALOR_ENV") == "dev" {
		cfg.Environment = alor.EnvDev
	}
	client := alor.New(cfg)
	defer client.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	guid, err := client.WS().SubscribeQuotes(ctx, alor.ExchangeMOEX, "SBER",
		func(guid string, data json.RawMessage) {
			var q alor.Symbol
			if err := json.Unmarshal(data, &q); err != nil {
				fmt.Println("raw:", string(data))
				return
			}
			fmt.Printf("%s %s last=%.4f\n", time.Now().Format(time.RFC3339), q.Symbol, q.LastPrice)
		},
		alor.WithInstrumentGroup("TQBR"),
		alor.WithFrequency(250),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.WS().Unsubscribe(context.Background(), guid) }()

	fmt.Println("подписка", guid, "— Ctrl+C для выхода")
	<-ctx.Done()
}
