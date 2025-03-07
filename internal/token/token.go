package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// TokenRequest defines the structure for the token post request
type TokenRequest struct {
	Address string `json:"address" binding:"required"`
}

// TokenMetadata represents the ERC20 token metadata
type TokenMetadata struct {
	Address  string `json:"address"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals"`
}

// API Key middleware for simple authentication
func apiKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("x-api-key")
		expectedApiKey := getEnvWithDefault("API_KEY", "default-api-key")
		
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing API key",
			})
			return
		}
		
		if apiKey != expectedApiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			return
		}
		
		c.Next()
	}
}

// connectRedis establishes a connection to Redis
func connectRedis() (*redis.Client, error) {
	redisURL := getEnvWithDefault("REDIS_URL", "localhost:6379")
	redisPassword := getEnvWithDefault("REDIS_PASSWORD", "")
	redisDB := 0
	
	client := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: redisPassword,
		DB:       redisDB,
	})
	
	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}
	
	return client, nil
}

// saveTokenMetadata saves the token metadata to Redis
func saveTokenMetadata(ctx context.Context, client *redis.Client, metadata TokenMetadata) error {
	// Use token address as key
	key := "token:" + metadata.Address
	
	// Marshal the metadata to JSON
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal token metadata: %v", err)
	}
	
	// Save to Redis
	err = client.Set(ctx, key, jsonData, 0).Err() // 0 = no expiration
	if err != nil {
		return fmt.Errorf("failed to save token metadata to Redis: %v", err)
	}
	
	return nil
}

// getTokenMetadataFromRedis retrieves token metadata from Redis
func getTokenMetadataFromRedis(ctx context.Context, client *redis.Client, address string) (*TokenMetadata, error) {
	key := "token:" + address
	
	// Get from Redis
	data, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Key does not exist
	} else if err != nil {
		return nil, fmt.Errorf("failed to get token metadata from Redis: %v", err)
	}
	
	// Unmarshal the data
	var metadata TokenMetadata
	err = json.Unmarshal([]byte(data), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token metadata: %v", err)
	}
	
	return &metadata, nil
}

// tokenPostHandler handles the /token POST endpoint
func tokenPostHandler(c *gin.Context) {
	// Use defer to recover from panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in tokenPostHandler: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error occurred",
			})
		}
	}()
	
	// Parse request body
	var request TokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}
	
	// Validate Ethereum address
	if !common.IsHexAddress(request.Address) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Ethereum address format",
		})
		return
	}
	
	tokenAddress := common.HexToAddress(request.Address)
	
	// Connect to Redis
	redisClient, err := connectRedis()
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect to database",
		})
		return
	}
	defer redisClient.Close()
	
	// Check if we already have this token in Redis
	ctx := context.Background()
	existingMetadata, err := getTokenMetadataFromRedis(ctx, redisClient, tokenAddress.String())
	if err != nil {
		log.Printf("Error checking Redis for token: %v", err)
	}
	
	if existingMetadata != nil {
		// Return existing metadata if found
		c.JSON(http.StatusOK, gin.H{
			"message": "Token metadata retrieved from cache",
			"token":   existingMetadata,
		})
		return
	}
	
	// Fetch token metadata from the blockchain
	name, symbol, decimals, err := uniswap.GetTokenMetadata(tokenAddress, config.NodeURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to fetch token metadata: %v", err),
		})
		return
	}
	
	// Create metadata object
	metadata := TokenMetadata{
		Address:  tokenAddress.String(),
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
	}
	
	// Save to Redis
	err = saveTokenMetadata(ctx, redisClient, metadata)
	if err != nil {
		log.Printf("Failed to save token metadata to Redis: %v", err)
		// Continue even if saving fails
	}
	
	// Return the metadata
	c.JSON(http.StatusOK, gin.H{
		"message": "Token metadata fetched and saved",
		"token":   metadata,
	})
}