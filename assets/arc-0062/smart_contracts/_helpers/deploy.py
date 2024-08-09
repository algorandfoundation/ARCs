# mypy: disable-error-code="no-untyped-call, misc"


import logging
from collections.abc import Callable
from pathlib import Path

from algokit_utils import (
    Account,
    ApplicationSpecification,
    EnsureBalanceParameters,
    ensure_funded,
    get_account,
    get_algod_client,
    get_indexer_client,
)
from algosdk.util import algos_to_microalgos
from algosdk.v2client.algod import AlgodClient
from algosdk.v2client.indexer import IndexerClient

logger = logging.getLogger(__name__)


def deploy(
    app_spec_path: Path,
    deploy_callback: Callable[
        [AlgodClient, IndexerClient, ApplicationSpecification, Account], None
    ],
    deployer_initial_funds: int = 2,
) -> None:
    # get clients
    # by default client configuration is loaded from environment variables
    algod_client = get_algod_client()
    indexer_client = get_indexer_client()

    # get app spec
    app_spec = ApplicationSpecification.from_json(app_spec_path.read_text())

    # get deployer account by name
    deployer = get_account(algod_client, "DEPLOYER", fund_with_algos=0)

    minimum_funds_micro_algos = algos_to_microalgos(deployer_initial_funds)
    ensure_funded(
        algod_client,
        EnsureBalanceParameters(
            account_to_fund=deployer,
            min_spending_balance_micro_algos=minimum_funds_micro_algos,
            min_funding_increment_micro_algos=minimum_funds_micro_algos,
        ),
    )

    # use provided callback to deploy the app
    deploy_callback(algod_client, indexer_client, app_spec, deployer)
