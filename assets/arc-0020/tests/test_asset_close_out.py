import pytest
from algokit_utils import OnCompleteCallParameters
from algokit_utils.beta.account_manager import AddressAndSigner
from algokit_utils.beta.algorand_client import AlgorandClient, AssetTransferParams
from algosdk.atomic_transaction_composer import TransactionWithSigner

from smart_contracts.artifacts.smart_asa.smart_asa_client import SmartAsaClient

from . import utils


@pytest.mark.parametrize("asa_config", [False], indirect=True)
def test_pass_regular_close_out(
    algorand_client: AlgorandClient,
    smart_asa_client: SmartAsaClient,
    account_with_supply: AddressAndSigner,
    receiver: AddressAndSigner,
) -> None:
    smart_asa = smart_asa_client.get_global_state()
    smart_asa_id = smart_asa.smart_asa_id
    account_asset_balance = utils.get_account_asset_balance(
        algorand_client.client.algod, account_with_supply.address, smart_asa_id
    )
    receiver_asset_balance = utils.get_account_asset_balance(
        algorand_client.client.algod, receiver.address, smart_asa_id
    )
    assert account_asset_balance == smart_asa.total
    assert receiver_asset_balance == 0
    sp = smart_asa_client.algod_client.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2
    close_out = smart_asa_client.compose().close_out_asset_close_out(
        close_asset=smart_asa_id,
        close_to=receiver.address,
        transaction_parameters=OnCompleteCallParameters(
            suggested_params=sp,
            signer=account_with_supply.signer,
            sender=account_with_supply.address,
        ),
    )
    close_out.atc.add_transaction(
        TransactionWithSigner(
            txn=algorand_client.transactions.asset_transfer(
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
    )
    close_out.atc.submit(algorand_client.client.algod)
    receiver_asset_balance = utils.get_account_asset_balance(
        algorand_client.client.algod, receiver.address, smart_asa_id
    )
    assert not utils.is_account_opted_in(
        algorand_client.client.algod, account_with_supply.address, smart_asa_id
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
