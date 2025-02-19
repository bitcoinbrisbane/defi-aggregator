package pancake

import (
	"flag"
	"fmt"
	"github.com/bitcoinbrisbane/defi-aggregator/internal/pairs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
	"log"
	"math/big"
)

var (
	// addrQuoter = w3.A("0xB048Bbc1Ee6b733FFfCFb9e9CeF7375518e25997")
	factorAddress = w3.A("0x0BFbCF9fa4f9C56B0F40a671Ad40E0805A091865")

	// funcQuoteExactInputSingle = w3.MustNewFunc("quoteExactInputSingle(address tokenIn, address tokenOut, uint24 fee, uint256 amountIn, uint160 sqrtPriceLimitX96)", "uint256 amountOut")
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

func Quote(tokenA, tokenB common.Address, fee *big.Int, rawUrl string) {
	// parse flags
	flag.TextVar(&amountIn, "amountIn", w3.I("1 ether"), "Token address")
	//flag.TextVar(amountIn, "amountIn", w3.I("1 ether"), "Token address")
	flag.TextVar(&addrTokenIn, "tokenIn", tokenA, "Token in")
	flag.TextVar(&addrTokenOut, "tokenOut", tokenB, "Token out")

	flag.Usage = func() {
		fmt.Println("pancake prints the Pancake V3 exchange rate to swap amountIn of tokenIn for tokenOut.")
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

	// amountOut := big.NewInt(0)
	// callFunc := eth.CallFunc(addrQuoter, funcQuoteExactInputSingle, addrTokenIn, addrTokenOut, big.NewInt(3000), amount, w3.Big0).Returns(&amountOut)

	// err := client.Call(callFunc)
	// callErr, ok := err.(w3.CallErrors)

	// if err != nil && !ok {
	// 	fmt.Printf("Failed to fetch quotes: %v\n", err)
	// 	return

	// }

	// // print quotes
	// fmt.Printf("Exchange %q for %q\n", tokenInName, tokenOutName)
	// fmt.Printf("Amount in:\n  %s %s\n", w3.FromWei(&amountIn, tokenInDecimals), tokenInSymbol)
	// fmt.Printf("Amount out:\n")

	// if ok && callErr != nil {
	// 	fmt.Printf("  Pool (fee=%5v): Pool does not exist\n", 3000)
	// }

	// fmt.Printf("  Pool (fee=%5v): %s %s\n", 3000, w3.FromWei(amountOut, tokenOutDecimals), tokenOutSymbol)

}

func Quotes(tokenA, tokenB common.Address, rawUrl string) {
	// parse flags
	flag.TextVar(&amountIn, "amountIn", w3.I("1 ether"), "Token address")
	flag.TextVar(&addrTokenIn, "tokenIn", tokenA, "Token in")
	flag.TextVar(&addrTokenOut, "tokenOut", tokenB, "Token out")

	flag.Usage = func() {
		fmt.Println("pancake prints the Pancake V3 exchange rate to swap amountIn of tokenIn for tokenOut.")
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

	// for i, fee := range fees {
	// 	// calls[i] = eth.CallFunc(factorAddress, funcQuoteExactInputSingle, addrTokenIn, addrTokenOut, fee, &amountIn, w3.Big0).Returns(&amountsOut[i])
	// }

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

	const factorAddress = "0x0BFbCF9fa4f9C56B0F40a671Ad40E0805A091865"
	fmt.Println(factorAddress)

	_factoryAddress := common.HexToAddress(factorAddress)

	fee := &big.Int{}
	fee.SetInt64(3000)

	getPool := w3.MustNewFunc("getPool(address,address,uint24)", "address")
	input, err := getPool.EncodeArgs(tokenIn, tokenOut, fee)
	fmt.Printf("getPool input: 0x%x\n", input)

	if err != nil {
		log.Fatalf("Failed to encode arguments: %v", err)
	}

	var poolAddress string

	if err := client.Call(
		eth.CallFunc(_factoryAddress, getPool, input).Returns(&poolAddress),
	); err != nil {
		fmt.Printf("Request failed: %v\n", err)
	}

	return common.HexToAddress(poolAddress)
}
