package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gourytch/loophole"
	"flag"
	"github.com/gourytch/yobiway/client"
	"github.com/gourytch/yobiway/exchange/livecoin"
	"sort"
	"github.com/gourytch/yobiway/exchange"
)


var YOBI_FEE float64 = 0.2
var YOBI_FEE_K float64 = (1.0 - YOBI_FEE/100.0)
var BITTREX_FEE float64 = 0.25
var BITTREX_FEE_K float64 = (1.0 - BITTREX_FEE/100.0)

var MAX_DISPERSION float64 = 0.2
var MIN_PRICE float64 = 0.00000100 // 0.0000001
var MIN_VOLUME float64 = 100       // 0.00001
var BEST_LIMIT int = 3

type NodeNames map[loophole.Node]string
type NameNodes map[string]loophole.Node

var session *client.Session

var nodenames NodeNames = NodeNames{}
var namenodes NameNodes = NameNodes{}
var graph loophole.Graph = loophole.Graph{}
var cur_fee float64
var cur_fee_k float64

var xcg exchange.Exchange

const (
	GREEDY_MODEL  = iota
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

func generate() {
	nodenames = NodeNames{}
	namenodes = NameNodes{}
	var nextnode loophole.Node = 1
	graph = loophole.Graph{}
	mp := xcg.GetMarketplace()
	log.Printf("generate graph from pricemap")
	for from, Vto := range mp.Pricemap {
		var ok bool
		var node_from loophole.Node
		var node_to loophole.Node

		node_from, ok = namenodes[from]
		if !ok {
			node_from = nextnode
			nextnode++
			namenodes[from] = node_from
			nodenames[node_from] = from
		}
		for to, price := range Vto {
			pair := xcg.GetTradePair(from + "/" + to)
			if pair == nil {
				pair = xcg.GetTradePair(to + "/" + from)
			}
			node_to, ok = namenodes[to]
			if !ok {
				node_to = nextnode
				nextnode++
				namenodes[to] = node_to
				nodenames[node_to] = to
			}
		// check for prices
			if price < MIN_PRICE {
				//log.Printf("skip ticker %s by price (sell=%.8f, buy=%.8f)", pairname, ticker.Sell, ticker.Buy)
				continue
			}
		// check distance
			if pair.Volume < MIN_VOLUME {
				//log.Printf("skip ticker %s by volume %.6f", pairname, ticker.CurrentVolume)
				continue
			}
			graph = append(graph, loophole.Edge{From:node_from, To: node_to, Weight: loophole.Weight(price)})
		}
		//log.Printf("add ticker %s to graph", pairname)
	}
	tokens := []string{}
	for token := range namenodes {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)
	log.Printf("graph tokens: %v", tokens)
}

func find_weight(from, to string) float64 {
	node_from := namenodes[from]
	node_to := namenodes[to]

	for _, n := range graph {
		if n.From == node_from && n.To == node_to {
			return float64(n.Weight)
		}
	}
	return -1
}

func find_indirect(from, to string) (tp *exchange.TradePair) {
	tp = xcg.GetTradePair(from + "/" + to)
	if tp != nil {
		return
	}
	tp = xcg.GetTradePair(to + "/" + from)
	if tp != nil {
		return
	}
	log.Fatalf("ticker %s,%s lost", from, to)
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

func tickers(mp MyPath) {
	fmt.Printf("LIST OF USED TICKERS:\n")
	from := mp.path[0]
	for _, to := range mp.path[1:] {
		T := find_indirect(from, to)
		fmt.Printf("%s-to-%s: token %s, curr %s, volume %f, avg %f\n",
			from, to, T.Token, T.Currency, T.Volume, T.Vwap)
		from = to
	}
}
func decode(mp MyPath) {
	from := mp.path[0]
	amount := 1.00
	fmt.Printf("AT START %.8f %s\n", amount, from)
	for _, to := range mp.path[1:] {
		T := find_indirect(from, to)
		var action string
		var price float64
		var result float64
		price = find_weight(T.Token, T.Currency)
		if T.Currency == to { // продаём from, получаем to
			result = amount * price * cur_fee_k
			action = fmt.Sprintf("[%s->%s] sell %s, amount=%.8f[%s] * price=%.8f - %.2f%% = %.8f %s",
				from, to, from, amount, from, price, cur_fee, result, to)
		} else { // покупаем to за from
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

type PathsMap map[int]*BestPaths

func (m *PathsMap) add(myp MyPath) {
	ix := len(myp.path)
	bp, ok := (*m)[ix]
	if !ok {
		bp = new(BestPaths)
	}
	bp.add(myp)
	(*m)[ix] = bp
}

func (m *PathsMap) show() {
	ixs := []int{}
	for ix := range *m {
		ixs = append(ixs, ix)
	}
	sort.Ints(ixs)
	for _, ix := range ixs {
		fmt.Printf("-= PATH LENGTH %d =-\n", ix)
		r :=(*m)[ix]
		r.show()
		tickers((*r)[0])
		decode((*r)[0])
		fmt.Println("")
	}
}

var best PathsMap

func path_processor(path *loophole.Path) bool {
	mypath := makeMyPath(path)
	// l := len(*path)
	// fmt.Printf("path:%v, Len:%v, best:%#v\n", mypath, l, best)
	(&best).add(mypath)
	return false
}

func Loop(token string) {
	fmt.Printf(" === LOOP FOR %s ===\n", token)
	node := namenodes[token]
	best = make(PathsMap)
	(&graph).Walk(node, node, path_processor)
	(&best).show()
}

/// MAIN ///
var CACHED = true

func main() {
	var xcgName string
	var token string
	flag.StringVar(&xcgName, "exchange", "livecoin", "exchange to analyze")
	flag.StringVar(&token,"token", "BTC", "token to cycle")
	flag.BoolVar(&CACHED, "cached", true, "load from cache")
	flag.Parse()
	xcgName = strings.ToUpper(xcgName)
	token = strings.ToUpper(token)
	log.Printf("exchange=%v, token=%v, cached=%v", xcgName, token, CACHED)

	var err error = client.BoltDB_init()
	if err != nil {
		log.Fatalf("database not initialized: %s", err)
	}
	defer client.BoltDB_close()
	session = client.NewSession()
	switch xcgName {
	//case "YOBIT":
	//	play_yobit(token)
	//case "BITTREX":
	//	play_bittrex(token)
	//case "CCEX":
	//	play_ccex(token)
	case "LIVECOIN":
		livecoin.Register()
		xcg = exchange.Registry["LIVECOIN"]
		cur_fee = BITTREX_FEE
		cur_fee_k = BITTREX_FEE_K
	default:
		log.Fatal("UNKNOWN EXCHANGE:", xcgName)
	}
	xcg.Refresh()
	fmt.Printf("\n### GENERATE ###\n\n")
	generate()
	log.Printf("%d edges", len(graph))
	Loop(token)

}
