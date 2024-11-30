package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// Supported broker types
const (
	BrokerTypeOanda    = "OANDA"
	BrokerTypeMT5_FTMO = "MT5_FTMO"
)

// BrokerAdapter defines the interface that all broker adapters must implement
// So we can do consistent monitoring across all brokers
type BrokerAdapter interface {
	// GetEquity returns the current equity of the broker account
	GetEquity(ctx context.Context, accountId string) (float64, error)
}

type OandaAdapter struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

type MT5Adapter struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

// NewAdapter is a factory function that returns a new broker adapter based on the given configuration
func NewAdapter(client *http.Client, brokerType string, config BrokersConfig) (BrokerAdapter, error) {
	switch brokerType {
	case BrokerTypeOanda:
		return newOandaAdapter(client, config.Oanda), nil
	case BrokerTypeMT5_FTMO:
		return newMT5Adapter(client, config.MT5), nil
	default:
		return nil, fmt.Errorf("unsupported broker type: %s", brokerType)
	}
}

// newOandaAdapter returns a new Oanda adapter based on the given configuration
func newOandaAdapter(client *http.Client, config OandaConfig) *OandaAdapter {
	return &OandaAdapter{
		client:  client,
		apiKey:  config.ApiKey,
		baseURL: config.BaseUrl,
	}
}

// newMT5Adapter returns a new MT5 adapter based on the given configuration
func newMT5Adapter(client *http.Client, config MT5Config) *MT5Adapter {
	return &MT5Adapter{
		client:  client,
		apiKey:  config.ApiKey,
		baseURL: config.BaseUrl,
	}
}

// Implementations

type OandaAccountResponse struct {
	Account OandaAccount `json:"account"`
}

type OandaAccount struct {
	Equity string `json:"NAV"`
}

func (o *OandaAdapter) GetEquity(ctx context.Context, accountId string) (float64, error) {
	url := o.baseURL + "/v3/accounts/" + accountId

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+o.apiKey)

	r, err := o.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error executing request: %v", err)
	}

	defer r.Body.Close()

	res := OandaAccountResponse{}
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		return 0, fmt.Errorf("error decoding response: %v", err)
	}

	equity, err := strconv.ParseFloat(res.Account.Equity, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing equity: %v", err)
	}

	return equity, nil
}

type MT5AccountResponse struct {
	Equity float64 `json:"equity"`
}

func (m *MT5Adapter) GetEquity(ctx context.Context, accountId string) (float64, error) {
	url := m.baseURL + "/accounts/" + accountId

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+m.apiKey)

	r, err := m.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error executing request: %v", err)
	}

	defer r.Body.Close()

	res := MT5AccountResponse{}
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		return 0, fmt.Errorf("error decoding response: %v", err)
	}

	return res.Equity, nil
}
