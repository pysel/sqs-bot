package osmocexfillertypes

import (
	"sort"
	"strconv"
	"sync"

	"github.com/osmosis-labs/osmosis/osmomath"
)

type OrderbookData struct {
	mu sync.Mutex

	Symbol string
	bids   map[string]string // Price: Size
	asks   map[string]string // Price: Size
}

func NewOrderbookData(symbol string, bids, asks map[string]string) *OrderbookData {
	return &OrderbookData{
		mu: sync.Mutex{},

		Symbol: symbol,
		bids:   bids,
		asks:   asks,
	}
}

func (o *OrderbookData) Bids() []OrderBasicI {
	o.mu.Lock()
	defer o.mu.Unlock()

	bids := make([]OrderBasicI, 0, len(o.bids))
	for price, size := range o.bids {
		bids = append(bids, &OrderbookEntry{
			Direction: "bid",
			Price:     price,
			Size:      size,
		})
	}
	return bids
}

func (o *OrderbookData) Asks() []OrderBasicI {
	o.mu.Lock()
	defer o.mu.Unlock()

	asks := make([]OrderBasicI, 0, len(o.asks))
	for price, size := range o.asks {
		asks = append(asks, &OrderbookEntry{
			Direction: "ask",
			Price:     price,
			Size:      size,
		})
	}
	return asks
}

// func (o *OrderbookData) GetBid(price string) (string, bool) {
// 	o.mu.Lock()
// 	defer o.mu.Unlock()

// 	size, ok := o.bids[price]
// 	return size, ok
// }

// func (o *OrderbookData) GetAsk(price string) (string, bool) {
// 	o.mu.Lock()
// 	defer o.mu.Unlock()

// 	size, ok := o.asks[price]
// 	return size, ok
// }

func (o *OrderbookData) SetBid(price, size string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.bids[price] = size
}

func (o *OrderbookData) SetAsk(price, size string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.asks[price] = size
}

func (o *OrderbookData) RemoveBid(price string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	delete(o.bids, price)
}

func (o *OrderbookData) RemoveAsk(price string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	delete(o.asks, price)
}

// ScaleSize scales the sizes in the orderbook to the same precision as interchain denoms and returns a deep copy of the orderbook
func (o *OrderbookData) ScaleSize(addedPrecision int) *OrderbookData {
	o.mu.Lock()
	defer o.mu.Unlock()

	copiedOrderbook := &OrderbookData{
		bids: make(map[string]string),
		asks: make(map[string]string),
	}

	for price, size := range o.bids {
		unscaled := osmomath.MustNewBigDecFromStr(size)
		base := osmomath.NewBigDec(10)
		multiplier := base.Power(osmomath.NewBigDec(int64(addedPrecision)))
		scaled := unscaled.Mul(multiplier)

		copiedOrderbook.bids[price] = scaled.String()
	}

	for price, size := range o.asks {
		unscaled := osmomath.MustNewBigDecFromStr(size)
		base := osmomath.NewBigDec(10)
		multiplier := base.Power(osmomath.NewBigDec(int64(addedPrecision)))
		scaled := unscaled.Mul(multiplier)

		copiedOrderbook.asks[price] = scaled.String()
	}

	return copiedOrderbook
}

func (o *OrderbookData) BidsDescending() []OrderBasicI {
	o.mu.Lock()
	defer o.mu.Unlock()

	bids := make([]OrderBasicI, 0, len(o.bids))
	for price, size := range o.bids {
		bids = append(bids, &OrderbookEntry{
			Direction: "bid",
			Price:     price,
			Size:      size,
		})
	}

	sort.Slice(bids, func(i, j int) bool {
		priceI, _ := strconv.ParseFloat(bids[i].GetPrice(), 64)
		priceJ, _ := strconv.ParseFloat(bids[j].GetPrice(), 64)
		return priceI > priceJ
	})
	return bids
}

func (o *OrderbookData) AsksAscending() []OrderBasicI {
	o.mu.Lock()
	defer o.mu.Unlock()

	asks := make([]OrderBasicI, 0, len(o.asks))
	for price, size := range o.asks {
		asks = append(asks, &OrderbookEntry{
			Direction: "ask",
			Price:     price,
			Size:      size,
		})
	}

	sort.Slice(asks, func(i, j int) bool {
		priceI, _ := strconv.ParseFloat(asks[i].GetPrice(), 64)
		priceJ, _ := strconv.ParseFloat(asks[j].GetPrice(), 64)
		return priceI < priceJ
	})
	return asks
}

type OrderbookEntry struct {
	Direction string
	Price     string
	Size      string
}

func (o OrderbookEntry) GetPrice() string {
	return o.Price
}

func (o OrderbookEntry) GetSize() string {
	return o.Size
}

func (o OrderbookEntry) GetDirection() string {
	return o.Direction
}

func (o *OrderbookEntry) SetSize(size string) {
	o.Size = size
}

var _ OrderBasicI = (*OrderbookEntry)(nil)

var (
	// remappings
	symbolToChainDenom = map[string]string{
		"USDC": "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4",                    // USDC
		"BTC":  "factory/osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3/alloyed/allBTC",  // alloyed bitcoin
		"USDT": "factory/osmo1em6xs47hd82806f5cxgyufguxrrc7l0aqx7nzzptjuqgswczk8csavdxek/alloyed/allUSDT", // alloyed USDT
		"ETH":  "factory/osmo1k6c8jln7ejuqwtqmay3yvzrg3kueaczl96pk067ldg8u835w0yhsw27twm/alloyed/allETH",  // alloyed ethereum
		"ATOM": "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",                    // ATOM
	}
)

type Pair struct {
	Base  string
	Quote string
}

func (p Pair) String() string {
	return p.Base + p.Quote
}

func (p Pair) BaseInterchainDenom() string {
	return symbolToChainDenom[p.Base]
}

func (p Pair) QuoteInterchainDenom() string {
	return symbolToChainDenom[p.Quote]
}