package uniswap

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
)

// Function signatures for Uniswap interactions
var (
	funcQuoteExactInputSingle = w3.MustNewFunc("quoteExactInputSingle(address tokenIn, address tokenOut, uint24 fee, uint256 amountIn, uint160 sqrtPriceLimitX96)", "uint256 amountOut")
	funcName                  = w3.MustNewFunc("name()", "string")
	funcSymbol                = w3.MustNewFunc("symbol()", "string")
	funcDecimals              = w3.MustNewFunc("decimals()", "uint8")
	funcGetPool               = w3.MustNewFunc("getPool(address,address,uint24)", "address")
)

// GetQuoteExactInputSingle gets a quote for a swap directly from the router contract
func GetQuoteExactInputSingle(
	tokenIn, tokenOut common.Address,
	fee *big.Int,
	amountIn *big.Int,
	routerAddress common.Address,
	nodeURL string,
) (*big.Int, error) {
	// Create a client
	client := w3.MustDial(nodeURL)
	defer client.Close()
	
	// Get quote
	var amountOut big.Int
	
	err := client.Call(
		eth.CallFunc(routerAddress, funcQuoteExactInputSingle, tokenIn, tokenOut, fee, amountIn, w3.Big0).Returns(&amountOut),
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %v", err)
	}
	
	return &amountOut, nil
}

// GetPoolAddress gets the pool address for a pair of tokens and a fee tier
func GetPoolAddress(
	tokenIn, tokenOut common.Address,
	fee *big.Int,
	factoryAddress common.Address,
	nodeURL string,
) (common.Address, error) {
	// Create a client
	client := w3.MustDial(nodeURL)
	defer client.Close()
	
	// Get pool address
	var poolAddress common.Address
	
	err := client.Call(
		eth.CallFunc(factoryAddress, funcGetPool, tokenIn, tokenOut, fee).Returns(&poolAddress),
	)
	
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to get pool address: %v", err)
	}
	
	// If the returned address is zero, the pool doesn't exist
	if poolAddress == (common.Address{}) {
		return common.Address{}, fmt.Errorf("pool doesn't exist")
	}
	
	return poolAddress, nil
}

// GetTokenMetadata gets token metadata (name, symbol, decimals)
func GetTokenMetadata(tokenAddress common.Address, nodeURL string) (string, string, uint8, error) {
	// Create a client
	client := w3.MustDial(nodeURL)
	defer client.Close()
	
	// Get token metadata
	var (
		name     string
		symbol   string
		decimals uint8
	)
	
	err := client.Call(
		eth.CallFunc(tokenAddress, funcName).Returns(&name),
		eth.CallFunc(tokenAddress, funcSymbol).Returns(&symbol),
		eth.CallFunc(tokenAddress, funcDecimals).Returns(&decimals),
	)
	
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get token metadata: %v", err)
	}
	
	return name, symbol, decimals, nil
}

// FromWei converts a wei amount to a human-readable decimal string based on the token's decimals
func FromWei(amount *big.Int, decimals uint8) string {
	if amount == nil {
		return "0"
	}
	
	// Create a copy of the amount
	wei := new(big.Int).Set(amount)
	
	// Convert to a decimal string based on decimals
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	
	// Integer part
	intPart := new(big.Int).Div(wei, divisor)
	
	// Fractional part
	fracPart := new(big.Int).Mod(wei, divisor)
	
	// Format fractional part with leading zeros
	fracStr := fmt.Sprintf("%0*s", decimals, fracPart.String())
	
	// Trim trailing zeros
	for len(fracStr) > 0 && fracStr[len(fracStr)-1] == '0' {
		fracStr = fracStr[:len(fracStr)-1]
	}
	
	if len(fracStr) > 0 {
		return fmt.Sprintf("%s.%s", intPart.String(), fracStr)
	}
	
	return intPart.String()
}

// ToWei converts a human-readable decimal string to wei based on the token's decimals
func ToWei(amount string, decimals uint8) (*big.Int, error) {
	// Parse the decimal amount
	parts := splitDecimal(amount)
	intPart, fracPart := parts[0], ""
	if len(parts) > 1 {
		fracPart = parts[1]
	}
	
	// Parse integer part
	intValue, ok := new(big.Int).SetString(intPart, 10)
	if !ok {
		return nil, fmt.Errorf("invalid integer part: %s", intPart)
	}
	
	// Multiply by 10^decimals
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	intValue.Mul(intValue, multiplier)
	
	// Add fractional part if it exists
	if fracPart != "" {
		// Pad or truncate fractional part to match decimals
		if len(fracPart) > int(decimals) {
			fracPart = fracPart[:decimals]
		} else {
			fracPart = fracPart + "0000000000000000000"[:int(decimals)-len(fracPart)]
		}
		
		// Parse fractional part
		fracValue, ok := new(big.Int).SetString(fracPart, 10)
		if !ok {
			return nil, fmt.Errorf("invalid fractional part: %s", fracPart)
		}
		
		// Add to result
		intValue.Add(intValue, fracValue)
	}
	
	return intValue, nil
}

// splitDecimal splits a decimal string into integer and fractional parts
func splitDecimal(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// GetAllQuotes gets quotes from all fee tiers for a token pair
func GetAllQuotes(
	tokenIn, tokenOut common.Address,
	amountIn *big.Int,
	routerAddress common.Address,
	feeTiers []uint64,
	nodeURL string,
) (map[uint64]*big.Int, error) {
	// Create a client
	client := w3.MustDial(nodeURL)
	defer client.Close()
	
	// Prepare calls for all fee tiers
	calls := make([]w3types.RPCCaller, 0, len(feeTiers))
	amountsOut := make([]*big.Int, len(feeTiers))
	
	for i := range feeTiers {
		amountsOut[i] = new(big.Int)
		calls = append(
			calls,
			eth.CallFunc(
				routerAddress,
				funcQuoteExactInputSingle,
				tokenIn,
				tokenOut,
				big.NewInt(int64(feeTiers[i])),
				amountIn,
				w3.Big0,
			).Returns(amountsOut[i]),
		)
	}
	
	// Execute batch request
	err := client.Call(calls...)
	callErrs, ok := err.(w3.CallErrors)
	
	// Handle complete failure
	if err != nil && !ok {
		return nil, fmt.Errorf("failed to batch fetch quotes: %v", err)
	}
	
	// Process results
	results := make(map[uint64]*big.Int)
	
	for i, feeTier := range feeTiers {
		// Skip failed calls
		if ok && callErrs[i] != nil {
			log.Printf("Failed to get quote for fee tier %d: %v", feeTier, callErrs[i])
			continue
		}
		
		// Store successful results
		results[feeTier] = amountsOut[i]
	}
	
	return results, nil
}