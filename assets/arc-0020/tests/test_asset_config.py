import pytest
from algokit_utils import LogicError, OnCompleteCallParameters
from algokit_utils.beta.account_manager import AddressAndSigner
from algosdk.constants import ZERO_ADDRESS
from algosdk.encoding import encode_address

import smart_contracts.errors as err
from smart_contracts.artifacts.smart_asa.smart_asa_client import SmartAsaClient

from .conftest import ASAConfig

ASA_CONFIG = ASAConfig(
    total=42,
    decimals=69,
    default_frozen=True,
    unit_name="FOO",
    name="Foo Asset",
    url="ipfs://config",
    metadata_hash=(420).to_bytes(length=32),
    manager_addr=ZERO_ADDRESS,
    reserve_addr=ZERO_ADDRESS,
    freeze_addr=ZERO_ADDRESS,
    clawback_addr=ZERO_ADDRESS,
)


def test_pass_asset_config(
    smart_asa_client: SmartAsaClient, manager: AddressAndSigner
) -> None:
    smart_asa_client.asset_config(
        config_asset=smart_asa_client.get_global_state().smart_asa_id,
        **ASA_CONFIG.dictify(),
        transaction_parameters=OnCompleteCallParameters(
            sender=manager.address, signer=manager.signer
        ),
    )

    state = smart_asa_client.get_global_state()

    assert state.total == ASA_CONFIG.total
    assert state.decimals == ASA_CONFIG.decimals
    assert state.default_frozen == ASA_CONFIG.default_frozen
    assert state.unit_name.as_str == ASA_CONFIG.unit_name
    assert state.name.as_str == ASA_CONFIG.name
    assert state.url.as_str == ASA_CONFIG.url
    assert state.metadata_hash.as_bytes == ASA_CONFIG.metadata_hash
    assert encode_address(state.manager_addr.as_bytes) == ASA_CONFIG.manager_addr
    assert encode_address(state.reserve_addr.as_bytes) == ASA_CONFIG.reserve_addr
    assert encode_address(state.freeze_addr.as_bytes) == ASA_CONFIG.freeze_addr
    assert encode_address(state.clawback_addr.as_bytes) == ASA_CONFIG.clawback_addr


def test_fail_missing_ctrl_asa(
    smart_asa_client_no_asset: SmartAsaClient, manager: AddressAndSigner, dummy_asa: int
) -> None:
    with pytest.raises(LogicError, match=err.MISSING_CTRL_ASA):
        smart_asa_client_no_asset.asset_config(
            config_asset=dummy_asa,
            **ASA_CONFIG.dictify(),
            transaction_parameters=OnCompleteCallParameters(
                sender=manager.address, signer=manager.signer
            ),
        )


def test_fail_invalid_ctrl_asa(
    smart_asa_client: SmartAsaClient, manager: AddressAndSigner, dummy_asa: int
) -> None:
    with pytest.raises(LogicError, match=err.INVALID_CTRL_ASA):
        smart_asa_client.asset_config(
            config_asset=dummy_asa,
            **ASA_CONFIG.dictify(),
            transaction_parameters=OnCompleteCallParameters(
                sender=manager.address, signer=manager.signer
            ),
        )


def test_fail_unauthorized_manager(smart_asa_client: SmartAsaClient) -> None:
    with pytest.raises(LogicError, match=err.UNAUTHORIZED_MANAGER):
        smart_asa_client.asset_config(
            config_asset=smart_asa_client.get_global_state().smart_asa_id,
            **ASA_CONFIG.dictify(),
        )


def test_fail_disabled_roles(
    smart_asa_client: SmartAsaClient, manager: AddressAndSigner
) -> None:
    smart_asa_id = smart_asa_client.get_global_state().smart_asa_id
    asa_config = ASA_CONFIG.dictify()
    asa_config["manager_addr"] = manager.address

    smart_asa_client.asset_config(
        config_asset=smart_asa_id,
        **asa_config,
        transaction_parameters=OnCompleteCallParameters(
            sender=manager.address, signer=manager.signer
        ),
    )

    asa_config["reserve_addr"] = smart_asa_client.app_address
    with pytest.raises(LogicError, match=err.DISABLED_RESERVE):
        smart_asa_client.asset_config(
            config_asset=smart_asa_client.get_global_state().smart_asa_id,
            **asa_config,
            transaction_parameters=OnCompleteCallParameters(
                sender=manager.address, signer=manager.signer
            ),
        )
    asa_config["reserve_addr"] = ASA_CONFIG.reserve_addr

    asa_config["freeze_addr"] = smart_asa_client.app_address
    with pytest.raises(LogicError, match=err.DISABLED_FREEZE):
        smart_asa_client.asset_config(
            config_asset=smart_asa_client.get_global_state().smart_asa_id,
            **asa_config,
            transaction_parameters=OnCompleteCallParameters(
                sender=manager.address, signer=manager.signer
            ),
        )
    asa_config["freeze_addr"] = ASA_CONFIG.freeze_addr

    asa_config["clawback_addr"] = smart_asa_client.app_address
    with pytest.raises(LogicError, match=err.DISABLED_CLAWBACK):
        smart_asa_client.asset_config(
            config_asset=smart_asa_client.get_global_state().smart_asa_id,
            **asa_config,
            transaction_parameters=OnCompleteCallParameters(
                sender=manager.address, signer=manager.signer
            ),
        )


def test_fail_invalid_total() -> None:
    # TODO: Mint something
    pass  # TODO
