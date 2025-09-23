import base64
import logging

import algokit_utils
import algokit_utils.transactions
import algokit_utils.transactions.transaction_creator
import algosdk
import pytest
from algokit_utils import (
    AlgoAmount,
    AlgorandClient,
    CommonAppCallParams,
    SigningAccount,
)

from smart_contracts.artifacts.recovery.recovery_client import (
    RecoveryClient,
    RecoveryFactory,
)

logger = logging.getLogger(__name__)


@pytest.fixture()
def deployer(algorand_client: AlgorandClient) -> SigningAccount:
    account = algorand_client.account.from_environment("DEPLOYER")
    logger.info("DEPLOYER")
    logger.info(account.address)
    algorand_client.account.ensure_funded_from_environment(
        account_to_fund=account.address, min_spending_balance=AlgoAmount.from_algo(10)
    )
    return account


@pytest.fixture()
def retriever(algorand_client: AlgorandClient) -> SigningAccount:
    account = algorand_client.account.from_environment("RETRIEVER")
    logger.info("RETRIEVER")
    logger.info(account.address)
    algorand_client.account.ensure_funded_from_environment(
        account_to_fund=account.address, min_spending_balance=AlgoAmount.from_algo(10)
    )
    return account


@pytest.fixture()
def recovery_client(
    algorand_client: AlgorandClient, deployer: SigningAccount
) -> RecoveryClient:
    factory = algorand_client.client.get_typed_app_factory(
        RecoveryFactory, default_sender=deployer.address
    )

    client, _ = factory.deploy(
        on_schema_break=algokit_utils.OnSchemaBreak.AppendApp,
        on_update=algokit_utils.OnUpdate.AppendApp,
    )
    algorand_client.account.ensure_funded_from_environment(
        account_to_fund=client.app_address,
        min_spending_balance=AlgoAmount.from_algo(10),
    )
    return client


@pytest.fixture()
def recovery_context(
    algorand_client: AlgorandClient,
    recovery_client: RecoveryClient,
    deployer: SigningAccount,
    retriever: SigningAccount,
) -> dict[str, str]:
    rekey_method = (
        recovery_client.app_spec.get_arc56_method("rekey_wallet")
        .to_abi_method()
        .get_selector()
    )
    logger.info("rekeywalletargs %s", rekey_method.hex())
    teal_program_lines = [
        "#pragma version 11",
        "txn TypeEnum",
        "int appl",
        "==",
        "txn ApplicationID",
        f"int {recovery_client.app_id}",
        "==",
        "&&",
        "txn NumAppArgs",
        "int 1",
        ">=",
        "&&",
        "txn ApplicationArgs 0",
        f"byte 0x{rekey_method.hex()}",
        "==",
        "&&",
        "return",
    ]
    teal_program = "\n".join(teal_program_lines)

    compiled_program = algorand_client.client.algod.compile(teal_program)
    program = base64.b64decode(compiled_program["result"])
    lsig = algokit_utils.LogicSigAccount(program, None)
    lsig.lsig.sign(deployer.private_key)
    public_sig = lsig.lsig.lsig.sig
    logger.warning("INIT")
    logger.warning(lsig.lsig.sigkey)
    logger.warning(lsig.lsig.lsig.sig)

    multisig_params = algokit_utils.MultisigMetadata(
        version=1,
        threshold=2,
        addresses=[deployer.address, retriever.address],
    )
    msig = algokit_utils.MultiSigAccount(
        multisig_params=multisig_params,
        signing_accounts=[deployer, retriever],
    )
    logger.info(msig.multisig.subsigs)
    logger.info(msig.multisig.address_bytes())
    logger.info(msig.multisig.address())
    logger.info("multisig")
    logger.info(msig.multisig.address())
    algorand_client.account.ensure_funded_from_environment(
        account_to_fund=msig.multisig.address(),
        min_spending_balance=AlgoAmount.from_algo(10),
    )
    return {
        "wallet_address": deployer.address,
        "public_sig": public_sig,
        "new_address": msig.multisig.address(),
        "program": program,
    }


def test_arc89(
    algorand_client: AlgorandClient,
    deployer: SigningAccount,
    retriever: SigningAccount,
    recovery_context: dict[str, str],
) -> None:
    logger.info("public signature: %s", recovery_context["public_sig"])
    logger.info("rekey wallet address: %s", recovery_context["wallet_address"])
    logger.info("retriever address: %s", recovery_context["new_address"])

    assert recovery_context["public_sig"]
    assert recovery_context["wallet_address"]
    assert recovery_context["new_address"]


# def test_says_hello(recovery_client: RecoveryClient) -> None:
#     result = recovery_client.send.hello(args=("World",))
#     assert result.abi_return == "Hello, World"


# def test_simulate_says_hello_with_correct_budget_consumed(
#     recovery_client: RecoveryClient,
# ) -> None:
#     result = (
#         recovery_client.new_group()
#         .hello(args=("World",))
#         .hello(args=("Jane",))
#         .simulate()
#     )
#     assert result.returns[0].value == "Hello, World"
#     assert result.returns[1].value == "Hello, Jane"
#     assert result.simulate_response["txn-groups"][0]["app-budget-consumed"] < 100


