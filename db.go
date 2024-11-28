package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
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

// GetAllActiveBrokers returns all active broker accounts
func (r *DBClient) GetAllActiveBrokers(ctx context.Context) ([]BrokerAccount, error) {
	query := `
        SELECT id, broker_name, broker_type, broker_env, account_id, 
               active, initial_balance, created_at, updated_at 
        FROM algotrade.broker_accounts_tb
        WHERE active = true
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []BrokerAccount
	for rows.Next() {
		var account BrokerAccount
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
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}
