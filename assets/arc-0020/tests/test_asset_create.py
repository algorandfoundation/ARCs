import pytest
from algokit_utils import LogicError, OnCompleteCallParameters
from algokit_utils.beta.account_manager import AddressAndSigner
from algosdk.encoding import encode_address

import smart_contracts.errors as err
import smart_contracts.smart_asa.config as cfg
from smart_contracts.artifacts.smart_asa.smart_asa_client import SmartAsaClient

from .conftest import ASAConfig


def test_pass_asset_create(
    smart_asa_client_no_asset: SmartAsaClient,
    asa_config: ASAConfig,
) -> None:
    sp = smart_asa_client_no_asset.algod_client.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2

    smart_asa_id = smart_asa_client_no_asset.asset_create(
        **asa_config.dictify(),
        transaction_parameters=OnCompleteCallParameters(
            suggested_params=sp,
        ),
    ).return_value

    # Verify Controlled ASA
    ctrl_asset = smart_asa_client_no_asset.algod_client.asset_info(smart_asa_id)
    assert ctrl_asset["index"] == smart_asa_id
    assert ctrl_asset["params"]["creator"] == smart_asa_client_no_asset.app_address
    assert ctrl_asset["params"]["total"] == cfg.TOTAL
    assert ctrl_asset["params"]["decimals"] == cfg.DECIMALS
    assert ctrl_asset["params"]["default-frozen"]
    assert ctrl_asset["params"]["unit-name"] == cfg.UNIT_NAME
    assert ctrl_asset["params"]["name"] == cfg.NAME
    assert ctrl_asset["params"]["url"] == cfg.APP_BINDING.decode() + str(
        smart_asa_client_no_asset.app_id
    )

    # Verify Smart ASA
    state = smart_asa_client_no_asset.get_global_state()
    assert state.total == asa_config.total
    assert state.decimals == asa_config.decimals
    assert state.default_frozen == asa_config.default_frozen
    assert state.unit_name.as_str == asa_config.unit_name
    assert state.name.as_str == asa_config.name
    assert state.url.as_str == asa_config.url
    # assert state.metadata_hash.as_bytes == asa_config.metadata_hash  # FIXME: Once default init is enabled
    assert encode_address(state.manager_addr.as_bytes) == asa_config.manager_addr
    assert encode_address(state.reserve_addr.as_bytes) == asa_config.reserve_addr
    assert encode_address(state.freeze_addr.as_bytes) == asa_config.freeze_addr
    assert encode_address(state.clawback_addr.as_bytes) == asa_config.clawback_addr
    assert state.smart_asa_id == smart_asa_id


def test_fail_unauthorized(
    smart_asa_client_no_asset: SmartAsaClient,
    eve: AddressAndSigner,
    asa_config: ASAConfig,
) -> None:
    sp = smart_asa_client_no_asset.algod_client.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2

    with pytest.raises(LogicError, match=err.UNAUTHORIZED):
        smart_asa_client_no_asset.asset_create(
            **asa_config.dictify(),
            transaction_parameters=OnCompleteCallParameters(
                suggested_params=sp,
                signer=eve.signer,
                sender=eve.address,
            ),
        )


def test_fail_asa_already_created(
    smart_asa_client: SmartAsaClient, asa_config: ASAConfig
) -> None:
    sp = smart_asa_client.algod_client.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2

    with pytest.raises(LogicError, match=err.EXISTING_CTRL_ASA):
        smart_asa_client.asset_create(
            **asa_config.dictify(),
            transaction_parameters=OnCompleteCallParameters(
                suggested_params=sp,
            ),
        )
