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
