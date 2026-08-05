/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  testTimeout: 60000,
  // temporary fix for bigint serialization issue
  // https://github.com/wormholelabs-xyz/wormhole-sdk-ts/commit/afab351e7ba90ec9abccdaf9edd220fe363f2399#diff-860bd1f15d1e0bafcfc6f62560524f588e6d6bf56d4ab1b0f6f8146461558730R15
  maxWorkers: "50%",
  workerThreads: true,
};
