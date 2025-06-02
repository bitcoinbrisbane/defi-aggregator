import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "@nomicfoundation/hardhat-verify";

import dotenv from "dotenv";
dotenv.config();

const PK = process.env.PK || "";

const config: HardhatUserConfig = {
	defaultNetwork: "hardhat",
	solidity: {
		version: "0.8.27",
		settings: {
			metadata: {
				bytecodeHash: "none", // disable ipfs
				useLiteralContent: true, // use source code
			},
		},
	},
	networks: {
		hardhat: {
			chainId: 1337,
			forking: {
				url: `${process.env.RPC_URL}`,
			},
		},
		monad: {
			url: "https://testnet-rpc.monad.xyz",
			chainId: 10143,
			accounts: PK ? [PK] : [],
		},
	},
	sourcify: {
		enabled: true,
		apiUrl: "https://sourcify-api-monad.blockvision.org",
		browserUrl: "https://testnet.monadexplorer.com",
	},
	// To avoid errors from Etherscan
	etherscan: {
		enabled: false,
	},
	paths: {
		sources: "./contracts",
		tests: "./test",
		cache: "./cache",
		artifacts: "./artifacts",
	},
	mocha: {
		timeout: 40000,
	},
};

export default config;
