/*
https://yobit.net/api/3/ticker/btc_rur-btc_usd
{
	btc_rur: {
		high: 145500,
		low: 136966.29477383,
		avg: 141233.14738691,
		vol: 12504725.43288324,
		vol_cur: 88.23927851,
		last: 141999.03135887,
		buy: 141990.00000001,
		sell: 141999.03135887,
		updated: 1500453527
	},
	btc_usd: {
		high: 2407.26403588,
		low: 2200,
		avg: 2303.63201794,
		vol: 200239.28162915,
		vol_cur: 85.79061219,
		last: 2378.20280836,
		buy: 2378.20280836,
		sell: 2394.5499999,
		updated: 1500453543
	}
}

 */
package yobit

import (
	"encoding/json"
	"strings"
	"github.com/gourytch/yobiway/exchange"
	"github.com/gourytch/yobiway/client"
)

const MAX_TICKERS_REQ = 50

type JYobitTicker struct {
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

type JYobitTickers map[string]JYobitTicker

func (x *YobitExchange) load_tickers() error {
	var data []byte
	var err error
	L := len(x.pairnames)
	offs := 0
	for offs < L {
		r := offs + MAX_TICKERS_REQ
		if L < r {
			r = L
		}
		//log.Printf("process slice [%d:%d]", offs, r)
		P := x.pairnames[offs:r]
		Ps := strings.Join(P, "-")
		if data, err = x.s.Get("https://yobit.net/api/3/ticker/"+Ps, client.CACHED_MODE); err != nil {
			return err
		}
		var j JYobitTickers
		err = json.Unmarshal(data, &j)
		if err != nil {
			return err
		}
		for jk, jv := range j {
			x.jtickers[jk] = jv
		}
		offs = r
	}
	return nil
}

func (x *YobitExchange) set_tickers() {
	for jk, jv := range x.jtickers {
		pairname, token, currency := parse_yobit_pairname(jk)
		x.marketplace.Add(&exchange.TradePair{
			Name:      pairname,
			URL:       "",
			Token:     token,
			Currency:  currency,
			Vwap:      jv.Average,
			Volume:    jv.CurrentVolume,
			Volume24H: jv.Volume,
			Max_Bid:   jv.Buy,
			Min_Ask:   jv.Sell,
		})
	}
}
