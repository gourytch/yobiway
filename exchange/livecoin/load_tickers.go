/*
 https://api.livecoin.net/exchange/ticker
[
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
	...
]
*/

package livecoin

import (
	"encoding/json"
	"github.com/gourytch/yobiway/exchange"
	"strings"
	"fmt"
	"github.com/gourytch/yobiway/client"
)

type JLivecoinTicker struct {
	Token   string  `json:"cur"`    // "USD"
	Symbol  string  `json:"symbol"` // "USD/RUR"
	Last    float64 `json:"last"`
	High    float64 `json:"high"`
	Low     float64 `json:"low"`
	Volume  float64 `json:"volume"`
	VWap    float64 `json:"vwap"` // volume weighted average price
	MaxBid  float64 `json:"max_bid"`
	MinAsk  float64 `json:"min_ask"`
	BestBid float64 `json:"best_bid"`
	BestAsk float64 `json:"best_ask"`
}

type JLivecoinTickers []JLivecoinTicker

func (x *LivecoinExchange) load_tickers() error {
	var data []byte
	var err error

	if data, err = x.s.Get("https://api.livecoin.net/exchange/ticker", client.CACHED_MODE); err != nil {
		return err
	}
	if err = json.Unmarshal(data, &x.jtickers); err != nil {
		return err
	}
	return nil
}

func (x *LivecoinExchange) parse_tradepairs() error {
	var J JLivecoinTicker
	x.pairs = map[string]*exchange.TradePair{}
	for _, J = range x.jtickers {
		V := strings.Split(J.Symbol, "/")
		tp := &exchange.TradePair{
			Name:      J.Symbol,
			URL:       fmt.Sprintf("https://www.livecoin.net/en/trade/index?currencyPair=%s", J.Symbol),
			Token:     J.Token,
			Currency:  V[1],
			Vwap:      J.VWap,
			Volume:    J.Volume,
			Volume24H: J.Volume,
			Max_Bid:   J.MaxBid,
			Min_Ask:   J.MinAsk,
			BuyFee:    0.002,
			SellFee:   0.002,
		}
		x.pairs[tp.Name] = tp
	}
	return nil
}

func (x *LivecoinExchange) generate_marketplace() {
	for _, tp := range x.pairs {
		x.marketplace.Add(tp)
	}

	x.tokens = make([]string, len(x.marketplace.Pricemap))
	ix := 0
	for token := range x.marketplace.Pricemap {
		x.tokens[ix] = token
		ix++
	}
	x.currencies = make([]string, len(x.marketplace.Currencies))
	ix = 0
	for token := range x.marketplace.Currencies {
		x.currencies[ix] = token
		ix++
	}
}
