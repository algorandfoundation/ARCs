import { Application, itxn, uint64, Contract } from "@algorandfoundation/algorand-typescript";
import { getSpendingAccount, rekeyAddress } from "../../utils/plugins";
import { Address } from "@algorandfoundation/algorand-typescript/arc4";

export class PayPlugin extends Contract {

  pay(walletID: uint64, rekeyBack: boolean, receiver: Address, asset: uint64, amount: uint64): void {
    const wallet = Application(walletID)
    const sender = getSpendingAccount(wallet)

    if (asset === 0) {
      itxn
        .payment({
          sender,
          receiver: receiver.native,
          amount,
          rekeyTo: rekeyAddress(rekeyBack, wallet)
        })
        .submit()
    } else {
      itxn
        .assetTransfer({
          sender,
          assetReceiver: receiver.native,
          assetAmount: amount,
          xferAsset: asset,
          rekeyTo: rekeyAddress(rekeyBack, wallet)
        })
        .submit()
    }
  }
}
