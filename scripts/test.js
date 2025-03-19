const { ethers } = require("ethers");

const erc20Abi = [
    /* ERC20 ABI */
    "function approve(address spender, uint256 amount) external returns (bool)",
    "function decimals() external view returns (uint8)",
    "function name() external view returns (string)",
    "function symbol() external view returns (string)",
];

const token0Address = "0x88b8E2161DEDC77EF4ab7585569D2415a1C1055D";

const provider = new ethers.JsonRpcProvider("https://monad-testnet.g.alchemy.com/v2/CrRBwY8ouIombWO9PolrlWsb0rEjVXaU");
const tokenContract = new ethers.Contract(token0Address, erc20Abi, provider);

async function main() {
    console.log("Testing token address: ", token0Address);
    const tokenDecimals = await tokenContract.decimals();
    const tokenName = await tokenContract.name();
    console.log("Token Decimals: ", tokenDecimals.toString());
    console.log("Token Name: ", tokenName);
}

main();