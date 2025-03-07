package protocols

import (
	"github.com/ethereum/go-ethereum/common"
)

// ProtocolConfig represents a DEX protocol configuration
type ProtocolConfig struct {
	Name           string         `json:"name"`
	FactoryAddress common.Address `json:"factoryAddress"`
	RouterAddress  common.Address `json:"routerAddress"`
	// Some protocols might need additional parameters
	FeeTiers      []uint64 `json:"feeTiers"`      // Available fee tiers (e.g., 500, 3000, 10000 for Uniswap V3)
	IsUniswapFork bool     `json:"isUniswapFork"` // Is this a Uniswap-compatible fork
}

// Protocols is a map of protocol configurations
var Protocols = map[string]ProtocolConfig{
	"uniswapv3": {
		Name:           "Uniswap V3",
		FactoryAddress: common.HexToAddress("0x961235a9020b05c44df1026d956d1f4d78014276"),
		RouterAddress:  common.HexToAddress("0x4c4eabd5fb1d1a7234a48692551eaecff8194ca7"), // monad
		FeeTiers:       []uint64{500, 3000, 10000},
		IsUniswapFork:  true,
	},
	"sushiswapv3": {
		Name:           "Sushiswap V3",
		FactoryAddress: common.HexToAddress("0xBACeb8eC6b9355Dfc0269C18bac9d6E2Bdc29C4F"),
		RouterAddress:  common.HexToAddress("0x8A21F6768C1f8075791D08546Bd61770d3F8a48F"),
		FeeTiers:       []uint64{100, 500, 3000, 10000},
		IsUniswapFork:  true,
	},
	"pancakeswapv3": {
		Name:           "PancakeSwap V3",
		FactoryAddress: common.HexToAddress("0x0BFbCF9fa4f9C56B0F40a671Ad40E0805A091865"),
		RouterAddress:  common.HexToAddress("0x13f4EA83D0bd40E75C8222255bc855a974568Dd4"),
		FeeTiers:       []uint64{100, 500, 2500, 10000},
		IsUniswapFork:  true,
	},
	// "kuru": {
	// 	Name:           "Kuru",
	// 	FactoryAddress: common.HexToAddress("0xc816865f172d640d93712C68a7E1F83F3fA63235"),
	// 	RouterAddress:  common.HexToAddress("0xc816865f172d640d93712C68a7E1F83F3fA63235"),
	// 	FeeTiers:       []uint64{100, 500, 2500, 10000},
	// 	IsUniswapFork:  true,
	// },
	// Add more protocols as needed
}

// GetSupportedProtocols returns a list of all supported protocol names
func GetSupportedProtocols() []string {
	protocols := make([]string, 0, len(Protocols))
	for key := range Protocols {
		protocols = append(protocols, key)
	}
	return protocols
}

// GetProtocolByName returns a protocol configuration by name
func GetProtocolByName(name string) (ProtocolConfig, bool) {
	protocol, exists := Protocols[name]
	return protocol, exists
}

// GetUniswapForks returns all Uniswap-compatible forks
func GetUniswapForks() []ProtocolConfig {
	forks := make([]ProtocolConfig, 0)
	for _, protocol := range Protocols {
		if protocol.IsUniswapFork {
			forks = append(forks, protocol)
		}
	}
	return forks
}