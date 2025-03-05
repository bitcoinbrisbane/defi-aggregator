package uniswap

import (
	"fmt"
	"github.com/bitcoinbrisbane/defi-aggregator/internal/pairs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
	"math/big"
)

const factorAddress = "0x1F98431c8aD98523631AE4a59f267346ea31F984"
const routerAddress = "0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6"

var (
	addrUniV3Quoter = w3.A(routerAddress)

	funcQuoteExactInputSingle = w3.MustNewFunc("quoteExactInputSingle(address tokenIn, address tokenOut, uint24 fee, uint256 amountIn, uint160 sqrtPriceLimitX96)", "uint256 amountOut")
	funcName                  = w3.MustNewFunc("name()", "string")
	funcSymbol                = w3.MustNewFunc("symbol()", "string")
	funcDecimals              = w3.MustNewFunc("decimals()", "uint8")

	// flags
	addrTokenIn  common.Address
	addrTokenOut common.Address
	amountIn     big.Int
)

// PairHandlerWrapper wraps pairs.PairHandler to allow method definitions
type PairHandlerWrapper struct {
	pairs.PairHandler
}

// In your uniswap.go file, update the QuoteResponse struct to include fee and route address
type QuoteResponse struct {
	ID            string `json:"id"`
	TokenIn       string `json:"tokenIn"`
	TokenOut      string `json:"tokenOut"`
	AmountIn      string `json:"amountIn"`
	AmountOut     string `json:"amountOut"`
	Fee           string `json:"fee"`          // Add fee information
	RouteAddress  string `json:"routeAddress"` // Add route address
}
// Add a new function to get the best quote from all quotes
func GetBestQuote(quotes []QuoteResponse) *QuoteResponse {
	if len(quotes) == 0 {
		return nil
	}
	
	// Find the quote with the highest amountOut
	bestIndex := 0
	bestAmountOut, _ := new(big.Float).SetString(quotes[0].AmountOut)
	
	for i := 1; i < len(quotes); i++ {
		currentAmountOut, _ := new(big.Float).SetString(quotes[i].AmountOut)
		if currentAmountOut.Cmp(bestAmountOut) > 0 {
			bestAmountOut = currentAmountOut
			bestIndex = i
		}
	}
	
	return &quotes[bestIndex]
}

// Update the Quote function to include fee and router address info
func Quote(tokenA, tokenB common.Address, amount big.Int, nodeUrl string) []QuoteResponse {
    // Use function parameters directly
    addrTokenIn = tokenA
    addrTokenOut = tokenB
    amountIn = amount
    
    // connect to RPC endpoint
    client := w3.MustDial(nodeUrl)
    defer client.Close()

    // fetch token details
    var (
        tokenInName      string
        tokenInSymbol    string
        tokenInDecimals  uint8
        tokenOutName     string
        tokenOutSymbol   string
        tokenOutDecimals uint8
    )

    quotes := make([]QuoteResponse, 0)

    if err := client.Call(
        eth.CallFunc(addrTokenIn, funcName).Returns(&tokenInName),
        eth.CallFunc(addrTokenIn, funcSymbol).Returns(&tokenInSymbol),
        eth.CallFunc(addrTokenIn, funcDecimals).Returns(&tokenInDecimals),
        eth.CallFunc(addrTokenOut, funcName).Returns(&tokenOutName),
        eth.CallFunc(addrTokenOut, funcSymbol).Returns(&tokenOutSymbol),
        eth.CallFunc(addrTokenOut, funcDecimals).Returns(&tokenOutDecimals),
    ); err != nil {
        fmt.Printf("Failed to fetch token details: %v\n", err)
        return quotes
    }

    // fetch quotes
    var (
        fees       = []*big.Int{big.NewInt(500), big.NewInt(3000), big.NewInt(10000)}
        calls      = make([]w3types.RPCCaller, len(fees))
        amountsOut = make([]big.Int, len(fees))
    )

    for i, fee := range fees {
        calls[i] = eth.CallFunc(addrUniV3Quoter, funcQuoteExactInputSingle, addrTokenIn, addrTokenOut, fee, &amountIn, w3.Big0).Returns(&amountsOut[i])
    }

    err := client.Call(calls...)
    callErrs, ok := err.(w3.CallErrors)

    if err != nil && !ok {
        fmt.Printf("Failed to fetch quotes: %v\n", err)
        return quotes
    }

    // print quotes
    fmt.Printf("Exchange %q for %q\n", tokenInName, tokenOutName)
    fmt.Printf("Amount in:\n  %s %s\n", w3.FromWei(&amountIn, tokenInDecimals), tokenInSymbol)
    fmt.Printf("Amount out:\n")

    for i, fee := range fees {
        if ok && callErrs[i] != nil {
            fmt.Printf("  Pool (fee=%5v): Pool does not exist\n", fee)
            continue
        }
        
        // Get pool address for this pair and fee
        poolAddress := GetPoolAddress(addrTokenIn, addrTokenOut, fee, nodeUrl)
        
        fmt.Printf("  Pool (fee=%5v): %s %s\n", fee, w3.FromWei(&amountsOut[i], tokenOutDecimals), tokenOutSymbol)
        quotes = append(quotes, QuoteResponse{
            ID:           fmt.Sprintf("%d", i),
            TokenIn:      tokenInSymbol,
            TokenOut:     tokenOutSymbol,
            AmountIn:     w3.FromWei(&amountIn, tokenInDecimals),
            AmountOut:    w3.FromWei(&amountsOut[i], tokenOutDecimals),
            Fee:          fmt.Sprintf("%d", fee.Uint64()), // Add fee information
            RouteAddress: poolAddress.String(),            // Add route address
        })
    }

    return quotes
}

func GetPoolAddress(tokenIn, tokenOut common.Address, fee *big.Int, nodeUrl string) common.Address {
    client := w3.MustDial(nodeUrl)
    defer client.Close()

    _factoryAddress := common.HexToAddress(factorAddress)
    
    getPool := w3.MustNewFunc("getPool(address,address,uint24)", "address")
    
    var poolAddress common.Address
    
    if err := client.Call(
        eth.CallFunc(_factoryAddress, getPool, tokenIn, tokenOut, fee).Returns(&poolAddress),
    ); err != nil {
        fmt.Printf("Failed to get pool address: %v\n", err)
        return common.Address{}
    }
    
    return poolAddress
}
