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
	"strings"
	"encoding/json"
)

type JLivecoinTicker struct {
	Token    string  `json:"cur"`    // "USD"
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

type JLivecoinTickers []JLivecoinTicker

type LivecoinExchange struct {
	marketplace *Marketplace
	currencies []string
	tokens []string
	s *Session
}

var livecoin *LivecoinExchange = &LivecoinExchange {
	marketplace: NewMarketplace(),
	currencies: make([]string,0),
	tokens: make([]string,0),
	s: NewSession(),
}

func (x *LivecoinExchange) GetName() string {
	return "LIVECOIN"
}

func (x *LivecoinExchange) load_tickers() error {
	var data []byte
	var err error

	if data, err = x.s.Get("https://api.livecoin.net/exchange/ticker", CACHED); err != nil {
		return err
	}
	var jresp JLivecoinTickers
	if err = json.Unmarshal(data, &jresp); err != nil {
		return err
	}
	x.marketplace.Clear()

	for _, J := range jresp {
		V := strings.Split(J.Symbol, "/")
		tp := &TradePair {
			Name:J.Symbol,
			URL:"",
			Token:J.Token,
			Currency: V[1],
			Vwap:J.VWap,
			Volume:J.Volume,
			Volume24H:J.Volume,
			Max_Bid:J.MaxBid,
			Min_Ask:J.MinAsk,
			Num_Bid:-1,
			Num_Ask:-1,
			Sum_Bid:-1,
			Sum_Ask:-1,
			Num_Trades:-1,
			BuyFee:0.002,
			SellFee:0.002,
			Min_Amount:-1,
		}
		x.marketplace.Add(tp)
	}
	x.tokens = make([]string,len(x.marketplace.Pricemap))
	ix := 0
	for token := range(x.marketplace.Pricemap) {
		x.tokens[ix] = token
		ix++
	}
	x.currencies = make([]string,len(x.marketplace.Currencies))
	ix = 0
	for token := range(x.marketplace.Currencies) {
		x.currencies[ix] = token
		ix++
	}
	return nil
}

func (x *LivecoinExchange) Refresh() error {
	var err error
	if err = x.load_tickers(); err != nil {
		return err
	}

	return nil
}

func (x *LivecoinExchange) GetAllTokens() []string {
	return x.tokens
}

func (x *LivecoinExchange) GetAllCurrencies() []string {
	return x.currencies
}

func (x *LivecoinExchange) GetMarketplace() *Marketplace {
	return x.marketplace
}

func (x *LivecoinExchange) GetTradePair(name string) *TradePair {
	if tp,ok := x.marketplace.Pairs[name]; !ok {
		return nil
	} else {
		return tp
	}
}
