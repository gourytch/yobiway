package livecoin

import (
	"github.com/gourytch/yobiway/exchange"
	"testing"
)

func TestRegister(t *testing.T) {
	Register()
	x, ok := exchange.ExchangesRegistry["LIVECOIN"]
	if !ok {
		t.Error("LIVECOIN not in ExchangesRegistry")
	}
	if x == nil {
		t.Error("ExchangesRegistry[LIVECOIN] is nil")
	}
	if s := x.GetName(); s != "LIVECOIN" {
		t.Errorf("ExchangesRegistry[LIVECOIN].GetName() = %v, != LIVECOIN", s)
	}
}
