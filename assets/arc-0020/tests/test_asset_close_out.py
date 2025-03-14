import pytest
from algokit_utils import (
    AlgoAmount,
    AlgorandClient,
    AssetTransferParams,
    CommonAppCallParams,
    SigningAccount,
)

from smart_contracts.artifacts.smart_asa.smart_asa_client import (
    AssetCloseOutArgs,
    SmartAsaClient,
)


@pytest.mark.parametrize("asa_config", [False], indirect=True)
def test_pass_regular_close_out(
    algorand: AlgorandClient,
    smart_asa_client: SmartAsaClient,
    account_with_supply: SigningAccount,
    receiver: SigningAccount,
) -> None:
    smart_asa = smart_asa_client.state.global_state
    smart_asa_id = smart_asa.smart_asa_id
    account_asset_balance = algorand.asset.get_account_information(
        account_with_supply, smart_asa_id
    ).balance
    receiver_asset_balance = algorand.asset.get_account_information(
        receiver, smart_asa_id
    ).balance
    assert account_asset_balance == smart_asa.total
    assert receiver_asset_balance == 0
    sp = smart_asa_client.algorand.client.algod.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2
    close_out = smart_asa_client.new_group().close_out.asset_close_out(
        AssetCloseOutArgs(
            close_asset=smart_asa_id,
            close_to=receiver.address,
        ),
        params=CommonAppCallParams(
            static_fee=AlgoAmount.from_micro_algo(sp.fee),
            signer=account_with_supply.signer,
            sender=account_with_supply.address,
        ),
    )
    close_out.add_transaction(
        txn=algorand.create_transaction.asset_transfer(
            AssetTransferParams(
                sender=account_with_supply.address,
                asset_id=smart_asa_id,
                receiver=smart_asa_client.app_address,
                close_asset_to=smart_asa_client.app_address,
                amount=0,
            )
        ),
        signer=account_with_supply.signer,
    )
    close_out.send()
    receiver_asset_balance = algorand.asset.get_account_information(
        receiver, smart_asa_id
    ).balance
    assert (
        smart_asa_id
        not in algorand.account.get_information(account_with_supply.address).assets
    )
    assert receiver_asset_balance == smart_asa.total


def test_pass_close_out_frozen_asset() -> None:
    pass  # TODO


def test_pass_close_out_account_frozen() -> None:
    pass  # TODO


def test_pass_destroyed_asset() -> None:
    pass  # TODO


def test_fail_wrong_on_complete() -> None:
    pass  # TODO


def test_fail_invalid_local_ctrl_asa() -> None:
    pass  # TODO


def test_fail_group_wrong_size() -> None:
    pass  # TODO


def test_fail_group_wrong_txn_type() -> None:
    pass  # TODO


def test_fail_group_wrong_asset() -> None:
    pass  # TODO


def test_fail_group_wrong_sender() -> None:
    pass  # TODO


def test_fail_group_wrong_amount() -> None:
    pass  # TODO


def test_fail_group_wrong_close_to() -> None:
    pass  # TODO


def test_fail_invalid_ctrl_asa() -> None:
    pass  # TODO


def test_fail_close_out_frozen_asset() -> None:
    pass  # TODO


def test_fail_close_out_sender_frozen() -> None:
    pass  # TODO


def test_fail_close_out_target_frozen() -> None:
    pass  # TODO
