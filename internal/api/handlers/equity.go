package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jwtly10/at4j-risk-manager/internal/db"
	"github.com/jwtly10/at4j-risk-manager/pkg/logger"
)

type EquityResponse struct {
	AccountId  string  `json:"accountId"`
	LastEquity float64 `json:"lastEquity"`
	// UpdatedAt is the last time the equity was updated in UTC
	UpdatedAt time.Time `json:"updatedAt"`
}

type EquityHandler struct {
	dbClient *db.Client
}

func NewEquityHandler(dbClient *db.Client) *EquityHandler {
	return &EquityHandler{
		dbClient: dbClient,
	}
}

func (h *EquityHandler) GetLatestEquity(w http.ResponseWriter, r *http.Request) {
	accountId := r.URL.Query().Get("accountId")
	if accountId == "" {
		http.Error(w, "accountId parameter is required", http.StatusBadRequest)
		return
	}

	data, err := h.dbClient.GetLatestEquity(r.Context(), accountId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "No equity data found for broker", http.StatusNotFound)
			return
		}

		logger.Errorf("Error getting latest equity: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := EquityResponse{
		AccountId:  accountId,
		LastEquity: data.Equity,
		UpdatedAt:  data.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		logger.Errorf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
