import {AlgorandClient, Config} from "@algorandfoundation/algokit-utils";
import {NetworkId, WalletId, WalletManager} from "@txnlab/use-wallet";

import {Semaphore} from "async-mutex";
import {faker} from "@faker-js/faker";


import {fromBoxes, toMBR} from "../objects";
import {deploy} from "../objects/deploy";

// Handle concurrency and total users created
const mutex = new Semaphore(25);
const TOTAL = 10;

Config.configure({
    debug: false,
    logger: {
        verbose(message: string, ...optionalParams) {
        },
        debug(message: string, ...optionalParams) {
        },
        info(message: string, ...optionalParams) {
        },
        warn(message: string, ...optionalParams) {
        },
        error(message: string, ...optionalParams) {
        },
    },
});

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
    }
}
export function createAnimal() {
    return {
        id: faker.number.int(),
        type: faker.animal.type(),
        avatar: faker.image.avatar(),
        petName: faker.animal.petName(),
        birthdate: faker.date.birthdate(),
        registeredAt: faker.date.past(),
        location: createRandomLocation()
    } as Animal;
}

export const animals = faker.helpers.multiple(createAnimal, {
    count: TOTAL,
});

export function createRandomLocation() {
    return {
        country: faker.location.country(),
        city: faker.location.city(),
        state: faker.location.state(),
    };
}


const data = animals.reduce(
    (acc, animal) => {
        acc[animal.id] = {
            total: 0n,
            user: toMBR(animal),
        };

        acc[animal.id].total = acc[animal.id].user;

        return acc;
    },
    {} as { [key: string]: { user: bigint; total: bigint } },
);

const totals = Object.keys(data).map((k) => data[k].total);
const totalMBR = totals.reduce((p, c) => p + c, BigInt(0));
const averageMBR = totalMBR / BigInt(totals.length);

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

console.log(
    `Deploying Store with ${animals.length} animals`,
);

try {
    const algorand = AlgorandClient.fromEnvironment();
    const deployer = await algorand.account.fromEnvironment("DEPLOYER");
    const balance = await algorand.client.algod
        .accountInformation(deployer.addr)
        .do()
        .then((info) => info.amount);


    // Grab the KMD using UseWallet just to show it's possible
    const manager = new WalletManager({
        wallets: [WalletId.KMD],
        defaultNetwork: NetworkId.LOCALNET,
    });
    globalThis.prompt = () => {
        return "";
    };
    await manager.wallets[0].connect();

    console.log(`Deploying with ${manager.activeAddress}`);
    console.log(`Account balance: ${balance.microAlgo().algo}`);

    if(balance < totalMBR){
        throw new Error(`Insufficient funds to deploy ${totalMBR.microAlgo().algo} MBR`);
    }
    console.log(
        `Deploying to ${await algorand.client.network().then((n) => n.genesisId)}`,
    );

    const appIds = await Promise.all(
        animals.map(async (animal) => {
            return await mutex.runExclusive(async () => {
                return await deploy(algorand, deployer, `animal-${animal.id.toString()}`, animal);
            });
        }),
    );

    console.log(`Deployed ${animals.length} animals`);

    const resultBalance = await algorand.client.algod
        .accountInformation(deployer.addr)
        .do()
        .then((info) => info.amount);
    const bDiff = balance - resultBalance;
    console.log(`Deployed with ${bDiff.microAlgo().algo} spent`);
    console.log(
        `Final Average: ${(bDiff / BigInt(animals.length)).microAlgo().algos}`,
    );

    await Promise.all(appIds.map(async (appId) => {
        console.log(await fromBoxes<Animal>(algorand, appId))
    }))
} catch (e) {
    console.log(`Ensure you have a localnet running and a wallet with funds.`)
    console.error(e);
}
