package main

import (
	"fmt"
	"strings"

	"github.com/gourytch/loophole"
	"flag"
	"github.com/gourytch/yobiway/client"
	"github.com/gourytch/yobiway/exchange/livecoin"
	"sort"
	"github.com/gourytch/yobiway/exchange"
	"os"
)

var YOBI_FEE float64 = 0.2
var YOBI_FEE_K float64 = (1.0 - YOBI_FEE/100.0)
var BITTREX_FEE float64 = 0.25
var BITTREX_FEE_K float64 = (1.0 - BITTREX_FEE/100.0)

var MIN_PRICE float64 = 0.00000100 // 0.0000001
var MIN_VOLUME float64 = 0.01      // 0.00001
var MIN_VOLUME24H float64 = 1.00   // 0.00001
var MIN_VOLUME_ASK float64 = 0.01  // 0.00001
var MIN_VOLUME_BID float64 = 0.01  // 0.00001
var MAX_WINNERS int = 3

type NodeNames map[loophole.Node]string
type NameNodes map[string]loophole.Node

var ban_tokens = map[string]bool{} // ignored tokens
var ban_pairs = map[string]bool{}  // ignored tradepairs

var nodenames = NodeNames{}
var namenodes = NameNodes{}
var graph = loophole.Graph{}
var cur_fee float64
var cur_fee_k float64

var xcg exchange.Exchange

///////////////////////////////////////////////////////////////////////////////
func check_pair(pair *exchange.TradePair) bool {
	if ban_tokens[pair.Token] {
		fmt.Printf("; hide trade pair %s by token %s\n", pair.Name, pair.Token)
		return false
	}
	if ban_tokens[pair.Currency] {
		fmt.Printf("; hide trade pair %s by currency token %s\n", pair.Name, pair.Currency)
		return false
	}
	// check for prices
	if pair.Vwap < MIN_PRICE {
		fmt.Printf("; hide trade pair %s by price %.8f\n", pair.Name, pair.Vwap)
		return false
	}
	if pair.Volume < MIN_VOLUME {
		fmt.Printf("; hide trade pair %s by total volume %f\n", pair.Name, pair.Volume)
		return false
	}
	if pair.Volume24H < MIN_VOLUME24H {
		fmt.Printf("; hide trade pair %s by daily volume %f\n", pair.Name, pair.Volume24H)
		return false
	}
	if pair.Volume_Asks < MIN_VOLUME_ASK {
		fmt.Printf("; hide trade pair %s by asks volume %f\n", pair.Name, pair.Volume_Asks)
		return false
	}
	if pair.Volume_Bids < MIN_VOLUME_BID {
		fmt.Printf("; hide trade pair %s by bids volume %f\n", pair.Name, pair.Volume_Bids)
		return false
	}
	return true
}

func check_pairs(tpnames []string) {
	fmt.Printf("; filter for %d trade pairs\n", len(tpnames))
	tps := xcg.GetMarketplace().Pairs
	for _, tpname := range tpnames {
		if !check_pair(tps[tpname]) {
			ban_pairs[tpname] = true
		}
	}
	fmt.Printf("; %d trade pairs banned\n", len(ban_pairs))
}

func generate() {
	fmt.Println("; prepare node maps")
	graph = loophole.Graph{}
	nodenames = NodeNames{}
	namenodes = NameNodes{}
	all_tokens := xcg.GetAllTokens()
	sort.Strings(all_tokens)
	for ix, token := range all_tokens {
		id := loophole.Node(ix)
		namenodes[token] = id
		nodenames[id] = token
		fmt.Printf("; %03v %s\n", id, token)
	}
	mp := xcg.GetMarketplace()
	tps := mp.Pairs
	tpnames_unfiltered := make([]string, 0, len(tps))
	for tpname, _ := range tps {
		tpnames_unfiltered = append(tpnames_unfiltered, tpname)
	}
	sort.Strings(tpnames_unfiltered)
	check_pairs(tpnames_unfiltered)
	tpnames := make([]string, 0, len(tpnames_unfiltered) - len(ban_pairs))
	for _, tpname := range tpnames_unfiltered {
		if ban_pairs[tpname] {
			continue
		}
		tpnames = append(tpnames, tpname)
	}
	fmt.Printf("; generate graph from %d filtered trade pairs\n", len(tpnames))
	for _, tpname := range tpnames {
		tp := tps[tpname]
		node_token := namenodes[tp.Token]
		node_currency := namenodes[tp.Currency]
		price_forward := mp.Pricemap[tp.Token][tp.Currency]
		price_backward := mp.Pricemap[tp.Currency][tp.Token]
		fmt.Printf("; add edges %d:%s<->%d:%s [%f/%f]\n",
			node_token, tp.Token, node_currency, tp.Currency,
			price_forward, price_backward)
		graph = append(graph, loophole.Edge{
			From:   node_token,
			To:     node_currency,
			Weight: loophole.Weight(price_forward),
		}, loophole.Edge{
			From:   node_currency,
			To:     node_token,
			Weight: loophole.Weight(price_backward),
		})
	}
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
	fmt.Printf("! ticker %s,%s lost\n", from, to)
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
	fmt.Println("= LIST OF USED TRADE PAIRS")
	from := mp.path[0]
	for _, to := range mp.path[1:] {
		T := find_indirect(from, to)
		fmt.Printf(" from %s to %s:\n", from, to)
		fmt.Printf("%s\n", T.Info())
		from = to
	}
}

