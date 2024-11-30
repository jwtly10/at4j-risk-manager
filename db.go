package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type DBClient struct {
	db *sql.DB
}

func NewDBConnection(config PostgresConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
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

func NewDBClient(db *sql.DB) *DBClient {
	return &DBClient{db: db}
}

// GetActiveBrokers returns all active broker accounts and when equity was last tracked
func (c *DBClient) GetActiveBrokers(ctx context.Context) ([]BrokerWithLastEquity, error) {
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

	var accounts []BrokerWithLastEquity
	for rows.Next() {
		var account BrokerWithLastEquity
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

// RecordEquityUpdate records the equity update for a broker account
func (c *DBClient) RecordEquity(ctx context.Context, brokerID int64, equity decimal.Decimal) error {
	query := `
        INSERT INTO algotrade.equity_tracking_tb 
        (broker_account_id, equity)
        VALUES ($1, $2)
    `
	_, err := c.db.ExecContext(ctx, query, brokerID, equity)
	return err
}
