package main

import (
	"fmt"
	"github.com/bitcoinbrisbane/defi-aggregator/internal/pairs"
)

func main() {
	pairHandler := pairs.NewPairHandler()

	// Example usage
	tokenA := pairs.ERC20Token{Address: "0x123...", Symbol: "TKNA", Decimals: 18}
	tokenB := pairs.ERC20Token{Address: "0x456...", Symbol: "TKNB", Decimals: 18}

	pairHandler.AddPair(tokenA, tokenB)

	// Adding protocol pairs
	pairHandler.AddProtocolPair("Uniswap", "0x00", pairs.TokenPair{Token0: tokenA, Token1: tokenB})
	pairHandler.AddProtocolPair("SushiSwap", "0x01", pairs.TokenPair{Token0: tokenA, Token1: tokenB})

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
}
