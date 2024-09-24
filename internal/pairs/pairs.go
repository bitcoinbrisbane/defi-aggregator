package pairs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
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
	redisClient   *redis.Client
	Pairs         map[string]TokenPair
	ProtocolPairs map[string][]ProtocolPair
}

func NewERC20Token(address, symbol string, decimals int) ERC20Token {
	return ERC20Token{Address: address, Symbol: symbol, Decimals: decimals}
}

func NewTokenPair(token0, token1 ERC20Token) TokenPair {
	return TokenPair{Token0: token0, Token1: token1}
}

func NewProtocolPair(protocolName, contractAddress string, pair TokenPair) ProtocolPair {
	return ProtocolPair{ProtocolName: protocolName, ContractAddress: contractAddress, Pair: pair}
}

// NewPairHandler creates a new PairHandler instance
func NewPairHandler(redisUrl string) *PairHandler {
	client := redis.NewClient(&redis.Options{
		Addr: redisUrl,
		DB:   0,
	})

	return &PairHandler{
		redisClient:   client,
		Pairs:         make(map[string]TokenPair),
		ProtocolPairs: make(map[string][]ProtocolPair),
	}
}

// AddProtocolPair adds a new protocol pair to the handler
func (ph *PairHandler) AddProtocolPair(ctx context.Context, protocolName, contractAddress string, pair TokenPair) {
	protocolPair := ProtocolPair{
		ProtocolName:    protocolName,
		ContractAddress: contractAddress,
		Pair:            pair,
	}

	pairKey := getPairKey(pair.Token0.Address, pair.Token1.Address)
	data, err := json.Marshal(protocolPair)

	if err != nil {
		fmt.Printf("failed to marshal protocol pair: %v\n", err)
		return
	}

	err = ph.redisClient.SAdd(ctx, pairKey, data).Err()
	if err != nil {
		fmt.Printf("failed to add protocol pair to Redis: %v", err)
		return
	}
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
