package livecoin

import (
	"github.com/gourytch/yobiway/client"
	"github.com/gourytch/yobiway/exchange"
)

func Register() {
	exchange.RegisterExchange(&LivecoinExchange{
		marketplace: exchange.NewMarketplace(),
		currencies:  make([]string, 0),
		tokens:      make([]string, 0),
		s:           client.NewSession(),
	})
}
