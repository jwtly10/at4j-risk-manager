package main

import (
	"context"
	"fmt"
)

// BrokerAdapter defines the interface that all broker adapters must implement
// So we can do consistent monitoring across all brokers
// Note the equity returned is int64, this is because we are dealing with money
// and the easiest way to avoid floating point errors is to use integers
type BrokerAdapter interface {
	// GetEquity returns the current equity of the broker account in cents (USD)
	GetEquity(ctx context.Context) (int64, error)
}

func NewAdapter(brokerType string) (BrokerAdapter, error) {
	switch brokerType {
	case "OANDA":
		return &OandaAdapter{}, nil
	case "MT5_FTMO":
		return &MT5Adapter{}, nil
	default:
		return nil, fmt.Errorf("unsupported broker type: %s", brokerType)
	}
}

type OandaAdapter struct {
}

func (o *OandaAdapter) GetEquity(ctx context.Context) (int64, error) {
	// TODO: Implement this
	return 0, nil
}

type MT5Adapter struct {
}

func (m *MT5Adapter) GetEquity(ctx context.Context) (int64, error) {
	// TODO: Implement this
	return 0, nil
}
