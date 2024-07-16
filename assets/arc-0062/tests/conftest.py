import pytest
from algokit_utils import (
    EnsureBalanceParameters,
    OnCompleteCallParameters,
    ensure_funded,
    get_algod_client,
    get_default_localnet_config,
    get_indexer_client,
)
from algokit_utils.beta.account_manager import AddressAndSigner
from algokit_utils.beta.algorand_client import (
    AlgorandClient,
    AssetCreateParams,
    AssetOptInParams,
    AssetTransferParams,
)
from algokit_utils.config import config
from algosdk.v2client.algod import AlgodClient
from algosdk.v2client.indexer import IndexerClient

from smart_contracts.artifacts.circulating_supply.circulating_supply_client import (
    CirculatingSupplyClient,
)

INITIAL_FUNDS = 100_000_000
ASA_TOTAL = 1000
RESERVE_BALANCE = 420
BURNED_BALANCE = 69
LOCKED_BALANCE = 42
GENERIC_BALANCE = 4


def get_asset_balance(
    algorand_client: AlgorandClient, address: str, asset_id: int
) -> int:
    asset_balance: int = algorand_client.account.get_asset_information(  # type: ignore
        sender=address, asset_id=asset_id
    )["asset-holding"]["amount"]
    return asset_balance


@pytest.fixture(scope="session")
def algod_client() -> AlgodClient:
    # by default we are using localnet algod
    client = get_algod_client(get_default_localnet_config("algod"))
    return client


@pytest.fixture(scope="session")
def indexer_client() -> IndexerClient:
    return get_indexer_client(get_default_localnet_config("indexer"))


@pytest.fixture(scope="session")
def algorand_client() -> AlgorandClient:
    client = AlgorandClient.default_local_net()
    client.set_suggested_params_timeout(0)
    return client


@pytest.fixture(scope="session")
def deployer(algorand_client: AlgorandClient) -> AddressAndSigner:
    acct = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=acct.address,
            min_spending_balance_micro_algos=INITIAL_FUNDS,
        ),
    )
    return acct


@pytest.fixture(scope="function")
def circulating_supply_client(
    algod_client: AlgodClient, indexer_client: IndexerClient, deployer: AddressAndSigner
) -> CirculatingSupplyClient:
    config.configure(debug=True)

    client = CirculatingSupplyClient(
        algod_client=algod_client,
        indexer_client=indexer_client,
        creator=deployer.address,
        signer=deployer.signer,
    )
    client.create_bare()
    return client


@pytest.fixture(scope="session")
def asset_creator(algorand_client: AlgorandClient) -> AddressAndSigner:
    acct = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=acct.address,
            min_spending_balance_micro_algos=INITIAL_FUNDS,
        ),
    )
    return acct


@pytest.fixture(scope="session")
def asset_manager(algorand_client: AlgorandClient) -> AddressAndSigner:
    acct = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=acct.address,
            min_spending_balance_micro_algos=INITIAL_FUNDS,
        ),
    )
    return acct


@pytest.fixture(scope="session")
def asset_reserve(algorand_client: AlgorandClient) -> AddressAndSigner:
    acct = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=acct.address,
            min_spending_balance_micro_algos=INITIAL_FUNDS,
        ),
    )
    return acct


@pytest.fixture(scope="session")
def asset_burning(algorand_client: AlgorandClient) -> AddressAndSigner:
    acct = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=acct.address,
            min_spending_balance_micro_algos=INITIAL_FUNDS,
        ),
    )
    return acct


@pytest.fixture(scope="session")
def asset_locking(algorand_client: AlgorandClient) -> AddressAndSigner:
    acct = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=acct.address,
            min_spending_balance_micro_algos=INITIAL_FUNDS,
        ),
    )
    return acct


@pytest.fixture(scope="session")
def asset_generic_not_circulating(algorand_client: AlgorandClient) -> AddressAndSigner:
    acct = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=acct.address,
            min_spending_balance_micro_algos=INITIAL_FUNDS,
        ),
    )
    return acct


