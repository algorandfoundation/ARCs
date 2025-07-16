import {Arc87Factory} from "./client";
import {AlgorandClient} from "@algorandfoundation/algokit-utils";
import {TransactionSignerAccount} from "@algorandfoundation/algokit-utils/types/account";
import {toMBR} from "./payments";
import { assemble } from "./state";
import {toChunks} from "./chunk";
import _ from "lodash";
import {PREFIX} from "./constants";

const {get} = _;
export async function deploy(algorand: AlgorandClient, deployer: TransactionSignerAccount, name: string, obj: any) {
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
        toChunks(obj).map(async (paths, idx) => {
            const atc = appClient!.newGroup();
            for (const path of paths) {
                console.log(
                    `%cGrouping ${path} with value: ${get(obj, path as string)} for app ${appClient?.appId}`,
                    "color: green;",
                );
                atc.set({
                    args: {
                        path: path as string,
                        value: get(obj, path as string).toString(),
                    },
                    boxReferences: [`${PREFIX}${path}`],
                    sender: deployer!.addr,
                    signer: deployer!.signer,
                });
            }
            await atc
                .send()
                .catch((e) => console.error(e));
        })
    );
    return appClient.appId
}
