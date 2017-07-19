package livecoin

import (
	"encoding/json"
	"strconv"
)

/*
 https://api.livecoin.net/exchange/all/order_book


{
	FST/BTC: {
		timestamp: 1500444199219,
		asks: [
				[
					"0.00000750",
					"74.39991576"
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

type Order struct {
	Price  float64
	Amount float64
}

type Orders []Order
type Orderbook struct {
	Asks Orders
	Bids Orders
}

func (x *LivecoinExchange) load_jorderbooks() error {
	var data []byte
	var err error

	if data, err = x.s.Get("https://api.livecoin.net/exchange/restrictions", false); err != nil {
		return err
	}

	if err = json.Unmarshal(data, &x.jrestrictions); err != nil {
		return err
	}
	return nil
}

func parse_order(J JOrder) Order {
	var R Order
	R.Price, _ = strconv.ParseFloat(J[0], 64)
	R.Amount, _ = strconv.ParseFloat(J[0], 64)
	return R
}

func parse_orders(J JOrders) Orders {
	V := make(Orders, len(J))
	for ix, j := range J {
		V[ix] = parse_order(j)
	}
	return V
}

func parse_orderbook(J JOrderbook) Orderbook {
	var B Orderbook
	B.Asks = parse_orders(J.Asks)
	B.Bids = parse_orders(J.Bids)
	return B
}

func summarize_orders(V Orders) (value, price float64) {
	value = 0.0
	price = 0.0
	for _, order := range V {
		value += order.Amount
		price += order.Amount * order.Price
	}
	return
}

func (x *LivecoinExchange) apply_jorderbooks() {
	var J JOrderbook
	var name string
	for name, J = range x.jorderbooks {
		B := parse_orderbook(J)
		P := x.GetTradePair(name)
		if P == nil {
			continue
		}
		P.Volume_Asks, P.Price_Asks = summarize_orders(B.Asks)
		P.Volume_Bids, P.Price_Bids = summarize_orders(B.Bids)
	}
}
