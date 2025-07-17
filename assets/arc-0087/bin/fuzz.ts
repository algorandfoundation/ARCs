import { AlgorandClient, Config } from "@algorandfoundation/algokit-utils";
import { faker } from "@faker-js/faker";
import { Semaphore } from "async-mutex";

// Reference for TypeScript using pure strings
import { deploy, fromBoxes, toMBR } from "../objects";

// Handle concurrency and total applications created
const mutex = new Semaphore(25);
const TOTAL = 10;

// Configure AlgoKit
Config.configure({
  debug: false,
  logger: {
    verbose() {},
    debug() {},
    info() {},
    warn() {},
    error() {},
  },
});

// Example Type for Testing
type Animal = {
  id: number;
  type: string;
  avatar: string;
  petName: string;
  birthdate: Date;
  registeredAt: Date;
  location: {
    country: string;
    city: string;
  };
};

// Factory for the example objects
export function createAnimal() {
  return {
    id: faker.number.int(),
    type: faker.animal.type(),
    avatar: faker.image.avatar(),
    petName: faker.animal.petName(),
    birthdate: faker.date.birthdate(),
    registeredAt: faker.date.past(),
    location: {
      country: faker.location.country(),
      city: faker.location.city(),
      state: faker.location.state(),
    },
  } as Animal;
}

// Create a sampling of data
export const animals = faker.helpers.multiple(createAnimal, {
  count: TOTAL,
});

// Calculate the totals
const data = animals.reduce(
  (acc, animal) => {
    acc[animal.id] = {
      total: 0n,
      animal: toMBR(animal),
    };

    acc[animal.id].total = acc[animal.id].animal;

    return acc;
  },
  {} as { [key: string]: { animal: bigint; total: bigint } },
);
const totals = Object.keys(data).map((k) => data[k].total);
const totalMBR = totals.reduce((p, c) => p + c, BigInt(0));
const averageMBR = totalMBR / BigInt(totals.length);

// Log the expected results
console.log(`Example Animal:`, faker.helpers.arrayElement(animals));
console.log(`\nTotal Animals: ${animals.length}`);
console.log(`Total MBR: ${totalMBR.microAlgo().algo}`);
console.log(`Average MBR: ${averageMBR.microAlgo().algo}`);
console.log(
  `Max MBR: ${totals.reduce((p, c) => (p > c ? p : c)).microAlgo().algo}`,
);
console.log(
  `Min MBR: ${totals.reduce((p, c) => (p < c ? p : c)).microAlgo().algo}\n`,
);
console.log(`Deploying ${animals.length} animals`);

// Attempt to deploy the objects and resolve their state on-chain
try {
  // Configure the client
  const algorand = AlgorandClient.fromEnvironment();
  const deployer = await algorand.account.fromEnvironment("DEPLOYER");
  const balance = await algorand.client.algod
    .accountInformation(deployer.addr)
    .do()
    .then((info) => info.amount);

  // Log accounts and preflight check
  console.log(`Deploying with ${deployer.addr}`);
  console.log(`Account balance: ${balance.microAlgo().algo}`);
  if (balance < totalMBR) {
    throw new Error(
      `Insufficient funds to deploy ${totalMBR.microAlgo().algo} MBR`,
    );
  }
  console.log(
    `Deploying to ${await algorand.client.network().then((n) => n.genesisId)}`,
  );

  // Deploy every object on-chain
  const appIds = await Promise.all(
    animals.map(async (animal) => {
      return await mutex.runExclusive(async () => {
        return await deploy(
          algorand,
          deployer,
          `animal-${animal.id.toString()}`,
          animal,
        );
      });
    }),
  );

  // Log the results of the deployment
  console.log(`\nDeployed ${animals.length} animals`);
  const resultBalance = await algorand.client.algod
    .accountInformation(deployer.addr)
    .do()
    .then((info) => info.amount);
  const bDiff = balance - resultBalance;
  console.log(`Deployed with ${bDiff.microAlgo().algo} spent`);
  console.log(
    `Final Average: ${(bDiff / BigInt(animals.length)).microAlgo().algos}\n`,
  );

  // Resolve all objects from their boxes on-chain
  await Promise.all(
    appIds.map(async (appId, idx) => {
      const obj = await mutex.runExclusive(async () => {
        return await fromBoxes<Animal>(algorand, appId);
      });
      if (idx === appIds.length - 1) {
        console.log(`Sample Object Resolved From Chain:`, obj);
      }
    }),
  );
} catch (e) {
  console.log(`Ensure you have a localnet running and a wallet with funds.`);
  console.error(e);
}
