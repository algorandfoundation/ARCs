from typing import TypeAlias

from algosdk.v2client.algod import AlgodClient

AccountBalance: TypeAlias = dict[int, int]


def get_account_balance(
    algod_client: AlgodClient, account_address: str
) -> AccountBalance:
    account_info = algod_client.account_info(account_address)
    balances = {a["asset-id"]: int(a["amount"]) for a in account_info["assets"]}
    balances[0] = int(account_info["amount"])
    return balances


def get_account_asset_balance(
    algod_client: AlgodClient, account_address: str, asset_id: int
) -> int:
    return get_account_balance(algod_client, account_address).get(asset_id, 0)


def is_account_opted_in(
    algod_client: AlgodClient, account_address: str, asset_id: int
) -> int:
    return asset_id in get_account_balance(algod_client, account_address)
