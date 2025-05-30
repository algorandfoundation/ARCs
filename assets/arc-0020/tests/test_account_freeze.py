import pytest
from algokit_utils import CommonAppCallParams, LogicError, SigningAccount

import smart_contracts.errors as err
from smart_contracts.artifacts.smart_asa.smart_asa_client import (
    AccountFreezeArgs,
    SmartAsaClient,
)


def test_pass_change_account_frozen(
    smart_asa_client: SmartAsaClient,
    freeze: SigningAccount,
    receiver: SigningAccount,
) -> None:
    smart_asa = smart_asa_client.state.global_state
    default_frozen = smart_asa.default_frozen
    smart_asa_client.send.account_freeze(
        AccountFreezeArgs(
            freeze_asset=smart_asa.smart_asa_id,
            freeze_account=receiver.address,
            asset_frozen=not default_frozen,
        ),
        params=CommonAppCallParams(
            signer=freeze.signer,
            sender=freeze.address,
        ),
    )
    assert (
        smart_asa_client.state.local_state(receiver.address).account_frozen
        != default_frozen
    )


def test_fail_unauthorized(
    smart_asa_client: SmartAsaClient, eve: SigningAccount, receiver: SigningAccount
) -> None:
    with pytest.raises(LogicError, match=err.UNAUTHORIZED_FREEZE):
        smart_asa_client.send.account_freeze(
            AccountFreezeArgs(
                freeze_asset=smart_asa_client.state.global_state.smart_asa_id,
                freeze_account=receiver.address,
                asset_frozen=True,
            ),
            params=CommonAppCallParams(
                signer=eve.signer,
                sender=eve.address,
            ),
        )


def test_fail_invalid_current_ctrl_asa() -> None:
    pass  # TODO
