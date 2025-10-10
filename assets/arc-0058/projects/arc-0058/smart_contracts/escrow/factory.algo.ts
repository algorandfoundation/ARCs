import { Account, Application, assert, assertMatch, BoxMap, Bytes, bytes, Contract, Global, GlobalState, gtxn, itxn, op, Txn, uint64 } from "@algorandfoundation/algorand-typescript";
import { abimethod, Address, compileArc4, DynamicArray, DynamicBytes, StaticBytes } from "@algorandfoundation/algorand-typescript/arc4";
import { Escrow } from "./contract.algo";
import { btoi, itob } from "@algorandfoundation/algorand-typescript/op";
import { ERR_ALREADY_REGISTERED, ERR_FORBIDDEN, ERR_INVALID_APP, ERR_INVALID_CREATOR } from "./errors";
import { ERR_DOESNT_EXIST } from './errors'
import { ERR_INVALID_PAYMENT } from "../utils/errors";
import { BoxCostPerByte, EscrowGlobalStateKeysCreator, GlobalStateKeyBytesCost, MinPages, MinWalletIDsByAccountsMbr } from "./constants";

function bytes16(acc: Account): bytes<16> {
  return acc.bytes.slice(0, 16).toFixed({ length: 16 })
}

export class EscrowFactory extends Contract {

  // 8 or 16 bytes
  walletIDsByAccounts = BoxMap<bytes<16>, bytes>({ keyPrefix: '' })

  private mbr(length: uint64): uint64 {
    return MinWalletIDsByAccountsMbr + (length * BoxCostPerByte)
  }

  private getCreator(): bytes {
    const nonAppCaller = Global.callerApplicationId === 0
    return nonAppCaller
      ? Bytes(bytes16(Txn.sender))
      : Bytes(itob(Global.callerApplicationId))
  }

  new(payment: gtxn.PaymentTxn): uint64 {
    const nonAppCaller = Global.callerApplicationId === 0
    const creator = this.getCreator()

    const escrow = compileArc4(Escrow);

    const childAppMBR: uint64 = (
      MinPages +
      GlobalStateKeyBytesCost
    )

    assertMatch(
      payment,
      {
        receiver: Global.currentApplicationAddress,
        amount: childAppMBR + Global.minBalance,
      },
      ERR_INVALID_PAYMENT
    )

    const newEscrow = escrow.call.create(
      {
        args: [creator],
      }
    ).itxn.createdApp

    itxn
      .payment({
        receiver: newEscrow.address,
        amount: Global.minBalance
      })
      .submit()

    escrow.call.rekey({
      appId: newEscrow.id,
      args: [
        nonAppCaller ? Txn.sender : Global.callerApplicationAddress
      ]
    })

    return newEscrow.id
  }

  register(payment: gtxn.PaymentTxn, app: uint64): void {
    // only apps can call this method
    assert(Global.callerApplicationId !== 0)
    // this way we ensure apps are always either
    // registering themselves or escrows they can prove they created
    assert(app === 0 || Application(app).creator === Global.currentApplicationAddress, ERR_INVALID_APP)

    let creator = Bytes('')
    if (Application(app).creator === Global.currentApplicationAddress) {
      creator = op.AppGlobal.getExBytes(app, Bytes(EscrowGlobalStateKeysCreator))[0]
      assert(btoi(creator) === Global.callerApplicationId, ERR_INVALID_CREATOR)
    } else {
      creator = Bytes(itob(Global.callerApplicationId))
      assert(app === Global.callerApplicationId, ERR_INVALID_APP)
    }

    const appAddress = bytes16(Application(app).address)
    assert(!this.walletIDsByAccounts(appAddress).exists, ERR_ALREADY_REGISTERED)

    assertMatch(
      payment,
      {
        receiver: Global.currentApplicationAddress,
        amount: this.mbr(creator.length),
      },
      ERR_INVALID_PAYMENT
    )

    this.walletIDsByAccounts(appAddress).value = creator
  }

  delete(id: uint64): void {
    const caller = Global.callerApplicationId
    const key = bytes16(Application(id).address)
    assert(this.walletIDsByAccounts(key).exists, ERR_DOESNT_EXIST)

    const creator = this.walletIDsByAccounts(key).value
    if (creator.length === 8) {
      assert(caller === btoi(creator), ERR_FORBIDDEN);
    } else {
      assert(Bytes(bytes16(Txn.sender)) === creator, ERR_FORBIDDEN);
    }

    const spendingAccount = compileArc4(Escrow);

    const refundAmount: uint64 = (
      MinPages +
      GlobalStateKeyBytesCost +
      this.mbr(creator.length)
    )

    spendingAccount.call.delete({ appId: id })

    this.walletIDsByAccounts(key).delete()

    itxn
      .payment({
        receiver: creator.length === 8 ? Global.callerApplicationAddress : Txn.sender,
        amount: refundAmount
      })
      .submit()
  }

  // @ts-ignore
  @abimethod({ readonly: true })
  cost(): uint64 {
    return (
      MinPages +
      GlobalStateKeyBytesCost +
      Global.minBalance
    )
  }

  // @ts-ignore
  @abimethod({ readonly: true })
  registerCost(): uint64 {
    return this.mbr(8)
  }

  // @ts-ignore
  @abimethod({ readonly: true })
  exists(address: Address): boolean {
    return this.walletIDsByAccounts(bytes16(address.native)).exists
  }

  // @ts-ignore
  @abimethod({ readonly: true })
  get(address: Address): bytes {
    if (!this.walletIDsByAccounts(bytes16(address.native)).exists) {
      return Bytes('')
    }
    return this.walletIDsByAccounts(bytes16(address.native)).value
  }

  // @ts-ignore
  @abimethod({ readonly: true })
  mustGet(address: Address): bytes {
    assert(this.walletIDsByAccounts(bytes16(address.native)).exists, 'Account not found')
    return this.walletIDsByAccounts(bytes16(address.native)).value
  }

  // @ts-ignore
  @abimethod({ readonly: true })
  getList(addresses: DynamicArray<Address>): DynamicArray<DynamicBytes> {
    const apps = new DynamicArray<DynamicBytes>()
    for (let i: uint64 = 0; i < addresses.length; i++) {
      const address = addresses[i]
      if (this.walletIDsByAccounts(bytes16(address.native)).exists) {
        apps.push(new DynamicBytes(this.walletIDsByAccounts(bytes16(address.native)).value))
      } else {
        apps.push(new DynamicBytes(''))
      }
    }
    return apps
  }

  // @ts-ignore
  @abimethod({ readonly: true })
  mustGetList(addresses: DynamicArray<Address>): DynamicArray<DynamicBytes> {
    const apps = new DynamicArray<DynamicBytes>()
    for (let i: uint64 = 0; i < addresses.length; i++) {
      const address = addresses[i]
      assert(this.walletIDsByAccounts(bytes16(address.native)).exists, 'Account not found')
      apps.push(new DynamicBytes(this.walletIDsByAccounts(bytes16(address.native)).value))
    }
    return apps
  }
}