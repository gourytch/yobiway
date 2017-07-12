package main

import (
	"fmt"
)

type TokenInfo struct {
	Token string
	Name  string
}

type Pair struct {
	Name       string  // пара в формате "ТОКЕН/ВАЛЮТА"
	Token      string  // символ токена
	Currency   string  // за какую валюту торгуется?
	Volume     float64 // текущий объём торгов в токенах
	Max_Bid    float64 // максимальная цена приказа покупки
	Min_Ask    float64 // минимальная цена приказа продажи
	Num_Bid    int64   // количество обозреваемых приказов покупки в стакане
	Num_Ask    int64   // количество обозреваемых приказов продажи в стакане
	Sum_Bid    float64 // суммарное количество токенов в обозреваемых приказах покупки в стакане
	Sum_Ask    float64 // суммарное количество токенов в обозреваемых приказах продажи в стакане
	Fee        float64 // процент комиссии на сделку
	Min_Amount float64 // минимальное количество токенов в приказе
}

type Pairs []Pair

type Exchange interface {
	GetName() string             // получить имя биржи
	Refresh() error              // обновить данные по бирже
	GetTokens() []string         // получить список всех активных токенов, пользующихся на бирже
	GetCurrencies() []string     // получить список всех активных валют, пользующихся на бирже
	GetAllPairs() Pairs          // получить список всех активных пар токен:валюта
	GetPairs(Token string) Pairs // получить список всех активных пар с участием токена
}

var XCGS map[string]Exchange

func RegisterExchange(xcg Exchange) error {
	name := xcg.GetName()
	if _, ok := XCGS[name]; ok {
		return fmt.Errorf("exchange already registered: `%s`", name)
	}
	XCGS[name] = xcg
	return nil
}
