package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nelfander/stock-go-app/shared"
	"github.com/redis/go-redis/v9"
)

func main() {
	//  Setup Structured Logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	//  Setup Graceful Shutdown Context
	// This context will be 'Done' when you press Ctrl+C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	//  Redis Connection
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Check connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	lastPrices := make(map[string]float64)
	slog.Info("🕵️ Analyzer started", "status", "listening", "stream", "stock-stream")

	//  Main Event Loop
	for {
		select {
		case <-ctx.Done():
			// Exit the loop if Ctrl+C is pressed
			slog.Info("Shutting down analyzer gracefully...")
			return
		default:
			// Read from Redis Stream with a timeout so it checks ctx.Done() frequently
			streams, err := rdb.XRead(ctx, &redis.XReadArgs{
				Streams: []string{"stock-stream", "$"},
				Block:   time.Second * 2, // Check for new messages every 2 seconds
				Count:   1,
			}).Result()

			if err == redis.Nil {
				continue // No new messages, loop again
			}
			if err != nil {
				// If error is because context was cancelled, just exit
				if ctx.Err() != nil {
					return
				}
				slog.Error("Redis read error", "error", err)
				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					processMessage(message, lastPrices)
				}
			}
		}
	}
}

// Separate logic into functions for cleaner code
func processMessage(message redis.XMessage, lastPrices map[string]float64) {
	rawEvent, exists := message.Values["event"]
	if !exists {
		slog.Warn("Received message with missing 'event' key")
		return
	}

	dataStr, ok := rawEvent.(string)
	if !ok {
		slog.Warn("Event data is not a string", "type", fmt.Sprintf("%T", rawEvent))
		return
	}

	var priceData shared.StockPrice
	if err := json.Unmarshal([]byte(dataStr), &priceData); err != nil {
		slog.Error("Failed to unmarshal JSON", "error", err)
		return
	}

	symbol := priceData.Symbol
	currentPrice := priceData.Price

	if oldPrice, exists := lastPrices[symbol]; exists {
		change := (currentPrice - oldPrice) / oldPrice
		absChange := math.Abs(change)

		// Threshold of 5% for the simulator
		if absChange > 0.05 {
			slog.Info("🚨 PRICE SPIKE DETECTED",
				"symbol", symbol,
				"change_percent", fmt.Sprintf("%.2f%%", change*100),
				"old_price", oldPrice,
				"new_price", currentPrice,
			)
		} else {
			slog.Debug("Price stable", "symbol", symbol, "price", currentPrice)
		}
	}

	lastPrices[symbol] = currentPrice
}
