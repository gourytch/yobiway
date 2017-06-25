package main

import (
	"log"
	//	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

const (
	DATABASE_FNAME  = "yobiway.db"
	MAX_TICKERS_REQ = 10
)

var bucketCACHE = []byte("CACHE")

var db *bolt.DB = nil

type Session struct {
	Client *http.Client
}

type Ticker struct {
	TokenName     string  `json:-` // имя фантика
	CurrencyName  string  `json:-` // имя валюты (btc/usd/rur/etc.)
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Average       float64 `json:"avg"`
	Volume        float64 `json:"vol"`
	CurrentVolume float64 `json:"vol_cur"`
	Last          float64 `json:"last"`
	Buy           float64 `json:"buy"`
	Sell          float64 `json:"sell"`
	Updated       float64 `json:"updated"`
	ServerTS      int64   `json:"server_time"`
}

type JTicker struct {
	T Ticker `json:"ticker"`
}

type JTickers map[string]Ticker

type PairDesc struct {
	TokenName     string  `json:-` // имя фантика
	CurrencyName  string  `json:-` // имя валюты (btc/usd/rur/etc.)
	DecimalPlaces int     `json:"decimal_places"`
	MinPrice      float64 `json:"min_price"`
	MaxPrice      float64 `json:"max_price"`
	MinAmount     float64 `json:"min_amount"`
	MinTotal      float64 `json:"min_total"`
	Hidden        int     `json:"hidden"`
	Fee           float64 `json:"fee"`
	FeeByer       float64 `json:"fee_buyer"`
	FeeSeller     float64 `json:"fee_seller"`
}

type JPairs struct {
	Pairs map[string]PairDesc `json:"pairs"`
}

type Edge struct { // обмен: [To] = [From] * K * (1.0 - Fee / 100)
	From string
	To   string
	K    float64 // Коэффициент обмена: N[To] = N[From] * K
	Fee  float64 // комиссия за обмен в процентах
}

type Edges []Edge         // цепочка обменов
type Ways map[string]Edge // список возможных путей

func (e Edge) str() string {
	return fmt.Sprintf("%s -> %s, k=%f, fee=%f", e.From, e.To, e.K, e.Fee)
}

func (w *Ways) DestinationNames(src string) (r []string) {
	for _, v := range *w {
		if v.From == src {
			r = append(r, v.To)
		}
	}
	sort.Sort(Alphabetically(r))
	return
}

func (w *Ways) SourceNames(dst string) (r []string) {
	for _, v := range *w {
		if v.To == dst {
			r = append(r, v.From)
		}
	}
	sort.Sort(Alphabetically(r))
	return
}

func (w *Ways) Destinations(src string) *Ways {
	d := &Ways{}
	for k, v := range *w {
		if v.From == src {
			(*d)[k] = v
		}
	}
	return d
}

// получаем множество путей в точки не указанные в параметре
func (w *Ways) ExcludeDestinations(ex []string) *Ways {
	d := &Ways{}
	for k, v := range *w {
		found := false
		for _, s := range ex {
			if v.To == s {
				found = true
				break
			}
		}
		if !found {
			(*d)[k] = v
		}
	}
	return d
}

func (w *Ways) _walk(
	from string, // откуда сейчас идём
	to string, // куда в итоге хотим попасть
	steps []string, // что уже прошли
	chain Edges, // и список всех участков
	callback func(chain Edges) bool) bool {
	//log.Printf("_walk(%v, %v, %v ...) {", from, to, steps)

	// вычисляем все пути ведущие к непосещённым узлам
	ww := w.Destinations(from).ExcludeDestinations(steps)
	if len(*ww) == 0 { // нет таких
		//log.Printf("} dead end")
		return false
	}
	for _, v := range *ww {
		// мастерим новые тропинки
		ss := append(steps, v.To)
		cc := append(chain, v)
		if v.To == to { // уткнулись в конец
			if callback(cc) {
				//log.Printf("} eject by callback")
				return true
			}
		} else { // роем дальше, исчерпывающим перебором
			if w._walk(v.To, to, ss, cc, callback) {
				//log.Printf("} ... eject by recursion")
				return true
			}
		}
	}
	//log.Printf("} _walk(%v, %v, %v ...) end of recursion reached", from, to, steps)
	return false
}

func (w *Ways) Walk(from string, to string, callback func(chain Edges) bool) {
	steps := []string(nil)
	if from != to {
		steps = append(steps, from)
	}
	w._walk(from, to, steps, Edges(nil), callback)
}

func (v Edges) Steps() (s []string) {
	if len(v) == 0 {
		return
	}
	s = append([]string(nil), v[0].From)
	for _, q := range v {
		s = append(s, q.To)
	}
	return
}

func (v Edges) Integral() (e Edge) {
	if len(v) == 0 {
		return
	}
	e.From = v[0].From
	e.To = v[len(v)-1].To
	e.K = 1.0
	fk := 1.0 // коэффициент возвращаемого после снятия мзды за транзакцию
	for _, q := range v {
		f := (1.0 - (q.Fee / 100.0)) // 0 < f <= 1 при 0 <= fee < 100.0
		e.K *= q.K * f
		fk *= f // уменьшается
	}
	e.Fee = (1.0 - fk) * 100.0 // переделаем обратно в проценты
	return
}

type ByK []Edges

func (a ByK) Len() int      { return len(a) }
func (a ByK) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByK) Less(i, j int) bool {
	return a[i].Integral().K > a[j].Integral().K
}

