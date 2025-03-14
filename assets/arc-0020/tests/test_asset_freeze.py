import pytest
from algokit_utils import CommonAppCallParams, LogicError, SigningAccount

import smart_contracts.errors as err
from smart_contracts.artifacts.smart_asa.smart_asa_client import (
    AssetFreezeArgs,
    SmartAsaClient,
)


@pytest.mark.parametrize("smart_asa_client", [False], indirect=True)
def test_pass_global_frozen(
    smart_asa_client: SmartAsaClient, freeze: SigningAccount
) -> None:
    smart_asa = smart_asa_client.state.global_state
    assert not smart_asa.global_frozen
    smart_asa_client.send.asset_freeze(
        AssetFreezeArgs(
            freeze_asset=smart_asa.smart_asa_id,
            asset_frozen=True,
        ),
        params=CommonAppCallParams(
            signer=freeze.signer,
            sender=freeze.address,
        ),
    )
    assert smart_asa_client.state.global_state.global_frozen


def test_fail_unauthorized(
    smart_asa_client: SmartAsaClient, eve: SigningAccount
) -> None:
    with pytest.raises(LogicError, match=err.UNAUTHORIZED_FREEZE):
        smart_asa_client.send.asset_freeze(
            AssetFreezeArgs(
                freeze_asset=smart_asa_client.state.global_state.smart_asa_id,
                asset_frozen=True,
            ),
            params=CommonAppCallParams(
                signer=eve.signer,
                sender=eve.address,
            ),
        )
