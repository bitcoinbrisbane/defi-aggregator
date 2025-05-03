// SPDX-License-Identifier: MIT
pragma solidity ^0.8.27;

import "@uniswap/v3-periphery/contracts/interfaces/ISwapRouter.sol";
import "@uniswap/v3-periphery/contracts/interfaces/IQuoter.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title Aggregator
 * @dev A smart contract that aggregates quotes from multiple Uniswap V3 protocols 
 * and finds the most favorable swap route.
 */
contract Aggregator is Ownable {
    // Struct to hold information about each DEX
    struct DexInfo {
        string name;
        address quoterAddress;
        bool enabled;
    }

    // private uint256 constant MAX_FEE = 3000; // 0.3% fee
    // private uint256 constant MIN_FEE = 500; // 0.05% fee
    // private uint256 constant DEFAULT_FEE = 1000; // 0.1% fee
    uint256 private fee;
    uint24[] public fees = [500, 3000, 10000]; // Fee tiers for Uniswap V3

    // Array to store all DEX information
    DexInfo[] public dexRegistry;

    constructor() Ownable(msg.sender) {
        fee = 2000;
    }

    /**
     * @dev Add a new DEX to the registry
     * @param _name Name of the DEX
     * @param _quoterAddress Address of the DEX's Quoter contract
     */
    function addDex(string memory _name, address _quoterAddress) external onlyOwner {
        require(_quoterAddress != address(0), "Invalid quoter address");
        
        uint256 index = dexRegistry.length;

        dexRegistry.push(DexInfo({
            name: _name,
            quoterAddress: _quoterAddress,
            enabled: true
        }));
        
        emit DexAdded(_name, _quoterAddress, index);
    }

    /**
     * @dev Update a DEX in the registry
     * @param _index Index of the DEX in the registry
     * @param _name New name for the DEX
     * @param _quoterAddress New address for the DEX's Quoter contract
     * @param _enabled Whether the DEX should be enabled or disabled
     */
    function updateDex(
        uint256 _index, 
        string memory _name, 
        address _quoterAddress, 
        bool _enabled
    ) external onlyOwner {
        require(_index < dexRegistry.length, "Index out of bounds");
        require(_quoterAddress != address(0), "Invalid quoter address");
        
        DexInfo storage dex = dexRegistry[_index];
        dex.name = _name;
        dex.quoterAddress = _quoterAddress;
        dex.enabled = _enabled;
        
        emit DexUpdated(_index, _name, _quoterAddress, _enabled);
    }

    /**
     * @dev Toggle the enabled status of a DEX
     * @param _index Index of the DEX in the registry
     */
    function toggleDexStatus(uint256 _index) external onlyOwner {
        require(_index < dexRegistry.length, "Index out of bounds");
        
        dexRegistry[_index].enabled = !dexRegistry[_index].enabled;
        
        emit DexUpdated(
            _index, 
            dexRegistry[_index].name, 
            dexRegistry[_index].quoterAddress, 
            dexRegistry[_index].enabled
        );
    }

    /**
     * @dev Get the number of DEXes in the registry
     * @return Number of DEXes
     */
    function getDexCount() external view returns (uint256) {
        return dexRegistry.length;
    }

    /**
     * @dev Get information about a specific DEX
     * @param _index Index of the DEX in the registry
     * @return DEX information
     */
    function getDexInfo(uint256 _index) external view returns (DexInfo memory) {
        require(_index < dexRegistry.length, "Index out of bounds");
        return dexRegistry[_index];
    }

    /**
     * @dev Find the best swap route across all enabled DEXes
     * @param _tokenIn Address of the input token
     * @param _tokenOut Address of the output token
     * @param _amountIn Amount of input token to swap
     * @param _fee Fee tier to use for the swap (3000 = 0.3%)
     * @return bestDexIndex Index of the best DEX
     * @return bestQuoterAddress Address of the best DEX's Quoter contract
     * @return bestAmountOut Amount of output token from the best DEX
     */
    function findBestRoute(
        address _tokenIn,
        address _tokenOut,
        uint256 _amountIn
    ) public view returns (
        uint256 bestDexIndex,
        address bestQuoterAddress,
        uint256 bestAmountOut
    ) {
        require(_tokenIn != address(0), "Invalid tokenIn");
        require(_tokenOut != address(0), "Invalid tokenOut");
        require(_amountIn > 0, "Amount must be > 0");
        
        bestAmountOut = 0;
        
        for (uint256 i = 0; i < dexRegistry.length; i++) {
            if (!dexRegistry[i].enabled) {
                continue;
            }
            
            for (uint256 j = 0; j < fees.length; j++) {
                IQuoter quoter = IQuoter(dexRegistry[i].quoterAddress);
                uint24 _fee = fees[j];
                
                try quoter.quoteExactInputSingle(
                    _tokenIn,
                    _tokenOut,
                    _fee,
                    _amountIn,
                    0 // sqrtPriceLimitX96 - set to 0 for no price limit
                ) returns (uint256 amountOut) {
                    if (amountOut > bestAmountOut) {
                        bestAmountOut = amountOut;
                        bestDexIndex = i;
                        bestQuoterAddress = dexRegistry[i].quoterAddress;
                    }
                } catch {
                    // If the quote fails, just continue to the next DEX
                    continue;
                }
            }
        }
        
        require(bestAmountOut > 0, "No valid route found");
    }

    /**
     * @dev Get a quote for the best swap route
     * @param _tokenIn Address of the input token
     * @param _tokenOut Address of the output token
     * @param _amountIn Amount of input token to swap
     * @param _fee Fee tier to use for the swap
     * @return dexName Name of the best DEX
     * @return quoterAddress Address of the best DEX's Quoter contract
     * @return amountOut Amount of output token from the best DEX
     */
    function getBestQuote(
        address _tokenIn,
        address _tokenOut,
        uint256 _amountIn,
        uint24 _fee
    ) external view returns (
        string memory dexName,
        address quoterAddress,
        uint256 amountOut
    ) {
        (uint256 bestDexIndex, address bestQuoterAddress, uint256 bestAmountOut) = 
            findBestRoute(_tokenIn, _tokenOut, _amountIn, _fee);
        
        dexName = dexRegistry[bestDexIndex].name;
        quoterAddress = bestQuoterAddress;
        amountOut = bestAmountOut;
    }

    /**
     * @dev Execute a swap through the best route
     * @param _tokenIn Address of the input token
     * @param _tokenOut Address of the output token
     * @param _amountIn Amount of input token to swap
     * @param _fee Fee tier to use for the swap
     * @param _amountOutMinimum Minimum amount of output token to receive
     * @param _recipient Address to receive the output tokens
     * @return amountOut Amount of output token received
     */
    function executeSwap(
        address _tokenIn,
        address _tokenOut,
        uint256 _amountIn,
        uint24 _fee,
        uint256 _amountOutMinimum,
        address _recipient
    ) external returns (uint256 amountOut) {
        require(_recipient != address(0), "Invalid recipient");
        
        (uint256 bestDexIndex, address bestQuoterAddress, uint256 bestAmountOut) = 
            findBestRoute(_tokenIn, _tokenOut, _amountIn, _fee);
        
        require(bestAmountOut >= _amountOutMinimum, "Insufficient output amount");
        
        // We need to transfer the input tokens from the user to this contract
        IERC20(_tokenIn).transferFrom(msg.sender, address(this), _amountIn);
        
        // Approve the router to spend the input tokens
        IERC20(_tokenIn).approve(bestQuoterAddress, _amountIn);
        
        address routerAddress = getRouterFromQuoter(bestQuoterAddress);
        ISwapRouter router = ISwapRouter(routerAddress);
        
        // Execute the swap
        ISwapRouter.ExactInputSingleParams memory params = ISwapRouter.ExactInputSingleParams({
            tokenIn: _tokenIn,
            tokenOut: _tokenOut,
            fee: _fee,
            recipient: _recipient,
            deadline: block.timestamp + 15 minutes,
            amountIn: _amountIn,
            amountOutMinimum: _amountOutMinimum,
            sqrtPriceLimitX96: 0
        });
        
        amountOut = router.exactInputSingle(params);
        
        emit BestRouteFound(
            dexRegistry[bestDexIndex].name,
            bestQuoterAddress,
            amountOut
        );
    }
    
    /**
     * @dev Get the router address from a quoter address
     * @param _quoterAddress Address of the quoter
     * @return Address of the corresponding router
     * @notice This is a placeholder function - in a real implementation, you would need to
     * either store router addresses alongside quoter addresses or have a way to derive one from the other
     */
    function getRouterFromQuoter(address _quoterAddress) internal pure returns (address) {
        // This is a placeholder - in reality, you would need to implement this properly
        // For example, you might have a mapping from quoter addresses to router addresses
        return address(0); // Replace with actual implementation
    }

    // Events
    event DexAdded(string name, address quoterAddress, uint256 index);
    event DexUpdated(uint256 index, string name, address quoterAddress, bool enabled);
    event BestRouteFound(string dexName, address dex, uint256 amountOut);
}