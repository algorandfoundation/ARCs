import pytest
from algokit_utils import SigningAccount

from smart_contracts.artifacts.smart_asa.smart_asa_client import (
    GetCirculatingSupplyArgs,
    SmartAsaClient,
)


def test_pass_no_circulating_supply(smart_asa_client: SmartAsaClient) -> None:
    circulating_supply = smart_asa_client.send.get_circulating_supply(
        GetCirculatingSupplyArgs(asset=smart_asa_client.state.global_state.smart_asa_id)
    ).abi_return
    assert circulating_supply == 0


@pytest.mark.parametrize("asa_config", [False], indirect=True)
def test_pass_circulating_supply(
    smart_asa_client: SmartAsaClient, reserve_with_supply: SigningAccount
) -> None:
    smart_asa_id = smart_asa_client.state.global_state.smart_asa_id
    circulating_supply = smart_asa_client.send.get_circulating_supply(
        GetCirculatingSupplyArgs(asset=smart_asa_id)
    ).abi_return
    assert (
        circulating_supply
        == smart_asa_client.algorand.asset.get_account_information(
            reserve_with_supply, smart_asa_id
        ).balance
    )
