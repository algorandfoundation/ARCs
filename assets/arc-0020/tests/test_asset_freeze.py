import pytest
from algokit_utils import LogicError, OnCompleteCallParameters
from algokit_utils.beta.account_manager import AddressAndSigner

import smart_contracts.errors as err
from smart_contracts.artifacts.smart_asa.smart_asa_client import SmartAsaClient


@pytest.mark.parametrize("smart_asa_client", [False], indirect=True)
def test_pass_global_frozen(
    smart_asa_client: SmartAsaClient, freeze: AddressAndSigner
) -> None:
    smart_asa = smart_asa_client.get_global_state()
    assert not smart_asa.global_frozen
    smart_asa_client.asset_freeze(
        freeze_asset=smart_asa.smart_asa_id,
        asset_frozen=True,
        transaction_parameters=OnCompleteCallParameters(
            signer=freeze.signer,
            sender=freeze.address,
        ),
    )
    assert smart_asa_client.get_global_state().global_frozen


def test_fail_unauthorized(
    smart_asa_client: SmartAsaClient, eve: AddressAndSigner
) -> None:
    with pytest.raises(LogicError, match=err.UNAUTHORIZED_FREEZE):
        smart_asa_client.asset_freeze(
            freeze_asset=smart_asa_client.get_global_state().smart_asa_id,
            asset_frozen=True,
            transaction_parameters=OnCompleteCallParameters(
                signer=eve.signer,
                sender=eve.address,
            ),
        )
