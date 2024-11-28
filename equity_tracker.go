package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

type brokerRepository interface {
	GetAllActiveBrokers(ctx context.Context) ([]BrokerAccount, error)
}

type BrokerTimeConfig struct {
	Timezone          string
	DailyUpdateHour   int // 0 - 23
	DailyUpdateMinute int // 0 - 59
}

type EquityTracker struct {
	brokerRepo    brokerRepository
	brokerConfigs map[string]BrokerTimeConfig
	checkInterval time.Duration
	stop          chan struct{}
	timeProvider  TimeProvider
}

func NewEquityTracker(
	brokerRepo brokerRepository,
	brokerConfigs map[string]BrokerTimeConfig,
	checkInterval time.Duration,
) *EquityTracker {
	return &EquityTracker{
		brokerRepo:    brokerRepo,
		brokerConfigs: brokerConfigs,
		checkInterval: checkInterval,
		stop:          make(chan struct{}),
		timeProvider:  RealTimeProvider{},
	}
}

// Start starts the equity tracker with the given configuration
func (et *EquityTracker) Start() error {
	ticket := time.NewTicker(et.checkInterval)
	defer ticket.Stop()

	for {
		select {
		case <-ticket.C:
			if err := et.checkAndUpdateEquity(); err != nil {
				log.Printf("Error checking and updating equity: %v", err)
			}
		case <-et.stop:
			return nil
		}
	}
}

// Stop stops the equity tracker
func (et *EquityTracker) Stop() {
	close(et.stop)
}

// checkAndUpdateEquity checks and updates the equity for all active brokers
// based on the configured check configurations
func (et *EquityTracker) checkAndUpdateEquity() error {
	log.Println("Checking and updating equity")
	accounts, err := et.brokerRepo.GetAllActiveBrokers(context.TODO())
	if err != nil {
		return fmt.Errorf("error getting all active brokers: %v", err)
	}

	log.Printf("Found %d active brokers", len(accounts))

	for _, account := range accounts {
		config, exists := et.brokerConfigs[account.BrokerType]
		if !exists {
			log.Printf("No configuration found for broker type %s. Skipping.", account.BrokerType)
			continue
		}

		location, err := time.LoadLocation(config.Timezone)
		if err != nil {
			return fmt.Errorf("error loading timezone %s: %v for broker type %v", config.Timezone, account.BrokerType, err)
		}

		now := et.timeProvider.Now().In(location)

		if isUpdateTime(now, config.DailyUpdateHour, config.DailyUpdateMinute) {
			log.Printf("Updating equity for broker %s", account.BrokerType)
		}
	}

	return nil
}

// isUpdateTime checks if the given time is the target time to update equity
// for a broker
func isUpdateTime(t time.Time, targetHour, targetMinute int) bool {
	// Check if we're within the first check interval after the target time
	// This prevents multiple updates within the same hour
	currentHour := t.Hour()
	currentMinute := t.Minute()

	return currentHour == targetHour &&
		currentMinute >= targetMinute &&
		currentMinute < targetMinute+1
}
