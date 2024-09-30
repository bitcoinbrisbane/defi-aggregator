package main

import (
	"context"
	"fmt"
	"log"
	"os"

	// "github.com/bitcoinbrisbane/defi-aggregator/internal/clients/uniswap"
	"github.com/bitcoinbrisbane/defi-aggregator/internal/clients/curvefi"
	"github.com/bitcoinbrisbane/defi-aggregator/internal/pairs"
	"github.com/ethereum/go-ethereum/common"

	// "github.com/ethereum/go-ethereum/node"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	redisUrl := os.Getenv("REDIS_URL")
	pairHandler := pairs.NewPairHandler(redisUrl)

	ctx := context.Background()

	// TODO: Call the ERC20 token for the metadata

	// Example usage
	tokenA := pairs.ERC20Token{Address: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", Symbol: "USDC", Decimals: 6}
	tokenB := pairs.ERC20Token{Address: "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", Symbol: "WBTC", Decimals: 18}

	// pairHandler.AddPair(tokenA, tokenB)

	// Adding protocol pairs
	pairHandler.AddProtocolPair(ctx, "Uniswap", "0x00", pairs.TokenPair{Token0: tokenA, Token1: tokenB})
	pairHandler.AddProtocolPair(ctx, "SushiSwap", "0x01", pairs.TokenPair{Token0: tokenA, Token1: tokenB})

	// Retrieving protocol pairs
	protocolPairs := pairHandler.GetProtocolPairs(tokenA.Address, tokenB.Address)
	for _, pp := range protocolPairs {
		fmt.Printf("Protocol: %s, Contract: %s, Pair: %s-%s\n",
			pp.ProtocolName, pp.ContractAddress, pp.Pair.Token0.Symbol, pp.Pair.Token1.Symbol)
	}

	// Finding protocols for a specific pair
	protocolsForAB := pairHandler.FindProtocolsForPair(tokenA.Address, tokenB.Address)
	fmt.Printf("Protocols supporting %s-%s pair:\n", tokenA.Symbol, tokenB.Symbol)
	for _, pp := range protocolsForAB {
		fmt.Printf("- %s (Contract: %s)\n", pp.ProtocolName, pp.ContractAddress)
	}

	token0 := common.HexToAddress(tokenA.Address)
	token1 := common.HexToAddress(tokenB.Address)

	nodeUrl := os.Getenv("NODE_URL")

	// do these in parallel

	// uniswap.Quote(token0, token1, nodeUrl)
	// curvefi.Quote(token0, token1, nodeUrl)
	// curvefi.GetPoolAddress(token0, token1, nodeUrl)
	curvefi.GetPrice(token0, token1, nodeUrl)
}
