import pytest
from algokit_utils import CommonAppCallParams, LogicError, SigningAccount

from smart_contracts.artifacts.circulating_supply.circulating_supply_client import (
    CirculatingSupplyClient,
    SetAssetArgs,
)
from smart_contracts.errors import std_errors as err


def test_pass_set_asset(
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: SigningAccount,
    asset: int,
) -> None:
    assert asset == asset_circulating_supply_client.state.global_state.asset_id


def test_fail_unauthorized_manager(
    circulating_supply_client: CirculatingSupplyClient,
    asset_creator: SigningAccount,
    asset: int,
) -> None:
    with pytest.raises(LogicError, match=err.UNAUTHORIZED):
        circulating_supply_client.send.set_asset(
            args=SetAssetArgs(asset_id=asset),
            params=CommonAppCallParams(sender=asset_creator.address),
        )


def test_fail_unauthorized_already_set(
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: SigningAccount,
    asset: int,
) -> None:
    with pytest.raises(LogicError, match=err.UNAUTHORIZED):
        asset_circulating_supply_client.send.set_asset(
            args=SetAssetArgs(asset_id=asset),
            params=CommonAppCallParams(sender=asset_manager.address),
        )
