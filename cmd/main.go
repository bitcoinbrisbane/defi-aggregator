package main

import (
	"fmt"
	"math/big"
)

// ERC20Token represents a single ERC20 token
type ERC20Token struct {
	Address string
	Symbol  string
	Decimals int
}

// TokenPair represents a pair of ERC20 tokens
type TokenPair struct {
	Token0 ERC20Token
	Token1 ERC20Token
}

// PairHandler manages operations related to token pairs
type PairHandler struct {
	Pairs map[string]TokenPair
}

// NewPairHandler creates a new PairHandler instance
func NewPairHandler() *PairHandler {
	return &PairHandler{
		Pairs: make(map[string]TokenPair),
	}
}

// AddPair adds a new token pair to the handler
func (ph *PairHandler) AddPair(token0, token1 ERC20Token) {
	pairKey := getPairKey(token0.Address, token1.Address)
	ph.Pairs[pairKey] = TokenPair{Token0: token0, Token1: token1}
}

// GetPair retrieves a token pair from the handler
func (ph *PairHandler) GetPair(token0Address, token1Address string) (TokenPair, bool) {
	pairKey := getPairKey(token0Address, token1Address)
	pair, exists := ph.Pairs[pairKey]
	return pair, exists
}

// Helper function to generate a unique key for a token pair
func getPairKey(address0, address1 string) string {
	if address0 < address1 {
		return address0 + "-" + address1
	}
	return address1 + "-" + address0
}

func main() {
	pairHandler := NewPairHandler()

	// Example usage
	tokenA := ERC20Token{Address: "0x123...", Symbol: "TKNA", Decimals: 18}
	tokenB := ERC20Token{Address: "0x456...", Symbol: "TKNB", Decimals: 18}

	pairHandler.AddPair(tokenA, tokenB)

	if pair, exists := pairHandler.GetPair(tokenA.Address, tokenB.Address); exists {
		fmt.Printf("Found pair: %s - %s\n", pair.Token0.Symbol, pair.Token1.Symbol)
	} else {
		fmt.Println("Pair not found")
	}
}