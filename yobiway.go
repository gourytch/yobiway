package main

import (
	"fmt"
	"log"
	"math"
	"sort"
	//	"sort"
	"strings"

	"github.com/gourytch/loophole"
)

var YOBI_FEE float64 = 0.2
var YOBI_FEE_K float64 = (1.0 - YOBI_FEE/100.0)
var BITTREX_FEE float64 = 0.25
var BITTREX_FEE_K float64 = (1.0 - BITTREX_FEE/100.0)

var MAX_DISPERSION float64 = 0.2
var MIN_PRICE float64 = 0.000001 // 0.0000001
var MIN_VOLUME float64 = 0.00001
var BEST_LIMIT int = 3

type NodeNames map[loophole.Node]string
type NameNodes map[string]loophole.Node

var session *Session

var nodenames NodeNames = NodeNames{}
var namenodes NameNodes = NameNodes{}
var graph loophole.Graph = loophole.Graph{}
var cur_fee float64
var cur_fee_k float64

//var all_pairs []string
var all_tickers []Ticker

const (
	GREEDY_MODEL = iota
	AVERAGE_MODEL
	SPEEDY_MODEL
)

///

func SplitPair(s string) (token, currency string, err error) {
	v := strings.Split(s, "_")
	if len(v) != 2 {
		err = fmt.Errorf("bad number of parts in the '%s'", s)
		return
	}
	return v[0], v[1], nil
}

///////////////////////////////////////////////////////////////////////////////

func generate(model int) {
	nodenames = NodeNames{}
	namenodes = NameNodes{}
	var nextnode loophole.Node = 1
	graph = loophole.Graph{}
	log.Printf("generate graph from %d tickers, model #%d", len(all_tickers), model)
	for _, ticker := range all_tickers {
		//		log.Printf("")
		//		(&ticker).log()
		var ok bool
		var node_from loophole.Node
		var node_to loophole.Node
		var weight_forw loophole.Weight
		var weight_back loophole.Weight

		node_from, ok = namenodes[ticker.TokenName]
		if !ok {
			node_from = nextnode
			nextnode++
			namenodes[ticker.TokenName] = node_from
			nodenames[node_from] = ticker.TokenName
		}
		node_to, ok = namenodes[ticker.CurrencyName]
		if !ok {
			node_to = nextnode
			nextnode++
			namenodes[ticker.CurrencyName] = node_to
			nodenames[node_to] = ticker.CurrencyName
		}
		//pairname := ticker.TokenName + "_" + ticker.CurrencyName
		// check for prices
		if ticker.Sell < MIN_PRICE || ticker.Buy < MIN_PRICE {
			//log.Printf("skip ticker %s by price (sell=%.8f, buy=%.8f)", pairname, ticker.Sell, ticker.Buy)
			continue
		}
		// check distance
		avg_price := (ticker.Sell + ticker.Buy) / 2.0
		dta_price := math.Abs(ticker.Sell - ticker.Buy)
		dsp_price := dta_price * dta_price / avg_price
		if MAX_DISPERSION < dsp_price {
			//log.Printf("skip ticker %s by dispersion (avg=%.6f, delta=%.6f, disp=%.6f", pairname, avg_price, dta_price, dsp_price)
			continue
		}
		if ticker.CurrentVolume < MIN_VOLUME {
			//log.Printf("skip ticker %s by volume %.6f", pairname, ticker.CurrentVolume)
			continue
		}
		switch model {
		case GREEDY_MODEL: // покупаем по цене закупа, продаём по продажной
			weight_forw = loophole.Weight(ticker.Sell * cur_fee_k)
			weight_back = loophole.Weight((1.0 / ticker.Buy) * cur_fee_k)
		case AVERAGE_MODEL: // покупаем и продаём по средней цене
			weight_forw = loophole.Weight(avg_price * cur_fee_k)
			weight_back = loophole.Weight((1.0 / avg_price) * cur_fee_k)
		case SPEEDY_MODEL: // покупаем что продадут и продаём что купят
			weight_forw = loophole.Weight(ticker.Buy * cur_fee_k)
			weight_back = loophole.Weight((1.0 / ticker.Sell) * cur_fee_k)
		}
		//loopcost := weight_forw * weight_back
		//if loopcost > 1.0 {
		//	log.Printf("skip ticker %s due OVERLOOP: %.8f * %.8f = %.8f", pairname, weight_forw, weight_back, loopcost)
		//	continue
		//}
		forw := loophole.Edge{
			From:   node_from,
			To:     node_to,
			Weight: weight_forw,
		}
		back := loophole.Edge{
			From:   node_to,
			To:     node_from,
			Weight: weight_back,
		}
		//log.Printf("add ticker %s to graph", pairname)
		graph = append(graph, forw, back)
	}
}

