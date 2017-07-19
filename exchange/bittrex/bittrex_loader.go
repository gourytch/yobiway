// API Description
// https://bittrex.com/Home/Api
//
// markets: https://bittrex.com/api/v1.1/public/getmarkets
//
// currencies: https://bittrex.com/api/v1.1/public/getcurrencies
//
// ticker: https://bittrex.com/api/v1.1/public/getticker?market={BASE}-{TRADE}
// example: https://bittrex.com/api/v1.1/public/getticker?market=BTC-LTC
//
// market summaries: https://bittrex.com/api/v1.1/public/getmarketsummaries
// market summary: https://bittrex.com/api/v1.1/public/getmarketsummary?market={BASE}-{TRADE}

package bittrex

import (
	"encoding/json"
	"strings"
)

type JBittrexMarketSummary struct {
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

type JBittrexMarketSummaries struct {
	Success bool                    `json:"success"` // true
	Message string                  `json:"message"` // ""
	Result  []JBittrexMarketSummary `json:"result"`  // [{.....}]
}

func (s *Session) GetBittrexTickers() (v []Ticker, err error) {
	var data []byte
	data, err = s.Get("https://bittrex.com/api/v1.1/public/getmarketsummaries", CACHED)
	if err != nil {
		return
	}
	var jresp JBittrexMarketSummaries
	err = json.Unmarshal(data, &jresp)
	if err != nil {
		return
	}
	for _, J := range jresp.Result {
		V := strings.Split(J.MarketName, "-")
		v = append(v, Ticker{
			TokenName:     V[1],
			CurrencyName:  V[0],
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
