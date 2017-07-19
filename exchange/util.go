package exchange

import (
	"fmt"
)

func (t *TradePair) Str() string {
	return fmt.Sprintf("name:%s, vwap:%f, vol:%f", t.Name, t.Vwap, t.Volume)
}

func (t *TradePair) Info() string {
	return "{\n" +
		fmt.Sprintf("   Name        : %s\n", t.Name) +
		fmt.Sprintf("   URL         : %s\n", t.URL) +
		fmt.Sprintf("   Token       : %s\n", t.Token) +
		fmt.Sprintf("   Currency    : %s\n", t.Currency) +
		fmt.Sprintf("   Vwap        : %.8f\n", t.Vwap) +
		fmt.Sprintf("   Volume      : %f\n", t.Volume) +
		fmt.Sprintf("   Volume24H   : %f\n", t.Volume24H) +
		fmt.Sprintf("   Max_Bid     : %.8f\n", t.Max_Bid) +
		fmt.Sprintf("   Min_Ask     : %.8f\n", t.Min_Ask) +
		fmt.Sprintf("   Volume_Bids : %f\n", t.Volume_Bids) +
		fmt.Sprintf("   Volume_Asks : %f\n", t.Volume_Asks) +
		fmt.Sprintf("   Price_Bids  : %f\n", t.Price_Bids) +
		fmt.Sprintf("   Price_Asks  : %f\n", t.Price_Asks) +
		fmt.Sprintf("   Num_Trades  : %d\n", t.Num_Trades) +
		fmt.Sprintf("   BuyFee      : %.5f\n", t.BuyFee) +
		fmt.Sprintf("   SellFee     : %.5f\n", t.SellFee) +
		fmt.Sprintf("   Min_Amount  : %.8f\n", t.Min_Amount) +
		"}"
}