func (w *Ways) BestWays(from string, to string, limit int) (best []Edges) {
	w.Walk(from, to, func(chain Edges) bool {
		log.Printf("add way: I=%s, S=%v", chain.Integral().str(), chain.Steps())
		best = append(best, chain)
		sort.Sort(ByK(best))
		if limit < len(best) {
			best = best[:limit]
		}
		return false
	})
	return
}

func (w *Ways) Circles(from string, limit int) (best []Edges) {
	return w.BestWays(from, from, limit)
}

///////////////////////////////////////////////////////////////////////

func (t *Ticker) log() {
	log.Printf("pair     : %s_%s", t.TokenName, t.CurrencyName)
	log.Printf("Lo/Hi,Avg: %f .. %f, %f", t.Low, t.High, t.Average)
	log.Printf("Vol/Cur  : %f / %f", t.Volume, t.CurrentVolume)
	log.Printf("Last     : %f", t.Last)
	log.Printf("Buy/Sell : %f / %f", t.Buy, t.Sell)
}

///////////////////////////////////////////////////////////////////////////////

func SplitPair(s string) (token, currency string, err error) {
	v := strings.Split(s, "_")
	if len(v) != 2 {
		err = fmt.Errorf("bad number of parts in the '%s'", s)
		return
	}
	return v[0], v[1], nil
}

type ByCurrency []string

func (a ByCurrency) Len() int      { return len(a) }
func (a ByCurrency) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByCurrency) Less(i, j int) bool {
	it, ic, _ := SplitPair(a[i])
	jt, jc, _ := SplitPair(a[j])
	return (ic < jc) || ((ic == jc) && (it < jt))
}

type ByToken []string

func (a ByToken) Len() int      { return len(a) }
func (a ByToken) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByToken) Less(i, j int) bool {
	it, ic, _ := SplitPair(a[i])
	jt, jc, _ := SplitPair(a[j])
	return (it < jt) || ((it == jt) && (ic < jc))
}

type Alphabetically []string

func (a Alphabetically) Len() int           { return len(a) }
func (a Alphabetically) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Alphabetically) Less(i, j int) bool { return a[i] < a[j] }

///////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////

func NewSession() *Session {
	s := &Session{}
	s.Client = new(http.Client)
	return s
}

func cache_get(url []byte) (body []byte) {
	db.View(func(tx *bolt.Tx) error {
		body = []byte(tx.Bucket(bucketCACHE).Get(url))
		if body == nil {
			log.Printf("+ %s not in cache", url)
		} else {
			log.Printf("+ %s get cached %d bytes", url, len(body))
		}
		return nil
	})
	return
}

func cache_put(url, body []byte) {
	log.Printf("+ %s <- %d bytes", url, len(body))
	db.Update(func(tx *bolt.Tx) error {
		log.Printf("... + %s <- %d bytes", url, len(body))
		err := tx.Bucket(bucketCACHE).Put(url, body)
		if err != nil {
			log.Printf("! not cached %s due error: %s", url, err)
		} else {
			log.Printf("+ %s cached", url)
		}
		return err
	})
}

func (s *Session) Get(url string) (body []byte, err error) {
	burl := []byte(url)

	if body = cache_get(burl); body != nil {
		log.Printf("... use cached: %s (%d bytes)", url, len(body))
		return
	}
	log.Printf("... cache miss. request: %s", url)
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("Accept-Encoding", "gzip")
	response, err := s.Client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	// Check that the server actually sent compressed data
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			return
		}
		defer reader.Close()
	default:
		reader = response.Body
	}
	body, err = ioutil.ReadAll(reader)
	fmt.Printf("... got: %d bytes\n", len(body))
	if err != nil {
		return
	}
	cache_put(burl, body)
	return
}

// pair :: 'ltc_btc'
func (s *Session) GetTicker(pair string) (t Ticker, err error) {
	token, currency, err := SplitPair(pair)
	if err != nil {
		return
	}
	data, err := s.Get("https://yobit.net/api/3/" + pair + "/ticker")
	if err != nil {
		return
	}
	var j JTicker
	err = json.Unmarshal(data, &j)
	if err != nil {
		return
	}
	t = j.T
	t.TokenName = token
	t.CurrencyName = currency
	return
}

