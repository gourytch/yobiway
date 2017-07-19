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


func (t *Ticker) log() {
	log.Print(t.str())
}
