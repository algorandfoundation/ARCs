import { Account, Application, Bytes, Global, op } from "@algorandfoundation/algorand-typescript";
import { AbstractAccountGlobalStateKeysControlledAddress, AbstractAccountGlobalStateKeysSpendingAddress } from "../abstracted_account/constants";

export type Arc58Accounts = {
  walletAddress: Account;
  origin: Account;
  sender: Account;
};

export function getAccounts(wallet: Application): Arc58Accounts {
  const origin = getOriginAccount(wallet)
  const sender = getSpendingAccount(wallet)
  return {
    walletAddress: wallet.address,
    origin,
    sender,
  }
}

/**
 * getOriginAddress returns the origin address of the contract
 * @param wallet The application to get the controlled address from
 * @returns The controlled address of the contract
 */
export function getOriginAccount(wallet: Application): Account {
  const [controlledAccountBytes] = op.AppGlobal.getExBytes(
    wallet,
    Bytes(AbstractAccountGlobalStateKeysControlledAddress)
  )
  return Account(Bytes(controlledAccountBytes))
}

export function getSpendingAccount(wallet: Application): Account {
  const [spendingAddressBytes] = op.AppGlobal.getExBytes(
    wallet,
    Bytes(AbstractAccountGlobalStateKeysSpendingAddress)
  )
  return Account(Bytes(spendingAddressBytes))
}

export function rekeyAddress(rekeyBack: boolean, wallet: Application): Account {
  if (!rekeyBack) {
    return Global.zeroAddress
  }

  return wallet.address
}