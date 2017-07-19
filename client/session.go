package client

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	REQ_INTERVAL = 1 * time.Second
)

var CACHED_MODE = false

type Session struct {
	Client *http.Client
	LastRq time.Time
}

func NewSession() *Session {
	s := &Session{}
	s.Client = new(http.Client)
	s.LastRq = time.Time{}
	return s
}

func (s *Session) Get(url string, cached bool) (body []byte, err error) {
	burl := []byte(url)

	if cached {
		if body = cache_get(burl); body != nil {
			//log.Printf("... use cached: %s (%d bytes)", url, len(body))
			return
		}
		//log.Printf("... cache miss. request: %s", url)
	}

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
