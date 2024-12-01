package main

import (
	"context"
	"github.com/jwtly10/at4j-risk-manager/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger.InitLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load environment configuration: %v", err)
	}

	db, err := NewDBConnection(cfg.DB)
	if err != nil {
		logger.Fatalf("Failed to connect connecting to database: %v", err)
	}
	defer db.Close()

	dbClient := NewDBClient(db)

	// Configure broker specific time configs
	configs := map[string]BrokerTimeConfig{
		"OANDA": {
			Timezone:          "UTC",
			DailyUpdateHour:   21,
			DailyUpdateMinute: 15,
		},
		"MT5_FTMO": {
			Timezone:          "Europe/Prague",
			DailyUpdateHour:   18,
			DailyUpdateMinute: 9,
		},
	}

	// Configure broker adapters
	brokerAdapters := make(map[string]BrokerAdapter)
	oandaAdapter, _ := NewAdapter(http.DefaultClient, Oanda, cfg.Brokers)
	brokerAdapters[Oanda] = oandaAdapter
	ftmoAdapter, _ := NewAdapter(http.DefaultClient, MT5FTMO, cfg.Brokers)
	brokerAdapters[MT5FTMO] = ftmoAdapter

	tracker := NewEquityTracker(dbClient, configs, brokerAdapters, time.Duration(cfg.Jobs.EquityCheckInterval)*time.Second)

	go func() {
		if err := tracker.Start(); err != nil {
			logger.Errorf("Error starting equity tracker: %v", err)
			cancel()
		}
	}()

	select {
	case sig := <-sigChan:
		logger.Infof("Received shutdown signal: %v", sig)
	case <-ctx.Done():
		logger.Infof("Shutting down due to error")
	}

	logger.Infof("Initiating graceful shutdown...")

	shutdownCtx, shutdownCancel :=
		context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()

	tracker.Stop()

	<-shutdownCtx.Done()

	logger.Infof("Service stopped")
}
