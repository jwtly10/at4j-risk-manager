package broker

import (
	"time"
)

type BrokerAccount struct {
	ID             int64     `db:"id"`
	BrokerName     string    `db:"broker_name"`
	BrokerType     string    `db:"broker_type"`
	BrokerEnv      string    `db:"broker_env"`
	AccountID      string    `db:"account_id"`
	Active         bool      `db:"active"`
	InitialBalance int       `db:"initial_balance"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type BrokerWithLastEquity struct {
	BrokerAccount
	// LastEquityUpdate is the last time the equity was updated in UTC. MUST BE CONVERTED TO THE REQURIED TIMEZONE WHEN USED
	LastEquityUpdate *time.Time `db:"last_equity_update"` // May be nil
}