func decode(mp MyPath) {
	from := mp.path[0]
	amount := 1.00
	fmt.Printf("= TRADE STRATEGY, AT START %.8f %s\n", amount, from)
	for _, to := range mp.path[1:] {
		T := find_indirect(from, to)
		var action string
		var price float64
		var result float64
		price = find_weight(T.Token, T.Currency)
		if T.Currency == to { // продаём from, получаем to
			result = amount * price * cur_fee_k
			action = fmt.Sprintf(" [%s->%s] sell %s, amount=%.8f[%s] * price=%.8f - %.2f%% = %.8f %s",
				from, to, from, amount, from, price, cur_fee, result, to)
		} else { // покупаем to за from
			result = amount / price * cur_fee_k
			action = fmt.Sprintf(" [%s<-%s] buy %s, amount=%.8f[%s] / price=%.8f - %.2f%% = %.8f %s",
				to, from, to, amount, from, price, cur_fee, result, to)
		}
		fmt.Printf("  %s\n", action)
		from = to
		amount = result
	}
	fmt.Printf("= AT END %.8f %s\n", amount, from)
}

type BestPaths []MyPath

func (a BestPaths) Len() int           { return len(a) }
func (a BestPaths) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BestPaths) Less(i, j int) bool { return a[i].weight < a[j].weight }

func (p *BestPaths) add(myp MyPath) {
	*p = append(*p, myp)
	sort.Sort(sort.Reverse(*p))
	if MAX_WINNERS < len(*p) {
		*p = (*p)[:MAX_WINNERS]
	}
}

func (p *BestPaths) show() {
	for _, v := range *p {
		fmt.Printf(" [%.4f] %v\n", v.weight, v.path)
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
		fmt.Printf("= PATHS WITH LENGTH %d\n", ix)
		r := (*m)[ix]
		r.show()
		tickers((*r)[0])
		decode((*r)[0])
		fmt.Println("")
	}
}

var best PathsMap
var botb BestPaths

func path_processor(path *loophole.Path) bool {
	mypath := makeMyPath(path)
	// l := len(*path)
	// fmt.Printf("path:%v, Len:%v, best:%#v\n", mypath, l, best)
	(&best).add(mypath)
	(&botb).add(mypath)
	return false
}

func Loop(token string) {
	fmt.Printf("= LOOP FOR %s\n", token)
	node := namenodes[token]
	best = make(PathsMap)
	(&graph).Walk(node, node, path_processor)
	(&best).show()
	fmt.Printf("= MOST PROFITABLE PATHS\n")
	(&botb).show()
}

func Way(from, to string) {
	fmt.Printf(" === WAY FROM %s TO %s ===\n", from, to)
	node_from := namenodes[from]
	node_to := namenodes[to]
	best = make(PathsMap)
	(&graph).Walk(node_from, node_to, path_processor)
	(&best).show()
	(&botb).show()
}

func main() {
	var xcgName string
	var from string
	var to string
	var exclude string
	flag.StringVar(&xcgName, "exchange", "livecoin", "exchange to analyze")
	flag.StringVar(&from, "from", "BTC", "token to start")
	flag.StringVar(&to, "to", "BTC", "token to finish")
	flag.StringVar(&exclude, "exclude", "", "tokens to exclude")
	flag.BoolVar(&client.CACHED_MODE, "cached", true, "use cached requests")
	flag.IntVar(&MAX_WINNERS, "max_winners", MAX_WINNERS, "limit number of winners in each category")
	flag.Float64Var(&MIN_PRICE, "min_price", MIN_PRICE, "filter by minimal price")
	flag.Float64Var(&MIN_VOLUME, "min_volume", MIN_VOLUME, "filter by minimal current volume")
	flag.Float64Var(&MIN_VOLUME24H, "min_volume24h", MIN_VOLUME24H, "filter by minimal daily volume")
	flag.Float64Var(&MIN_VOLUME_BID, "min_volume_bid", MIN_VOLUME_BID, "filter by minimal bid volume")
	flag.Float64Var(&MIN_VOLUME_ASK, "min_volume_ask", MIN_VOLUME_ASK, "filter by minimal ask volume")
	flag.Parse()
	xcgName = strings.ToUpper(xcgName)
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)
	exclude = strings.ToUpper(exclude)
	fmt.Printf("= INITIAL PARAMETERS\n")
	fmt.Printf(" exchange       = %v\n", xcgName)
	fmt.Printf(" search way for = %v->%v\n", from, to)
	fmt.Printf(" exclude tokens = %v\n", exclude)
	fmt.Printf(" use cached req = %v\n", client.CACHED_MODE)
	fmt.Printf(" MAX_WINNERS    = %d\n", MAX_WINNERS)
	fmt.Printf(" MIN_PRICE      = %.8f\n", MIN_PRICE)
	fmt.Printf(" MIN_VOLUME     = %f\n", MIN_VOLUME)
	fmt.Printf(" MIN_VOLUME24H  = %f\n", MIN_VOLUME24H)
	fmt.Printf(" MIN_VOLUME_BID = %f\n", MIN_VOLUME_BID)
	fmt.Printf(" MIN_VOLUME_ASK = %f\n", MIN_VOLUME_ASK)

	for _, s := range strings.Split(exclude, ",") {
		ban_tokens[s] = true
	}

	var err error = client.BoltDB_init()
	if err != nil {
		fmt.Printf("! database not initialized: %s", err)
		os.Exit(1)
	}
	defer client.BoltDB_close()
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
		fmt.Printf("! UNKNOWN EXCHANGE: %v\n", xcgName)
		os.Exit(1)
	}
	if err = xcg.Refresh(); err != nil {
		fmt.Printf("! DATA LOAD ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("; generate graph")
	generate()
	fmt.Printf("; graph generated; size=%d\n", len(graph))
	if from == to {
		fmt.Printf("; search for the best cycles with %s\n", from)
		Loop(from)
	} else {
		fmt.Printf("; search for the best ways %s->%s\n", from, to)
		Way(from, to)
	}
	fmt.Println("; done.")
}
