package main

import (
	"log"
	"fmt"
)

/*
const (
	EXC_LIVECOIN = "livecoin"
	EXC_BITTREX  = "bittrex"
	EXC_CCEX     = "ccex"
	EXC_YOBIT    = "yobit"
)
*/

type Ticker struct {
	TokenName     string  `json:"-"` // имя фантика
	CurrencyName  string  `json:"-"` // имя валюты (btc/usd/rur/etc.)
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Average       float64 `json:"avg"`
	Volume        float64 `json:"vol"`
	CurrentVolume float64 `json:"vol_cur"`
	Last          float64 `json:"last"`
	Buy           float64 `json:"buy"`
	Sell          float64 `json:"sell"`
	Updated       float64 `json:"updated"`
	ServerTS      int64   `json:"server_time"`
}

type JTicker struct {
	T Ticker `json:"ticker"`
}

type JTickers map[string]Ticker

type PairDesc struct {
	TokenName     string  `json:"-"` // имя фантика
	CurrencyName  string  `json:"-"` // имя валюты (btc/usd/rur/etc.)
	DecimalPlaces int     `json:"decimal_places"`
	MinPrice      float64 `json:"min_price"`
	MaxPrice      float64 `json:"max_price"`
	MinAmount     float64 `json:"min_amount"`
	MinTotal      float64 `json:"min_total"`
	Hidden        int     `json:"hidden"`
	Fee           float64 `json:"fee"`
	FeeByer       float64 `json:"fee_buyer"`
	FeeSeller     float64 `json:"fee_seller"`
}

type JPairs struct {
	Pairs map[string]PairDesc `json:"pairs"`
}

func (t *Ticker) str() string {
	return fmt.Sprintf("pair     : %s_%s\n", t.TokenName, t.CurrencyName) +
		fmt.Sprintf("Lo/Hi    : %f .. %f\n", t.Low, t.High) +
		fmt.Sprintf("Vol/Cur  : %f / %f\n", t.Volume, t.CurrentVolume) +
		fmt.Sprintf("Last     : %f\n", t.Last) +
		fmt.Sprintf("Average  : %f\n", t.Average) +
		fmt.Sprintf("Buy/Sell : %f / %f\n", t.Buy, t.Sell)
}

func (t *Ticker) log() {
	log.Print(t.str())
}

type Alphabetically []string

func (a Alphabetically) Len() int           { return len(a) }
func (a Alphabetically) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Alphabetically) Less(i, j int) bool { return a[i] < a[j] }
