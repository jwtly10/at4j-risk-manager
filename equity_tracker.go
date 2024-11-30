package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/shopspring/decimal"
)

type brokerRepository interface {
	GetActiveBrokers(ctx context.Context) ([]BrokerWithLastEquity, error)
	RecordEquity(ctx context.Context, brokerID int64, equity decimal.Decimal) error
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
	accounts, err := et.brokerRepo.GetActiveBrokers(context.TODO())
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
			// If theres NO last equity, we should always update it
			if account.LastEquityUpdate != nil {
				// Else handle the case when equity has already been tracked today
				lastUpdateLocal := account.LastEquityUpdate.In(location)

				if isSameDay(now, lastUpdateLocal) {
					// TODO: Remove this log statement
					log.Printf("Equity already updated for broker %s today", account.BrokerType)
					continue
				}
			}
			log.Printf("Updating equity for broker %s", account.BrokerType)

			// TODO: Implement  get real equity from the brokers api

			// Random number betwen 0 and 100
			equity := 100 * rand.Float64()
			err := et.brokerRepo.RecordEquity(context.TODO(), account.ID, decimal.NewFromFloat(equity))
			if err != nil {
				return fmt.Errorf("error recording equity for broker %s: %v", account.BrokerType, err)
			}
		}
	}

	return nil
}

// isUpdateTime checks if the given time is the target time to update equity for the broker
func isUpdateTime(t time.Time, targetHour, targetMinute int) bool {
	// Check if we're within the first check interval after the target time
	// This prevents multiple updates within the same hour
	currentHour := t.Hour()
	currentMinute := t.Minute()

	return currentHour == targetHour &&
		currentMinute >= targetMinute &&
		currentMinute < targetMinute+1
}

// isSameDay checks if two time instances are on the same day
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
