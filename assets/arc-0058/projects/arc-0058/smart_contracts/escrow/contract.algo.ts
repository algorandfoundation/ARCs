import { Account, assert, Contract, Global, itxn, Txn } from "@algorandfoundation/algorand-typescript"
import { ERR_ONLY_CREATOR_CAN_REKEY, ERR_ONLY_FACTORY_CAN_DELETE } from "./errors"
import { abimethod } from "@algorandfoundation/algorand-typescript/arc4"
import { fee } from "../utils/constants";

export class Escrow extends Contract {

  rekey(rekeyTo: Account): void {
    assert(Txn.sender === Global.creatorAddress, ERR_ONLY_CREATOR_CAN_REKEY)

    itxn
      .payment({
        amount: 0,
        receiver: Global.currentApplicationAddress,
        rekeyTo,
        fee,
      })
      .submit()
  }

  @abimethod({ allowActions: 'DeleteApplication' })
  delete(): void {
    assert(Txn.sender === Global.creatorAddress, ERR_ONLY_FACTORY_CAN_DELETE)
    
    itxn
      .payment({ closeRemainderTo: Global.creatorAddress })
      .submit()
  }
}