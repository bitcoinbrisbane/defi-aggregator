package tokens

import (
	"fmt"
	"log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

// Function signatures for Uniswap interactions
var (
	funcName                  = w3.MustNewFunc("name()", "string")
	funcSymbol                = w3.MustNewFunc("symbol()", "string")
	funcDecimals              = w3.MustNewFunc("decimals()", "uint8")
)

// GetTokenMetadata gets token metadata (name, symbol, decimals)
func GetTokenMetadata(tokenAddress common.Address, nodeURL string) (string, string, uint8, error) {
	// Create a client
	client := w3.MustDial(nodeURL)
	defer client.Close()
	
	// Get token metadata
	var (
		name     string
		symbol   string
		decimals uint8
	)
	
	err := client.Call(
		eth.CallFunc(tokenAddress, funcName).Returns(&name),
		eth.CallFunc(tokenAddress, funcSymbol).Returns(&symbol),
		eth.CallFunc(tokenAddress, funcDecimals).Returns(&decimals),
	)
	
	if err != nil {
		log.Printf("Error getting token metadata: %v", err)
		return "", "", 0, fmt.Errorf("failed to get token metadata: %v", err)
	}
	
	return name, symbol, decimals, nil
}