from algokit_utils import (
    AlgorandClient,
    AssetConfigParams,
    AssetTransferParams,
    CommonAppCallParams,
    SigningAccount,
)

from smart_contracts.artifacts.circulating_supply.circulating_supply_client import (
    Arc62GetCirculatingSupplyArgs,
    CirculatingSupplyClient,
    SetNotCirculatingAddressArgs,
)
from smart_contracts.circulating_supply import config as cfg


def test_pass_get_circulating_supply(
    algorand: AlgorandClient,
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: SigningAccount,
    asset: int,
    reserve_with_balance: SigningAccount,
    not_circulating_balance_1: SigningAccount,
    not_circulating_balance_2: SigningAccount,
    not_circulating_balance_3: SigningAccount,
) -> None:
    total = algorand.asset.get_by_id(asset).total
    reserve_balance = algorand.asset.get_account_information(
        reserve_with_balance, asset
    ).balance
    nc_balance_1 = algorand.asset.get_account_information(
        not_circulating_balance_1, asset
    ).balance
    nc_balance_2 = algorand.asset.get_account_information(
        not_circulating_balance_2, asset
    ).balance
    nc_balance_3 = algorand.asset.get_account_information(
        not_circulating_balance_3, asset
    ).balance

    print("\nASA Total: ", total)
    print("Reserve Balance: ", reserve_balance)
    print(f"{cfg.NOT_CIRCULATING_LABEL_1.capitalize()} Balance: ", nc_balance_1)
    print(f"{cfg.NOT_CIRCULATING_LABEL_2.capitalize()} Balance: ", nc_balance_2)
    print(f"{cfg.NOT_CIRCULATING_LABEL_3.capitalize()} Balance: ", nc_balance_3)

    circulating_supply = (
        asset_circulating_supply_client.send.arc62_get_circulating_supply(
            args=Arc62GetCirculatingSupplyArgs(asset_id=asset)
        ).abi_return
    )
    assert circulating_supply == total - reserve_balance

    asset_circulating_supply_client.send.set_not_circulating_address(
        args=SetNotCirculatingAddressArgs(
            address=not_circulating_balance_1.address,
            label=cfg.NOT_CIRCULATING_LABEL_1,
        ),
        params=CommonAppCallParams(sender=asset_manager.address),
    )
    circulating_supply = (
        asset_circulating_supply_client.send.arc62_get_circulating_supply(
            args=Arc62GetCirculatingSupplyArgs(asset_id=asset),
        ).abi_return
    )
    assert circulating_supply == total - reserve_balance - nc_balance_1

    asset_circulating_supply_client.send.set_not_circulating_address(
        args=SetNotCirculatingAddressArgs(
            address=not_circulating_balance_2.address,
            label=cfg.NOT_CIRCULATING_LABEL_2,
        ),
        params=CommonAppCallParams(sender=asset_manager.address),
    )
    circulating_supply = (
        asset_circulating_supply_client.send.arc62_get_circulating_supply(
            args=Arc62GetCirculatingSupplyArgs(asset_id=asset),
        ).abi_return
    )
    assert circulating_supply == total - reserve_balance - nc_balance_1 - nc_balance_2

    asset_circulating_supply_client.send.set_not_circulating_address(
        args=SetNotCirculatingAddressArgs(
            address=not_circulating_balance_3.address,
            label=cfg.NOT_CIRCULATING_LABEL_3,
        ),
        params=CommonAppCallParams(sender=asset_manager.address),
    )
    circulating_supply = (
        asset_circulating_supply_client.send.arc62_get_circulating_supply(
            args=Arc62GetCirculatingSupplyArgs(asset_id=asset),
        ).abi_return
    )
    assert (
        circulating_supply
        == total - reserve_balance - nc_balance_1 - nc_balance_2 - nc_balance_3
    )
    print("Circulating Supply: ", circulating_supply)


def test_pass_no_reserve(
    algorand: AlgorandClient,
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: SigningAccount,
    asset: int,
) -> None:
    total = algorand.asset.get_by_id(asset).total
    algorand.send.asset_config(
        AssetConfigParams(
            sender=asset_manager.address,
            signer=asset_manager.signer,
            asset_id=asset,
            manager=asset_manager.address,
            reserve="",
        ),
    )
    circulating_supply = (
        asset_circulating_supply_client.send.arc62_get_circulating_supply(
            args=Arc62GetCirculatingSupplyArgs(asset_id=asset),
        ).abi_return
    )
    assert circulating_supply == total


def test_pass_closed_address(
    algorand: AlgorandClient,
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_creator: SigningAccount,
    asset_manager: SigningAccount,
    reserve_with_balance: SigningAccount,
    not_circulating_balance_1: SigningAccount,
    asset: int,
) -> None:
    total = algorand.asset.get_by_id(asset).total

    asset_circulating_supply_client.send.set_not_circulating_address(
        args=SetNotCirculatingAddressArgs(
            address=not_circulating_balance_1.address,
            label=cfg.NOT_CIRCULATING_LABEL_1,
        ),
        params=CommonAppCallParams(sender=asset_manager.address),
    )

    algorand.send.asset_transfer(
        AssetTransferParams(
            sender=not_circulating_balance_1.address,
            signer=not_circulating_balance_1.signer,
            asset_id=asset,
            amount=0,
            receiver=asset_creator.address,
            close_asset_to=asset_creator.address,
        ),
    )

    algorand.send.asset_transfer(
        AssetTransferParams(
            sender=reserve_with_balance.address,
            signer=reserve_with_balance.signer,
            asset_id=asset,
            amount=0,
            receiver=asset_creator.address,
            close_asset_to=asset_creator.address,
        ),
    )

    circulating_supply = (
        asset_circulating_supply_client.send.arc62_get_circulating_supply(
            args=Arc62GetCirculatingSupplyArgs(asset_id=asset),
        ).abi_return
    )
    assert circulating_supply == total
