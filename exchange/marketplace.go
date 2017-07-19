package exchange

import "fmt"

func NewMarketplace() *Marketplace {
	mp := new(Marketplace)
	mp.Clear()
	return mp
}

func (mp *Marketplace) Clear() {
	mp.Currencies = make(map[string]bool)
	mp.Pairs = make(map[string]*TradePair)
	mp.Pricemap = make(map[string]map[string]float64)
}

func (mp *Marketplace) SetPrice(from, to string, price float64) {
	if mp.Pricemap[from] == nil {
		mp.Pricemap[from] = make(map[string]float64)
	}
	mp.Pricemap[from][to] = price
}

func (mp *Marketplace) GetPrice(from, to string) (price float64, err error) {
	M, ok := mp.Pricemap[from]
	if !ok {
		err = fmt.Errorf("token %v not exists", from)
		return
	}
	price, ok = M[to]
	if !ok {
		err = fmt.Errorf("no way for direction %v->%v", from, to)
		return
	}
	return
}

func (mp *Marketplace) Add(tp *TradePair) {
	mp.Pairs[tp.Name] = tp
	mp.Currencies[tp.Currency] = true
	price := tp.Avg_Price
	mp.SetPrice(tp.Token, tp.Currency, price)
	var reverse_price float64
	if 0.0 < price {
		reverse_price = 1.0 / price
	} else {
		reverse_price = 0.0
	}
	mp.SetPrice(tp.Currency, tp.Token, reverse_price)
}

func (mp *Marketplace) FilterByToken(token string) []*TradePair {
	dsts := mp.Pricemap[token]
	V := make([]*TradePair, len(dsts))
	ix := 0
	for dst := range dsts {
		V[ix] = mp.Pairs[token+"/"+dst]
		ix++
	}
	return V
}

