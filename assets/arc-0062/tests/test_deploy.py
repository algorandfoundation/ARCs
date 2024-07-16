from algosdk.constants import ZERO_ADDRESS
from algosdk.encoding import encode_address

from smart_contracts.artifacts.circulating_supply.circulating_supply_client import (
    CirculatingSupplyClient,
)


def test_pass_create(circulating_supply_client: CirculatingSupplyClient) -> None:
    state = circulating_supply_client.get_global_state()

    assert state.asset_id == 0
    assert encode_address(state.burned.as_bytes) == ZERO_ADDRESS  # type: ignore
    assert encode_address(state.locked.as_bytes) == ZERO_ADDRESS  # type: ignore
    assert encode_address(state.generic.as_bytes) == ZERO_ADDRESS  # type: ignore