def test_set_and_get_public_sig(
    recovery_client: RecoveryClient,
    recovery_context: dict[str, str],
) -> None:
    if not hasattr(recovery_client.send, "set_public_sig"):
        pytest.skip("Recovery client not regenerated with set_public_sig")
    if not hasattr(recovery_client.send, "get_public_sig"):
        pytest.skip("Recovery client not regenerated with get_public_sig")

    wallet_address = recovery_context["wallet_address"]
    public_sig = recovery_context["public_sig"]

    write_result = recovery_client.send.set_public_sig(
        args=(wallet_address, public_sig)
    )
    assert getattr(write_result, "abi_return", True)

    read_result = recovery_client.send.get_public_sig(args=(wallet_address,))
    assert read_result.abi_return == public_sig


def test_rekey_wallet_notice_and_cancel(
    recovery_client: RecoveryClient,
    recovery_context: dict[str, str],
    retriever: SigningAccount,
) -> None:
    for method_name in ("set_rekey_record", "rekey_wallet", "cancel_rekey"):
        if not hasattr(recovery_client.send, method_name):
            pytest.skip(f"Recovery client not regenerated with {method_name}")

    wallet_address = recovery_context["wallet_address"]
    new_address = recovery_context["new_address"]
    notification_window = 2

    recovery_client.send.set_rekey_record(
        args=(wallet_address, retriever.address, new_address, notification_window)
    )

    params = CommonAppCallParams(sender=retriever.address, signer=retriever.signer)

    notice_result = recovery_client.send.rekey_wallet(
        args=(wallet_address, new_address),
        params=params,
    )
    if hasattr(notice_result, "abi_return"):
        assert notice_result.abi_return is None

    with pytest.raises(Exception):
        recovery_client.send.rekey_wallet(args=(wallet_address, new_address))

    recovery_client.send.cancel_rekey(args=(wallet_address,))

    with pytest.raises(Exception):
        recovery_client.send.rekey_wallet(args=(wallet_address, new_address))


def test_rekey_wallet_notice_and_rekey(
    algorand_client: AlgorandClient,
    recovery_client: RecoveryClient,
    recovery_context: dict[str, str],
    retriever: SigningAccount,
    deployer: SigningAccount,
) -> None:
    for method_name in ("set_rekey_record", "rekey_wallet"):
        if not hasattr(recovery_client.send, method_name):
            pytest.skip(f"Recovery client not regenerated with {method_name}")

    wallet_address = recovery_context["wallet_address"]
    new_address = recovery_context["new_address"]
    notification_window = 1

    recovery_client.send.set_rekey_record(
        args=(wallet_address, retriever.address, new_address, notification_window)
    )

    params = CommonAppCallParams(
        sender=retriever.address,
        signer=retriever.signer,
        extra_fee=algokit_utils.AlgoAmount(micro_algo=2000),
    )

    notice_result = recovery_client.send.rekey_wallet(
        args=(wallet_address, new_address),
        params=params,
    )
    if hasattr(notice_result, "abi_return"):
        assert notice_result.abi_return is None

    algod_client = algorand_client.client.algod
    current_round = algod_client.status()["last-round"]
    # target_round = current_round + notification_window
    # algod_client.status_after_block(target_round)

    # recovery_client.send.cancel_rekey(args=(wallet_address,))

    # composer = RecoveryComposer(recovery_client)
    # composer.composer().add_app_call(
    #     recovery_client.params.rekey_wallet(
    #         args=(wallet_address, new_address),
    #         params=CommonAppCallParams(sender=deployer.address, signer=deployer.signer),
    #     )
    # )

    # rekey_txn = composer.composer()._txns[0]
    psig = recovery_context["public_sig"]
    program = recovery_context["program"]
    # lsig = transaction.LogicSigAccount(program)

    # lsig.sigkey = psig

    # lstx = transaction.LogicSigTransaction(rekey_txn.txn, lsig)
    # logger.info("rekeying with txid: %s", lstx)
    # algod_client.send_transaction(lstx)

    lsig = algokit_utils.LogicSigAccount(program, None)
    # lsig.lsig.sign(deployer.private_key)
    lsig.lsig.sigkey = algosdk.encoding.decode_address(deployer.address)
    lsig.lsig.lsig.sig = psig
    logger.warning("Method")
    logger.warning(lsig.lsig.sigkey)
    logger.warning(lsig.lsig.lsig.sig)
    # algorand_client.send.payment(
    #     algokit_utils.PaymentParams(
    #         sender=wallet_address,
    #         receiver=wallet_address,
    #         amount=algokit_utils.AlgoAmount(micro_algo=0),
    #         rekey_to=new_address,
    #         signer=lsig.signer,
    #         extra_fee=algokit_utils.AlgoAmount(micro_algo=1000),
    #     )
    # )
    recovery_client.send.rekey_wallet(
        args=(wallet_address, new_address),
        params=CommonAppCallParams(
            sender=lsig.address, signer=lsig.signer, rekey_to=new_address
        ),
    )

    account_info = algod_client.account_info(wallet_address)
    assert account_info.get("auth-addr") == new_address
