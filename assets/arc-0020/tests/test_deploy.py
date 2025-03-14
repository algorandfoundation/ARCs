from algosdk.constants import ZERO_ADDRESS
from algosdk.encoding import encode_address

from smart_contracts.artifacts.smart_asa.smart_asa_client import SmartAsaClient


def test_pass_create(smart_asa_client_no_asset: SmartAsaClient) -> None:
    state = smart_asa_client_no_asset.state.global_state

    assert state.total == 0
    assert state.decimals == 0
    assert not state.default_frozen
    assert state.unit_name == ""
    assert state.name == ""
    assert state.url == ""
    assert state.metadata_hash == ""
    assert encode_address(state.manager_addr) == ZERO_ADDRESS
    assert encode_address(state.reserve_addr) == ZERO_ADDRESS
    assert encode_address(state.freeze_addr) == ZERO_ADDRESS
    assert encode_address(state.clawback_addr) == ZERO_ADDRESS
    assert state.smart_asa_id == 0
    assert not state.global_frozen


def test_fail_update() -> None:
    pass  # TODO


def test_fail_delete() -> None:
    pass  # TODO


def test_wrong_state_schema() -> None:
    pass  # TODO
