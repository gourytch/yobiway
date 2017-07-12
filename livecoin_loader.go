// API URL: https://api.livecoin.net/
// API DOC: https://www.livecoin.net/api?lang=ru
/*
https://api.livecoin.net/info/coinInfo
https://api.livecoin.net/exchange/restrictions
https://api.livecoin.net/exchange/ticker
https://api.livecoin.net/exchange/all/order_book
https://api.livecoin.net/exchange/order_book?currencyPair={token}/{currency}
https://api.livecoin.net/exchange/order_book?currencyPair=DIME/USD
https://api.livecoin.net/exchange/maxbid_minask
*/

// https://api.livecoin.net/exchange/ticker
/* [
{
cur: "USD",
symbol: "USD/RUR",
last: 59.99999,
high: 61.9,
low: 59.1,
volume: 922.72676008,
vwap: 60.24952267,
max_bid: 61.9,
min_ask: 59.1,
best_bid: 59.1,
best_ask: 59.99999
},
]
*/

package main

import (
	"encoding/json"
	"strings"
)

type JLivecoinExchangeTicker struct {
	Currency string  `json:"cur"`    // "USD"
	Symbol   string  `json:"symbol"` // "USD/RUR"
	Last     float64 `json:"last"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Volume   float64 `json:"volume"`
	VWap     float64 `json:"vwap"` // volume weighted average price
	MaxBid   float64 `json:"max_bid"`
	MinAsk   float64 `json:"min_ask"`
	BestBid  float64 `json:"best_bid"`
	BestAsk  float64 `json:"best_ask"`
}

type JLivecoinExchangeTickers []JLivecoinExchangeTicker

func (s *Session) GetLivecoinTickers() (v []Ticker, err error) {
	var data []byte
	data, err = s.Get("https://api.livecoin.net/exchange/ticker", CACHED)
	if err != nil {
		return
	}
	var jresp JLivecoinExchangeTickers
	err = json.Unmarshal(data, &jresp)
	if err != nil {
		return
	}
	for _, J := range jresp {
		V := strings.Split(J.Symbol, "/")
		var token, currency string
		if V[0] == J.Currency { // FIXME who is who ???
			currency, token = V[1], V[0]
		} else {
			currency, token = V[0], V[1]
		}
		v = append(v, Ticker{
			TokenName:     token,
			CurrencyName:  currency,
			High:          J.High,
			Low:           J.Low,
			Average:       J.VWap,
			Volume:        J.Volume,
			CurrentVolume: J.Volume,
			Last:          J.Last,
			Buy:           J.MaxBid,
			Sell:          J.MinAsk,
			Updated:       0.0,
			ServerTS:      0,
		})
	}
	return
}
