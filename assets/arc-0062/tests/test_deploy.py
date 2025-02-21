from algosdk.constants import ZERO_ADDRESS

from smart_contracts.artifacts.circulating_supply.circulating_supply_client import (
    CirculatingSupplyClient,
)


def test_pass_create(circulating_supply_client: CirculatingSupplyClient) -> None:
    state = circulating_supply_client.state.global_state

    assert state.asset_id == 0
    assert state.not_circulating_label_1 == ZERO_ADDRESS
    assert state.not_circulating_label_2 == ZERO_ADDRESS
    assert state.not_circulating_label_3 == ZERO_ADDRESS
