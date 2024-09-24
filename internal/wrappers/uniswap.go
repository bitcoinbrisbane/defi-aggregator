package wrappers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"

	"github.com/bitcoinbrisbane/defi-aggregator/internal/pairs"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// IERC20ABI is the ABI for the ERC20 interface
const IERC20ABI = `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"type":"function"}]`
const IUNISWAPV3QUOTABI = `[{"constant":true,"inputs":[{"name":"tokenIn","type":"address"},{"name":"tokenOut","type":"address"},{"name":"amountIn","type":"uint256"}],"name":"quote","outputs":[{"name":"amountOut","type":"uint256"}],"type":"function"}]`

// MockQuoter represents a mock contract for getting quotes
type MockQuoter struct{}

// IERC20 is a simplified interface for ERC20 tokens
type IERC20 struct {
	Name  func(opts *bind.CallOpts) (string, error)
}

// GetQuote is a mock method to simulate getting a quote from a smart contract
func (mq *MockQuoter) GetQuote(tokenIn, tokenOut common.Address, amountIn *big.Int) (*big.Int, error) {
	// This is a mock implementation. In a real scenario, this would interact with the blockchain.
	// For simplicity, we're just returning a dummy value based on the input amount.

	// // Connect to an Ethereum node
	// client, err := ethclient.Dial("https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
	// if err != nil {
	// 	log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	// }

	// pairAddress := common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F")


	// // Create a new instance of the ERC20 contract
	// token, err := NewIERC20(tokenAddress, client)
	// if err != nil {
	// 	log.Fatalf("Failed to instantiate ERC20 contract: %v", err)
	// }

	// // Call the name() method
	// name, err := token.Name(&bind.CallOpts{})
	// if err != nil {
	// 	log.Fatalf("Failed to retrieve token name: %v", err)
	// }

	// fmt.Printf("Token name: %s\n", name)

	return new(big.Int).Mul(amountIn, big.NewInt(2)), nil
}

// // NewIERC20 creates a new instance of IERC20, bound to a specific deployed contract
// func NewIERC20(address common.Address, backend bind.ContractBackend) (*IERC20, error) {
// 	contract := bind.NewBoundContract(address, common.HexToHash(IERC20ABI), backend, backend, backend)
// 	return &IERC20{
// 		Name: func(opts *bind.CallOpts) (string, error) {
// 			var out []interface{}
// 			err := contract.Call(opts, &out, "name")
// 			if err != nil {
// 				return "", err
// 			}
// 			return *abi.ConvertType(out[0], new(string)).(*string), nil
// 		},
// 	}, nil
// }

// PairHandlerWrapper wraps pairs.PairHandler to allow method definitions
type PairHandlerWrapper struct {
	pairs.PairHandler
}

// QuoteHandler handles HTTP requests for quotes
func (ph *PairHandlerWrapper) QuoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	tokenInAddress := query.Get("tokenIn")
	tokenOutAddress := query.Get("tokenOut")
	protocolName := query.Get("protocol")
	amountInStr := query.Get("amountIn")

	// Validate input
	if tokenInAddress == "" || tokenOutAddress == "" || protocolName == "" || amountInStr == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Find the protocol pair
	protocolPairs := ph.FindProtocolsForPair(tokenInAddress, tokenOutAddress)
	var targetPair pairs.ProtocolPair
	for _, pp := range protocolPairs {
		if pp.ProtocolName == protocolName {
			targetPair = pp
			break
		}
	}

	if targetPair.ContractAddress == "" {
		http.Error(w, "Protocol not found for the given token pair", http.StatusNotFound)
		return
	}

	// Parse amountIn
	amountIn, ok := new(big.Int).SetString(amountInStr, 10)
	if !ok {
		http.Error(w, "Invalid amountIn value", http.StatusBadRequest)
		return
	}

	// Create a mock quoter (in a real scenario, this would connect to the blockchain)
	quoter := &MockQuoter{}

	// Get the quote
	quoteAmount, err := quoter.GetQuote(common.HexToAddress(tokenInAddress), common.HexToAddress(tokenOutAddress), amountIn)
	if err != nil {
		http.Error(w, "Error getting quote: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]string{
		"tokenIn":    tokenInAddress,
		"tokenOut":   tokenOutAddress,
		"protocol":   protocolName,
		"amountIn":   amountIn.String(),
		"quoteAmount": quoteAmount.String(),
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}