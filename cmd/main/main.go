package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/bitcoinbrisbane/defi-aggregator/internal/aggregator"
	"github.com/bitcoinbrisbane/defi-aggregator/internal/clients/uniswap"
	"github.com/bitcoinbrisbane/defi-aggregator/internal/protocols"

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
		NodeURL:  getEnvWithDefault("NODE_URL", "https://eth-mainnet.g.alchemy.com/v2/fmiJslJk8E60f0Ni9QLq5nsnjm-lUzn1"),
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

	// Create aggregator service
	aggregatorService := aggregator.NewService(config.NodeURL)

	// Add routes
	router.GET("/", helloHandler)
	router.GET("/pairs", func(c *gin.Context) { pairHandler(c, aggregatorService) })
	router.GET("/protocols", protocolsHandler)

	// Start the server
	router.Run(":" + config.Port)
}

func helloHandler(c *gin.Context) {
	response := Response{
		Message: "Hello, World!",
	}
	c.JSON(http.StatusOK, response)
}

func protocolsHandler(c *gin.Context) {
	protocols := protocols.GetSupportedProtocols()
	c.JSON(http.StatusOK, gin.H{
		"protocols": protocols,
	})
}

func pairHandler(c *gin.Context, aggregatorService *aggregator.Service) {
	// Use defer to recover from panics in this handler
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in pairHandler: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error occurred",
			})
		}
	}()

	// Get token addresses from query parameters
	tokenAAddress := c.Query("tokena")
	if tokenAAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameter: tokena",
		})
		return
	}
	
	tokenBAddress := c.Query("tokenb")
	if tokenBAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameter: tokenb",
		})
		return
	}
	
	// Get amount from query parameter
	amountStr := c.DefaultQuery("amount", "10000")
	
	// Convert amount string to big.Int
	amount, success := new(big.Int).SetString(amountStr, 10)
	if !success {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid amount parameter",
		})
		return
	}
	
	// Get token metadata
	tokenA := common.HexToAddress(tokenAAddress)
	tokenB := common.HexToAddress(tokenBAddress)
	
	// Get token metadata
	tokenAName, tokenASymbol, tokenADecimals, err := uniswap.GetTokenMetadata(tokenA, config.NodeURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to get tokenA metadata: %v", err),
		})
		return
	}
	
	tokenBName, tokenBSymbol, tokenBDecimals, err := uniswap.GetTokenMetadata(tokenB, config.NodeURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to get tokenB metadata: %v", err),
		})
		return
	}
	
	// Log token information
	log.Printf("TokenA: %s (%s) - %d decimals", tokenAName, tokenASymbol, tokenADecimals)
	log.Printf("TokenB: %s (%s) - %d decimals", tokenBName, tokenBSymbol, tokenBDecimals)
	
	// Option to return all routes or just the best
	showAllRoutes := c.DefaultQuery("all", "false") == "true"
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	
	// Find the best route across all protocols
	result, err := aggregatorService.FindBestRoute(
		ctx,
		tokenA,
		tokenB,
		amount,
		tokenADecimals,
		tokenBDecimals,
		tokenASymbol,
		tokenBSymbol,
	)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to find routes: %v", err),
		})
		return
	}
	
	// Return the result
	if showAllRoutes {
		// Return all routes
		c.JSON(http.StatusOK, gin.H{
			"bestRoute": result.BestRoute,
			"allRoutes": result.AllRoutes,
		})
	} else {
		// Return only the best route
		c.JSON(http.StatusOK, gin.H{
			"result": result.BestRoute,
		})
	}
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

// func pairHandler(c *gin.Context) {
// 	// Use defer to recover from panics in this handler
// 	defer func() {
// 		if r := recover(); r != nil {
// 			log.Printf("Recovered from panic in pairHandler: %v", r)
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"error": "Internal server error occurred",
// 			})
// 		}
// 	}()

// 	tokenAAddress := c.Query("tokena")
// 	if tokenAAddress == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Missing required parameter: tokena",
// 		})
// 		return
// 	}
	
// 	tokenBAddress := c.Query("tokenb")
// 	if tokenBAddress == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Missing required parameter: tokenb",
// 		})
// 		return
// 	}
	
// 	// Add the amount parameter from query string with default value of 10000
// 	amountStr := c.DefaultQuery("amount", "10000")
	
// 	// Convert the amount string to a big.Int
// 	amount, success := new(big.Int).SetString(amountStr, 10)
// 	if !success {
// 		// Handle invalid amount parameter
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid amount parameter",
// 		})
// 		return
// 	}
	
// 	// Get option to return all quotes or just best
// 	bestOnly := c.DefaultQuery("best", "true") == "true"

// 	token0 := getMetadata(common.HexToAddress(tokenAAddress))
// 	token1 := getMetadata(common.HexToAddress(tokenBAddress))

// 	nodeUrl := config.NodeURL

// 	quotes := uniswap.Quote(token0.Address, token1.Address, *amount, nodeUrl)
	
// 	if bestOnly && len(quotes) > 0 {
// 		// Get the best quote only
// 		bestQuote := uniswap.GetBestQuote(quotes)
// 		c.JSON(http.StatusOK, gin.H{
// 			"result": bestQuote,
// 		})
// 	} else {
// 		// Return all quotes
// 		c.JSON(http.StatusOK, gin.H{
// 			"result": quotes,
// 		})
// 	}
// }

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
