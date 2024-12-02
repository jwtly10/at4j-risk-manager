# AT4J Risk Management Service

A Go-based service for monitoring and managing account equity across multiple broker accounts, designed as a critical risk management component of the AlgoTrade4j platform.

Used internally as a reliable way to track account equity and ensure that trading strategies are operating within risk limits.

## Overview

This service provides continuous monitoring of account equity across configured brokers, operating independently of trading strategy states. It's designed to support prop firm requirements and enhance the platform's risk management capabilities.

## Features
- Continuous equity monitoring across multiple brokers, even when strategy is not 'LIVE'
- Historical equity data tracking
- Timezone-aware prop firm equity tracking (supports FTMO)
- Independent operation alongside existing Java services

## API Endpoints

### GET /api/v1/equity/latest
Retrieves the latest equity value for a specified trading account.

**Query Parameters:**
- `accountId` (required): The ID of the trading account

**Response:**
```json
{
    "accountId": "string",
    "lastEquity": 1000.00,
    "updatedAt": "2024-12-02T12:00:00Z"
}
```

## Configuration

The service uses environment variables for configuration:
- API keys for broker connections
- Database credentials
- Service port
- Other broker-specific configurations

## Development Setup

1. Ensure you have Go installed
2. Clone the repository
3. Set up required environment variables [.env-example]
4. Run the service:
   ```bash
   go run cmd/main.go
   ```

## Future Enhancements

- Automated risk limit monitoring
- Trade closure capabilities for risk violations
- Telegram notifications for risk events
- Smart data storage optimization
- Additional broker integrations