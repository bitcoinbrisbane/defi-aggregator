package uniswapv2

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

// Constants for Uniswap V2
const (
	UniswapV2FactoryAddress = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f" // Mainnet V2 Factory
	UniswapV2RouterAddress  = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D" // Mainnet V2 Router
)

// Function signatures for Uniswap V2 interactions
var (
	funcGetPair        = w3.MustNewFunc("getPair(address,address)", "address")
	funcGetReserves    = w3.MustNewFunc("getReserves()", "uint112,uint112,uint32")
	funcGetAmountOut   = w3.MustNewFunc("getAmountOut(uint256,uint256,uint256)", "uint256")
	funcToken0         = w3.MustNewFunc("token0()", "address")
	funcToken1         = w3.MustNewFunc("token1()", "address")
	funcName           = w3.MustNewFunc("name()", "string")
	funcSymbol         = w3.MustNewFunc("symbol()", "string")
	funcDecimals       = w3.MustNewFunc("decimals()", "uint8")
)

// GetPairAddress gets the pool address for a pair of tokens
func GetPairAddress(
	tokenA, tokenB common.Address,
	factoryAddress common.Address,
	nodeURL string,
) (common.Address, error) {
	// Create a client
	client := w3.MustDial(nodeURL)
	defer client.Close()
	
	// Get pair address
	var pairAddress common.Address
	
	err := client.Call(
		eth.CallFunc(factoryAddress, funcGetPair, tokenA, tokenB).Returns(&pairAddress),
	)
	
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to get pair address: %v", err)
	}
	
	// If the returned address is zero, the pair doesn't exist
	if pairAddress == (common.Address{}) {
		return common.Address{}, fmt.Errorf("pair doesn't exist")
	}
	
	return pairAddress, nil
}

// GetReserves gets the reserves for a pair
func GetReserves(
	pairAddress common.Address,
	nodeURL string,
) (*big.Int, *big.Int, uint32, error) {
	// Create a client
	client := w3.MustDial(nodeURL)
	defer client.Close()
	
	// Get reserves
	var reserve0, reserve1 big.Int
	var blockTimestampLast uint32
	
	err := client.Call(
		eth.CallFunc(pairAddress, funcGetReserves).Returns(&reserve0, &reserve1, &blockTimestampLast),
	)
	
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get reserves: %v", err)
	}
	
	return &reserve0, &reserve1, blockTimestampLast, nil
}

// GetTokenOrder gets the token order in the pair
func GetTokenOrder(
	pairAddress common.Address,
	nodeURL string,
) (common.Address, common.Address, error) {
	// Create a client
	client := w3.MustDial(nodeURL)
	defer client.Close()
	
	// Get token0 and token1
	var token0, token1 common.Address
	
	err := client.Call(
		eth.CallFunc(pairAddress, funcToken0).Returns(&token0),
		eth.CallFunc(pairAddress, funcToken1).Returns(&token1),
	)
	
	if err != nil {
		return common.Address{}, common.Address{}, fmt.Errorf("failed to get token order: %v", err)
	}
	
	return token0, token1, nil
}

// CalculateAmountOut calculates the output amount for a given input amount
// Using the Uniswap V2 formula: amountOut = (amountIn * reserveOut * 997) / (reserveIn * 1000 + amountIn * 997)
func CalculateAmountOut(
	amountIn *big.Int,
	reserveIn, reserveOut *big.Int,
) *big.Int {
	if amountIn.Sign() <= 0 || reserveIn.Sign() <= 0 || reserveOut.Sign() <= 0 {
		return big.NewInt(0)
	}
	
	// amountIn * reserveOut
	amountInReserveOut := new(big.Int).Mul(amountIn, reserveOut)
	
	// (amountIn * reserveOut * 997)
	amountInReserveOutWithFee := new(big.Int).Mul(amountInReserveOut, big.NewInt(997))
	
	// reserveIn * 1000
	reserveInScaled := new(big.Int).Mul(reserveIn, big.NewInt(1000))
	
	// amountIn * 997
	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
	
	// (reserveIn * 1000 + amountIn * 997)
	denominator := new(big.Int).Add(reserveInScaled, amountInWithFee)
	
	// (amountIn * reserveOut * 997) / (reserveIn * 1000 + amountIn * 997)
	amountOut := new(big.Int).Div(amountInReserveOutWithFee, denominator)
	
	return amountOut
}

// GetQuote gets a quote for a swap
func GetQuote(
	tokenIn, tokenOut common.Address,
	amountIn *big.Int,
	factoryAddress common.Address,
	nodeURL string,
) (*big.Int, error) {
	// Get pair address
	pairAddress, err := GetPairAddress(tokenIn, tokenOut, factoryAddress, nodeURL)
	if err != nil {
		return nil, err
	}
	
	// Get token order
	token0, token1, err := GetTokenOrder(pairAddress, nodeURL)
	if err != nil {
		return nil, err
	}
	
	// Get reserves
	reserve0, reserve1, _, err := GetReserves(pairAddress, nodeURL)
	if err != nil {
		return nil, err
	}
	
	// Determine which token is the input token and which is the output token
	var reserveIn, reserveOut *big.Int
	if tokenIn == token0 {
		reserveIn = reserve0
		reserveOut = reserve1
	} else {
		reserveIn = reserve1
		reserveOut = reserve0
	}
	
	// Calculate amount out
	amountOut := CalculateAmountOut(amountIn, reserveIn, reserveOut)
	
	return amountOut, nil
}
