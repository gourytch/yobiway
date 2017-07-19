package livecoin

import (
	"encoding/json"
	"strconv"
	"github.com/gourytch/yobiway/client"
	"github.com/gourytch/yobiway/exchange"
)

/*
 https://api.livecoin.net/exchange/all/order_book


{
	FST/BTC: {
		timestamp: 1500444199219,
		asks: [
				[
					"0.00000750", // price
					"74.39991576" // quantity
				],
				...
			],
		bids: [
				[
					"0.00000740",
					"201.97428412"
				],
				...
			]
		}
	},
	...
}
*/

type JOrder []string
type JOrders []JOrder
type JOrderbook struct {
	Timestamp int64   `json:"timestamp"`
	Asks      JOrders `json:"asks"`
	Bids      JOrders `json:"bids"`
}

type JLivecoinOrderbooks map[string]JOrderbook

func (x *LivecoinExchange) load_orderbooks() error {
	var data []byte
	var err error

	if data, err = x.s.Get("https://api.livecoin.net/exchange/all/order_book", client.CACHED_MODE); err != nil {
		return err
	}

	if err = json.Unmarshal(data, &x.jorderbooks); err != nil {
		return err
	}
	return nil
}

func parse_order(J JOrder) exchange.Order {
	var R exchange.Order
	R.Price, _ = strconv.ParseFloat(J[0], 64)
	R.Amount, _ = strconv.ParseFloat(J[1], 64)
	return R
}

func parse_orders(J JOrders) exchange.Orders {
	V := make(exchange.Orders, len(J))
	for ix, j := range J {
		V[ix] = parse_order(j)
	}
	return V
}

func parse_orderbook(J JOrderbook) exchange.Orderbook {
	var B exchange.Orderbook
	B.Asks = parse_orders(J.Asks)
	B.Bids = parse_orders(J.Bids)
	return B
}

func summarize_orders(V exchange.Orders) (value, price float64) {
	value = 0.0
	price = 0.0
	for _, order := range V {
		value += order.Amount
		price += order.Amount * order.Price
	}
	return
}

func (x *LivecoinExchange) apply_orderbooks() {
	var J JOrderbook
	var name string
	for name, J = range x.jorderbooks {
		P := x.pairs[name]
		if P == nil {
			continue
		}
		P.Orderbook = parse_orderbook(J)
		P.Volume_Asks, P.Price_Asks = summarize_orders(P.Orderbook.Asks)
		P.Volume_Bids, P.Price_Bids = summarize_orders(P.Orderbook.Bids)
		if len(P.Orderbook.Asks) > 0 {
			P.Min_Ask = P.Orderbook.Asks[0].Price
			for _, r := range P.Orderbook.Asks {
				if r.Price < P.Min_Ask {
					P.Min_Ask = r.Price
				}
			}
		}
		if len(P.Orderbook.Bids) > 0 {
			P.Max_Bid = P.Orderbook.Bids[0].Price
			for _, r := range P.Orderbook.Bids {
				if P.Max_Bid < r.Price {
					P.Max_Bid = r.Price
				}
			}
		}
		P.Avg_Price = (P.Max_Bid + P.Min_Ask) / 2.0
	}
}
