package yobit

import "testing"

func Test_parse_yobit_pairname(t *testing.T) {
	pairname, token, currency := parse_yobit_pairname("btc_usd")
	if pairname != "BTC/USD" {
		t.Errorf("pairname=%v, not BTC/USD", pairname)
	}
	if token != "BTC" {
		t.Errorf("token=%v, not BTC", token)
	}
	if currency != "USD" {
		t.Errorf("currency=%v, not USD", currency)
	}
}