////

func find_direct(from, to string) (v []Ticker) {
	for _, t := range all_tickers {
		if from != "" && from != t.TokenName {
			continue
		}
		if to != "" && to != t.CurrencyName {
			continue
		}
		v = append(v, t)
	}
	return
}

func find_indirect(from, to string) (T Ticker) {
	for _, t := range all_tickers {
		if from == t.TokenName && to == t.CurrencyName {
			return t
		}
		if from == t.CurrencyName && to == t.TokenName {
			return t
		}
	}
	log.Fatalf("========== ticker names ===========")
	for _, t := range all_tickers {
		log.Printf("    %s_%s", t.CurrencyName, t.TokenName)
	}
	log.Fatalf("=========-= --- end --- ===========")

	log.Fatalf("ticker %s_%s lost", from, to)
	return
}

type MyPath struct {
	path   []string
	weight float64
}

func makeMyPath(path *loophole.Path) (r MyPath) {
	r.path = make([]string, 0, len(*path)+1)
	r.path = append(r.path, nodenames[(*path)[0].From])
	r.weight = 1.0
	for _, e := range *path {
		r.path = append(r.path, nodenames[e.To])
		r.weight *= float64(e.Weight)
	}
	return
}

func decode(mp MyPath, model int) {
	from := mp.path[0]
	amount := 1.00
	fmt.Printf("AT START %.8f %s, MODEL #%v\n", amount, from, model)
	for _, to := range mp.path[1:] {
		T := find_indirect(from, to)
		var action string
		var price float64
		var result float64
		if T.CurrencyName == to { // продаём from, получаем to
			switch model {
			case GREEDY_MODEL:
				price = T.Sell
			case AVERAGE_MODEL:
				price = (T.Sell + T.Buy) / 2.0
			case SPEEDY_MODEL:
				price = T.Buy
			}
			result = amount * price * cur_fee_k
			action = fmt.Sprintf("[%s->%s] sell %s, amount=%.8f[%s] * price=%.8f - %.2f%% = %.8f %s",
				from, to, from, amount, from, price, cur_fee, result, to)
		} else { // покупаем to за from
			switch model {
			case GREEDY_MODEL:
				price = T.Buy
			case AVERAGE_MODEL:
				price = (T.Sell + T.Buy) / 2.0
			case SPEEDY_MODEL:
				price = T.Sell
			}
			result = amount / price * cur_fee_k
			action = fmt.Sprintf("[%s<-%s] buy %s, amount=%.8f[%s] / price=%.8f - %.2f%% = %.8f %s",
				to, from, to, amount, from, price, cur_fee, result, to)
		}
		fmt.Printf("  %s\n", action)
		from = to
		amount = result
	}
	fmt.Printf("AT END %.8f %s\n", amount, from)
}

type BestPaths []MyPath

func (a BestPaths) Len() int           { return len(a) }
func (a BestPaths) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BestPaths) Less(i, j int) bool { return a[i].weight < a[j].weight }

