package livecoin

import "testing"

func TestParse_order(t *testing.T) {
	J := []string{"0.00000750", "74.39991576"}
	r := parse_order(J)
	if r.Price != 0.00000750 {
		t.Errorf("Price=%v, != 0.00000750", r.Price)
	}
	if r.Amount != 74.39991576 {
		t.Errorf("Amount=%v, != 74.39991576", r.Amount)
	}
}

func TestParse_orders(t *testing.T) {
	J := [][]string{
		{"0.14000000", "79.99999000"},
		{"0.14000000", "10.00000000"},
		{"0.15000000", "105.00000000"},
	}
	v := parse_orders(J)
	if l := len(v); l != 3 {
		t.Errorf("len(v) = %v, != 3", l)

	}
	if v[0].Price != 0.14000000 {
		t.Errorf("[0]Price=%v, != 0.14000000", v[0].Price)
	}
	if v[0].Amount != 79.99999000 {
		t.Errorf("[0]Amount=%v, != 79.99999000", v[0].Amount)
	}
	if v[1].Price != 0.14000000 {
		t.Errorf("[1]Price=%v, != 0.14000000", v[1].Price)
	}
	if v[1].Amount != 10.00000000 {
		t.Errorf("[1]Amount=%v, != 10.00000000", v[1].Amount)
	}
	if v[2].Price != 0.15000000 {
		t.Errorf("[2]Price=%v, != 0.15000000", v[2].Price)
	}
	if v[2].Amount != 105.00000000 {
		t.Errorf("[2]Amount=%v, != 105.00000000", v[2].Amount)
	}
}
