import pytest
from algokit_utils import AlgoAmount, CommonAppCallParams, LogicError, SigningAccount
from algosdk.encoding import encode_address

import smart_contracts.errors as err
import smart_contracts.smart_asa.config as cfg
from smart_contracts.artifacts.smart_asa.smart_asa_client import (
    AssetCreateArgs,
    SmartAsaClient,
)

from .conftest import ASAConfig


def test_pass_asset_create(
    smart_asa_client_no_asset: SmartAsaClient,
    asa_config: ASAConfig,
) -> None:
    sp = smart_asa_client_no_asset.algorand.client.algod.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2

    smart_asa_id = smart_asa_client_no_asset.send.asset_create(
        AssetCreateArgs(**asa_config.dictify()),
        params=CommonAppCallParams(static_fee=AlgoAmount.from_micro_algo(sp.fee)),
    ).abi_return

    # Verify Controlled ASA
    ctrl_asset = smart_asa_client_no_asset.algorand.asset.get_by_id(smart_asa_id)
    assert ctrl_asset.asset_id == smart_asa_id
    assert ctrl_asset.creator == smart_asa_client_no_asset.app_address
    assert ctrl_asset.total == cfg.TOTAL
    assert ctrl_asset.decimals == cfg.DECIMALS
    assert ctrl_asset.default_frozen
    assert ctrl_asset.unit_name == cfg.UNIT_NAME
    assert ctrl_asset.asset_name == cfg.NAME
    assert ctrl_asset.url == cfg.APP_BINDING.decode() + str(
        smart_asa_client_no_asset.app_id
    )

    # Verify Smart ASA
    state = smart_asa_client_no_asset.state.global_state
    assert state.total == asa_config.total
    assert state.decimals == asa_config.decimals
    assert state.default_frozen == asa_config.default_frozen
    assert state.unit_name == asa_config.unit_name
    assert state.name == asa_config.name
    assert state.url == asa_config.url
    assert state.metadata_hash == asa_config.metadata_hash
    assert encode_address(state.manager_addr) == asa_config.manager_addr
    assert encode_address(state.reserve_addr) == asa_config.reserve_addr
    assert encode_address(state.freeze_addr) == asa_config.freeze_addr
    assert encode_address(state.clawback_addr) == asa_config.clawback_addr
    assert state.smart_asa_id == smart_asa_id


def test_fail_unauthorized(
    smart_asa_client_no_asset: SmartAsaClient,
    eve: SigningAccount,
    asa_config: ASAConfig,
) -> None:
    sp = smart_asa_client_no_asset.algorand.client.algod.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2

    with pytest.raises(LogicError, match=err.UNAUTHORIZED):
        smart_asa_client_no_asset.send.asset_create(
            AssetCreateArgs(**asa_config.dictify()),
            params=CommonAppCallParams(
                static_fee=AlgoAmount.from_micro_algo(sp.fee),
                signer=eve.signer,
                sender=eve.address,
            ),
        )


def test_fail_asa_already_created(
    smart_asa_client: SmartAsaClient, asa_config: ASAConfig
) -> None:
    sp = smart_asa_client.algorand.client.algod.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2

    with pytest.raises(LogicError, match=err.EXISTING_CTRL_ASA):
        smart_asa_client.send.asset_create(
            AssetCreateArgs(**asa_config.dictify()),
            params=CommonAppCallParams(static_fee=AlgoAmount.from_micro_algo(sp.fee)),
        )
