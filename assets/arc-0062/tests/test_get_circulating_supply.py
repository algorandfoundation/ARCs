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
    burning_with_balance: AddressAndSigner,
    locking_with_balance: AddressAndSigner,
    generic_not_circulating_with_balance: AddressAndSigner,
) -> None:
    total: int = algorand_client.client.algod.asset_info(asset)["params"]["total"]  # type: ignore
    reserve_balance: int = get_asset_balance(
        algorand_client, reserve_with_balance.address, asset
    )
    burned_balance: int = get_asset_balance(
        algorand_client, burning_with_balance.address, asset
    )
    locked_balance: int = get_asset_balance(
        algorand_client, locking_with_balance.address, asset
    )
    generic_balance: int = get_asset_balance(
        algorand_client, generic_not_circulating_with_balance.address, asset
    )

    print("\nASA Total: ", total)
    print("Reserve Balance: ", reserve_balance)
    print("Burned Balance: ", burned_balance)
    print("Locked Balance: ", locked_balance)
    print("Generic Not-Circulating Balance: ", generic_balance)

    not_circulating_addresses = [
        reserve_with_balance.address,
        burning_with_balance.address,
        locking_with_balance.address,
        generic_not_circulating_with_balance.address,
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
        address=burning_with_balance.address,
        label=cfg.BURNED,
        transaction_parameters=OnCompleteCallParameters(
            sender=asset_manager.address,
            signer=asset_manager.signer,
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=[burning_with_balance.address],
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
    assert circulating_supply == total - reserve_balance - burned_balance

    asset_circulating_supply_client.set_not_circulating_address(
        address=locking_with_balance.address,
        label=cfg.LOCKED,
        transaction_parameters=OnCompleteCallParameters(
            sender=asset_manager.address,
            signer=asset_manager.signer,
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=[locking_with_balance.address],
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
        circulating_supply == total - reserve_balance - burned_balance - locked_balance
    )

    asset_circulating_supply_client.set_not_circulating_address(
        address=generic_not_circulating_with_balance.address,
        label=cfg.GENERIC,
        transaction_parameters=OnCompleteCallParameters(
            sender=asset_manager.address,
            signer=asset_manager.signer,
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=[generic_not_circulating_with_balance.address],
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
        == total - reserve_balance - burned_balance - locked_balance - generic_balance
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
    burning_with_balance: AddressAndSigner,
    asset: int,
) -> None:
    total: int = algorand_client.client.algod.asset_info(asset)["params"]["total"]  # type: ignore

    asset_circulating_supply_client.set_not_circulating_address(
        address=burning_with_balance.address,
        label=cfg.BURNED,
        transaction_parameters=OnCompleteCallParameters(
            sender=asset_manager.address,
            signer=asset_manager.signer,
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
            accounts=[burning_with_balance.address],
        ),
    )

    algorand_client.send.asset_transfer(
        AssetTransferParams(
            sender=burning_with_balance.address,
            signer=burning_with_balance.signer,
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
            accounts=[reserve_with_balance.address, burning_with_balance.address],
        ),
    ).return_value
    assert circulating_supply == total
