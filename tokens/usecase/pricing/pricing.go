package pricing

import (
	"fmt"
	"time"

	"github.com/osmosis-labs/sqs/domain"
	"github.com/osmosis-labs/sqs/domain/mvc"
	chainpricing "github.com/osmosis-labs/sqs/tokens/usecase/pricing/chain"
)

// NewPricingStrategy is a factory method to create the pricing strategy based on the desired source.
func NewPricingStrategy(config domain.PricingConfig, tokensUsecase mvc.TokensUsecase, routerUseCase mvc.RouterUsecase) (domain.PricingStrategy, error) {
	if config.DefaultSource == domain.ChainPricingSource {
		cacheExpiryNs := config.CacheExpiryMs * time.Millisecond
		return chainpricing.New(routerUseCase, tokensUsecase, cacheExpiryNs), nil
	}

	return nil, fmt.Errorf("pricing source (%d) is not supported", config.DefaultSource)
}
