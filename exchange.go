package main

import (
	"fmt"
)

type TokenInfo struct {
	Token string
	Name  string
}

type TradePair struct {
	Name       string  // пара в формате "ТОКЕН/ВАЛЮТА" капсом (req)
	URL        string  // адрес торговой пары на бирже (если есть)
	Token      string  // символ токена (req)
	Currency   string  // за какую валюту торгуется? (req)
	Vwap       float64 // Volume Weighted Average Price (Средневзвешенная цена)
	Volume     float64 // текущий объём торгов в токенах
	Volume24H  float64 // объём торгов в токенах за последние сутки
	Max_Bid    float64 // максимальная цена приказа покупки
	Min_Ask    float64 // минимальная цена приказа продажи
	Num_Bid    int64   // количество обозреваемых приказов покупки в стакане
	Num_Ask    int64   // количество обозреваемых приказов продажи в стакане
	Sum_Bid    float64 // суммарное количество токенов в обозреваемых приказах покупки в стакане
	Sum_Ask    float64 // суммарное количество токенов в обозреваемых приказах продажи в стакане
	Num_Trades int64   // число совершенных сделок за последний час
	BuyFee     float64 // комиссиионный процент на покупку
	SellFee    float64 // комиссиионный процент на продажу
	Min_Amount float64 // минимальное количество токенов в приказе
}

type Marketplace struct {
	Pairs      map[string]*TradePair         // каталог всех торговых пар
	Currencies map[string]bool               // перечень всех токенов-валют на рынке
	Pricemap   map[string]map[string]float64 // словарь токен->имена торговых пар в которых его можно купить-продать
}

type Exchange interface {
	GetName() string                     // получить имя биржи
	Refresh() error                      // обновить данные по бирже
	GetAllTokens() []string              // получить список всех активных токенов, пользующихся на бирже
	GetAllCurrencies() []string          // получить список всех активных валют, пользующихся на бирже
	GetMarketplace() *Marketplace        //  получить описание рынка
	GetTradePair(name string) *TradePair // получить отдельную пару
}

var ExchangesRegistry map[string]Exchange = make(map[string]Exchange)

func RegisterExchange(xcg Exchange) error {
	name := xcg.GetName()
	if _, ok := ExchangesRegistry[name]; ok {
		return fmt.Errorf("exchange already registered: `%s`", name)
	}
	ExchangesRegistry[name] = xcg
	return nil
}

func NewMarketplace() *Marketplace {
	mp := new(Marketplace)
	mp.Currencies = make(map[string]bool)
	mp.Pairs = make(map[string]*TradePair)
	mp.Pricemap = make(map[string]map[string]float64)
	return mp
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
	mp.SetPrice(tp.Token, tp.Currency, tp.Vwap)
	mp.SetPrice(tp.Currency, tp.Token, 1.0/tp.Vwap)
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