func (p *BestPaths) add(myp MyPath) {
	*p = append(*p, myp)
	sort.Sort(sort.Reverse(*p))
	if BEST_LIMIT < len(*p) {
		*p = (*p)[:BEST_LIMIT]
	}
}

func (p *BestPaths) show() {
	for _, v := range *p {
		fmt.Printf("[%.8f] %v\n", v.weight, v.path)
	}
}

var best BestPaths

func path_processor(path *loophole.Path) bool {
	(&best).add(makeMyPath(path))
	return false
}

func Loop(token string, model int) {
	fmt.Printf(" === LOOP FOR %s ===\n", token)
	node := namenodes[token]
	best = BestPaths(nil)
	(&graph).Walk(node, node, path_processor)
	(&best).show()
	if len(best) > 0 {
		//decode(best[0], GREEDY_MODEL)
		//decode(best[0], AVERAGE_MODEL)
		//decode(best[0], SPEEDY_MODEL)
		decode(best[0], model)
	}
	fmt.Println("")
}

func load_yobiway() {
	pairs, err := session.GetPairs()
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return
	} else {
		log.Printf("%d pairs\n", len(pairs))
	}
	sort.Sort(Alphabetically(pairs))

	all_tickers, err = session.GetTickers(pairs)
	if err != nil {
		log.Printf("ERROR: %s", err)
	} else {
		log.Printf("%d tickers\n", len(all_tickers))
	}
}

func load_bittrex() {
	var err error
	all_tickers, err = session.GetBittrexTickers()
	if err != nil {
		log.Printf("ERROR: %s", err)
	} else {
		log.Printf("%d tickers\n", len(all_tickers))
	}
}

func load_ccex() {
	var err error
	all_tickers, err = session.GetCCexTickers()
	if err != nil {
		log.Printf("ERROR: %s", err)
	} else {
		log.Printf("%d tickers\n", len(all_tickers))
	}
}

func load_livecoin() {
	var err error
	all_tickers, err = session.GetLivecoinTickers()
	if err != nil {
		log.Printf("ERROR: %s", err)
	} else {
		log.Printf("%d tickers\n", len(all_tickers))
	}
}

func play_yobit() {
	load_yobiway()
	cur_fee = YOBI_FEE
	cur_fee_k = YOBI_FEE_K
	var model int
	fmt.Printf("\n### GENERATE AVERAGE MODEL\n\n")
	model = AVERAGE_MODEL
	generate(model)
	log.Printf("%d edges", len(graph))
	Loop("rur", model)
	Loop("usd", model)
	Loop("btc", model)
}

func play_bittrex() {
	load_bittrex()
	cur_fee = BITTREX_FEE
	cur_fee_k = BITTREX_FEE_K
	model := AVERAGE_MODEL
	fmt.Printf("\n### GENERATE MODEL #%d\n\n", model)
	generate(model)
	log.Printf("%d edges", len(graph))
	Loop("BTC", model)
}

func play_ccex() {
	load_ccex()
	cur_fee = BITTREX_FEE
	cur_fee_k = BITTREX_FEE_K
	model := AVERAGE_MODEL
	fmt.Printf("\n### GENERATE MODEL #%d\n\n", model)
	generate(model)
	log.Printf("%d edges", len(graph))
	Loop("BTC", model)
}

func play_livecoin() {
	load_livecoin()
	cur_fee = BITTREX_FEE
	cur_fee_k = BITTREX_FEE_K
	model := AVERAGE_MODEL
	fmt.Printf("\n### GENERATE MODEL #%d\n\n", model)
	generate(model)
	log.Printf("%d edges", len(graph))
	Loop("BTC", model)
}

/// MAIN ///

func main() {
	var err error = initdb()
	if err != nil {
		log.Fatalf("database not initialized: %s", err)
	}
	defer closedb()
	session = NewSession()
	//play_yobit()
	//play_bittrex()
	//play_ccex()
	play_livecoin()
}
