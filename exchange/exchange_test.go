package exchange

import "testing"

func TestNewMarketplace(t *testing.T) {
	mp := NewMarketplace()
	if mp == nil {
		t.Error("NewMarketplace returned nil")
	}
	if mp.Pairs == nil {
		t.Error("Pairs is nil")
	}
	if len(mp.Pairs) != 0 {
		t.Error("Pairs not empty")
	}
	if mp.Currencies == nil {
		t.Error("Currencies is nil")
	}
	if len(mp.Currencies) != 0 {
		t.Error("Currencies not empty")
	}
	if mp.Pricemap == nil {
		t.Error("Pricemap is nil")
	}
	if len(mp.Pricemap) != 0 {
		t.Error("Pricemap not empty")
	}
}

func TestMarketplace_SetPrice(t *testing.T) {
	mp := NewMarketplace()
	mp.SetPrice("FOO", "BAR", 123.456)
	FOO_Map, ok := mp.Pricemap["FOO"]
	if !ok {
		t.Error("Pricemap[FOO] not found")
	}
	if FOO_Map == nil {
		t.Error("Pricemap[FOO] is nil")
	}
	price, ok := FOO_Map["BAR"]
	if !ok {
		t.Error("Pricemap[FOO][BAR] not found")
	}
	if FOO_Map == nil {
		t.Error("Pricemap[FOO][BAR] is nil")
	}
	if price != 123.456 {
		t.Error("Pricemap[FOO][BAR] not match")
	}
}

func TestMarketplace_GetPrice(t *testing.T) {
	mp := NewMarketplace()
	mp.SetPrice("FOO", "BAR", 123.456)
	var price float64
	var err error

	price, err = mp.GetPrice("FOO", "BAR")
	if err != nil {
		t.Error("GetPrice(FOO,BAR) must return no error")
	}
	if price != 123.456 {
		t.Error("GetPrice(FOO,BAR) must return 123.456")
	}
	price, err = mp.GetPrice("BAR", "FOO")
	if err == nil {
		t.Error("GetPrice(BAR,FOO) must return error")
	}

	mp.SetPrice("BAR", "FOO", 654.321)
	price, err = mp.GetPrice("BAR", "FOO")
	if err != nil {
		t.Error("GetPrice(BAR,FOO) must return no error")
	}
	if price != 654.321 {
		t.Error("GetPrice(BAR,FOO) must return 654.321")
	}

	price, err = mp.GetPrice("KAKA", "BYAKA")
	if err == nil {
		t.Error("GetPrice(KAKA,BYAKA) must return error")
	}
	price, err = mp.GetPrice("FOO", "BYAKA")
	if err == nil {
		t.Error("GetPrice(FOO,BYAKA) must return error")
	}
	price, err = mp.GetPrice("KAKA", "BAR")
	if err == nil {
		t.Error("GetPrice(FOO,BYAKA) must return error")
	}
}

func TestMarketplace_Add(t *testing.T) {
	mp := NewMarketplace()
	tp := &TradePair{
		Token:    "TOKN",
		Currency: "CURN",
		Name:     "TOKN/CURN",
		Volume:   1000.0,
		Vwap:     10.00,
	}
	mp.Add(tp)
	var ok bool
	var price float64
	var TOKN_Map, CURN_Map map[string]float64
	var tp2 *TradePair

	tp2, ok = mp.Pairs["TOKN/CURN"]
	if !ok {
		t.Error("Pairs[TOKN/CURN] not found")
	}
	if tp2 == nil {
		t.Error("Pairs[TOKN/CURN] is nil")
	}
	if tp2.Token != "TOKN" {
		t.Errorf("Pairs[TOKN/CURN].Token = %v, != TOKN", tp2.Token)
	}
	if tp2.Currency != "CURN" {
		t.Error("Pairs[TOKN/CURN].Currency = %v, != CURN", tp2.Currency)
	}
	_, ok = mp.Currencies["CURN"]
	if !ok {
		t.Error("Currencies[CURN] not set")
	}

	TOKN_Map, ok = mp.Pricemap["TOKN"]
	if !ok {
		t.Error("Pricemap[TOKN] not found")
	}
	if TOKN_Map == nil {
		t.Error("Pricemap[TOKN] is nil")
	}
	price, ok = TOKN_Map["CURN"]
	if !ok {
		t.Error("Pricemap[TOKN][CURN] not found")
	}
	if price != 10.00 {
		t.Error("Pricemap[TOKN][CURN] = %v, != 10.00", price)
	}

	CURN_Map, ok = mp.Pricemap["CURN"]
	if !ok {
		t.Error("Pricemap[CURN] not found")
	}
	if CURN_Map == nil {
		t.Error("Pricemap[CURN] is nil")
	}
	price, ok = CURN_Map["TOKN"]
	if !ok {
		t.Error("Pricemap[CURN][TOKN] not found")
	}
	if price != 1.0/10.00 {
		t.Error("Pricemap[CURN][TOKN] != 1.0 / 10.00")
	}
}

func TestMarketplace_FilterByToken(t *testing.T) {
	mp := NewMarketplace()
	mp.Add(&TradePair{Token: "FOO", Currency: "CURN", Name: "FOO/CURN", Vwap: 10.00})
	mp.Add(&TradePair{Token: "BAR", Currency: "CURN", Name: "BAR/CURN", Vwap: 0.01})
	Vbad := mp.FilterByToken("BAD")
	if Vbad == nil {
		t.Error("FilterByToken(BAD) returned nil")
	}
	if len(Vbad) != 0 {
		t.Error("FilterByToken(BAD) returned non-empty")
	}
	Vfoo := mp.FilterByToken("FOO")
	if Vfoo == nil {
		t.Error("FilterByToken(FOO) returned nil")
	}
	if len(Vfoo) != 1 {
		t.Errorf("FilterByToken(FOO) must be single, not %v", len(Vfoo))
	}
	if Vfoo[0].Token != "FOO" {
		t.Errorf("FilterByToken(FOO) return wrong tradepair: %v", Vfoo[0])
	}
	Vbar := mp.FilterByToken("BAR")
	if Vfoo == nil {
		t.Error("FilterByToken(BAR) returned nil")
	}
	if len(Vbar) != 1 {
		t.Errorf("FilterByToken(BAR) must be single, not %v", len(Vbar))
	}
	if Vbar[0].Token != "BAR" {
		t.Errorf("FilterByToken(BAR) return wrong tradepair: %v", Vbar[0])
	}
	Vcurn := mp.FilterByToken("CURN")
	if Vcurn == nil {
		t.Error("FilterByToken(CURN) returned nil")
	}
	if len(Vcurn) != 2 {
		t.Errorf("FilterByToken(BAR) must be two, not %v", len(Vcurn))
	}
}

// test Exchange //

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
	if ExchangesRegistry == nil {
		t.Error("ExchangesRegistry is nil")
	}
	if len(ExchangesRegistry) != 0 {
		t.Error("ExchangesRegistry not empty")
	}
	err := RegisterExchange(ACME)
	if err != nil {
		t.Errorf("RegisterExchange(ACME) returned error: %v", err)
	}
	acme, ok := ExchangesRegistry["ACME"]
	if !ok {
		t.Error("ExchangesRegistry[ACME] not found")
	}
	if acme == nil {
		t.Error("ExchangesRegistry[ACME] is nil")
	}
	if acme.GetName() != "ACME" {
		t.Error("acme has wrong name")
	}
	err = RegisterExchange(ACME)
	if err == nil {
		t.Error("RegisterExchange(ACME) must returned error")
	}

}
