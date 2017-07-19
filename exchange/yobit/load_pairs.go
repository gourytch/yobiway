/*
https://yobit.net/api/3/info
{
	server_time: 1500453748,
	pairs: {
		olymp_btc: {
			decimal_places: 8,
			min_price: 1e-8,
			max_price: 10000,
			min_amount: 0.0001,
			min_total: 0.0001,
			hidden: 0,
			fee: 0.2,
			fee_buyer: 0.2,
			fee_seller: 0.2
		},
		btc_rur: {
			decimal_places: 8,
			min_price: 1e-8,
			max_price: 10000,
			min_amount: 0.0001,
			min_total: 0.0001,
			hidden: 0,
			fee: 0.2,
			fee_buyer: 0.2,
			fee_seller: 0.2
		}...
}
 */
package yobit


import (
	"encoding/json"
	"sort"
	"github.com/gourytch/yobiway/client"
)

type JYobitPair struct {
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

type JYobitPairs struct {
	ServerTime int64 	`json:"server_time"`
	Pairs map[string]JYobitPair `json:"pairs"`
}

func (x *YobitExchange) load_pairs() error {
	var data []byte
	var err error
	if data, err = x.s.Get("https://yobit.net/api/3/info", client.CACHED_MODE); err != nil {
		return err
	}
	if err = json.Unmarshal(data, &x.jpairs); err != nil {
		return err
	}
	return nil
}

func (x *YobitExchange) collect_pairs() {
	x.pairnames = make([]string, len(x.jpairs.Pairs))
	ix := 0
	for k, _ := range x.jpairs.Pairs {
		x.pairnames[ix] = k
		ix++
	}
	sort.Strings(x.pairnames)
}

func (x *YobitExchange) parse_pairs() {
	for k, v := range x.jpairs.Pairs {
		if v.Hidden == 0 {

			tpname,_,_ := parse_yobit_pairname(k)
			tp := x.GetTradePair(tpname)
			if tp == nil {
				continue
			}
			tp.Min_Amount = v.MinAmount
		}
	}
	return
}
