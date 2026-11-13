package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/maksim-zakharov/go-alor-api/alor"
)

func main() {
	refresh := os.Getenv("ALOR_REFRESH_TOKEN")
	cfg := alor.Config{Environment: alor.EnvProduction, RefreshToken: refresh}
	if os.Getenv("ALOR_ENV") == "dev" {
		cfg.Environment = alor.EnvDev
	}
	client := alor.New(cfg)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	sec, err := client.GetTime(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("server time:", time.Unix(sec, 0).UTC())

	if refresh == "" {
		fmt.Println("ALOR_REFRESH_TOKEN не задан — только публичный GetTime")
		return
	}
	quotes, err := client.GetQuotes(ctx, "MOEX:SBER", nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, q := range quotes {
		fmt.Printf("%s last=%.4f bid=%.4f ask=%.4f\n", q.Symbol, q.LastPrice, q.Bid, q.Ask)
	}
}
