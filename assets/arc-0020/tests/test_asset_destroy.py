import pytest
from algokit_utils import AlgoAmount, CommonAppCallParams, LogicError, SigningAccount
from algosdk.constants import ZERO_ADDRESS
from algosdk.error import AlgodHTTPError

import smart_contracts.errors as err
from smart_contracts.artifacts.smart_asa.smart_asa_client import (
    AssetDestroyArgs,
    SmartAsaClient,
)


def test_pass_destroy(
    smart_asa_client: SmartAsaClient, manager: SigningAccount
) -> None:
    smart_asa = smart_asa_client.state.global_state
    sp = smart_asa_client.algorand.client.algod.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2
    smart_asa_client.send.asset_destroy(
        AssetDestroyArgs(destroy_asset=smart_asa.smart_asa_id),
        params=CommonAppCallParams(
            static_fee=AlgoAmount.from_micro_algo(sp.fee),
            signer=manager.signer,
            sender=manager.address,
        ),
    )
    with pytest.raises(AlgodHTTPError, match="asset does not exist"):
        smart_asa_client.algorand.asset.get_by_id(smart_asa.smart_asa_id)
    smart_asa = smart_asa_client.state.global_state
    assert smart_asa.total == 0
    assert smart_asa.decimals == 0
    assert not smart_asa.default_frozen
    assert smart_asa.unit_name == ""
    assert smart_asa.name == ""
    assert smart_asa.url == ""
    assert smart_asa.metadata_hash == ""
    assert smart_asa.manager_addr == ZERO_ADDRESS
    assert smart_asa.reserve_addr == ZERO_ADDRESS
    assert smart_asa.freeze_addr == ZERO_ADDRESS
    assert smart_asa.clawback_addr == ZERO_ADDRESS
    assert smart_asa.smart_asa_id == 0
    assert not smart_asa.global_frozen


def test_fail_missing_ctrl_asa() -> None:
    pass  # TODO


def test_fail_invalid_ctrl_asa() -> None:
    pass  # TODO


def test_fail_unauthorized_manager(
    smart_asa_client: SmartAsaClient, eve: SigningAccount
) -> None:
    smart_asa = smart_asa_client.state.global_state
    sp = smart_asa_client.algorand.client.algod.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2
    with pytest.raises(LogicError, match=err.UNAUTHORIZED_MANAGER):
        smart_asa_client.send.asset_destroy(
            AssetDestroyArgs(destroy_asset=smart_asa.smart_asa_id),
            params=CommonAppCallParams(
                static_fee=AlgoAmount.from_micro_algo(sp.fee),
                signer=eve.signer,
                sender=eve.address,
            ),
        )


@pytest.mark.parametrize("asa_config", [False], indirect=True)
def test_fail_still_in_circulation(
    smart_asa_client: SmartAsaClient,
    manager: SigningAccount,
    account_with_supply: SigningAccount,  # To have circulating supply
) -> None:
    smart_asa = smart_asa_client.state.global_state
    sp = smart_asa_client.algorand.client.algod.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2
    with pytest.raises(LogicError, match="creator is holding only"):
        smart_asa_client.send.asset_destroy(
            AssetDestroyArgs(destroy_asset=smart_asa.smart_asa_id),
            params=CommonAppCallParams(
                static_fee=AlgoAmount.from_micro_algo(sp.fee),
                signer=manager.signer,
                sender=manager.address,
            ),
        )
