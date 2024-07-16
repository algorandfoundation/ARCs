import pytest
from algokit_utils import LogicError, OnCompleteCallParameters
from algokit_utils.beta.account_manager import AddressAndSigner

from smart_contracts.artifacts.circulating_supply.circulating_supply_client import (
    CirculatingSupplyClient,
)
from smart_contracts.errors import std_errors as err


def test_pass_set_asset(
    asset_circulating_supply_client: CirculatingSupplyClient,
    asset_manager: AddressAndSigner,
    asset: int,
) -> None:
    assert asset == asset_circulating_supply_client.get_global_state().asset_id


def test_fail_unauthorized(
    circulating_supply_client: CirculatingSupplyClient,
    asset_creator: AddressAndSigner,
    asset: int,
) -> None:
    with pytest.raises(LogicError, match=err.UNAUTHORIZED):  # type: ignore
        circulating_supply_client.set_asset(
            asset_id=asset,
            transaction_parameters=OnCompleteCallParameters(
                sender=asset_creator.address,
                signer=asset_creator.signer,
                # TODO: Foreign resources should be auto-populated
                foreign_assets=[asset],
            ),
        )
