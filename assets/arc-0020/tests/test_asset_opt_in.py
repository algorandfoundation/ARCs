import pytest
from algokit_utils import (
    AlgorandClient,
    AssetOptInParams,
    AssetTransferParams,
    LogicError,
    SigningAccount,
)
from algosdk.atomic_transaction_composer import TransactionWithSigner

import smart_contracts.errors as err
from smart_contracts.artifacts.smart_asa.smart_asa_client import (
    AssetOptInArgs,
    SmartAsaClient,
)


def test_pass_asset_opt_in(
    algorand: AlgorandClient,
    smart_asa_client: SmartAsaClient,
    creator: SigningAccount,
) -> None:
    state = smart_asa_client.state.global_state
    smart_asa_id = state.smart_asa_id
    smart_asa_client.send.opt_in.asset_opt_in(
        AssetOptInArgs(
            asset=smart_asa_id,
            ctrl_asa_opt_in=TransactionWithSigner(
                txn=algorand.create_transaction.asset_opt_in(
                    AssetOptInParams(sender=creator.address, asset_id=smart_asa_id)
                ),
                signer=creator.signer,
            ),
        ),
    )

    local_state = smart_asa_client.state.local_state(creator.address)
    assert local_state.account_smart_asa_id == smart_asa_id
    if state.default_frozen:
        assert local_state.account_frozen
    else:
        assert not local_state.account_frozen


def test_fail_missing_ctrl_asa(
    algorand: AlgorandClient,
    smart_asa_client_no_asset: SmartAsaClient,
    creator: SigningAccount,
    dummy_asa: int,
) -> None:
    with pytest.raises(LogicError, match=err.MISSING_CTRL_ASA):
        smart_asa_client_no_asset.send.opt_in.asset_opt_in(
            AssetOptInArgs(
                asset=dummy_asa,
                ctrl_asa_opt_in=TransactionWithSigner(
                    txn=algorand.create_transaction.asset_opt_in(
                        AssetOptInParams(sender=creator.address, asset_id=dummy_asa)
                    ),
                    signer=creator.signer,
                ),
            ),
        )


def test_fail_invalid_ctrl_asa(
    algorand: AlgorandClient,
    smart_asa_client: SmartAsaClient,
    creator: SigningAccount,
    dummy_asa: int,
) -> None:
    smart_asa_id = smart_asa_client.state.global_state.smart_asa_id
    with pytest.raises(LogicError, match=err.INVALID_CTRL_ASA):
        smart_asa_client.send.opt_in.asset_opt_in(
            AssetOptInArgs(
                asset=dummy_asa,
                ctrl_asa_opt_in=TransactionWithSigner(
                    txn=algorand.create_transaction.asset_opt_in(
                        AssetOptInParams(sender=creator.address, asset_id=smart_asa_id)
                    ),
                    signer=creator.signer,
                ),
            ),
        )


def test_fail_opt_in_wrong_type(
    smart_asa_client: SmartAsaClient, creator: SigningAccount
) -> None:
    pass  # TODO: Require using low level SDK since the ATC catch the incorrect txn type error first.


def test_fail_opt_in_wrong_asa(
    algorand: AlgorandClient,
    smart_asa_client: SmartAsaClient,
    creator: SigningAccount,
    dummy_asa: int,
) -> None:
    smart_asa_id = smart_asa_client.state.global_state.smart_asa_id
    with pytest.raises(LogicError, match=err.OPT_IN_WRONG_ASA):
        smart_asa_client.send.opt_in.asset_opt_in(
            AssetOptInArgs(
                asset=smart_asa_id,
                ctrl_asa_opt_in=TransactionWithSigner(
                    txn=algorand.create_transaction.asset_opt_in(
                        AssetOptInParams(sender=creator.address, asset_id=dummy_asa)
                    ),
                    signer=creator.signer,
                ),
            ),
        )


def test_fail_opt_in_wrong_sender(
    algorand: AlgorandClient,
    smart_asa_client: SmartAsaClient,
    eve: SigningAccount,
) -> None:
    smart_asa_id = smart_asa_client.state.global_state.smart_asa_id
    with pytest.raises(LogicError, match=err.OPT_IN_WRONG_SENDER):
        smart_asa_client.send.opt_in.asset_opt_in(
            AssetOptInArgs(
                asset=smart_asa_id,
                ctrl_asa_opt_in=TransactionWithSigner(
                    txn=algorand.create_transaction.asset_opt_in(
                        AssetOptInParams(sender=eve.address, asset_id=smart_asa_id)
                    ),
                    signer=eve.signer,
                ),
            ),
        )


def test_fail_opt_in_wrong_receiver(
    algorand: AlgorandClient,
    smart_asa_client: SmartAsaClient,
    creator: SigningAccount,
    eve: SigningAccount,
) -> None:
    smart_asa_id = smart_asa_client.state.global_state.smart_asa_id
    with pytest.raises(LogicError, match=err.OPT_IN_WRONG_RECEIVER):
        smart_asa_client.send.opt_in.asset_opt_in(
            AssetOptInArgs(
                asset=smart_asa_id,
                ctrl_asa_opt_in=TransactionWithSigner(
                    txn=algorand.create_transaction.asset_transfer(
                        AssetTransferParams(
                            sender=creator.address,
                            asset_id=smart_asa_id,
                            receiver=eve.address,
                            amount=0,
                        )
                    ),
                    signer=creator.signer,
                ),
            ),
        )


# NOTE: Test opt-in with wrong amount skipped since the Controlled ASA is default frozen, any transfer > 0 would fail.


def test_fail_opt_in_forbidden_close_out(
    algorand: AlgorandClient,
    smart_asa_client: SmartAsaClient,
    creator: SigningAccount,
) -> None:
    smart_asa_id = smart_asa_client.state.global_state.smart_asa_id
    with pytest.raises(LogicError, match=err.OPT_IN_WRONG_CLOSE_TO):
        smart_asa_client.send.opt_in.asset_opt_in(
            AssetOptInArgs(
                asset=smart_asa_id,
                ctrl_asa_opt_in=TransactionWithSigner(
                    txn=algorand.create_transaction.asset_transfer(
                        AssetTransferParams(
                            sender=creator.address,
                            asset_id=smart_asa_id,
                            receiver=creator.address,
                            amount=0,
                            close_asset_to=smart_asa_client.app_address,
                        )
                    ),
                    signer=creator.signer,
                ),
            ),
        )


def test_fail_wrong_on_complete() -> None:
    pass  # TODO: Once txn `OnComplete` can be passed as `OnCompleteCallParameters` to AppClient
