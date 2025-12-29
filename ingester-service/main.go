package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nelfander/stock-go-app/shared"
	"github.com/redis/go-redis/v9"
)

func main() {
	//  Text Logger for readability
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	stocks := []string{"AAPL", "TSLA", "NVDA", "MSFT", "GOOGL"}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	slog.Info("📈 Ingester started", "interval", "1s")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down ingester...")
			return
		case <-ticker.C:
			priceUpdate := shared.StockPrice{
				Symbol:    stocks[rand.Intn(len(stocks))],
				Price:     100.0 + rand.Float64()*500.0,
				Timestamp: time.Now().Unix(),
			}

			payload, _ := json.Marshal(priceUpdate)

			err := rdb.XAdd(ctx, &redis.XAddArgs{
				Stream: "stock-stream",
				Values: map[string]interface{}{"event": payload},
			}).Err()

			if err != nil {
				slog.Error("Redis publish error", "error", err)
			} else {
				slog.Info("Published price", "symbol", priceUpdate.Symbol, "price", fmt.Sprintf("%.2f", priceUpdate.Price))
			}
		}
	}
}
