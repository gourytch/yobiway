package main

import (
	"compress/gzip"
	"encoding/json"
	//"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

const (
	DATABASE_FNAME  = "yobiway.db"
	MAX_TICKERS_REQ = 50
	REQ_INTERVAL    = 1 * time.Second
)

var bucketCACHE = []byte("CACHE")

var db *bolt.DB = nil

type Session struct {
	Client *http.Client
	LastRq time.Time
}

func initdb() error {
	var err error
	db, err = bolt.Open(DATABASE_FNAME, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Printf("database %v open error: %s", DATABASE_FNAME, err)
		return err
	}
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
			//log.Printf("bucket %s exists", bucketCACHE)
		}
		return err
	})
	return err
}

func closedb() {
	db.Close()
}

func NewSession() *Session {
	s := &Session{}
	s.Client = new(http.Client)
	s.LastRq = time.Time{}
	return s
}

func cache_get(url []byte) (body []byte) {
	db.View(func(tx *bolt.Tx) error {
		body = []byte(tx.Bucket(bucketCACHE).Get(url))
		if body == nil {
			//log.Printf("+ %s not in cache", url)
		} else {
			//log.Printf("+ %s get cached %d bytes", url, len(body))
		}
		return nil
	})
	return
}

func cache_put(url, body []byte) {
	//log.Printf("+ %s <- %d bytes", url, len(body))
	db.Update(func(tx *bolt.Tx) error {
		//log.Printf("... + %s <- %d bytes", url, len(body))
		err := tx.Bucket(bucketCACHE).Put(url, body)
		if err != nil {
			log.Printf("! not cached %s due error: %s", url, err)
		} else {
			//log.Printf("+ %s cached", url)
		}
		return err
	})
}

func (s *Session) Get(url string) (body []byte, err error) {
	burl := []byte(url)

	if body = cache_get(burl); body != nil {
		//log.Printf("... use cached: %s (%d bytes)", url, len(body))
		return
	}
	//log.Printf("... cache miss. request: %s", url)

	dt := time.Now().Sub(s.LastRq)
	if dt < REQ_INTERVAL {
		//log.Printf("... wait for a little ...")
		time.Sleep(REQ_INTERVAL - dt)
	}
	s.LastRq = time.Now()
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
	//fmt.Printf("... got: %d bytes\n", len(body))
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
	sorted_pairs := append([]string(nil), pairs...)
	sort.Sort(Alphabetically(sorted_pairs))

	L := len(sorted_pairs)
	offs := 0
	for offs < L {
		r := offs + MAX_TICKERS_REQ
		if L < r {
			r = L
		}
		//log.Printf("process slice [%d:%d]", offs, r)
		P := sorted_pairs[offs:r]
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
