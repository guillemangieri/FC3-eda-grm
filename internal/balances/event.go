package balances

import (
	"encoding/json"
	"fmt"
	"math"
)

// BalanceUpdated matches the envelope published by Wallet Core.
type BalanceUpdated struct {
	Name    string `json:"Name"`
	Payload struct {
		AccountFrom string      `json:"account_id_from"`
		AccountTo   string      `json:"account_id_to"`
		BalanceFrom json.Number `json:"balance_account_id_from"`
		BalanceTo   json.Number `json:"balance_account_id_to"`
	} `json:"Payload"`
}

func Decode(data []byte) (*BalanceUpdated, error) {
	var event BalanceUpdated
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if event.Name != "BalanceUpdated" || event.Payload.AccountFrom == "" || event.Payload.AccountTo == "" {
		return nil, fmt.Errorf("invalid BalanceUpdated event name or account IDs")
	}
	for _, value := range []json.Number{event.Payload.BalanceFrom, event.Payload.BalanceTo} {
		number, err := value.Float64()
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			return nil, fmt.Errorf("missing or invalid balance: %q", value)
		}
	}
	return &event, nil
}
