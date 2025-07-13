// SPDX-License-Identifier: MIT
pragma solidity ^0.8.27;

import "@uniswap/v3-periphery/contracts/interfaces/ISwapRouter.sol";
import { IUniswapV2Router } from "Interfaces.sol";
import { IRC20 } from "Interfaces.sol";

contract V2Adaptor is IV3 {
    // This contract is a placeholder for the V2 Adaptor
    // It implements the ISwapRouter interface but does not provide any functionality
    // The actual implementation would be done in a separate contract

    // The address of the Uniswap V2 router
    address public immutable swapRouter;
    address public immutable quoter;
    address public immutable factory;
    string public name;

    constructor(address _swapRouter, address _quoter, address _factory, string memory _name) {
        swapRouter = _swapRouter;
        quoter = _quoter;
        factory = _factory;
        name = _name;
    }

    function exactInputSingle(ExactInputSingleParams calldata params) external payable override returns (uint256 amountOut) {
        // exactInputSingle
        IUniswapV2Router router = IUniswapV2Router(swapRouter);
        router.approve(params.tokenIn, params.amountIn);

        uint256 balanceBefore = IERC20(params.tokenOut).balanceOf(address(this));

        IERC20(params.tokenIn).transferFrom(msg.sender, address(this), params.amountIn);
        // Call the Uniswap V2 router to swap tokens
        router.swapExactTokensForTokens(
            params.amountIn,
            params.amountOutMinimum,
            [params.tokenIn, params.tokenOut],
            params.recipient,
            params.deadline
        );

        uint256 balanceAfter = IERC20(params.tokenOut).balanceOf(address(this));
        amountOut = balanceAfter - balanceBefore;
        require(amountOut >= params.amountOutMinimum, "Insufficient output amount");

        // Transfer the output tokens to the recipient
        IERC20(params.tokenOut).transfer(params.recipient, amountOut);
    }

    function exactInput(ExactInputParams calldata params) external payable override returns (uint256 amountOut) {
        IUniswapV2Pair pair = IUniswapV2Pair(factory);
        
        revert("Not implemented");
    }

    function exactOutputSingle(ExactOutputSingleParams calldata params) external payable override returns (uint256 amountIn) {
        revert("Not implemented");
    }

    function exactOutput(ExactOutputParams calldata params) external payable override returns (uint256 amountIn) {
        revert("Not implemented");
    }
}