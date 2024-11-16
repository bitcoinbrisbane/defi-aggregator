package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"

	"github.com/bitcoinbrisbane/defi-aggregator/internal/clients/uniswap"
	
	// "github.com/bitcoinbrisbane/defi-aggregator/internal/clients/uniswap"
	// "github.com/bitcoinbrisbane/defi-aggregator/internal/clients/curvefi"
	"github.com/bitcoinbrisbane/defi-aggregator/internal/pairs"
	"github.com/ethereum/go-ethereum/common"

	// "github.com/ethereum/go-ethereum/node"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
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
	NodeURL  string
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
		NodeURL:  getEnvWithDefault("NODE_URL", "https://eth-mainnet.g.alchemy.com/v2/-Lh1_OMuwKGBKgoU4nk07nz98TYeUZxj"),
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

	router.GET("/", helloHandler)
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

func setupPairs() {
	_tokenA := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	_tokenB := common.HexToAddress("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599")
	tokenA := pairs.ERC20Token{Address: _tokenA, Symbol: "USDC", Decimals: 6}
	tokenB := pairs.ERC20Token{Address: _tokenB, Symbol: "WBTC", Decimals: 18}

	redisUrl := config.RedisURL
	pairHandler := pairs.NewPairHandler(redisUrl)

	ctx := context.Background()

	pairHandler.AddPair(tokenA, tokenB)

	// Adding protocol pairs
	pairHandler.AddProtocolPair(ctx, "UniswapV3_10000", "0xCBFB0745b8489973Bf7b334d54fdBd573Df7eF3c", pairs.TokenPair{Token0: tokenA, Token1: tokenB})
	pairHandler.AddProtocolPair(ctx, "UniswapV3_30000", "0xCBFB0745b8489973Bf7b334d54fdBd573Df7eF3c", pairs.TokenPair{Token0: tokenA, Token1: tokenB})
	pairHandler.AddProtocolPair(ctx, "SushiSwap", "0x01", pairs.TokenPair{Token0: tokenA, Token1: tokenB})

	// Retrieving protocol pairs
	protocolPairs := pairHandler.GetProtocolPairs(tokenA.Address.String(), tokenB.Address.String())
	for _, pp := range protocolPairs {
		fmt.Printf("Protocol: %s, Contract: %s, Pair: %s-%s\n",
			pp.ProtocolName, pp.ContractAddress, pp.Pair.Token0.Symbol, pp.Pair.Token1.Symbol)
	}

	// Finding protocols for a specific pair
	protocolsForAB := pairHandler.FindProtocolsForPair(tokenA.Address.String(), tokenB.Address.String())
	fmt.Printf("Protocols supporting %s-%s pair:\n", tokenA.Symbol, tokenB.Symbol)
	for _, pp := range protocolsForAB {
		fmt.Printf("- %s (Contract: %s)\n", pp.ProtocolName, pp.ContractAddress)
	}
}

func getMetadata(token common.Address) pairs.ERC20Token {

	nodeUrl := config.NodeURL
	// redisUrl := config.RedisURL

	// // Check to see if the token metadata is in Redis
	// client := redis.NewClient(&redis.Options{
	// 	Addr:     redisUrl,
	// 	DB:       0,
	// 	Password: "Test1234!",
	// })

	// // client.Set(context.Background(), token.String(), "metadata", 0)
	// client.Get(context.Background(), token.String())

	// return &PairHandler{
	// 	redisClient:   client,
	// 	Pairs:         make(map[string]TokenPair),
	// 	ProtocolPairs: make(map[string][]ProtocolPair),
	// }

	client := w3.MustDial(nodeUrl)
	defer client.Close()

	var (
		funcName     = w3.MustNewFunc("name()", "string")
		funcSymbol   = w3.MustNewFunc("symbol()", "string")
		funcDecimals = w3.MustNewFunc("decimals()", "uint8")
	)

	// fetch token details
	var (
		name     string
		symbol   string
		decimals uint8
		// address common.Address
	)

	_token := pairs.ERC20Token{
		Address: token,
	}

	if err := client.Call(
		eth.CallFunc(token, funcName).Returns(&name),
		eth.CallFunc(token, funcSymbol).Returns(&symbol),
		eth.CallFunc(token, funcDecimals).Returns(&decimals),
	); err != nil {
		fmt.Printf("Failed to fetch token details: %v\n", err)

		// Set the token metadata
		// TODO: Probably a better way to handle this
		_token.Name = name
		_token.Symbol = symbol
		_token.Decimals = decimals

	}

	return _token
}

func pairHandler(c *gin.Context) {

	tokenAAddress := c.DefaultQuery("tokena", "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	tokenBAddress := c.DefaultQuery("tokenb", "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599")

	// // Initialize the pair handler
	// redisUrl := config.RedisURL
	// pairHandler := pairs.NewPairHandler(redisUrl)

	// ctx := context.Background()

	// // TODO: Call the ERC20 token for the metadata

	// // Example usage
	// tokenA := pairs.ERC20Token{Address: tokenAAddress, Symbol: "USDC", Decimals: 6}
	// tokenB := pairs.ERC20Token{Address: tokenBAddress, Symbol: "WBTC", Decimals: 18}

	// pairHandler.AddPair(tokenA, tokenB)

	// // Adding protocol pairs
	// pairHandler.AddProtocolPair(ctx, "UniswapV3_10000", "0xCBFB0745b8489973Bf7b334d54fdBd573Df7eF3c", pairs.TokenPair{Token0: tokenA, Token1: tokenB})
	// pairHandler.AddProtocolPair(ctx, "UniswapV3_30000", "0xCBFB0745b8489973Bf7b334d54fdBd573Df7eF3c", pairs.TokenPair{Token0: tokenA, Token1: tokenB})
	// pairHandler.AddProtocolPair(ctx, "SushiSwap", "0x01", pairs.TokenPair{Token0: tokenA, Token1: tokenB})

	// // Retrieving protocol pairs
	// protocolPairs := pairHandler.GetProtocolPairs(tokenA.Address, tokenB.Address)
	// for _, pp := range protocolPairs {
	// 	fmt.Printf("Protocol: %s, Contract: %s, Pair: %s-%s\n",
	// 		pp.ProtocolName, pp.ContractAddress, pp.Pair.Token0.Symbol, pp.Pair.Token1.Symbol)
	// }

	// // Finding protocols for a specific pair
	// protocolsForAB := pairHandler.FindProtocolsForPair(tokenA.Address, tokenB.Address)
	// fmt.Printf("Protocols supporting %s-%s pair:\n", tokenA.Symbol, tokenB.Symbol)
	// for _, pp := range protocolsForAB {
	// 	fmt.Printf("- %s (Contract: %s)\n", pp.ProtocolName, pp.ContractAddress)
	// }

	token0 := getMetadata(common.HexToAddress(tokenAAddress))
	token1 := getMetadata(common.HexToAddress(tokenBAddress))

	nodeUrl := config.NodeURL

	// do these in parallel
	// pancake.Quote(token0, token1, fee, nodeUrl)

	// $1,000 USDC
	amount := big.NewInt(1000000000)
	quoteResponse := uniswap.Quote(token0.Address, token1.Address, *amount, nodeUrl)

	c.JSON(http.StatusOK, gin.H{
		"result": quoteResponse,
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