@pytest.fixture(scope="function")
def asset(
    algorand_client: AlgorandClient,
    asset_creator: AddressAndSigner,
    asset_manager: AddressAndSigner,
    asset_reserve: AddressAndSigner,
) -> int:
    txn_result = algorand_client.send.asset_create(  # type: ignore
        AssetCreateParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            total=ASA_TOTAL,
            manager=asset_manager.address,
            reserve=asset_reserve.address,
        )
    )
    return txn_result["confirmation"]["asset-index"]  # type: ignore


@pytest.fixture(scope="function")
def reserve_with_balance(
    algorand_client: AlgorandClient,
    asset_creator: AddressAndSigner,
    asset_reserve: AddressAndSigner,
    asset: int,
) -> AddressAndSigner:
    algorand_client.send.asset_opt_in(
        AssetOptInParams(
            sender=asset_reserve.address,
            signer=asset_reserve.signer,
            asset_id=asset,
        )
    )
    algorand_client.send.asset_transfer(
        AssetTransferParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            asset_id=asset,
            amount=RESERVE_BALANCE,
            receiver=asset_reserve.address,
        )
    )
    assert (
        get_asset_balance(algorand_client, asset_reserve.address, asset)
        == RESERVE_BALANCE
    )
    return asset_reserve


@pytest.fixture(scope="function")
def burning_with_balance(
    algorand_client: AlgorandClient,
    asset_creator: AddressAndSigner,
    asset_burning: AddressAndSigner,
    asset: int,
) -> AddressAndSigner:
    algorand_client.send.asset_opt_in(
        AssetOptInParams(
            sender=asset_burning.address,
            signer=asset_burning.signer,
            asset_id=asset,
        )
    )
    algorand_client.send.asset_transfer(
        AssetTransferParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            asset_id=asset,
            amount=BURNED_BALANCE,
            receiver=asset_burning.address,
        )
    )
    assert (
        get_asset_balance(algorand_client, asset_burning.address, asset)
        == BURNED_BALANCE
    )
    return asset_burning


@pytest.fixture(scope="function")
def locking_with_balance(
    algorand_client: AlgorandClient,
    asset_creator: AddressAndSigner,
    asset_locking: AddressAndSigner,
    asset: int,
) -> AddressAndSigner:
    algorand_client.send.asset_opt_in(
        AssetOptInParams(
            sender=asset_locking.address,
            signer=asset_locking.signer,
            asset_id=asset,
        )
    )
    algorand_client.send.asset_transfer(
        AssetTransferParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            asset_id=asset,
            amount=LOCKED_BALANCE,
            receiver=asset_locking.address,
        )
    )
    assert (
        get_asset_balance(algorand_client, asset_locking.address, asset)
        == LOCKED_BALANCE
    )
    return asset_locking


@pytest.fixture(scope="function")
def generic_not_circulating_with_balance(
    algorand_client: AlgorandClient,
    asset_creator: AddressAndSigner,
    asset_generic_not_circulating: AddressAndSigner,
    asset: int,
) -> AddressAndSigner:
    algorand_client.send.asset_opt_in(
        AssetOptInParams(
            sender=asset_generic_not_circulating.address,
            signer=asset_generic_not_circulating.signer,
            asset_id=asset,
        )
    )
    algorand_client.send.asset_transfer(
        AssetTransferParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            asset_id=asset,
            amount=GENERIC_BALANCE,
            receiver=asset_generic_not_circulating.address,
        )
    )
    assert (
        get_asset_balance(algorand_client, asset_generic_not_circulating.address, asset)
        == GENERIC_BALANCE
    )
    return asset_generic_not_circulating


@pytest.fixture(scope="function")
def asset_circulating_supply_client(
    circulating_supply_client: CirculatingSupplyClient,
    asset_manager: AddressAndSigner,
    asset: int,
) -> CirculatingSupplyClient:
    circulating_supply_client.set_asset(
        asset_id=asset,
        transaction_parameters=OnCompleteCallParameters(
            sender=asset_manager.address,
            signer=asset_manager.signer,
            # TODO: Foreign resources should be auto-populated
            foreign_assets=[asset],
        ),
    )
    return circulating_supply_client
