# defi-aggregator

## DeFi Aggregator Contracts

# DeFi Aggregator Smart Contract

A Solidity smart contract that aggregates quotes from multiple Uniswap V3-compatible DEXes and automatically finds the most favorable swap routes for users.

## 🚀 Features

- **Multi-DEX Support**: Aggregates quotes from multiple Uniswap V3-compatible protocols
- **Best Route Finding**: Automatically finds the most favorable swap route across all enabled DEXes
- **Multiple Fee Tiers**: Supports different fee tiers (0.05%, 0.3%, 1%) for optimal pricing
- **Protocol Fee**: Built-in fee mechanism (2%) for protocol sustainability
- **Owner Controls**: Admin functions for DEX management and fee claiming
- **Gas Optimized**: Efficient quote comparison and route selection

## 📋 Contract Overview

The `Aggregator` contract maintains a registry of DEX protocols and their quoter/router addresses, allowing users to:
1. Get quotes from multiple DEXes simultaneously
2. Execute swaps through the best available route
3. Benefit from competitive pricing across protocols

## 🔧 Core Functions

### Administrative Functions

#### `addDex(string memory _name, address _quoterAddress, address _routerAddress)`
Adds a new DEX to the aggregator registry.

**Parameters:**
- `_name`: Human-readable name of the DEX (e.g., "Uniswap V3", "PancakeSwap V3")
- `_quoterAddress`: Address of the DEX's Quoter contract
- `_routerAddress`: Address of the DEX's Router contract

**Example:**
```solidity
// Add Uniswap V3 to the registry
aggregator.addDex(
    "Uniswap V3",
    0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6, // Uniswap V3 Quoter
    0xE592427A0AEce92De3Edee1F18E0157C05861564  // Uniswap V3 Router
);
```

#### `updateDex(uint256 _index, string memory _name, address _quoterAddress, bool _enabled)`
Updates an existing DEX in the registry.

**Parameters:**
- `_index`: Index of the DEX in the registry
- `_name`: Updated name for the DEX
- `_quoterAddress`: Updated quoter address
- `_enabled`: Whether the DEX should be enabled for quotes

**Example:**
```solidity
// Update DEX at index 0
aggregator.updateDex(0, "Uniswap V3 Updated", quoterAddress, true);
```

#### `toggleDexStatus(uint256 _index)`
Toggles the enabled/disabled status of a DEX.

**Example:**
```solidity
// Disable DEX at index 1
aggregator.toggleDexStatus(1);
```

### Query Functions

#### `getDexCount() → uint256`
Returns the total number of DEXes in the registry.

**Example:**
```solidity
uint256 totalDexes = aggregator.getDexCount();
// Returns: 3 (if 3 DEXes are registered)
```

#### `getDexInfo(uint256 _index) → DexInfo`
Returns detailed information about a specific DEX.

**Returns:**
- `name`: DEX name
- `quoterAddress`: Quoter contract address
- `routerAddress`: Router contract address  
- `enabled`: Whether the DEX is enabled

**Example:**
```solidity
DexInfo memory dexInfo = aggregator.getDexInfo(0);
// Returns: ("Uniswap V3", 0xb27..., 0xE59..., true)
```

### Quote Functions

#### `getBestQuote(address _tokenIn, address _tokenOut, uint256 _amountIn)`
Gets the best quote across all enabled DEXes without executing a swap.

**Parameters:**
- `_tokenIn`: Address of the input token
- `_tokenOut`: Address of the output token  
- `_amountIn`: Amount of input token to swap

**Returns:**
- `dexName`: Name of the DEX offering the best quote
- `quoterAddress`: Address of the best DEX's quoter
- `amountOut`: Expected output amount
- `bestFee`: Optimal fee tier for the swap

**Example:**
```solidity
// Get best quote for swapping 1000 USDC to WETH
(string memory dexName, 
 address quoterAddress, 
 uint256 amountOut, 
 uint24 bestFee) = aggregator.getBestQuote(
    0xA0b86a33E6441b8e49F1B7Ec72e1cEBe3F6a6C8e, // USDC
    0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2, // WETH
    1000 * 10**6 // 1000 USDC
);
// Returns: ("Uniswap V3", 0xb27..., 285000000000000000, 3000)
```

#### `findBestRoute(address _tokenIn, address _tokenOut, uint256 _amountIn)`
Internal function that finds the best route across all DEXes and fee tiers.

