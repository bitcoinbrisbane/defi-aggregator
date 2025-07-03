const { ethers } = require("ethers");

// Contract ABI - only including the functions we need
const aggregatorABI = [
	{
		inputs: [
			{ internalType: "address", name: "_tokenIn", type: "address" },
			{ internalType: "address", name: "_tokenOut", type: "address" },
			{ internalType: "uint256", name: "_amountIn", type: "uint256" },
		],
		name: "getBestQuote",
		outputs: [
			{ internalType: "string", name: "dexName", type: "string" },
			{ internalType: "address", name: "quoterAddress", type: "address" },
			{ internalType: "uint256", name: "amountOut", type: "uint256" },
			{ internalType: "uint24", name: "bestFee", type: "uint24" },
		],
		stateMutability: "nonpayable",
		type: "function",
	},
];

async function getBestQuote() {
	try {
		// Configuration
		const config = {
			// Replace with your RPC URL (Infura, Alchemy, or local node)
			rpcUrl: "https://testnet-rpc.monad.xyz",
			// Replace with your contract address
			contractAddress: "0xEd7C8b67CBE408a04D3eaba163e24f844834300B",
			// Example token addresses (replace with actual addresses)
			tokenIn: "0xaEef2f6B429Cb59C9B2D7bB2141ADa993E8571c3", // gmon
			tokenOut: "0x760AfE86e5de5fa0Ee542fc7B7B713e1c5425701", // wmon
			// Amount in wei (1 USDC = 1e6 for USDC with 6 decimals)
			amountIn: ethers.parseUnits("1", 18),
		};

		// Create provider
		const provider = new ethers.JsonRpcProvider(config.rpcUrl);

		// Create contract instance
		const contract = new ethers.Contract(config.contractAddress, aggregatorABI, provider);

		console.log("Getting best quote...");
		console.log("Token In:", config.tokenIn);
		console.log("Token Out:", config.tokenOut);
		console.log("Amount In:", ethers.formatUnits(config.amountIn, 6), "tokens");

		// Call getBestQuote function
		// Since this function modifies state, we use staticCall to simulate without sending a transaction
		const result = await contract.getBestQuote.staticCall(config.tokenIn, config.tokenOut, config.amountIn);

		// Parse results
		const [dexName, quoterAddress, amountOut, bestFee] = result;

		console.log("\n=== Best Quote Results ===");
		console.log("DEX Name:", dexName);
		console.log("Quoter Address:", quoterAddress);
		console.log("Amount Out:", ethers.formatEther(amountOut), "ETH"); // Adjust decimals as needed
		console.log("Best Fee:", bestFee.toString(), "basis points");
		console.log("Fee Percentage:", Number(bestFee) / 10000 + "%");

		return {
			dexName,
			quoterAddress,
			amountOut: amountOut.toString(),
			bestFee: bestFee.toString(),
		};
	} catch (error) {
		console.error("Error getting best quote:", error);

		// Handle specific error cases
		if (error.message.includes("No valid route found")) {
			console.log("No valid trading route found for the given token pair");
		} else if (error.message.includes("Invalid token")) {
			console.log("Invalid token address provided");
		} else if (error.message.includes("Amount must be > 0")) {
			console.log("Amount must be greater than 0");
		}

		throw error;
	}
}

// Main execution
getBestQuote()
	.then((result) => {
		console.log("\nQuote retrieved successfully!");
		process.exit(0);
	})
	.catch((error) => {
		console.error("Script failed:", error.message);
		process.exit(1);
	});
