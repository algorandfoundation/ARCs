from collections.abc import Callable
from dataclasses import asdict, dataclass
from typing import Final

import pytest
from algokit_utils import (
    AlgoAmount,
    AlgorandClient,
    AssetCreateParams,
    AssetOptInParams,
    CommonAppCallParams,
    SigningAccount,
)
from algokit_utils.config import config
from algosdk.atomic_transaction_composer import TransactionWithSigner

from smart_contracts.artifacts.smart_asa.smart_asa_client import (
    AssetConfigArgs,
    AssetCreateArgs,
    AssetOptInArgs,
    AssetTransferArgs,
    SmartAsaClient,
    SmartAsaFactory,
)

INITIAL_FUNDS: Final[AlgoAmount] = AlgoAmount.from_algo(100)


@dataclass
class ASAConfig:
    manager_addr: str
    reserve_addr: str
    freeze_addr: str
    clawback_addr: str
    total: int = 100
    decimals: int = 2
    default_frozen: bool = False
    unit_name: str = "TST"
    name: str = "Test"
    url: str = "ipfs://..."
    metadata_hash: bytes = b"SmartASA"

    def dictify(self) -> dict:
        return asdict(self)  # type: ignore


@pytest.fixture(scope="session")
def algorand() -> AlgorandClient:
    client = AlgorandClient.default_localnet()
    client.set_suggested_params_cache_timeout(0)
    return client


@pytest.fixture(scope="session")
def creator(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="session")
def manager(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="session")
def reserve(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="session")
def freeze(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="session")
def clawback(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="session")
def eve(algorand: AlgorandClient) -> SigningAccount:
    account = algorand.account.random()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=account.address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return account


@pytest.fixture(scope="session")
def dummy_asa(algorand: AlgorandClient, creator: SigningAccount) -> int:
    return algorand.send.asset_create(
        AssetCreateParams(sender=creator.address, signer=creator.signer, total=1)
    ).asset_id


@pytest.fixture(scope="function")
def smart_asa_client_no_asset(
    algorand: AlgorandClient, creator: SigningAccount
) -> SmartAsaClient:
    config.configure(
        debug=False,
        populate_app_call_resources=True,
        # trace_all=True,
    )

    factory = algorand.client.get_typed_app_factory(
        SmartAsaFactory,
        default_sender=creator.address,
        default_signer=creator.signer,
    )
    client, _ = factory.send.create.bare()
    algorand.account.ensure_funded_from_environment(
        account_to_fund=client.app_address,
        min_spending_balance=INITIAL_FUNDS,
    )
    return client


@pytest.fixture(
    scope="function", params=[False, True], ids=["Not Default Frozen", "Default Frozen"]
)
def asa_config(
    manager: SigningAccount,
    reserve: SigningAccount,
    freeze: SigningAccount,
    clawback: SigningAccount,
    request,  # noqa: ANN001
) -> ASAConfig:
    return ASAConfig(
        manager_addr=manager.address,
        reserve_addr=reserve.address,
        freeze_addr=freeze.address,
        clawback_addr=clawback.address,
        default_frozen=request.param,
    )


@pytest.fixture(scope="function")
def smart_asa_client(
    smart_asa_client_no_asset: SmartAsaClient,
    asa_config: ASAConfig,
) -> SmartAsaClient:
    sp = smart_asa_client_no_asset.algorand.client.algod.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2

    smart_asa_client_no_asset.send.asset_create(
        AssetCreateArgs(**asa_config.dictify()),
        params=CommonAppCallParams(static_fee=AlgoAmount.from_micro_algo(sp.fee)),
    )
    return smart_asa_client_no_asset


@pytest.fixture(scope="function")
def opted_in_account_factory(
    algorand: AlgorandClient, smart_asa_client: SmartAsaClient
) -> Callable[..., SigningAccount]:
    def _factory() -> SigningAccount:
        account = algorand.account.random()
        algorand.account.ensure_funded_from_environment(
            account_to_fund=account.address,
            min_spending_balance=INITIAL_FUNDS,
        )
        smart_asa_id = smart_asa_client.state.global_state.smart_asa_id

        smart_asa_client.send.opt_in.asset_opt_in(
            AssetOptInArgs(
                asset=smart_asa_id,
                ctrl_asa_opt_in=TransactionWithSigner(
                    txn=algorand.create_transaction.asset_opt_in(
                        AssetOptInParams(asset_id=smart_asa_id, sender=account.address)
                    ),
                    signer=account.signer,
                ),
            ),
            params=CommonAppCallParams(
                signer=account.signer,
                sender=account.address,
            ),
        )
        return account

    return _factory


@pytest.fixture(scope="function")
def receiver(
    opted_in_account_factory: Callable[..., SigningAccount],
) -> SigningAccount:
    return opted_in_account_factory()


@pytest.fixture(scope="function")
def reserve_and_clawback(
    manager: SigningAccount,
    reserve: SigningAccount,
    asa_config: ASAConfig,
    smart_asa_client: SmartAsaClient,
) -> SigningAccount:
    asa_config.clawback_addr = asa_config.reserve_addr
    smart_asa_client.send.asset_config(
        AssetConfigArgs(
            config_asset=smart_asa_client.state.global_state.smart_asa_id,
            **asa_config.dictify(),
        ),
        params=CommonAppCallParams(
            signer=manager.signer,
            sender=manager.address,
        ),
    )
    return reserve


@pytest.fixture(scope="function")
def reserve_with_supply(
    algorand: AlgorandClient,
    reserve: SigningAccount,
    smart_asa_client: SmartAsaClient,
) -> SigningAccount:
    smart_asa = smart_asa_client.state.global_state
    smart_asa_id = smart_asa.smart_asa_id
    ctrl_asa_opt_in = TransactionWithSigner(
        txn=algorand.create_transaction.asset_opt_in(
            AssetOptInParams(asset_id=smart_asa_id, sender=reserve.address)
        ),
        signer=reserve.signer,
    )
    smart_asa_client.send.opt_in.asset_opt_in(
        AssetOptInArgs(
            asset=smart_asa_id,
            ctrl_asa_opt_in=ctrl_asa_opt_in,
        ),
        params=CommonAppCallParams(
            signer=reserve.signer,
            sender=reserve.address,
        ),
    )
    sp = smart_asa_client.algorand.client.algod.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2
    smart_asa_client.send.asset_transfer(
        AssetTransferArgs(
            xfer_asset=smart_asa_id,
            asset_amount=smart_asa.total,
            asset_sender=smart_asa_client.app_address,
            asset_receiver=reserve.address,
        ),
        params=CommonAppCallParams(
            static_fee=AlgoAmount.from_micro_algo(sp.fee),
            signer=reserve.signer,
            sender=reserve.address,
        ),
    )
    return reserve


@pytest.fixture(scope="function")
def account_with_supply(
    reserve: SigningAccount,
    smart_asa_client: SmartAsaClient,
    opted_in_account_factory: Callable[..., SigningAccount],
) -> SigningAccount:
    account = opted_in_account_factory()
    smart_asa = smart_asa_client.state.global_state
    sp = smart_asa_client.algorand.client.algod.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2
    smart_asa_client.send.asset_transfer(
        AssetTransferArgs(
            xfer_asset=smart_asa.smart_asa_id,
            asset_amount=smart_asa.total,
            asset_sender=smart_asa_client.app_address,
            asset_receiver=account.address,
        ),
        params=CommonAppCallParams(
            static_fee=AlgoAmount.from_micro_algo(sp.fee),
            signer=reserve.signer,
            sender=reserve.address,
        ),
    )
    return account
