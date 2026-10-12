package balances

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Balance struct {
	AccountID string      `json:"account_id"`
	Balance   json.Number `json:"balance"`
}

// NewHTTPHandler reads only the Balances projection, never Wallet Core's DB.
func NewHTTPHandler(db *sql.DB) http.Handler {
	router := chi.NewRouter()
	router.Get("/balances/{account_id}", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		var result Balance
		var amount string
		err := db.QueryRowContext(ctx, `SELECT account_id, balance FROM account_balances WHERE account_id = ?`, chi.URLParam(r, "account_id")).Scan(&result.AccountID, &amount)
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "account not found"})
			return
		}
		if err != nil {
			log.Printf("Balance query failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "unable to retrieve balance"})
			return
		}
		// Keep DECIMAL precision while returning a JSON number rather than a string.
		result.Balance = json.Number(amount)
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Printf("Balance response failed: %v", err)
		}
	})
	return router
}
