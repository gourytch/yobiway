package main

import "testing"

func TestNewMarketplace(t *testing.T) {
	mp := NewMarketplace()
	if mp == nil {
		t.Error("NewMarketplace returned nil")
	}
	if len(mp.Pairs) != 0 {
		t.Error("Pairs not empty")
	}
	if len(mp.Currencies) != 0 {
		t.Error("Currencies not empty")
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

func TestMarketplace_Add(t *testing.T) {
	mp := NewMarketplace()
	tp := new(TradePair)
	tp.Token = "TOKN"
	tp.Currency = "CURN"
	tp.Name = "TOKN/CURN"
	tp.Vwap = 10.00
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
		t.Error("Pairs[TOKN/CURN].Token != TOKN")
	}
	if tp2.Currency != "CURN" {
		t.Error("Pairs[TOKN/CURN].Currency != CURN")
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
		t.Error("Pricemap[TOKN][CURN] != 10.00")
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
