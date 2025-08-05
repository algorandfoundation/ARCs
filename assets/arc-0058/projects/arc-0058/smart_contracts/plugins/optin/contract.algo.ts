import { Application, Asset, Global, gtxn, itxn, assertMatch, uint64, assert, Contract } from "@algorandfoundation/algorand-typescript";
import { ERR_INVALID_PAYMENT } from "../../utils/errors";
import { ERR_ALREADY_OPTED_IN } from "./errors";
import { getSpendingAccount, rekeyAddress } from "../../utils/plugins";
import { fee } from "../../utils/constants";

export class OptInPlugin extends Contract {

  optInToAsset(walletID: uint64, rekeyBack: boolean, assets: uint64[], mbrPayment: gtxn.PaymentTxn): void {
    const wallet = Application(walletID)
    const sender = getSpendingAccount(wallet)

    assertMatch(
      mbrPayment,
      {
        receiver: sender,
        amount: Global.assetOptInMinBalance * assets.length
      },
      ERR_INVALID_PAYMENT
    )

    for (let i: uint64 = 0; i < assets.length; i++) {
      assert(!sender.isOptedIn(Asset(assets[i])), ERR_ALREADY_OPTED_IN)

      itxn
        .assetTransfer({
          sender,
          assetReceiver: sender,
          assetAmount: 0,
          xferAsset: Asset(assets[i]),
          rekeyTo: rekeyAddress(rekeyBack, wallet),
          fee,
        })
        .submit();
    }
  }
}
