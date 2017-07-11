package main

import (
	"encoding/json"
	"sort"
	"strings"
)

const MAX_TICKERS_REQ = 50

// pair :: 'ltc_btc'
func (s *Session) GetTicker(pair string) (t Ticker, err error) {
	token, currency, err := SplitPair(pair)
	if err != nil {
		return
	}
	data, err := s.Get("https://yobit.net/api/3/" + pair + "/ticker", CACHED)
	if err != nil {
		return
	}
	var j JTicker
	err = json.Unmarshal(data, &j)
	if err != nil {
		return
	}
	t = j.T
	t.TokenName = token
	t.CurrencyName = currency
	return
}

// pair :: 'ltc_btc'
func (s *Session) GetTickers(pairs []string) (v []Ticker, err error) {
	sorted_pairs := append([]string(nil), pairs...)
	sort.Sort(Alphabetically(sorted_pairs))

	L := len(sorted_pairs)
	offs := 0
	for offs < L {
		r := offs + MAX_TICKERS_REQ
		if L < r {
			r = L
		}
		//log.Printf("process slice [%d:%d]", offs, r)
		P := sorted_pairs[offs:r]
		Ps := strings.Join(P, "-")
		var data []byte
		data, err = s.Get("https://yobit.net/api/3/ticker/" + Ps, CACHED)
		if err != nil {
			return
		}
		var j JTickers
		err = json.Unmarshal(data, &j)
		if err != nil {
			return
		}
		for jk, jv := range j {
			var token, currency string
			token, currency, err = SplitPair(jk)
			if err != nil {
				return
			}
			jv.TokenName = token
			jv.CurrencyName = currency
			v = append(v, jv)
		}
		offs = r
	}
	return
}

func (s *Session) GetPairs() (pairs []string, err error) {
	data, err := s.Get("https://yobit.net/api/3/info", CACHED)
	if err != nil {
		return
	}
	var j JPairs
	err = json.Unmarshal(data, &j)
	if err != nil {
		return
	}
	pairs = nil
	for k, v := range j.Pairs {
		if v.Hidden == 0 {
			pairs = append(pairs, k)
		}
	}
	return
}
