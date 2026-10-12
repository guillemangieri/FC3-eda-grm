package balances

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestHTTPMySQL(t *testing.T) {
	dsn := os.Getenv("BALANCES_TEST_DSN")
	if dsn == "" {
		t.Skip("set BALANCES_TEST_DSN to run MySQL integration test")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	id := uuid.NewString()
	if _, err := db.Exec("INSERT INTO account_balances (account_id, balance) VALUES (?, ?)", id, "9007199254740991.12345678"); err != nil {
		t.Fatal(err)
	}
	defer db.Exec("DELETE FROM account_balances WHERE account_id = ?", id)
	handler := NewHTTPHandler(db)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/balances/"+id, nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status/headers: %d %v", response.Code, response.Header())
	}
	var result Balance
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.AccountID != id || result.Balance.String() != "9007199254740991.12345678" {
		t.Fatalf("lost decimal precision: %s", response.Body.String())
	}
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{http.MethodGet, "/balances/" + uuid.NewString(), http.StatusNotFound},
		{http.MethodGet, "/balances/%27%20OR%201=1--", http.StatusNotFound},
		{http.MethodPost, "/balances/" + id, http.StatusMethodNotAllowed},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))
		if response.Code != tc.status {
			t.Errorf("%s %s: got %d, want %d", tc.method, tc.path, response.Code, tc.status)
		}
	}
}

func TestHTTPDatabaseFailure(t *testing.T) {
	db, err := sql.Open("mysql", "test:test@tcp(localhost:1)/test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	response := httptest.NewRecorder()
	NewHTTPHandler(db).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/balances/test", nil))
	if response.Code != http.StatusInternalServerError || response.Body.String() != "{\"error\":\"unable to retrieve balance\"}\n" {
		t.Fatalf("unexpected error response: %d %s", response.Code, response.Body.String())
	}
}
