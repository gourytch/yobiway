package livecoin

import (
	"github.com/gourytch/yobiway/exchange"
	"testing"
)

func TestRegister(t *testing.T) {
	Register()
	x, ok := exchange.Registry["LIVECOIN"]
	if !ok {
		t.Error("LIVECOIN not in Registry")
	}
	if x == nil {
		t.Error("Registry[LIVECOIN] is nil")
	}
	if s := x.GetName(); s != "LIVECOIN" {
		t.Errorf("Registry[LIVECOIN].GetName() = %v, != LIVECOIN", s)
	}
}
