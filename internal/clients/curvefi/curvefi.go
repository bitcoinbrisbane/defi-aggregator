package curvefi

import (
	"flag"
	"fmt"
	"github.com/bitcoinbrisbane/defi-aggregator/internal/pairs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
	"math/big"
)

// ERC20Token represents an ERC20 token
var (
	addrUniV3Quoter = w3.A("0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6")

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

func Quote(tokenA, tokenB common.Address, rawUrl string) {

	// parse flags
	flag.TextVar(&amountIn, "amountIn", w3.I("1 ether"), "Token address")
	flag.TextVar(&addrTokenIn, "tokenIn", tokenA, "Token in")
	flag.TextVar(&addrTokenOut, "tokenOut", tokenB, "Token out")

	flag.Usage = func() {
		fmt.Println("curve.fi exchange rate to swap amountIn of tokenIn for tokenOut.")
		flag.PrintDefaults()
	}

	flag.Parse()

	// connect to RPC endpoint
	client := w3.MustDial(rawUrl)
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

	if err := client.Call(
		eth.CallFunc(addrTokenIn, funcName).Returns(&tokenInName),
		eth.CallFunc(addrTokenIn, funcSymbol).Returns(&tokenInSymbol),
		eth.CallFunc(addrTokenIn, funcDecimals).Returns(&tokenInDecimals),
		eth.CallFunc(addrTokenOut, funcName).Returns(&tokenOutName),
		eth.CallFunc(addrTokenOut, funcSymbol).Returns(&tokenOutSymbol),
		eth.CallFunc(addrTokenOut, funcDecimals).Returns(&tokenOutDecimals),
	); err != nil {
		fmt.Printf("Failed to fetch token details: %v\n", err)
		return
	}

	// fetch quotes
	var (
		fees       = []*big.Int{big.NewInt(100), big.NewInt(500), big.NewInt(3000), big.NewInt(10000)}
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
		return

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
		fmt.Printf("  Pool (fee=%5v): %s %s\n", fee, w3.FromWei(&amountsOut[i], tokenOutDecimals), tokenOutSymbol)
	}
}

func GetPoolAddress(tokenIn, tokenOut common.Address, nodeUrl string) common.Address {
	client := w3.MustDial(nodeUrl)
	defer client.Close()

	// https://etherscan.io/token/0x0c0e5f2fF0ff18a3be9b835635039256dC4B4963#readContract
	factorAddress := "0x0c0e5f2fF0ff18a3be9b835635039256dC4B4963"
	fmt.Println(factorAddress)

	_factoryAddress := common.HexToAddress(factorAddress)

	getPool := w3.MustNewFunc("find_pool_for_coins(address,address)", "address")

	var poolAddress string

	if err := client.Call(
		eth.CallFunc(_factoryAddress, getPool, tokenIn, tokenOut).Returns(&poolAddress),
	); err != nil {
		fmt.Printf("Request failed: %v\n", err)
	}

	return common.HexToAddress(poolAddress)
}

func GetPrice(tokenIn, tokenOut common.Address, nodeUrl string) big.Int {
	client := w3.MustDial(nodeUrl)
	defer client.Close()

	// poolAddress := GetPoolAddress(tokenIn, tokenOut, nodeUrl)
	poolAddress := common.HexToAddress("0x7F86Bf177Dd4F3494b841a37e810A34dD56c829B")
	getPrice := w3.MustNewFunc("get_virtual_price()", "uint256")

	var price big.Int

	if err := client.Call(
		eth.CallFunc(poolAddress, getPrice).Returns(&price),
	); err != nil {
		fmt.Printf("Request failed: %v\n", err)
	}

	fmt.Printf("Price: %s\n", w3.FromWei(&price, 18))

	return price
}
