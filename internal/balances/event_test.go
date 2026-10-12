package balances

import "testing"

func TestDecode(t *testing.T) {
	valid := `{"Name":"BalanceUpdated","Payload":{"account_id_from":"a","account_id_to":"b","balance_account_id_from":900.25,"balance_account_id_to":0}}`
	event, err := Decode([]byte(valid))
	if err != nil || event.Payload.BalanceFrom.String() != "900.25" || event.Payload.BalanceTo.String() != "0" {
		t.Fatalf("valid Wallet event was not decoded: %+v, %v", event, err)
	}
	for _, invalid := range []string{
		`{`,
		`{"Name":"TransactionCreated","Payload":{}}`,
		`{"Name":"BalanceUpdated","Payload":{"account_id_from":"a","account_id_to":"b","balance_account_id_from":900}}`,
		`{"Name":"BalanceUpdated","Payload":{"account_id_from":"","account_id_to":"b","balance_account_id_from":900,"balance_account_id_to":600}}`,
	} {
		if _, err := Decode([]byte(invalid)); err == nil {
			t.Errorf("accepted invalid event: %s", invalid)
		}
	}
}
