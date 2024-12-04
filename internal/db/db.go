package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jwtly10/at4j-risk-manager/internal/broker"
	"github.com/jwtly10/at4j-risk-manager/internal/config"
	"time"

	_ "github.com/lib/pq"
)

type Client struct {
	db *sql.DB
}

func NewDBConnection(config config.PostgresConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		config.URL, config.Port, config.Username, config.Password, config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening db connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging db: %v", err)
	}

	return db, nil

}

func NewDBClient(db *sql.DB) *Client {
	return &Client{db: db}
}

// GetActiveBrokers returns all active broker accounts and when equity was last tracked
func (c *Client) GetActiveBrokers(ctx context.Context) ([]broker.BrokerWithLastEquity, error) {
	query := `
        SELECT 
            b.id, 
            b.broker_name, 
            b.broker_type, 
            b.broker_env, 
            b.account_id, 
            b.active, 
            b.initial_balance, 
            b.created_at, 
            b.updated_at,
            e.created_at as last_equity_update
        FROM algotrade.broker_accounts_tb b
        LEFT JOIN (
            SELECT broker_account_id, MAX(created_at) as created_at
            FROM algotrade.equity_tracking_tb
            GROUP BY broker_account_id
        ) e ON b.id = e.broker_account_id
        WHERE b.active = true
    `

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []broker.BrokerWithLastEquity
	for rows.Next() {
		var account broker.BrokerWithLastEquity
		if err := rows.Scan(
			&account.ID,
			&account.BrokerName,
			&account.BrokerType,
			&account.BrokerEnv,
			&account.AccountID,
			&account.Active,
			&account.InitialBalance,
			&account.CreatedAt,
			&account.UpdatedAt,
			&account.LastEquityUpdate,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

// RecordEquity records the equity update for a broker account
func (c *Client) RecordEquity(ctx context.Context, brokerID int64, equity float64) error {
	query := `
        INSERT INTO algotrade.equity_tracking_tb 
        (broker_account_id, equity)
        VALUES ($1, $2)
    `
	_, err := c.db.ExecContext(ctx, query, brokerID, equity)
	return err
}

type EquityData struct {
	Equity    float64
	UpdatedAt time.Time
}

// GetLatestEquity returns the latest equity data for a broker account
func (c *Client) GetLatestEquity(ctx context.Context, brokerId string) (*EquityData, error) {
	query := `
        SELECT et.equity, et.created_at
        FROM algotrade.equity_tracking_tb et
        INNER JOIN algotrade.broker_accounts_tb ba ON et.broker_account_id = ba.id
        WHERE ba.account_id = $1
        ORDER BY et.created_at DESC
        LIMIT 1
    `

	var data EquityData
	err := c.db.QueryRowContext(ctx, query, brokerId).Scan(&data.Equity, &data.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no equity data found for broker ID %s", brokerId)
		}
		return nil, fmt.Errorf("error fetching equity data: %w", err)
	}

	return &data, nil
}
