package main

import (
	"net"
	"net/rpc/jsonrpc"
	"net/rpc"
	"log"
	"sort"
	"fmt"
)

// КРИПТОЦЫКЛ! ВЖЖЖЖЖ!!!!!
type CryptoCycle struct {}


type Args struct {
	Exchange string // на какой бирже
	InitialToken string // с какого фантика начнём
	FinalToken string // и на каком фантике закончим
}

type NoArgs struct {}
type ExchangeList []string

// добыть список использующихся у нас бирж
func (cc *CryptoCycle) ListExchanges(args *NoArgs, result *ExchangeList) error {
	var L ExchangeList
	for name := range XCGS {
		L = append(L, name)
	}
	sort.Sort(Alphabetically(L))
	result = &L
	return nil
}

type ExchangeArgs struct {
	Exchange string // какую биржу смотреть
	Currencies bool // true ::= отфильтровать и вернуть валюты; false ::= все токены
}

type TokenList []string

// добыть список использующихся на бирже токенов
func (cc *CryptoCycle) ListTokens(args *ExchangeArgs, result *TokenList) error {
	var T TokenList
	xcg, ok := XCGS[args.Exchange]
	if !ok {
		result = &T
		return fmt.Errorf("unknown exchange: `%s`", args.Exchange)
	}
	if args.Currencies {
		T = TokenList(xcg.GetCurrencies())
	} else {
		T = TokenList(xcg.GetTokens())
	}
	return nil
}


// сервис JSONRPC поверх HTTP
func serve() error {
	cryptocycle := new(CryptoCycle)
	server := rpc.NewServer()
	server.Register(cryptocycle)
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Print("listen error:", err)
		return err
	}
	for {
		if conn, err := listener.Accept(); err != nil {
			log.Print("accept error: " + err.Error())
			return err
		} else {
			log.Print("new connection established\n")
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}
}