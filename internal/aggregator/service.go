package aggregator

import (
	"context"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/bitcoinbrisbane/defi-aggregator/internal/utils"
	"github.com/bitcoinbrisbane/defi-aggregator/internal/clients/uniswap"
	"github.com/bitcoinbrisbane/defi-aggregator/internal/protocols"
	"github.com/ethereum/go-ethereum/common"
)

// RouteQuote represents a single quote from a specific protocol and pool
type RouteQuote struct {
	Protocol     string  `json:"protocol"`     // Protocol name (e.g., "Uniswap V3")
	PoolAddress  string  `json:"poolAddress"`  // Pool address
	Fee          uint64  `json:"fee"`          // Fee tier (e.g., 500, 3000, 10000)
	TokenIn      string  `json:"tokenIn"`      // Input token symbol
	TokenOut     string  `json:"tokenOut"`     // Output token symbol
	AmountIn     string  `json:"amountIn"`     // Input amount (human-readable)
	AmountOut    string  `json:"amountOut"`    // Output amount (human-readable)
	AmountOutRaw *big.Float `json:"-"`         // Raw output amount for sorting (not serialized)
}

// AggregatorResult contains the best routes across all protocols
type AggregatorResult struct {
	BestRoute RouteQuote   `json:"bestRoute"`  // Best overall route
	AllRoutes []RouteQuote `json:"allRoutes"`  // All available routes sorted by output amount
}

// Service handles DEX aggregation logic
type Service struct {
	nodeURL string
}

// NewService creates a new aggregator service
func NewService(nodeURL string) *Service {
	return &Service{
		nodeURL: nodeURL,
	}
}

// FindBestRoute finds the best route for a swap across all supported DEX protocols
func (s *Service) FindBestRoute(
	ctx context.Context, 
	tokenIn, tokenOut common.Address, 
	amountIn *big.Int,
	tokenInDecimals, tokenOutDecimals uint8,
	tokenInSymbol, tokenOutSymbol string,
) (*AggregatorResult, error) {
	// Get all Uniswap-compatible forks
	forks := protocols.GetUniswapForks()
	
	// Create a channel to receive quotes from each protocol
	quotesChan := make(chan []RouteQuote, len(forks))
	
	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	
	// Launch a goroutine for each protocol to get quotes in parallel
	for _, protocol := range forks {
		wg.Add(1)
		go func(protocol protocols.ProtocolConfig) {
			defer wg.Done()
			
			// Create a context with timeout for this query
			queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			
			// Get quotes for this protocol
			quotes, err := s.getProtocolQuotes(
				queryCtx, 
				protocol, 
				tokenIn, 
				tokenOut, 
				amountIn,
				tokenInDecimals,
				tokenOutDecimals,
				tokenInSymbol,
				tokenOutSymbol,
			)
			
			if err != nil {
				log.Printf("Error getting quotes from %s: %v", protocol.Name, err)
				quotesChan <- []RouteQuote{} // Send empty quotes on error
				return
			}
			
			quotesChan <- quotes
		}(protocol)
	}
	
	// Create a goroutine to close the channel when all workers are done
	go func() {
		wg.Wait()
		close(quotesChan)
	}()
	
	// Collect all quotes
	var allRoutes []RouteQuote
	
	// Receive quotes from each protocol
	for quotes := range quotesChan {
		allRoutes = append(allRoutes, quotes...)
	}
	
	// Sort routes by output amount (highest first)
	sortRoutesByOutput(allRoutes)
	
	result := &AggregatorResult{
		AllRoutes: allRoutes,
	}
	
	// Set the best route if we have any
	if len(allRoutes) > 0 {
		result.BestRoute = allRoutes[0]
	}
	
	return result, nil
}

// getProtocolQuotes gets quotes from a specific protocol
func (s *Service) getProtocolQuotes(
	_ context.Context,
	protocol protocols.ProtocolConfig,
	tokenIn, tokenOut common.Address,
	amountIn *big.Int,
	tokenInDecimals, tokenOutDecimals uint8,
	tokenInSymbol, tokenOutSymbol string,
) ([]RouteQuote, error) {
	var routes []RouteQuote
	
	// For each fee tier in the protocol
	for _, feeTier := range protocol.FeeTiers {
		// Get the pool address for this pair and fee tier
		poolAddress, err := uniswap.GetPoolAddress(
			tokenIn, 
			tokenOut, 
			big.NewInt(int64(feeTier)),
			protocol.FactoryAddress,
			s.nodeURL,
		)
		
		// Skip if pool doesn't exist or error occurs
		if err != nil || poolAddress == (common.Address{}) {
			continue
		}
		
		// Get quote for this pool
		amountOut, err := uniswap.GetQuoteExactInputSingle(
			tokenIn,
			tokenOut,
			big.NewInt(int64(feeTier)),
			amountIn,
			protocol.RouterAddress,
			s.nodeURL,
		)
		
		if err != nil {
			continue
		}
		
		// Convert amount out to human-readable format
		amountOutStr := utils.FromWei(amountOut, tokenOutDecimals)
		amountInStr := utils.FromWei(amountIn, tokenInDecimals)
		
		// Parse the output amount as a big.Float for sorting
		amountOutFloat, _ := new(big.Float).SetString(amountOutStr)
		
		// Create route quote
		route := RouteQuote{
			Protocol:     protocol.Name,
			PoolAddress:  poolAddress.String(),
			Fee:          feeTier,
			TokenIn:      tokenInSymbol,
			TokenOut:     tokenOutSymbol,
			AmountIn:     amountInStr,
			AmountOut:    amountOutStr,
			AmountOutRaw: amountOutFloat,
		}
		
		routes = append(routes, route)
	}
	
	return routes, nil
}

// sortRoutesByOutput sorts routes by output amount (highest first)
func sortRoutesByOutput(routes []RouteQuote) {
	for i := 0; i < len(routes); i++ {
		for j := i + 1; j < len(routes); j++ {
			// Compare AmountOutRaw values (higher is better)
			if routes[i].AmountOutRaw.Cmp(routes[j].AmountOutRaw) < 0 {
				// Swap if j has a higher output
				routes[i], routes[j] = routes[j], routes[i]
			}
		}
	}
}