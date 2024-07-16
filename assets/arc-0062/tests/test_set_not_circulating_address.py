import pytest
from algokit_utils import LogicError, OnCompleteCallParameters
from algokit_utils.beta.account_manager import AddressAndSigner
from algosdk.encoding import encode_address

from smart_contracts.artifacts.circulating_supply.circulating_supply_client import (
    CirculatingSupplyClient,
)
from smart_contracts.circulating_supply import config as cfg
from smart_contracts.errors import std_errors as err


def test_pass_set_not_circulating_address(
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: AddressAndSigner,
    asset: int,
    burning_with_balance: AddressAndSigner,
    locking_with_balance: AddressAndSigner,
    generic_not_circulating_with_balance: AddressAndSigner,
) -> None:
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

    state = asset_circulating_supply_client.get_global_state()

    assert encode_address(state.burned.as_bytes) == burning_with_balance.address  # type: ignore
    assert encode_address(state.locked.as_bytes) == locking_with_balance.address  # type: ignore
    assert (
        encode_address(state.generic.as_bytes)  # type: ignore
        == generic_not_circulating_with_balance.address
    )


def test_fail_unauthorized(
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_creator: AddressAndSigner,
    asset: int,
    burning_with_balance: AddressAndSigner,
) -> None:
    with pytest.raises(LogicError, match=err.UNAUTHORIZED):  # type: ignore
        asset_circulating_supply_client.set_not_circulating_address(
            address=burning_with_balance.address,
            label=cfg.BURNED,
            transaction_parameters=OnCompleteCallParameters(
                sender=asset_creator.address,
                signer=asset_creator.signer,
                # TODO: Foreign resources should be auto-populated
                foreign_assets=[asset],
                accounts=[burning_with_balance.address],
            ),
        )


def test_fail_not_opted_in(
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: AddressAndSigner,
    asset: int,
) -> None:
    with pytest.raises(LogicError, match=err.NOT_OPTED_IN):  # type: ignore
        asset_circulating_supply_client.set_not_circulating_address(
            address=asset_manager.address,
            label=cfg.BURNED,
            transaction_parameters=OnCompleteCallParameters(
                sender=asset_manager.address,
                signer=asset_manager.signer,
                # TODO: Foreign resources should be auto-populated
                foreign_assets=[asset],
                accounts=[asset_manager.address],
            ),
        )


def test_fail_invalid_label(
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: AddressAndSigner,
    asset: int,
    burning_with_balance: AddressAndSigner,
) -> None:
    with pytest.raises(LogicError, match=err.INVALID_LABEL):  # type: ignore
        asset_circulating_supply_client.set_not_circulating_address(
            address=burning_with_balance.address,
            label="spam",
            transaction_parameters=OnCompleteCallParameters(
                sender=asset_manager.address,
                signer=asset_manager.signer,
                # TODO: Foreign resources should be auto-populated
                foreign_assets=[asset],
                accounts=[burning_with_balance.address],
            ),
        )
