package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"

	"github.com/bitcoinbrisbane/defi-aggregator/internal/clients/pancake"

	// "github.com/bitcoinbrisbane/defi-aggregator/internal/clients/uniswap"
	// "github.com/bitcoinbrisbane/defi-aggregator/internal/clients/curvefi"
	"github.com/bitcoinbrisbane/defi-aggregator/internal/pairs"
	"github.com/ethereum/go-ethereum/common"

	// "github.com/ethereum/go-ethereum/node"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"net/http"
)

type Quote struct {
	TokenIn   string
	TokenOut  string
	AmountIn  *big.Int
	AmountOut *big.Int
}

type Response struct {
	Message string `json:"message"`
}

// Config holds all our environment configurations
type Config struct {
	Port     string
	RedisURL string
}

// Global config variable
var config Config

func loadConfig() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize config with environment variables
	config = Config{
		Port:     getEnvWithDefault("PORT", "8080"),
		RedisURL: getEnvWithDefault("REDIS_URL", "localhost:6379"),
	}

	log.Printf("Config loaded. Port: %s", config.Port)
}

// getEnvWithDefault gets an environment variable or returns a default value
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	// Create a default gin router
	router := gin.Default()

	// Load the configuration
	loadConfig()

	// Add middleware for logging and recovery
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/hello", helloHandler)
	router.GET("/pairs", pairHandler)

	// Start the server
	router.Run(":" + config.Port)
}

func helloHandler(c *gin.Context) {
	response := Response{
		Message: "Hello, World!",
	}
	c.JSON(http.StatusOK, response)
}

func pairHandler(c *gin.Context) {
	// Initialize the pair handler
	redisUrl := os.Getenv("REDIS_URL")
	pairHandler := pairs.NewPairHandler(redisUrl)

	ctx := context.Background()

	// TODO: Call the ERC20 token for the metadata

	// Example usage
	tokenA := pairs.ERC20Token{Address: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", Symbol: "USDC", Decimals: 6}
	tokenB := pairs.ERC20Token{Address: "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", Symbol: "WBTC", Decimals: 18}

	// pairHandler.AddPair(tokenA, tokenB)

	// Adding protocol pairs
	pairHandler.AddProtocolPair(ctx, "UniswapV3_10000", "0xCBFB0745b8489973Bf7b334d54fdBd573Df7eF3c", pairs.TokenPair{Token0: tokenA, Token1: tokenB})
	pairHandler.AddProtocolPair(ctx, "UniswapV3_10000", "0xCBFB0745b8489973Bf7b334d54fdBd573Df7eF3c", pairs.TokenPair{Token0: tokenA, Token1: tokenB})
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
	fee := big.NewInt(5000)
	pancake.Quote(token0, token1, fee, nodeUrl)

	// uniswap.Quote(token0, token1, nodeUrl)
	// curvefi.Quote(token0, token1, nodeUrl)
	// curvefi.GetPoolAddress(token0, token1, nodeUrl)
	// curvefi.GetPrice(token0, token1, nodeUrl)

	c.JSON(http.StatusOK, gin.H{
		"message": "Pair handler",
	})
}

func test() {
	quote1 := Quote{
		TokenIn:   "USDC",
		TokenOut:  "USDT",
		AmountIn:  big.NewInt(1000000),
		AmountOut: big.NewInt(999999),
	}

	quote2 := Quote{
		TokenIn:   "USDC",
		TokenOut:  "WBTC",
		AmountIn:  big.NewInt(1000000),
		AmountOut: big.NewInt(999999),
	}

	_distance := distance(quote1, quote2)
	fmt.Printf("Distance: %f\n", _distance)
}

func distance(quote1, quote2 Quote) float64 {
	dx := quote1.AmountIn.Int64() - quote2.AmountIn.Int64()
	dy := quote1.AmountOut.Int64() - quote2.AmountOut.Int64()

	return math.Sqrt(float64(dx*dx + dy*dy))
}
