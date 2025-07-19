// SPDX-License-Identifier: MIT
pragma solidity ^0.8.27;

import { IV3Router, IV3Quoter, IUniswapV2Router } from "Interfaces.sol";
import { IRC20 } from "Interfaces.sol";

contract V2Adaptor is IV3Router, IV3Quoter {
    // This contract is a placeholder for the V2 Adaptor
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

    function quoteExactInputSingle(
        address tokenIn,
        address tokenOut,
        uint24 fee,
        uint256 amountIn
    ) external returns (uint256 amountOut) {
        // Call the Uniswap V2 quoter to get the output amount for a given input amount
        require(tokenIn != address(0) && tokenOut != address(0), "Invalid token addresses");
        require(amountIn > 0, "Amount in must be greater than zero");
    }
}