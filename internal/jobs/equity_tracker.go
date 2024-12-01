package jobs

import (
	"context"
	"fmt"
	"github.com/jwtly10/at4j-risk-manager/internal/broker"
	"github.com/jwtly10/at4j-risk-manager/internal/utils"
	"github.com/jwtly10/at4j-risk-manager/pkg/logger"
	"time"
)

type brokerRepository interface {
	GetActiveBrokers(ctx context.Context) ([]broker.BrokerWithLastEquity, error)
	RecordEquity(ctx context.Context, brokerID int64, equity float64) error
}

type BrokerTimeConfig struct {
	Timezone          string
	DailyUpdateHour   int // 0 - 23
	DailyUpdateMinute int // 0 - 59
}

type EquityTracker struct {
	brokerRepo     brokerRepository
	brokerConfigs  map[string]BrokerTimeConfig
	brokerAdapters map[string]broker.BrokerAdapter
	checkInterval  time.Duration
	stop           chan struct{}
	timeProvider   utils.TimeProvider
}

func NewEquityTracker(
	brokerRepo brokerRepository,
	brokerConfigs map[string]BrokerTimeConfig,
	brokerAdapters map[string]broker.BrokerAdapter,
	checkInterval time.Duration,
) *EquityTracker {
	return &EquityTracker{
		brokerRepo:     brokerRepo,
		brokerConfigs:  brokerConfigs,
		brokerAdapters: brokerAdapters,
		checkInterval:  checkInterval,
		stop:           make(chan struct{}),
		timeProvider:   utils.RealTimeProvider{},
	}
}

// Start starts the equity tracker with the given configuration
func (et *EquityTracker) Start() error {

	logger.Infof("Starting equity tracker with check interval %v", et.checkInterval)

	ticket := time.NewTicker(et.checkInterval)
	defer ticket.Stop()

	ctx := context.Background()

	for {
		select {
		case <-ticket.C:
			if err := et.checkAndUpdateEquity(ctx); err != nil {
				logger.Infof("Error checking and updating equity: %v", err)
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
func (et *EquityTracker) checkAndUpdateEquity(ctx context.Context) error {
	logger.Infof("Running equity check job")
	accounts, err := et.brokerRepo.GetActiveBrokers(ctx)
	if err != nil {
		return fmt.Errorf("error getting all active brokers: %v", err)
	}

	logger.Debugf("Found %d active brokers", len(accounts))

	for _, account := range accounts {
		config, exists := et.brokerConfigs[account.BrokerType]
		if !exists {
			logger.Warnf("No configuration found for broker type %s. Skipping.", account.BrokerType)
			continue
		}

		adapter, exists := et.brokerAdapters[account.BrokerType]
		if !exists {
			logger.Warnf("No adapter found for broker type %s. Skipping.", account.BrokerType)
			continue
		}

		location, err := time.LoadLocation(config.Timezone)
		if err != nil {
			logger.Errorf("error loading timezone %s: %v for broker type %v", config.Timezone, account.BrokerType, err)
			continue
		}

		now := et.timeProvider.Now().In(location)

		if isUpdateTime(now, config.DailyUpdateHour, config.DailyUpdateMinute) {
			// If there's NO last equity, we should always update it
			if account.LastEquityUpdate != nil {
				// Else handle the case when equity has already been tracked today
				lastUpdateLocal := account.LastEquityUpdate.In(location)

				if isSameDay(now, lastUpdateLocal) {
					logger.Debugf("Equity already updated for broker %s today", account.BrokerType)
					continue
				}
			}
			logger.Infof("Updating equity for broker %s", account.BrokerType)

			equity, err := adapter.GetEquity(ctx, account.AccountID)
			if err != nil {
				logger.Errorf("Error getting equity for broker %s: %v", account.BrokerType, err)
				continue
			}

			err = et.brokerRepo.RecordEquity(ctx, account.ID, equity)
			if err != nil {
				logger.Errorf("error recording equity for broker %s: %v", account.BrokerType, err)
			}

			logger.Infof("Equity updated for broker %s: %.2f", account.BrokerType, equity)
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