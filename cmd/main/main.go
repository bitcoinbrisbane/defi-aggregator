package main

import (
	"fmt"
)

// ERC20Token represents a single ERC20 token
type ERC20Token struct {
	Address  string
	Symbol   string
	Decimals int
}

// TokenPair represents a pair of ERC20 tokens
type TokenPair struct {
	Token0 ERC20Token
	Token1 ERC20Token
}

// ProtocolPair represents a DeFi protocol contract address and its associated token pair
type ProtocolPair struct {
	ProtocolName    string
	ContractAddress string
	Pair            TokenPair
}

// PairHandler manages operations related to token pairs
type PairHandler struct {
	Pairs         map[string]TokenPair
	ProtocolPairs map[string][]ProtocolPair
}

// NewPairHandler creates a new PairHandler instance
func NewPairHandler() *PairHandler {
	return &PairHandler{
		Pairs:         make(map[string]TokenPair),
		ProtocolPairs: make(map[string][]ProtocolPair),
	}
}

// AddProtocolPair adds a new protocol pair to the handler
func (ph *PairHandler) AddProtocolPair(protocolName, contractAddress string, pair TokenPair) {
	protocolPair := ProtocolPair{
		ProtocolName:    protocolName,
		ContractAddress: contractAddress,
		Pair:            pair,
	}
	pairKey := getPairKey(pair.Token0.Address, pair.Token1.Address)
	ph.ProtocolPairs[pairKey] = append(ph.ProtocolPairs[pairKey], protocolPair)
}

// GetProtocolPairs retrieves all protocol pairs for a given token pair
func (ph *PairHandler) GetProtocolPairs(token0Address, token1Address string) []ProtocolPair {
	pairKey := getPairKey(token0Address, token1Address)
	return ph.ProtocolPairs[pairKey]
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

// FindProtocolsForPair finds all protocols that include the specified token pair
func (ph *PairHandler) FindProtocolsForPair(token0Address, token1Address string) []ProtocolPair {
	pairKey := getPairKey(token0Address, token1Address)
	protocols := ph.ProtocolPairs[pairKey]

	// Also check for the reverse pair
	reversePairKey := getPairKey(token1Address, token0Address)
	reverseProtocols := ph.ProtocolPairs[reversePairKey]

	// Combine and return all found protocols
	return append(protocols, reverseProtocols...)
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

	// Adding protocol pairs
	pairHandler.AddProtocolPair("Uniswap", "0x00", TokenPair{Token0: tokenA, Token1: tokenB})
	pairHandler.AddProtocolPair("SushiSwap", "0x01", TokenPair{Token0: tokenA, Token1: tokenB})

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
