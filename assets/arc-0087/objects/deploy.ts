import type { AlgorandClient } from "@algorandfoundation/algokit-utils";
import type { TransactionSignerAccount } from "@algorandfoundation/algokit-utils/types/account";
import _ from "lodash";

// Client and interfaces
import { Arc87Factory } from "./client.js";
import { GROUP_SIZE, PREFIX } from "./constants.js";
import { toPaths } from "./paths.js";
import { toMBR } from "./payments.js";

/**
 * This is just for demonstration purposes
 *
 * @protected
 */
export async function deploy<T>(
  algorand: AlgorandClient,
  deployer: TransactionSignerAccount,
  name: string,
  obj: T,
) {
  const factory = algorand.client.getTypedAppFactory(Arc87Factory, {
    deletable: true,
    defaultSender: deployer.addr,
    defaultSigner: deployer.signer,
  });
  const { appClient } = await factory.deploy({
    onUpdate: "append",
    onSchemaBreak: "append",
    appName: name,
  });
  await appClient.appClient.fundAppAccount({
    amount: toMBR(obj).microAlgo(),
    sender: deployer.addr,
  });

  // Save State On-Chain
  await Promise.all(
    toChunks(obj).map(async (paths) => {
      const atc = appClient.newGroup();
      for (const path of paths) {
        console.log(
          `%cGrouping ${path} with value: ${_.get(obj, path as string)} for app ${appClient?.appId}`,
          "color: green;",
        );
        atc.set({
          args: {
            path: path as string,
            value: _.get(obj, path as string).toString(),
          },
          boxReferences: [`${PREFIX}${path}`],
          sender: deployer.addr,
          signer: deployer.signer,
        });
      }
      await atc.send().catch((e) => console.error(e));
    }),
  );
  return appClient.appId;
}

/**
 * Splits the paths of an object into smaller chunks of a specified size.
 *
 * @protected
 */
function toChunks(obj: unknown, size: number = GROUP_SIZE): string[][] {
  return toPaths(obj).reduce((accumulator, currentValue, index) => {
    const chunkIndex = Math.floor(index / size);
    if (!accumulator[chunkIndex]) {
      accumulator[chunkIndex] = []; // Start a new chunk
    }

    accumulator[chunkIndex].push(currentValue);

    return accumulator;
  }, [] as string[][]);
}
