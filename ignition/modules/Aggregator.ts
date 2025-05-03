// This setup uses Hardhat Ignition to manage smart contract deployments.
// Learn more about it at https://hardhat.org/ignition

import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const AggregatorModule = buildModule("AggregatorModule", (m) => {
  const ag = m.contract("Aggregator", [], {
  });

  return { ag };
});

export default AggregatorModule;
