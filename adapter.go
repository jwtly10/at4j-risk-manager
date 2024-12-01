package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

type MT5AccountResponse struct {
	Equity float64 `json:"equity"`
}

func (o *OandaAdapter) GetEquity(ctx context.Context, accountId string) (float64, error) {
	url := o.baseURL + "/v3/accounts/" + accountId

	headers := map[string]string{
		"Authorization": "Bearer " + o.apiKey,
	}

	response, err := makeGET[OandaAccountResponse](ctx, o.client, url, headers)

	if err != nil {
		return 0, err
	}

	equity, err := strconv.ParseFloat(response.Account.Equity, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing equity: %v", err)
	}

	return equity, nil
}

func (m *MT5Adapter) GetEquity(ctx context.Context, accountId string) (float64, error) {
	url := m.baseURL + "/accounts/" + accountId

	fmt.Println(url)

	headers := map[string]string{
		"x-api-key": m.apiKey,
	}

	response, err := makeGET[MT5AccountResponse](ctx, m.client, url, headers)
	if err != nil {
		return 0, err
	}

	return response.Equity, nil
}

// makeGET is a helper function to handle GET requests, and resolve generic response types & errors
func makeGET[T any](ctx context.Context, client *http.Client, url string, headers map[string]string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	r, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %v", err)
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", r.StatusCode, string(body))
	}

	var result T
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}
