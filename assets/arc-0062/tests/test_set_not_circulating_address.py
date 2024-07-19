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
    not_circulating_balance_1: AddressAndSigner,
    not_circulating_balance_2: AddressAndSigner,
    not_circulating_balance_3: AddressAndSigner,
) -> None:
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

    state = asset_circulating_supply_client.get_global_state()

    assert encode_address(state.not_circulating_label_1.as_bytes) == not_circulating_balance_1.address  # type: ignore
    assert encode_address(state.not_circulating_label_2.as_bytes) == not_circulating_balance_2.address  # type: ignore
    assert encode_address(state.not_circulating_label_3.as_bytes) == not_circulating_balance_3.address  # type: ignore


def test_fail_unauthorized(
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_creator: AddressAndSigner,
    asset: int,
    not_circulating_balance_1: AddressAndSigner,
) -> None:
    with pytest.raises(LogicError, match=err.UNAUTHORIZED):  # type: ignore
        asset_circulating_supply_client.set_not_circulating_address(
            address=not_circulating_balance_1.address,
            label=cfg.NOT_CIRCULATING_LABEL_1,
            transaction_parameters=OnCompleteCallParameters(
                sender=asset_creator.address,
                signer=asset_creator.signer,
                # TODO: Foreign resources should be auto-populated
                foreign_assets=[asset],
                accounts=[not_circulating_balance_1.address],
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
            label=cfg.NOT_CIRCULATING_LABEL_1,
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
    not_circulating_balance_1: AddressAndSigner,
) -> None:
    with pytest.raises(LogicError, match=err.INVALID_LABEL):  # type: ignore
        asset_circulating_supply_client.set_not_circulating_address(
            address=not_circulating_balance_1.address,
            label=cfg.NOT_CIRCULATING_LABEL_1 + "spam",
            transaction_parameters=OnCompleteCallParameters(
                sender=asset_manager.address,
                signer=asset_manager.signer,
                # TODO: Foreign resources should be auto-populated
                foreign_assets=[asset],
                accounts=[not_circulating_balance_1.address],
            ),
        )
