import { Account, Application, assert, assertMatch, BoxMap, Bytes, bytes, Contract, Global, GlobalState, gtxn, itxn, Txn, uint64 } from "@algorandfoundation/algorand-typescript";
import { abimethod, Address, compileArc4, DynamicArray, DynamicBytes, StaticBytes } from "@algorandfoundation/algorand-typescript/arc4";
import { Escrow } from "./contract.algo";
import { btoi, itob } from "@algorandfoundation/algorand-typescript/op";
import { ERR_FORBIDDEN } from "./errors";
import { ERR_DOESNT_EXIST} from './errors'
import { fee } from "../utils/constants";
import { ERR_INVALID_PAYMENT } from "../utils/errors";

function bytes16(acc: Account): bytes<16> {
  return acc.bytes.slice(0, 16).toFixed({ length: 16})
}

const GlobalMinPages: uint64 = 100_000

export class EscrowFactory extends Contract {

  childMBR = GlobalState<uint64>({ key: 'child_mbr', initialValue: GlobalMinPages })

  // 8 or 16 bytes
  walletIDsByAccounts = BoxMap<bytes<16>, bytes>({ keyPrefix: '' })

  private mbr(length: uint64): uint64 {
    return 8_900 + (length * 400)
  }

  new(payment: gtxn.PaymentTxn): uint64 {
    const appCaller = Global.callerApplicationId === 0
    const creator = appCaller
      ? Bytes(bytes16(Txn.sender))
      : Bytes(itob(Global.callerApplicationId))

    const escrow = compileArc4(Escrow);

    const childAppMBR: uint64 = this.childMBR.value + this.mbr(creator.length)

    assertMatch(
      payment,
      {
        receiver: Global.currentApplicationAddress,
        amount: childAppMBR + Global.minBalance,
      },
      ERR_INVALID_PAYMENT
    )

    const newEscrow = escrow.bareCreate({ fee }).createdApp

    const id = newEscrow.id
    const spendAccount = bytes16(newEscrow.address)

    this.walletIDsByAccounts(spendAccount).value = creator

    itxn
      .payment({
        receiver: newEscrow.address,
        amount: Global.minBalance,
        fee,
      })
      .submit()

    escrow.call.rekey({
      appId: id,
      args: [
        appCaller ? Global.callerApplicationAddress : Txn.sender
      ],
      fee
    })

    return id
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

    const childAppMBR: uint64 = Global.minBalance + this.mbr(creator.length)

    spendingAccount.call.delete({ appId: id, fee })

    this.walletIDsByAccounts(key).delete()

    itxn
      .payment({
        receiver: creator.length === 8 ? Global.callerApplicationAddress : Txn.sender,
        amount: childAppMBR,
        fee,
      })
      .submit()
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