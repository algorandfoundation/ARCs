import algokit_utils
import pytest
import base64
from algokit_utils import (
    AlgoAmount,
    AlgorandClient,
    SigningAccount,
)

from smart_contracts.artifacts.recovery.recovery_client import (
    RecoveryClient,
    RecoveryFactory,
)

import logging

logger = logging.getLogger(__name__)

@pytest.fixture()
def deployer(algorand_client: AlgorandClient) -> SigningAccount:
    account = algorand_client.account.from_environment("DEPLOYER")
    algorand_client.account.ensure_funded_from_environment(
        account_to_fund=account.address, min_spending_balance=AlgoAmount.from_algo(10)
    )
    return account

@pytest.fixture()
def retriever(algorand_client: AlgorandClient) -> SigningAccount:
    account = algorand_client.account.from_environment("DEPLOYER")
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
    return client

def test_arc89(algorand_client: AlgorandClient, deployer: SigningAccount, retriever: SigningAccount) -> None:
    # Arrange

    teal_program = """
    #pragma version 11
    txn TypeEnum
    pushint 4
    ==
    txn AssetAmount
    pushint 0
    ==
    &&
    txn AssetCloseTo
    global ZeroAddress
    ==
    &&
    txn Fee
    global MinTxnFee
    ==
    &&
    return
    """

    compiled_program = algorand_client.client.algod.compile(teal_program)
    program = base64.b64decode(compiled_program["result"])
    lsig = algokit_utils.LogicSigAccount(program, None)

    lsig.lsig.sign(deployer.private_key)
    psig = lsig.lsig.lsig.sig
    multisig_params = algokit_utils.MultisigMetadata(
    version=1,
    threshold=2,
    addresses=[deployer.address, retriever.address],
)

    msig = algokit_utils.MultiSigAccount(multisig_params=multisig_params,signing_accounts=[deployer, retriever])
    logger.info(msig.multisig.subsigs)
    logger.info(msig.multisig.address_bytes())
    logger.info(msig.multisig.address())
    msig.multisig.
    
    # Assert
    assert True

def test_says_hello(recovery_client: RecoveryClient) -> None:
    result = recovery_client.send.hello(args=("World",))
    assert result.abi_return == "Hello, World"


def test_simulate_says_hello_with_correct_budget_consumed(
    recovery_client: RecoveryClient,
) -> None:
    result = (
        recovery_client.new_group()
        .hello(args=("World",))
        .hello(args=("Jane",))
        .simulate()
    )
    assert result.returns[0].value == "Hello, World"
    assert result.returns[1].value == "Hello, Jane"
    assert result.simulate_response["txn-groups"][0]["app-budget-consumed"] < 100
