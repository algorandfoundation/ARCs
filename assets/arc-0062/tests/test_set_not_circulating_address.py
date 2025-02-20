import pytest
from algokit_utils import CommonAppCallParams, LogicError, SigningAccount

from smart_contracts.artifacts.circulating_supply.circulating_supply_client import (
    CirculatingSupplyClient,
    SetNotCirculatingAddressArgs,
)
from smart_contracts.circulating_supply import config as cfg
from smart_contracts.errors import std_errors as err


def test_pass_set_not_circulating_address(
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: SigningAccount,
    asset: int,
    not_circulating_balance_1: SigningAccount,
    not_circulating_balance_2: SigningAccount,
    not_circulating_balance_3: SigningAccount,
) -> None:
    asset_circulating_supply_client.send.set_not_circulating_address(
        args=SetNotCirculatingAddressArgs(
            address=not_circulating_balance_1.address,
            label=cfg.NOT_CIRCULATING_LABEL_1,
        ),
        params=CommonAppCallParams(sender=asset_manager.address),
    )

    asset_circulating_supply_client.send.set_not_circulating_address(
        args=SetNotCirculatingAddressArgs(
            address=not_circulating_balance_2.address,
            label=cfg.NOT_CIRCULATING_LABEL_2,
        ),
        params=CommonAppCallParams(sender=asset_manager.address),
    )

    asset_circulating_supply_client.send.set_not_circulating_address(
        args=SetNotCirculatingAddressArgs(
            address=not_circulating_balance_3.address,
            label=cfg.NOT_CIRCULATING_LABEL_3,
        ),
        params=CommonAppCallParams(sender=asset_manager.address),
    )

    state = asset_circulating_supply_client.state.global_state

    assert state.not_circulating_label_1 == not_circulating_balance_1.address
    assert state.not_circulating_label_2 == not_circulating_balance_2.address
    assert state.not_circulating_label_3 == not_circulating_balance_3.address


def test_fail_unauthorized(
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_creator: SigningAccount,
    asset: int,
    not_circulating_balance_1: SigningAccount,
) -> None:
    with pytest.raises(LogicError, match=err.UNAUTHORIZED):
        asset_circulating_supply_client.send.set_not_circulating_address(
            args=SetNotCirculatingAddressArgs(
                address=not_circulating_balance_1.address,
                label=cfg.NOT_CIRCULATING_LABEL_1,
            ),
            params=CommonAppCallParams(sender=asset_creator.address),
        )


def test_fail_not_opted_in(
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: SigningAccount,
    asset: int,
) -> None:
    with pytest.raises(LogicError, match=err.NOT_OPTED_IN):
        asset_circulating_supply_client.send.set_not_circulating_address(
            args=SetNotCirculatingAddressArgs(
                address=asset_manager.address,
                label=cfg.NOT_CIRCULATING_LABEL_1,
            ),
            params=CommonAppCallParams(sender=asset_manager.address),
        )


def test_fail_invalid_label(
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: SigningAccount,
    asset: int,
    not_circulating_balance_1: SigningAccount,
) -> None:
    with pytest.raises(LogicError, match=err.INVALID_LABEL):
        asset_circulating_supply_client.send.set_not_circulating_address(
            args=SetNotCirculatingAddressArgs(
                address=not_circulating_balance_1.address,
                label=cfg.NOT_CIRCULATING_LABEL_1 + "spam",
            ),
            params=CommonAppCallParams(sender=asset_manager.address),
        )
