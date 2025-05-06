import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "@nomicfoundation/hardhat-verify";

import dotenv from "dotenv";
dotenv.config();

const PK = process.env.PK || "";

const config: HardhatUserConfig = {
	defaultNetwork: "hardhat",
	networks: {
		hardhat: {
			chainId: 1337,
			forking: {
				url: `${process.env.RPC_URL}`,
			},
		},
		sepolia: {
			url: `https://sepolia.infura.io/v3/${process.env.INFURA_API_KEY}`,
			accounts: PK ? [PK] : [],
			chainId: 11155111,
		},
		monad: {
			chainId: 10143,
			url: `${process.env.RPC_URL}`,
			accounts: PK ? [PK] : [],
		},
	},
	solidity: "0.8.27",
	paths: {
		sources: "./contracts",
		tests: "./test",
		cache: "./cache",
		artifacts: "./artifacts",
	},
	mocha: {
		timeout: 40000,
	},
	etherscan: {
		apiKey: {
			sepolia: process.env.ETHERSCAN_API_KEY || "",
			monad: process.env.ETHERSCAN_API_KEY || "",
		},
		customChains: [
			{
				network: "base",
				chainId: 8453,
				urls: {
					apiURL: "https://api.basescan.org/api",
					browserURL: "https://basescan.org",
				},
			},
		],
	},
};

export default config;
