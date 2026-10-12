package balances

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

func TestApplyMySQL(t *testing.T) {
	dsn := os.Getenv("BALANCES_TEST_DSN")
	if dsn == "" {
		t.Skip("set BALANCES_TEST_DSN to run MySQL integration test")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	topic := "test-" + uuid.NewString()
	a, b := uuid.NewString(), uuid.NewString()
	defer db.Exec("DELETE FROM account_balances WHERE account_id IN (?, ?)", a, b)
	defer db.Exec("DELETE FROM consumer_offsets WHERE topic = ?", topic)
	event := &BalanceUpdated{Name: "BalanceUpdated"}
	event.Payload.AccountFrom, event.Payload.AccountTo = a, b
	event.Payload.BalanceFrom, event.Payload.BalanceTo = json.Number("900.25"), json.Number("600.75")
	if applied, err := Apply(ctx, db, topic, 0, 10, event); err != nil || !applied {
		t.Fatalf("initial apply: %t, %v", applied, err)
	}
	event.Payload.BalanceFrom = json.Number("1")
	for _, offset := range []int64{10, 9} {
		if applied, err := Apply(ctx, db, topic, 0, offset, event); err != nil || applied {
			t.Fatalf("replayed offset %d: %t, %v", offset, applied, err)
		}
	}
	// The second write fails; the first balance and the offset must roll back.
	event.Payload.AccountTo = strings.Repeat("x", 256)
	if _, err := Apply(ctx, db, topic, 0, 11, event); err == nil {
		t.Fatal("expected oversized account ID to fail")
	}
	var balance string
	if err := db.QueryRow("SELECT balance FROM account_balances WHERE account_id = ?", a).Scan(&balance); err != nil || balance != "900.25000000" {
		t.Fatalf("balance rollback: %q, %v", balance, err)
	}
	if err := db.QueryRow("SELECT balance FROM account_balances WHERE account_id = ?", b).Scan(&balance); err != nil || balance != "600.75000000" {
		t.Fatalf("destination balance: %q, %v", balance, err)
	}
	var offset int64
	if err := db.QueryRow("SELECT last_offset FROM consumer_offsets WHERE topic = ? AND partition_id = 0", topic).Scan(&offset); err != nil || offset != 10 {
		t.Fatalf("offset rollback: %d, %v", offset, err)
	}
	event.Payload.AccountTo = b
	if applied, err := Apply(ctx, db, topic, 0, 11, event); err != nil || !applied {
		t.Fatalf("next offset: %t, %v", applied, err)
	}
}
