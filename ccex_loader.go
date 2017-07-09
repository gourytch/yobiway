// API dox https://c-cex.com/?id=api
// https://c-cex.com/t/prices.json
// market prices: https://c-cex.com/t/api_pub.html?a=getmarketsummaries

package main

import (
	"encoding/json"
	"strings"
)

type JCCexMarketSummary struct {
	MarketName        string  `json:"MarketName"`        // "BTC-888"
	High              float64 `json:"High"`              // 0.00000919,
	Low               float64 `json:"Low"`               // 0.00000820,
	Volume            float64 `json:"Volume"`            // 74339.61396015,
	Last              float64 `json:"Last"`              // 0.00000820,
	BaseVolume        float64 `json:"BaseVolume"`        // 0.64966963,
	TimeStamp         string  `json:"TimeStamp"`         // "2014-07-09T07:19:30.15",
	Bid               float64 `json:"Bid"`               // 0.00000820,
	Ask               float64 `json:"Ask"`               // 0.00000831,
	OpenBuyOrders     int64   `json:"OpenBuyOrders"`     // 15,
	OpenSellOrders    int64   `json:"OpenSellOrders"`    // 15,
	PrevDay           float64 `json:"PrevDay"`           // 0.00000821,
	Created           string  `json:"Created"`           // "2014-03-20T06:00:00",
	DisplayMarketName string  `json:"DisplayMarketName"` // null
}

type JCCexMarketSummaries struct {
	Success bool                 `json:"success"` // true
	Message string               `json:"message"` // ""
	Result  []JCCexMarketSummary `json:"result"`  // [{.....}]
}

func (s *Session) GetCCexTickers() (v []Ticker, err error) {
	var data []byte
	data, err = s.Get("https://c-cex.com/t/api_pub.html?a=getmarketsummaries")
	if err != nil {
		return
	}
	var jresp JCCexMarketSummaries
	err = json.Unmarshal(data, &jresp)
	if err != nil {
		return
	}
	for _, J := range jresp.Result {
		V := strings.Split(J.MarketName, "-")
		v = append(v, Ticker{
			TokenName:     V[0],
			CurrencyName:  V[1],
			High:          J.High,
			Low:           J.Low,
			Average:       (J.High + J.Low) / 2.0,
			Volume:        J.BaseVolume,
			CurrentVolume: J.Volume,
			Last:          J.Last,
			Buy:           J.Bid,
			Sell:          J.Ask,
			Updated:       0.0,
			ServerTS:      0,
		})
	}
	return
}
