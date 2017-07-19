package exchange

type TokenInfo struct {
	Token string
	Name  string
}

type Order struct {
	Price  float64 // по какой цене за токен поставили
	Amount float64 // сколько токенов поставили
}

type Orders []Order // список ордеров

type Orderbook struct {
	Asks Orders // выставленные на продажу
	Bids Orders // выставленные на покупку
}

type TradePair struct {
	Name        string    // пара в формате "ТОКЕН/ВАЛЮТА" капсом (req)
	URL         string    // адрес торговой пары на бирже (если есть)
	Token       string    // символ токена (req)
	Currency    string    // за какую валюту торгуется? (req)
	Vwap        float64   // Volume Weighted Average Price (Средневзвешенная цена)
	Volume      float64   // текущий объём торгов в токенах
	Volume24H   float64   // объём торгов в токенах за последние сутки
	Max_Bid     float64   // текущая максимальная цена приказа покупки
	Min_Ask     float64   // текущая минимальная цена приказа продажи
	Avg_Price   float64   // текущая средняя цена (Max_Bid + Min_Ask) / 2
	Volume_Bids float64   // суммарное количество токенов в обозреваемых приказах покупки в стакане
	Volume_Asks float64   // суммарное количество токенов в обозреваемых приказах продажи в стакане
	Price_Bids  float64   // суммарная стоимость обозреваемых приказов покупки в стакане
	Price_Asks  float64   // суммарная стоимость обозреваемых приказов продажи в стакане
	Num_Trades  int64     // число совершенных сделок за последний час
	BuyFee      float64   // комиссиионный процент на покупку
	SellFee     float64   // комиссиионный процент на продажу
	Min_Amount  float64   // минимальное количество токенов в приказе
	Orderbook   Orderbook // стакан торгов
}

type Marketplace struct {
	Pairs      map[string]*TradePair         // каталог всех торговых пар
	Currencies map[string]bool               // перечень всех токенов-валют на рынке
	Pricemap   map[string]map[string]float64 // словарь токен->имена торговых пар в которых его можно купить-продать
}

type Exchange interface {
	GetName() string                     // получить имя биржи
	Refresh() error                      // обновить данные по бирже
	GetAllTokens() []string              // получить список всех активных токенов, пользующихся на бирже
	GetAllCurrencies() []string          // получить список всех активных валют, пользующихся на бирже
	GetMarketplace() *Marketplace        //  получить описание рынка
	GetTradePair(name string) *TradePair // получить отдельную пару
}
