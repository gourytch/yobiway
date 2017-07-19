// API URL: https://api.livecoin.net/
// API DOC: https://www.livecoin.net/api?lang=ru
/*
https://api.livecoin.net/info/coinInfo
https://api.livecoin.net/exchange/restrictions
https://api.livecoin.net/exchange/ticker
https://api.livecoin.net/exchange/all/order_book
https://api.livecoin.net/exchange/order_book?currencyPair={token}/{currency}
https://api.livecoin.net/exchange/order_book?currencyPair=DIME/USD
https://api.livecoin.net/exchange/maxbid_minask
*/

package livecoin

import (
	"github.com/gourytch/yobiway/client"
	"github.com/gourytch/yobiway/exchange"
)

type LivecoinExchange struct {
	marketplace   *exchange.Marketplace
	currencies    []string
	tokens        []string
	s             *client.Session
	jtickers      JLivecoinTickers
	jrestrictions JLivecoinRestrictions
	jorderbooks   JLivecoinOrderbooks
}

func (x *LivecoinExchange) GetName() string {
	return "LIVECOIN"
}

func (x *LivecoinExchange) Refresh() error {
	var err error
	if err = x.load_jtickers(); err != nil {
		return err
	}
	x.refresh_tokens()
	if err = x.load_restrictions(); err != nil {
		return err
	}
	x.apply_restrictions()
	if err = x.load_orderbooks(); err != nil {
		return err
	}
	x.apply_orderbooks()
	return nil
}

func (x *LivecoinExchange) GetAllTokens() []string {
	return x.tokens
}

func (x *LivecoinExchange) GetAllCurrencies() []string {
	return x.currencies
}

func (x *LivecoinExchange) GetMarketplace() *exchange.Marketplace {
	return x.marketplace
}

func (x *LivecoinExchange) GetTradePair(name string) *exchange.TradePair {
	if tp, ok := x.marketplace.Pairs[name]; !ok {
		return nil
	} else {
		return tp
	}
}
