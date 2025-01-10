import pytest
from algokit_utils.beta.account_manager import AddressAndSigner

from smart_contracts.artifacts.smart_asa.smart_asa_client import SmartAsaClient

from . import utils


def test_pass_no_circulating_supply(smart_asa_client: SmartAsaClient) -> None:
    circulating_supply = smart_asa_client.get_circulating_supply(
        asset=smart_asa_client.get_global_state().smart_asa_id
    ).return_value
    assert circulating_supply == 0


@pytest.mark.parametrize("asa_config", [False], indirect=True)
def test_pass_circulating_supply(
    smart_asa_client: SmartAsaClient, reserve_with_supply: AddressAndSigner
) -> None:
    smart_asa_id = smart_asa_client.get_global_state().smart_asa_id
    circulating_supply = smart_asa_client.get_circulating_supply(
        asset=smart_asa_id
    ).return_value
    assert circulating_supply == utils.get_account_asset_balance(
        smart_asa_client.algod_client, reserve_with_supply.address, smart_asa_id
    )
