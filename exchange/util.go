package exchange

import "fmt"

func (t *TradePair) str() string {
	return fmt.Sprintf("name:%s, vwap:%f, vol:%f", t.Name, t.Vwap, t.Volume)
}

