package usecase

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/sqs/domain"
	"github.com/osmosis-labs/sqs/log"
	"github.com/osmosis-labs/sqs/router/usecase/route"
	"github.com/osmosis-labs/sqs/sqsdomain"
)

type (
	RouterUseCaseImpl = routerUseCaseImpl

	QuoteImpl = quoteExactAmountIn

	CandidatePoolWrapper  = candidatePoolWrapper
	CandidateRouteWrapper = candidateRouteWrapper
)

const (
	NoPoolLiquidityCapError = noPoolLiquidityCapError
)

func ValidateAndFilterRoutes(candidateRoutes []candidateRouteWrapper, tokenInDenom string, logger log.Logger) (sqsdomain.CandidateRoutes, error) {
	return validateAndFilterRoutes(candidateRoutes, tokenInDenom, logger)
}

func (r *routerUseCaseImpl) HandleRoutes(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string, candidateRouteSearchOptions domain.CandidateRouteSearchOptions) (candidateRoutes sqsdomain.CandidateRoutes, err error) {
	return r.handleCandidateRoutes(ctx, tokenIn, tokenOutDenom, candidateRouteSearchOptions)
}

func (r *routerUseCaseImpl) EstimateAndRankSingleRouteQuote(ctx context.Context, routes []route.RouteImpl, tokenIn sdk.Coin, logger log.Logger) (domain.Quote, []RouteWithOutAmount, error) {
	return r.estimateAndRankSingleRouteQuote(ctx, routes, tokenIn, logger)
}

func FilterDuplicatePoolIDRoutes(rankedRoutes []RouteWithOutAmount) []route.RouteImpl {
	return filterAndConvertDuplicatePoolIDRankedRoutes(rankedRoutes)
}

func ConvertRankedToCandidateRoutes(rankedRoutes []route.RouteImpl) sqsdomain.CandidateRoutes {
	return convertRankedToCandidateRoutes(rankedRoutes)
}

func FormatRankedRouteCacheKey(tokenInDenom string, tokenOutDenom string, tokenIOrderOfMagnitude int) string {
	return formatRankedRouteCacheKey(tokenInDenom, tokenOutDenom, tokenIOrderOfMagnitude)
}

func FormatRouteCacheKey(tokenInDenom string, tokenOutDenom string) string {
	return formatRouteCacheKey(tokenInDenom, tokenOutDenom)
}

func FormatCandidateRouteCacheKey(tokenInDenom string, tokenOutDenom string) string {
	return formatCandidateRouteCacheKey(tokenInDenom, tokenOutDenom)
}

func SortPools(pools []sqsdomain.PoolI, transmuterCodeIDs map[uint64]struct{}, totalTVL osmomath.Int, preferredPoolIDsMap map[uint64]struct{}, logger log.Logger) []sqsdomain.PoolI {
	return sortPools(pools, transmuterCodeIDs, totalTVL, preferredPoolIDsMap, logger)
}

func GetSplitQuote(ctx context.Context, routes []route.RouteImpl, tokenIn sdk.Coin) (domain.Quote, error) {
	return getSplitQuote(ctx, routes, tokenIn)
}

func (r *routerUseCaseImpl) RankRoutesByDirectQuote(ctx context.Context, candidateRoutes sqsdomain.CandidateRoutes, tokenIn sdk.Coin, tokenOutDenom string, maxRoutes int) (domain.Quote, []route.RouteImpl, error) {
	return r.rankRoutesByDirectQuote(ctx, candidateRoutes, tokenIn, tokenOutDenom, maxRoutes)
}

func CutRoutesForSplits(maxSplitRoutes int, routes []route.RouteImpl) []route.RouteImpl {
	return cutRoutesForSplits(maxSplitRoutes, routes)
}

func (r *routerUseCaseImpl) SetCandidateRouteCacheToMock(tokenInDenom, tokenOutDenom string) {
	r.candidateRouteCache.Set(formatCandidateRouteCacheKey(tokenInDenom, tokenOutDenom), sqsdomain.CandidateRoutes{
		// Note: some mock dummy values
		Routes: []sqsdomain.CandidateRoute{
			{}, {},
		}}, 0)
}

func (r *routerUseCaseImpl) SetRankedRouteCacheToMock(tokenInDenom, tokenOutDenom string, orderOfMagnitude int) {
	r.rankedRouteCache.Set(formatRankedRouteCacheKey(tokenInDenom, tokenOutDenom, orderOfMagnitude), sqsdomain.CandidateRoutes{
		// Note: some mock dummy values
		Routes: []sqsdomain.CandidateRoute{
			{}, {},
		}}, 0)

}
