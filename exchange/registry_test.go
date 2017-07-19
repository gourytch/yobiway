package exchange
import "testing"

type ACMEExchange struct{}

func (x *ACMEExchange) GetName() string {
	return "ACME"
}

func (x *ACMEExchange) Refresh() error {
	return nil
}

func (x *ACMEExchange) GetAllTokens() []string {
	return []string{"TOKN", "CURN"}
}

func (x *ACMEExchange) GetAllCurrencies() []string {
	return []string{"BAR"}
}

func (x *ACMEExchange) GetMarketplace() *Marketplace {
	mp := new(Marketplace)
	tp := new(TradePair)
	tp.Token = "TOKN"
	tp.Currency = "CURN"
	tp.Name = "TOKN/CURN"
	tp.Vwap = 10.00
	mp.Add(tp)
	return mp
}

func (x *ACMEExchange) GetTradePair(name string) *TradePair {
	if name != "TOKN/CURN" {
		return nil
	}
	tp := new(TradePair)
	tp.Token = "TOKN"
	tp.Currency = "CURN"
	tp.Name = "TOKN/CURN"
	tp.Vwap = 10.00
	return tp
}

var ACME *ACMEExchange = new(ACMEExchange)

func TestRegisterExchange(t *testing.T) {
	if Registry == nil {
		t.Error("Registry is nil")
	}
	if len(Registry) != 0 {
		t.Error("Registry not empty")
	}
	err := Register(ACME)
	if err != nil {
		t.Errorf("Register(ACME) returned error: %v", err)
	}
	acme, ok := Registry["ACME"]
	if !ok {
		t.Error("Registry[ACME] not found")
	}
	if acme == nil {
		t.Error("Registry[ACME] is nil")
	}
	if acme.GetName() != "ACME" {
		t.Error("acme has wrong name")
	}
	err = Register(ACME)
	if err == nil {
		t.Error("Register(ACME) must returned error")
	}

}

