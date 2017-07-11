package main

import (
	"fmt"
)

type Pair struct {
	Token    string
	Currency string
	Low      float64
	Avg      float64
	High     float64
	Volume   float64
	Fee      float64
}

type Pairs []Pair

type Exchange interface {
	GetName() string // получить имя биржи
	Refresh(cached bool) error // обновить данные по бирже
	GetTokens() []string // получить список всех активных токенов, пользующихся на бирже
	GetCurrencies() []string // получить список всех активных валют, пользующихся на бирже
	GetAllPairs() Pairs // получить список всех активных пар токен:валюта
	GetPairs(Token string) Pairs // получить список всех активных пар, отфильтрованных по токену (или валюте)
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
