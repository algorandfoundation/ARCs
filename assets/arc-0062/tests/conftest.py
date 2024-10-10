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
NOT_CIRCULATING_BALANCE_1 = 69
NOT_CIRCULATING_BALANCE_2 = 42
NOT_CIRCULATING_BALANCE_3 = 4


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
    config.configure(debug=False)

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
def not_circulating_address_1(algorand_client: AlgorandClient) -> AddressAndSigner:
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
def not_circulating_address_2(algorand_client: AlgorandClient) -> AddressAndSigner:
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
def not_circulating_address_3(algorand_client: AlgorandClient) -> AddressAndSigner:
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
def not_circulating_balance_1(
    algorand_client: AlgorandClient,
    asset_creator: AddressAndSigner,
    not_circulating_address_1: AddressAndSigner,
    asset: int,
) -> AddressAndSigner:
    algorand_client.send.asset_opt_in(
        AssetOptInParams(
            sender=not_circulating_address_1.address,
            signer=not_circulating_address_1.signer,
            asset_id=asset,
        )
    )
    algorand_client.send.asset_transfer(
        AssetTransferParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            asset_id=asset,
            amount=NOT_CIRCULATING_BALANCE_1,
            receiver=not_circulating_address_1.address,
        )
    )
    assert (
        get_asset_balance(algorand_client, not_circulating_address_1.address, asset)
        == NOT_CIRCULATING_BALANCE_1
    )
    return not_circulating_address_1


@pytest.fixture(scope="function")
def not_circulating_balance_2(
    algorand_client: AlgorandClient,
    asset_creator: AddressAndSigner,
    not_circulating_address_2: AddressAndSigner,
    asset: int,
) -> AddressAndSigner:
    algorand_client.send.asset_opt_in(
        AssetOptInParams(
            sender=not_circulating_address_2.address,
            signer=not_circulating_address_2.signer,
            asset_id=asset,
        )
    )
    algorand_client.send.asset_transfer(
        AssetTransferParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            asset_id=asset,
            amount=NOT_CIRCULATING_BALANCE_2,
            receiver=not_circulating_address_2.address,
        )
    )
    assert (
        get_asset_balance(algorand_client, not_circulating_address_2.address, asset)
        == NOT_CIRCULATING_BALANCE_2
    )
    return not_circulating_address_2


@pytest.fixture(scope="function")
def not_circulating_balance_3(
    algorand_client: AlgorandClient,
    asset_creator: AddressAndSigner,
    not_circulating_address_3: AddressAndSigner,
    asset: int,
) -> AddressAndSigner:
    algorand_client.send.asset_opt_in(
        AssetOptInParams(
            sender=not_circulating_address_3.address,
            signer=not_circulating_address_3.signer,
            asset_id=asset,
        )
    )
    algorand_client.send.asset_transfer(
        AssetTransferParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            asset_id=asset,
            amount=NOT_CIRCULATING_BALANCE_3,
            receiver=not_circulating_address_3.address,
        )
    )
    assert (
        get_asset_balance(algorand_client, not_circulating_address_3.address, asset)
        == NOT_CIRCULATING_BALANCE_3
    )
    return not_circulating_address_3


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
