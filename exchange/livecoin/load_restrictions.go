/*
{
	success: true,
	minBtcVolume: 0.0001,
	restrictions: [
		{
			currencyPair: "BTC/USD",
			minLimitQuantity: 0.0001,
			priceScale: 5
		},
		...
	]
}
*/

package livecoin

import (
	"encoding/json"
	"github.com/gourytch/yobiway/exchange"
	"github.com/gourytch/yobiway/client"
)

type JLivecoinRestriction struct {
	CurrencyPair     string  `json:"currencyPair"`
	MinLimitQuantity float64 `json:"minLimitQuantity"`
	PriceScale       int     `json:"priceScale"`
}

type JLivecoinRestrictions struct {
	Success      bool                   `json:"success"`
	MinBTCVolume float64                `json:"minBtcVolume"`
	Restrictions []JLivecoinRestriction `json:"restrictions"`
}

func (x *LivecoinExchange) load_restrictions() error {
	var data []byte
	var err error

	if data, err = x.s.Get("https://api.livecoin.net/exchange/restrictions", client.CACHED_MODE); err != nil {
		return err
	}
	if err = json.Unmarshal(data, &x.jrestrictions); err != nil {
		return err
	}
	return nil
}

func (x *LivecoinExchange) apply_restrictions() {
	var R JLivecoinRestriction
	for _, R = range x.jrestrictions.Restrictions {
		var P *exchange.TradePair = x.GetTradePair(R.CurrencyPair)
		if P == nil {
			continue // missing pair
		}
		P.Min_Amount = R.MinLimitQuantity
	}
}