// pair :: 'ltc_btc'
func (s *Session) GetTickers(pairs []string) (v []Ticker, err error) {
	L := len(pairs)
	offs := 0
	for offs < L {
		r := offs + MAX_TICKERS_REQ
		if L < r {
			r = L
		}
		log.Printf("process slice [%d:%d]", offs, r)
		P := pairs[offs:r]
		Ps := strings.Join(P, "-")
		var data []byte
		data, err = s.Get("https://yobit.net/api/3/ticker/" + Ps)
		if err != nil {
			return
		}
		var j JTickers
		err = json.Unmarshal(data, &j)
		if err != nil {
			return
		}
		for jk, jv := range j {
			var token, currency string
			token, currency, err = SplitPair(jk)
			if err != nil {
				return
			}
			jv.TokenName = token
			jv.CurrencyName = currency
			v = append(v, jv)
		}
		offs = r
	}
	return
}

func (s *Session) GetPairs() (pairs []string, err error) {
	data, err := s.Get("https://yobit.net/api/3/info")
	if err != nil {
		return
	}
	var j JPairs
	err = json.Unmarshal(data, &j)
	if err != nil {
		return
	}
	pairs = nil
	for k, v := range j.Pairs {
		if v.Hidden == 0 {
			pairs = append(pairs, k)
		}
	}
	return
}

func MakeWays(tickers []Ticker) *Ways {
	ways := &Ways{}
	for _, ticker := range tickers {
		log.Printf("")
		(&ticker).log()
		avg := (ticker.Sell + ticker.Buy) / 2.0
		(*ways)[ticker.TokenName+"_"+ticker.CurrencyName] = Edge{
			From: ticker.TokenName,
			To:   ticker.CurrencyName,
			K:    avg,
			Fee:  0.2, // FIXME надо брать из списка описания токенов
		}
		(*ways)[ticker.CurrencyName+"_"+ticker.TokenName] = Edge{
			From: ticker.CurrencyName,
			To:   ticker.TokenName,
			K:    1.0 / avg,
			Fee:  0.2, // FIXME надо брать из списка описания токенов
		}
	}
	return ways
}

/// MAIN ///

func main() {
	var err error
	db, err = bolt.Open(DATABASE_FNAME, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Printf("database %v open error: %s", DATABASE_FNAME, err)
		return
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCACHE)
		if b == nil {
			log.Printf("create bucket %s", bucketCACHE)
			_, err := tx.CreateBucket(bucketCACHE)
			if err != nil {
				log.Fatalf("bucket creation error: %s", err)
				return err
			}
		} else {
			log.Printf("bucket %s exists", bucketCACHE)
		}
		return err
	})
	s := NewSession()
	v, err := s.GetPairs()
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return
	}
	log.Printf("GOT %d PAIRS\n", len(v))
	sort.Sort(ByCurrency(v))

	t, err := s.GetTickers(v)
	if err != nil {
		log.Printf("ERROR: %s", err)
	} else {
		log.Printf("%d tickers\n", len(t))
	}
	ways := MakeWays(t)
	/*
		log.Printf("BTC->%v", ways.DestinationNames("btc"))
		log.Printf("RUR->%v", ways.DestinationNames("rur"))
		log.Printf("USD->%v", ways.DestinationNames("usd"))
		log.Printf("BTC->%v", ways.Destinations("btc"))
		log.Printf("RUR->%v", ways.Destinations("rur"))
		log.Printf("USD->%v", ways.Destinations("usd"))

		dst_rur := ways.Destinations("rur")
		log.Printf("RUR->%v", dst_rur)
		dst_usd := ways.Destinations("usd")
		log.Printf("USD->%v", dst_usd)
		rur_no_usd := dst_rur.ExcludeDestinations(ways.DestinationNames("usd"))
		usd_no_rur := dst_usd.ExcludeDestinations(ways.DestinationNames("rur"))
		log.Printf("RUR/USD=%v", rur_no_usd)
		log.Printf("USD/RUR=%v", usd_no_rur)

		log.Printf("USD->BTC K = %f", (*ways)["usd_btc"].K)
		log.Printf("BTC->USD K = %f", (*ways)["btc_usd"].K)
		log.Printf("RUR->USD K = %f", (*ways)["rur_usd"].K)
		log.Printf("USD->RUR K = %f", (*ways)["usd_rur"].K)
	*/

	/*
		ways.Walk("rur", "usd", func(chain Edges) bool {
			s := []string(nil)
			k := 1.0
			if len(chain) != 0 {
				s = append(s, chain[0].From)
				for _, v := range chain {
					s = append(s, v.To)
					k *= v.K * (1.0 - v.Fee/100.0)
				}
			}
			log.Printf("K=%f CHAIN=%v", k, s)
			return false
		})
		ways.Walk("usd", "rur", func(chain Edges) bool {
			s := []string(nil)
			k := 1.0
			if len(chain) != 0 {
				s = append(s, chain[0].From)
				for _, v := range chain {
					s = append(s, v.To)
					k *= v.K * (1.0 - v.Fee/100.0)
				}
			}
			log.Printf("K=%f CHAIN=%v", k, s)
			return false
		})
	*/
	C10 := ways.Circles("rur", 10)
	for _, chain := range C10 {
		log.Printf("I:%s steps:%v", chain.Integral().str(), chain.Steps())
	}

}
