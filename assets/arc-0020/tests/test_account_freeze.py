import pytest
from algokit_utils import LogicError, OnCompleteCallParameters
from algokit_utils.beta.account_manager import AddressAndSigner

import smart_contracts.errors as err
from smart_contracts.artifacts.smart_asa.smart_asa_client import SmartAsaClient


def test_pass_change_account_frozen(
    smart_asa_client: SmartAsaClient,
    freeze: AddressAndSigner,
    receiver: AddressAndSigner,
) -> None:
    smart_asa = smart_asa_client.get_global_state()
    default_frozen = smart_asa.default_frozen
    smart_asa_client.account_freeze(
        freeze_asset=smart_asa.smart_asa_id,
        freeze_account=receiver.address,
        asset_frozen=not default_frozen,
        transaction_parameters=OnCompleteCallParameters(
            signer=freeze.signer,
            sender=freeze.address,
        ),
    )
    assert (
        smart_asa_client.get_local_state(receiver.address).account_frozen
        != default_frozen
    )


def test_fail_unauthorized(
    smart_asa_client: SmartAsaClient, eve: AddressAndSigner, receiver: AddressAndSigner
) -> None:
    with pytest.raises(LogicError, match=err.UNAUTHORIZED_FREEZE):
        smart_asa_client.account_freeze(
            freeze_asset=smart_asa_client.get_global_state().smart_asa_id,
            freeze_account=receiver.address,
            asset_frozen=True,
            transaction_parameters=OnCompleteCallParameters(
                signer=eve.signer,
                sender=eve.address,
            ),
        )


def test_fail_invalid_current_ctrl_asa() -> None:
    pass  # TODO
