package yobit

import (
	"github.com/gourytch/yobiway/client"
	"github.com/gourytch/yobiway/exchange"
)

func Register() {
	exchange.Register(&YobitExchange{
		marketplace: exchange.NewMarketplace(),
		currencies:  make([]string, 0),
		tokens:      make([]string, 0),
		s:           client.NewSession(),
	})
}

