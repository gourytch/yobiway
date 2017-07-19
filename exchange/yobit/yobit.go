package yobit

import (
	"github.com/gourytch/yobiway/client"
	"github.com/gourytch/yobiway/exchange"
)

type YobitExchange struct {
	marketplace   *exchange.Marketplace
	currencies    []string
	tokens        []string
	s             *client.Session
	pairnames   []string // in yobit form "token_currency"
	jpairs 		JYobitPairs
	jtickers    JYobitTickers
}

func (x *YobitExchange) GetName() string {
	return "YOBIT"
}

func (x *YobitExchange) Refresh() error {
	var err error
	if err = x.load_pairs(); err != nil {
		return err
	}
	x.collect_pairs()
	if err = x.load_tickers(); err != nil {
		return err
	}
	x.set_tickers()
	x.parse_pairs()
	return nil
}

func (x *YobitExchange) GetAllTokens() []string {
	return x.tokens
}

func (x *YobitExchange) GetAllCurrencies() []string {
	return x.currencies
}

func (x *YobitExchange) GetMarketplace() *exchange.Marketplace {
	return x.marketplace
}

func (x *YobitExchange) GetTradePair(name string) *exchange.TradePair {
	if tp, ok := x.marketplace.Pairs[name]; !ok {
		return nil
	} else {
		return tp
	}
}
