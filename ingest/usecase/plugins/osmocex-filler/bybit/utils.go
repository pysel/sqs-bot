package bybit

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	cometrpc "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	wsbybit "github.com/hirokisan/bybit/v2"
	"github.com/osmosis-labs/osmosis/osmomath"
	clmath "github.com/osmosis-labs/osmosis/v25/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/sqs/domain"
	orderbookplugindomain "github.com/osmosis-labs/sqs/domain/orderbook/plugin"
	osmocexfillertypes "github.com/osmosis-labs/sqs/ingest/usecase/plugins/osmocex-filler/types"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

func parseBybitOrderbook(data wsbybit.V5WebsocketPublicOrderBookData) *osmocexfillertypes.OrderbookData {
	bids := make(map[string]string)
	asks := make(map[string]string)

	for _, bid := range data.Bids {
		bids[bid.Price] = bid.Size
	}

	for _, ask := range data.Asks {
		asks[ask.Price] = ask.Size
	}

	return osmocexfillertypes.NewOrderbookData(string(data.Symbol), bids, asks)
}

func (be *BybitExchange) updateBybitOrderbook(data wsbybit.V5WebsocketPublicOrderBookData) {
	orderbookAny, ok := be.orderbooks.Load(data.Symbol)
	if !ok {
		be.logger.Error("orderbook not found", zap.String("symbol", string(data.Symbol)))
		return
	}

	orderbook := orderbookAny.(*osmocexfillertypes.OrderbookData)

	for _, bid := range data.Bids {
		if bid.Size == "0" {
			orderbook.RemoveBid(bid.Price)
		} else {
			orderbook.SetBid(bid.Price, bid.Size)
		}
	}

	for _, ask := range data.Asks {
		if ask.Size == "0" {
			orderbook.RemoveAsk(ask.Price)
		} else {
			orderbook.SetAsk(ask.Price, ask.Size)
		}
	}

	be.orderbooks.Store(string(data.Symbol), orderbook)
}

func (be *BybitExchange) getOsmoOrderbookForPair(pair osmocexfillertypes.Pair) (domain.CanonicalOrderBooksResult, error) {
	base := SymbolToChainDenom[pair.Base]
	quote := SymbolToChainDenom[pair.Quote]

	osmoPoolId, contractAddress, err := (*be.osmoPoolsUsecase).GetCanonicalOrderbookPool(base, quote)
	if err != nil {
		be.logger.Error("failed to get canonical orderbook pool", zap.Error(err))
		return domain.CanonicalOrderBooksResult{}, err
	}

	return domain.CanonicalOrderBooksResult{
		Base:            pair.Base,
		Quote:           pair.Quote,
		PoolID:          osmoPoolId,
		ContractAddress: contractAddress,
	}, nil
}

func (be *BybitExchange) getBybitOrderbookForPair(pair osmocexfillertypes.Pair) (*osmocexfillertypes.OrderbookData, error) {
	orderbookAny, ok := be.orderbooks.Load(pair.String())
	if !ok {
		be.logger.Error("orderbook not found", zap.String("pair", pair.String()))
		return nil, errors.New("orderbook not found")
	}

	orderbook := orderbookAny.(*osmocexfillertypes.OrderbookData)

	return orderbook, nil
}

func (be *BybitExchange) getUnscaledPriceForOrder(pair osmocexfillertypes.Pair, order orderbookplugindomain.Order) (osmomath.BigDec, error) {
	// get osmo highest bid price from tick
	osmoHighestBidPrice, err := clmath.TickToPrice(order.TickId)
	if err != nil {
		return osmomath.NewBigDec(-1), err
	}

	// unscale osmoHighestBidPrice
	osmoHighestBidPrice, err = be.unscalePrice(osmoHighestBidPrice, pair.Base, pair.Quote)
	if err != nil {
		return osmomath.NewBigDec(-1), err
	}

	return osmoHighestBidPrice, nil
}

type AccountInfo struct {
	Sequence      string `json:"sequence"`
	AccountNumber string `json:"account_number"`
}

type AccountResult struct {
	Account AccountInfo `json:"account"`
}

func getInitialSequence(ctx context.Context, address string) (uint64, uint64) {
	resp, err := httpGet(ctx, LCD+"/cosmos/auth/v1beta1/accounts/"+address)
	if err != nil {
		log.Printf("Failed to get initial sequence: %v", err)
		return 0, 0
	}

	var accountRes AccountResult
	err = json.Unmarshal(resp, &accountRes)
	if err != nil {
		log.Printf("Failed to unmarshal account result: %v", err)
		return 0, 0
	}

	seqint, err := strconv.ParseUint(accountRes.Account.Sequence, 10, 64)
	if err != nil {
		log.Printf("Failed to convert sequence to int: %v", err)
		return 0, 0
	}

	accnum, err := strconv.ParseUint(accountRes.Account.AccountNumber, 10, 64)
	if err != nil {
		log.Printf("Failed to convert account number to int: %v", err)
		return 0, 0
	}

	return seqint, accnum
}

var client = &http.Client{
	Timeout:   10 * time.Second, // Adjusted timeout to 10 seconds
	Transport: otelhttp.NewTransport(http.DefaultTransport),
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		netErr, ok := err.(net.Error)
		if ok && netErr.Timeout() {
			log.Printf("Request to %s timed out, continuing...", url)
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// broadcastTransaction broadcasts a transaction to the chain.
// Returning the result and error.
func broadcastTransaction(ctx context.Context, txBytes []byte, rpcEndpoint string) (*coretypes.ResultBroadcastTx, error) {
	cmtCli, err := cometrpc.New(rpcEndpoint, "/websocket")
	if err != nil {
		return nil, err
	}

	t := tmtypes.Tx(txBytes)

	res, err := cmtCli.BroadcastTxSync(ctx, t)
	if err != nil {
		return nil, err
	}

	return res, nil
}