**Returns:**
- `bestDexIndex`: Index of the optimal DEX
- `bestQuoterAddress`: Address of the optimal quoter
- `bestAmountOut`: Best output amount found
- `bestFee`: Optimal fee tier

### Swap Execution

#### `executeSwap(address _tokenIn, address _tokenOut, uint256 _amountIn, uint256 _amountOutMinimum, address _recipient)`
Executes a swap through the best available route.

**Parameters:**
- `_tokenIn`: Input token address
- `_tokenOut`: Output token address
- `_amountIn`: Amount of input token to swap
- `_amountOutMinimum`: Minimum acceptable output amount (slippage protection)
- `_recipient`: Address to receive the output tokens

**Returns:**
- `amountOut`: Actual amount of output tokens received

**Example:**
```solidity
// First approve the aggregator to spend your tokens
IERC20(usdcAddress).approve(aggregatorAddress, 1000 * 10**6);

// Execute swap: 1000 USDC → WETH with 1% slippage tolerance
uint256 amountOut = aggregator.executeSwap(
    0xA0b86a33E6441b8e49F1B7Ec72e1cEBe3F6a6C8e, // USDC
    0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2, // WETH  
    1000 * 10**6,                                 // 1000 USDC
    280000000000000000,                           // Min 0.28 WETH (1% slippage)
    msg.sender                                    // Send WETH to caller
);
```

### Fee Management

#### `claimFees(address token)`
Allows the contract owner to claim accumulated protocol fees.

**Parameters:**
- `token`: Address of the token to claim fees for

**Example:**
```solidity
// Claim accumulated USDC fees
aggregator.claimFees(0xA0b86a33E6441b8e49F1B7Ec72e1cEBe3F6a6C8e);
```

## 📊 Events

The contract emits several events for tracking and monitoring:

```solidity
event BestRouteFound(string dexName, address dex, uint256 amountOut);
event DexAdded(string name, address indexed quoterAddress, uint256 index);
event DexUpdated(uint256 index, string name, address indexed quoterAddress, bool enabled);
event FeesClaimed(address indexed token, uint256 amount);
```

## 💡 Usage Examples

### Complete Swap Flow
```solidity
// 1. Check available DEXes
uint256 dexCount = aggregator.getDexCount();

// 2. Get best quote
(string memory bestDex, , uint256 expectedOut, ) = aggregator.getBestQuote(
    tokenA, tokenB, amountIn
);

// 3. Approve tokens
IERC20(tokenA).approve(aggregatorAddress, amountIn);

// 4. Execute swap with slippage protection
uint256 minOut = expectedOut * 99 / 100; // 1% slippage
uint256 actualOut = aggregator.executeSwap(
    tokenA, tokenB, amountIn, minOut, recipient
);
```

### Adding Multiple DEXes
```solidity
// Add Uniswap V3
aggregator.addDex("Uniswap V3", uniQuoter, uniRouter);

// Add PancakeSwap V3  
aggregator.addDex("PancakeSwap V3", cakeQuoter, cakeRouter);

// Add SushiSwap V3
aggregator.addDex("SushiSwap V3", sushiQuoter, sushiRouter);
```

## ⚙️ Configuration

- **Protocol Fee**: 2% (20 basis points) taken from input amount
- **Supported Fee Tiers**: 0.05% (500), 0.3% (3000), 1% (10000)
- **Swap Deadline**: 15 minutes from execution
- **Owner**: Deployer of the contract (transferable)

## 🔒 Security Features

- **Access Control**: Admin functions restricted to contract owner
- **Input Validation**: Comprehensive checks on addresses and amounts
- **Slippage Protection**: Minimum output amount enforcement
- **Error Handling**: Graceful handling of failed quotes
- **Reentrancy Safe**: No external calls that could enable reentrancy

## 🚀 Deployment

1. Deploy the contract (no constructor parameters needed)
2. Add DEX protocols using `addDex()`
3. Test with small amounts first
4. Monitor events for successful operations


## Defi Aggregator API

This project is a decentralized finance (DeFi) aggregator that allows users to view and compare DeFi projects and find the best trading route.

## Requirements

## Install Go

```bash

```

## Copy system 

## Starting

Start the Redis server:

```bash
docker compose up
```

## 📝 License

This project is licensed under the MIT License.