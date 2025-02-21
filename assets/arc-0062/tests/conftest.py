from typing import Final

import pytest
from algokit_utils import (
    AlgoAmount,
    AlgorandClient,
    AssetCreateParams,
    AssetOptInParams,
    AssetTransferParams,
    CommonAppCallParams,
    SigningAccount,
)
from algokit_utils.config import config

from smart_contracts.artifacts.circulating_supply.circulating_supply_client import (
    CirculatingSupplyClient,
    CirculatingSupplyFactory,
    SetAssetArgs,
)

INITIAL_FUNDS: Final[AlgoAmount] = AlgoAmount.from_algo(100)
ASA_TOTAL = 1000
RESERVE_BALANCE = 420
NOT_CIRCULATING_BALANCE_1 = 69
NOT_CIRCULATING_BALANCE_2 = 42
NOT_CIRCULATING_BALANCE_3 = 4


@pytest.fixture(scope="session")
def algorand() -> AlgorandClient:
    client = AlgorandClient.default_localnet()
    client.set_suggested_params_cache_timeout(0)
    return client


@pytest.fixture(scope="session")
def deployer(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="function")
def circulating_supply_client(
    algorand: AlgorandClient, deployer: SigningAccount
) -> CirculatingSupplyClient:
    config.configure(
        debug=False,
        populate_app_call_resources=True,
        # trace_all=True,
    )

    factory = algorand.client.get_typed_app_factory(
        CirculatingSupplyFactory,
        default_sender=deployer.address,
        default_signer=deployer.signer,
    )
    client, _ = factory.send.create.bare()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=client.app_address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return client


@pytest.fixture(scope="session")
def asset_creator(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="session")
def asset_manager(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="session")
def asset_reserve(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="session")
def not_circulating_address_1(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="session")
def not_circulating_address_2(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="session")
def not_circulating_address_3(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="function")
def asset(
    algorand: AlgorandClient,
    asset_creator: SigningAccount,
    asset_manager: SigningAccount,
    asset_reserve: SigningAccount,
) -> int:
    return algorand.send.asset_create(
        AssetCreateParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            total=ASA_TOTAL,
            manager=asset_manager.address,
            reserve=asset_reserve.address,
        )
    ).asset_id


@pytest.fixture(scope="function")
def reserve_with_balance(
    algorand: AlgorandClient,
    asset_creator: SigningAccount,
    asset_reserve: SigningAccount,
    asset: int,
) -> SigningAccount:
    algorand.send.asset_opt_in(
        AssetOptInParams(
            sender=asset_reserve.address,
            signer=asset_reserve.signer,
            asset_id=asset,
        )
    )
    algorand.send.asset_transfer(
        AssetTransferParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            asset_id=asset,
            amount=RESERVE_BALANCE,
            receiver=asset_reserve.address,
        )
    )
    assert (
        algorand.asset.get_account_information(asset_reserve, asset).balance
        == RESERVE_BALANCE
    )
    return asset_reserve


@pytest.fixture(scope="function")
def not_circulating_balance_1(
    algorand: AlgorandClient,
    asset_creator: SigningAccount,
    not_circulating_address_1: SigningAccount,
    asset: int,
) -> SigningAccount:
    algorand.send.asset_opt_in(
        AssetOptInParams(
            sender=not_circulating_address_1.address,
            signer=not_circulating_address_1.signer,
            asset_id=asset,
        )
    )
    algorand.send.asset_transfer(
        AssetTransferParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            asset_id=asset,
            amount=NOT_CIRCULATING_BALANCE_1,
            receiver=not_circulating_address_1.address,
        )
    )
    assert (
        algorand.asset.get_account_information(not_circulating_address_1, asset).balance
        == NOT_CIRCULATING_BALANCE_1
    )
    return not_circulating_address_1


@pytest.fixture(scope="function")
def not_circulating_balance_2(
    algorand: AlgorandClient,
    asset_creator: SigningAccount,
    not_circulating_address_2: SigningAccount,
    asset: int,
) -> SigningAccount:
    algorand.send.asset_opt_in(
        AssetOptInParams(
            sender=not_circulating_address_2.address,
            signer=not_circulating_address_2.signer,
            asset_id=asset,
        )
    )
    algorand.send.asset_transfer(
        AssetTransferParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            asset_id=asset,
            amount=NOT_CIRCULATING_BALANCE_2,
            receiver=not_circulating_address_2.address,
        )
    )
    assert (
        algorand.asset.get_account_information(not_circulating_address_2, asset).balance
        == NOT_CIRCULATING_BALANCE_2
    )
    return not_circulating_address_2


@pytest.fixture(scope="function")
def not_circulating_balance_3(
    algorand: AlgorandClient,
    asset_creator: SigningAccount,
    not_circulating_address_3: SigningAccount,
    asset: int,
) -> SigningAccount:
    algorand.send.asset_opt_in(
        AssetOptInParams(
            sender=not_circulating_address_3.address,
            signer=not_circulating_address_3.signer,
            asset_id=asset,
        )
    )
    algorand.send.asset_transfer(
        AssetTransferParams(
            sender=asset_creator.address,
            signer=asset_creator.signer,
            asset_id=asset,
            amount=NOT_CIRCULATING_BALANCE_3,
            receiver=not_circulating_address_3.address,
        )
    )
    assert (
        algorand.asset.get_account_information(not_circulating_address_3, asset).balance
        == NOT_CIRCULATING_BALANCE_3
    )
    return not_circulating_address_3


@pytest.fixture(scope="function")
def asset_circulating_supply_client(
    circulating_supply_client: CirculatingSupplyClient,
    asset_manager: SigningAccount,
    asset: int,
) -> CirculatingSupplyClient:
    circulating_supply_client.send.set_asset(
        args=SetAssetArgs(asset_id=asset),
        params=CommonAppCallParams(sender=asset_manager.address),
    )
    return circulating_supply_client
