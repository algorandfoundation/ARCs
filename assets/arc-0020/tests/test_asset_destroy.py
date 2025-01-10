import pytest
from algokit_utils import LogicError, OnCompleteCallParameters
from algokit_utils.beta.account_manager import AddressAndSigner
from algosdk.constants import ZERO_ADDRESS
from algosdk.encoding import encode_address
from algosdk.error import AlgodHTTPError

import smart_contracts.errors as err
from smart_contracts.artifacts.smart_asa.smart_asa_client import SmartAsaClient


def test_pass_destroy(
    smart_asa_client: SmartAsaClient, manager: AddressAndSigner
) -> None:
    smart_asa = smart_asa_client.get_global_state()
    sp = smart_asa_client.algod_client.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2
    smart_asa_client.asset_destroy(
        destroy_asset=smart_asa.smart_asa_id,
        transaction_parameters=OnCompleteCallParameters(
            suggested_params=sp,
            signer=manager.signer,
            sender=manager.address,
        ),
    )
    with pytest.raises(AlgodHTTPError, match="asset does not exist"):
        smart_asa_client.algod_client.asset_info(smart_asa.smart_asa_id)
    smart_asa = smart_asa_client.get_global_state()
    assert smart_asa.total == 0
    assert smart_asa.decimals == 0
    assert not smart_asa.default_frozen
    assert smart_asa.unit_name.as_str == ""
    assert smart_asa.name.as_str == ""
    assert smart_asa.url.as_str == ""
    # assert smart_asa.metadata_hash.as_bytes == bytes(HASH_LEN)  # FIXME: Once default init is enabled
    assert encode_address(smart_asa.manager_addr.as_bytes) == ZERO_ADDRESS
    assert encode_address(smart_asa.reserve_addr.as_bytes) == ZERO_ADDRESS
    assert encode_address(smart_asa.freeze_addr.as_bytes) == ZERO_ADDRESS
    assert encode_address(smart_asa.clawback_addr.as_bytes) == ZERO_ADDRESS
    assert smart_asa.smart_asa_id == 0
    assert not smart_asa.global_frozen


def test_fail_missing_ctrl_asa() -> None:
    pass  # TODO


def test_fail_invalid_ctrl_asa() -> None:
    pass  # TODO


def test_fail_unauthorized_manager(
    smart_asa_client: SmartAsaClient, eve: AddressAndSigner
) -> None:
    smart_asa = smart_asa_client.get_global_state()
    sp = smart_asa_client.algod_client.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2
    with pytest.raises(LogicError, match=err.UNAUTHORIZED_MANAGER):
        smart_asa_client.asset_destroy(
            destroy_asset=smart_asa.smart_asa_id,
            transaction_parameters=OnCompleteCallParameters(
                suggested_params=sp,
                signer=eve.signer,
                sender=eve.address,
            ),
        )


@pytest.mark.parametrize("asa_config", [False], indirect=True)
def test_fail_still_in_circulation(
    smart_asa_client: SmartAsaClient,
    manager: AddressAndSigner,
    account_with_supply: AddressAndSigner,  # To have circulating supply
) -> None:
    smart_asa = smart_asa_client.get_global_state()
    sp = smart_asa_client.algod_client.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2
    with pytest.raises(LogicError, match="creator is holding only"):
        smart_asa_client.asset_destroy(
            destroy_asset=smart_asa.smart_asa_id,
            transaction_parameters=OnCompleteCallParameters(
                suggested_params=sp,
                signer=manager.signer,
                sender=manager.address,
            ),
        )
