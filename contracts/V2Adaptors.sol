// SPDX-License-Identifier: MIT
pragma solidity ^0.8.27;

import "@uniswap/v3-periphery/contracts/interfaces/ISwapRouter.sol";

contract V2Adaptor is ISwapRouter {
    // This contract is a placeholder for the V2 Adaptor
    // It implements the ISwapRouter interface but does not provide any functionality
    // The actual implementation would be done in a separate contract

    // The address of the Uniswap V2 router
    address private immutable swapRouter;

    constructor(address _swapRouter) {
        swapRouter = _swapRouter;
    }

    function exactInputSingle(ExactInputSingleParams calldata params) external payable override returns (uint256 amountOut) {
        revert("Not implemented");
    }

    function exactInput(ExactInputParams calldata params) external payable override returns (uint256 amountOut) {
        revert("Not implemented");
    }

    function exactOutputSingle(ExactOutputSingleParams calldata params) external payable override returns (uint256 amountIn) {
        revert("Not implemented");
    }

    function exactOutput(ExactOutputParams calldata params) external payable override returns (uint256 amountIn) {
        revert("Not implemented");
    }
}