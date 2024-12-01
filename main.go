package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load environment configuration: %v", err)
	}

	db, err := NewDBConnection(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect connecting to database: %v", err)
	}
	defer db.Close()

	dbClient := NewDBClient(db)

	// Configure broker specific time configs
	configs := map[string]BrokerTimeConfig{
		"OANDA": {
			Timezone:          "UTC",
			DailyUpdateHour:   18,
			DailyUpdateMinute: 54,
		},
		"MT5_FTMO": {
			Timezone:          "Europe/Prague",
			DailyUpdateHour:   16,
			DailyUpdateMinute: 23,
		},
	}

	// Configure broker adapters
	brokerAdapters := make(map[string]BrokerAdapter)
	oandaAdapter, _ := NewAdapter(http.DefaultClient, BrokerTypeOanda, cfg.Brokers)
	brokerAdapters[BrokerTypeOanda] = oandaAdapter
	ftmoAdapter, _ := NewAdapter(http.DefaultClient, BrokerTypeMT5_FTMO, cfg.Brokers)
	brokerAdapters[BrokerTypeMT5_FTMO] = ftmoAdapter

	tracker := NewEquityTracker(dbClient, configs, brokerAdapters, 1*time.Second)

	go func() {
		if err := tracker.Start(); err != nil {
			log.Printf("Error starting equity tracker: %v", err)
			cancel()
		}
	}()

	log.Println("Equity tracking service started. Press Ctrl+C to stop")

	select {
	case sig := <-sigChan:
		log.Printf("Received shutdown signal: %v", sig)
	case <-ctx.Done():
		log.Println("Shutting down due to error")
	}

	log.Println("Initiating graceful shutdown...")

	shutdownCtx, shutdownCancel :=
		context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()

	tracker.Stop()

	<-shutdownCtx.Done()

	log.Println("Service stopped")
}
