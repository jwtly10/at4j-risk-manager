//go:build integration
// +build integration

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func setupTestEnv(t *testing.T) BrokersConfig {
	t.Helper()

	if err := godotenv.Load(); err != nil {
		t.Fatalf("failed to load environment variables: %v", err)
	}

	// Validate required env vars are present
	requiredVars := []string{"OANDA_API_KEY", "OANDA_API_URL", "TEST_OANDA_ACCOUNT_ID", "MT5_API_KEY", "MT5_API_URL", "TEST_MT5_ACCOUNT_ID"}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			t.Fatalf("required environment variable %s is not set", v)
		}
	}

	o := OandaConfig{
		ApiKey:  os.Getenv("OANDA_API_KEY"),
		BaseUrl: os.Getenv("OANDA_API_URL"),
	}

	m := MT5Config{
		ApiKey:  os.Getenv("MT5_API_KEY"),
		BaseUrl: os.Getenv("MT5_API_URL"),
	}

	return BrokersConfig{
		Oanda: o,
		MT5:   m,
	}
}

func TestGetAccountFromOandaIntegration(t *testing.T) {
	config := setupTestEnv(t)
	accountIdUnderTest := os.Getenv("TEST_OANDA_ACCOUNT_ID")

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	client, err := NewAdapter(httpClient, "OANDA", BrokersConfig{Oanda: config.Oanda})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	e, err := client.GetEquity(ctx, accountIdUnderTest)
	if err != nil {
		t.Fatalf("failed to get equity: %v", err)
	}

	// We can just assert 0 because if the account is valid, it should have some equity
	if e <= 0 {
		t.Errorf("expected equity to be greater than 0, got %f", e)
	}

	fmt.Println("Oanda Equity: ", e)
}

func TestGetAccountFromMT5Integration(t *testing.T) {
	config := setupTestEnv(t)
	accountIdUnderTest := os.Getenv("TEST_MT5_ACCOUNT_ID")

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	client, err := NewAdapter(httpClient, "MT5_FTMO", BrokersConfig{MT5: config.MT5})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	e, err := client.GetEquity(ctx, accountIdUnderTest)
	if err != nil {
		t.Fatalf("failed to get equity: %v", err)
	}

	if e <= 0 {
		t.Errorf("expected equity to be greater than 0, got %f", e)
	}

	fmt.Println("MT5 Equity: ", e)
}
