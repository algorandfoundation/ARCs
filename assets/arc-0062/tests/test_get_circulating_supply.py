from algokit_utils import OnCompleteCallParameters
from algokit_utils.beta.account_manager import AddressAndSigner
from algokit_utils.beta.algorand_client import (
    AlgorandClient,
    AssetConfigParams,
    AssetTransferParams,
)

from smart_contracts.artifacts.circulating_supply.circulating_supply_client import (
    CirculatingSupplyClient,
)
from smart_contracts.circulating_supply import config as cfg

from .conftest import get_asset_balance


def test_pass_get_circulating_supply(
    algorand_client: AlgorandClient,
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: AddressAndSigner,
    asset: int,
    reserve_with_balance: AddressAndSigner,
    not_circulating_balance_1: AddressAndSigner,
    not_circulating_balance_2: AddressAndSigner,
    not_circulating_balance_3: AddressAndSigner,
) -> None:
    total: int = algorand_client.client.algod.asset_info(asset)["params"]["total"]  # type: ignore
    reserve_balance: int = get_asset_balance(
        algorand_client, reserve_with_balance.address, asset
    )
    nc_balance_1: int = get_asset_balance(
        algorand_client, not_circulating_balance_1.address, asset
    )
    nc_balance_2: int = get_asset_balance(
        algorand_client, not_circulating_balance_2.address, asset
    )
    nc_balance_3: int = get_asset_balance(
        algorand_client, not_circulating_balance_3.address, asset
    )

    print("\nASA Total: ", total)
    print("Reserve Balance: ", reserve_balance)
    print(f"{cfg.NOT_CIRCULATING_LABEL_1.capitalize()} Balance: ", nc_balance_1)
    print(f"{cfg.NOT_CIRCULATING_LABEL_2.capitalize()} Balance: ", nc_balance_2)
    print(f"{cfg.NOT_CIRCULATING_LABEL_3.capitalize()} Balance: ", nc_balance_3)

    not_circulating_addresses = [
        reserve_with_balance.address,
        not_circulating_balance_1.address,
        not_circulating_balance_2.address,
        not_circulating_balance_3.address,
    ]

    circulating_supply = asset_circulating_supply_client.arc62_get_circulating_supply(
        asset_id=asset,
        transaction_parameters=OnCompleteCallParameters(
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=[reserve_with_balance.address],
        ),
    ).return_value
    assert circulating_supply == total - reserve_balance

    asset_circulating_supply_client.set_not_circulating_address(
        address=not_circulating_balance_1.address,
        label=cfg.NOT_CIRCULATING_LABEL_1,
        transaction_parameters=OnCompleteCallParameters(
            sender=asset_manager.address,
            signer=asset_manager.signer,
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=[not_circulating_balance_1.address],
        ),
    )
    circulating_supply = asset_circulating_supply_client.arc62_get_circulating_supply(
        asset_id=asset,
        transaction_parameters=OnCompleteCallParameters(
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=not_circulating_addresses,
        ),
    ).return_value
    assert circulating_supply == total - reserve_balance - nc_balance_1

    asset_circulating_supply_client.set_not_circulating_address(
        address=not_circulating_balance_2.address,
        label=cfg.NOT_CIRCULATING_LABEL_2,
        transaction_parameters=OnCompleteCallParameters(
            sender=asset_manager.address,
            signer=asset_manager.signer,
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=[not_circulating_balance_2.address],
        ),
    )
    circulating_supply = asset_circulating_supply_client.arc62_get_circulating_supply(
        asset_id=asset,
        transaction_parameters=OnCompleteCallParameters(
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=not_circulating_addresses,
        ),
    ).return_value
    assert circulating_supply == total - reserve_balance - nc_balance_1 - nc_balance_2

    asset_circulating_supply_client.set_not_circulating_address(
        address=not_circulating_balance_3.address,
        label=cfg.NOT_CIRCULATING_LABEL_3,
        transaction_parameters=OnCompleteCallParameters(
            sender=asset_manager.address,
            signer=asset_manager.signer,
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=[not_circulating_balance_3.address],
        ),
    )
    circulating_supply = asset_circulating_supply_client.arc62_get_circulating_supply(
        asset_id=asset,
        transaction_parameters=OnCompleteCallParameters(
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=not_circulating_addresses,
        ),
    ).return_value
    assert (
        circulating_supply
        == total - reserve_balance - nc_balance_1 - nc_balance_2 - nc_balance_3
    )
    print("Circulating Supply: ", circulating_supply)


def test_pass_no_reserve(
    algorand_client: AlgorandClient,
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: AddressAndSigner,
    asset: int,
) -> None:
    total: int = algorand_client.client.algod.asset_info(asset)["params"]["total"]  # type: ignore
    algorand_client.send.asset_config(
        AssetConfigParams(
            sender=asset_manager.address,
            signer=asset_manager.signer,
            asset_id=asset,
            manager=asset_manager.address,
            reserve="",
        ),
    )
    circulating_supply = asset_circulating_supply_client.arc62_get_circulating_supply(
        asset_id=asset,
        transaction_parameters=OnCompleteCallParameters(
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
        ),
    ).return_value
    assert circulating_supply == total


def test_pass_closed_address(
    algorand_client: AlgorandClient,
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_creator: AddressAndSigner,
    asset_manager: AddressAndSigner,
    reserve_with_balance: AddressAndSigner,
    not_circulating_balance_1: AddressAndSigner,
    asset: int,
) -> None:
    total: int = algorand_client.client.algod.asset_info(asset)["params"]["total"]  # type: ignore

    asset_circulating_supply_client.set_not_circulating_address(
        address=not_circulating_balance_1.address,
        label=cfg.NOT_CIRCULATING_LABEL_1,
        transaction_parameters=OnCompleteCallParameters(
            sender=asset_manager.address,
            signer=asset_manager.signer,
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=[not_circulating_balance_1.address],
        ),
    )

    algorand_client.send.asset_transfer(
        AssetTransferParams(
            sender=not_circulating_balance_1.address,
            signer=not_circulating_balance_1.signer,
            asset_id=asset,
            amount=0,
            receiver=asset_creator.address,
            close_asset_to=asset_creator.address,
        ),
    )

    algorand_client.send.asset_transfer(
        AssetTransferParams(
            sender=reserve_with_balance.address,
            signer=reserve_with_balance.signer,
            asset_id=asset,
            amount=0,
            receiver=asset_creator.address,
            close_asset_to=asset_creator.address,
        ),
    )

    circulating_supply = asset_circulating_supply_client.arc62_get_circulating_supply(
        asset_id=asset,
        transaction_parameters=OnCompleteCallParameters(
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=[reserve_with_balance.address, not_circulating_balance_1.address],
        ),
    ).return_value
    assert circulating_supply == total
