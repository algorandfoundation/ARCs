from dataclasses import asdict, dataclass
from typing import Callable, Final

import pytest
from algokit_utils import (
    EnsureBalanceParameters,
    OnCompleteCallParameters,
    ensure_funded,
)
from algokit_utils.beta.account_manager import AddressAndSigner
from algokit_utils.beta.algorand_client import (
    AlgorandClient,
    AssetCreateParams,
    AssetOptInParams,
)
from algokit_utils.config import config
from algosdk.atomic_transaction_composer import TransactionWithSigner
from algosdk.util import algos_to_microalgos

from smart_contracts.artifacts.smart_asa.smart_asa_client import SmartAsaClient

INITIAL_ALGO_FUNDS: Final[int] = algos_to_microalgos(100)


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
    metadata_hash: bytes = bytes(32)

    def dictify(self) -> dict:
        return asdict(self)  # type: ignore


@pytest.fixture(scope="session")
def algorand_client() -> AlgorandClient:
    client = AlgorandClient.default_local_net()
    client.set_suggested_params_timeout(0)
    return client


@pytest.fixture(scope="session")
def creator(algorand_client: AlgorandClient) -> AddressAndSigner:
    account = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=account.address,
            min_spending_balance_micro_algos=INITIAL_ALGO_FUNDS,
        ),
    )
    return account


@pytest.fixture(scope="session")
def manager(algorand_client: AlgorandClient) -> AddressAndSigner:
    account = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=account.address,
            min_spending_balance_micro_algos=INITIAL_ALGO_FUNDS,
        ),
    )
    return account


@pytest.fixture(scope="session")
def reserve(algorand_client: AlgorandClient) -> AddressAndSigner:
    account = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=account.address,
            min_spending_balance_micro_algos=INITIAL_ALGO_FUNDS,
        ),
    )
    return account


@pytest.fixture(scope="session")
def freeze(algorand_client: AlgorandClient) -> AddressAndSigner:
    account = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=account.address,
            min_spending_balance_micro_algos=INITIAL_ALGO_FUNDS,
        ),
    )
    return account


@pytest.fixture(scope="session")
def clawback(algorand_client: AlgorandClient) -> AddressAndSigner:
    account = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=account.address,
            min_spending_balance_micro_algos=INITIAL_ALGO_FUNDS,
        ),
    )
    return account


@pytest.fixture(scope="session")
def eve(algorand_client: AlgorandClient) -> AddressAndSigner:
    account = algorand_client.account.random()

    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=account.address,
            min_spending_balance_micro_algos=INITIAL_ALGO_FUNDS,
        ),
    )
    return account


@pytest.fixture(scope="session")
def dummy_asa(algorand_client: AlgorandClient, creator: AddressAndSigner) -> int:
    return algorand_client.send.asset_create(
        AssetCreateParams(sender=creator.address, signer=creator.signer, total=1)
    )["confirmation"]["asset-index"]


@pytest.fixture(scope="function")
def smart_asa_client_void(
    algorand_client: AlgorandClient, creator: AddressAndSigner
) -> SmartAsaClient:
    config.configure(
        debug=True,
        # trace_all=True,
    )
    return SmartAsaClient(algorand_client.client.algod, signer=creator.signer)


@pytest.fixture(scope="function")
def smart_asa_client_no_asset(
    algorand_client: AlgorandClient,
    smart_asa_client_void: SmartAsaClient,
) -> SmartAsaClient:
    smart_asa_client_void.create_bare()
    ensure_funded(
        algorand_client.client.algod,
        EnsureBalanceParameters(
            account_to_fund=smart_asa_client_void.app_address,
            min_spending_balance_micro_algos=INITIAL_ALGO_FUNDS,
        ),
    )
    return smart_asa_client_void


@pytest.fixture(
    scope="function", params=[False, True], ids=["Not Default Frozen", "Default Frozen"]
)
def asa_config(
    manager: AddressAndSigner,
    reserve: AddressAndSigner,
    freeze: AddressAndSigner,
    clawback: AddressAndSigner,
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
    sp = smart_asa_client_no_asset.algod_client.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2

    smart_asa_client_no_asset.asset_create(
        **asa_config.dictify(),
        transaction_parameters=OnCompleteCallParameters(suggested_params=sp),
    )
    return smart_asa_client_no_asset


@pytest.fixture(scope="function")
def opted_in_account_factory(
    algorand_client: AlgorandClient, smart_asa_client: SmartAsaClient
) -> Callable[..., AddressAndSigner]:
    def _factory() -> AddressAndSigner:
        account = algorand_client.account.random()
        ensure_funded(
            algorand_client.client.algod,
            EnsureBalanceParameters(
                account_to_fund=account.address,
                min_spending_balance_micro_algos=INITIAL_ALGO_FUNDS,
            ),
        )
        smart_asa_id = smart_asa_client.get_global_state().smart_asa_id

        smart_asa_client.opt_in_asset_opt_in(
            asset=smart_asa_id,
            ctrl_asa_opt_in=TransactionWithSigner(
                txn=algorand_client.transactions.asset_opt_in(
                    AssetOptInParams(asset_id=smart_asa_id, sender=account.address)
                ),
                signer=account.signer,
            ),
            transaction_parameters=OnCompleteCallParameters(
                signer=account.signer,
                sender=account.address,
            ),
        )
        return account

    return _factory


@pytest.fixture(scope="function")
def receiver(
    opted_in_account_factory: Callable[..., AddressAndSigner]
) -> AddressAndSigner:
    return opted_in_account_factory()


@pytest.fixture(scope="function")
def reserve_and_clawback(
    manager: AddressAndSigner,
    reserve: AddressAndSigner,
    asa_config: ASAConfig,
    smart_asa_client: SmartAsaClient,
) -> AddressAndSigner:
    asa_config.clawback_addr = asa_config.reserve_addr
    smart_asa_client.asset_config(
        config_asset=smart_asa_client.get_global_state().smart_asa_id,
        **asa_config.dictify(),
        transaction_parameters=OnCompleteCallParameters(
            signer=manager.signer,
            sender=manager.address,
        ),
    )
    return reserve


@pytest.fixture(scope="function")
def reserve_with_supply(
    algorand_client: AlgorandClient,
    reserve: AddressAndSigner,
    smart_asa_client: SmartAsaClient,
) -> AddressAndSigner:
    smart_asa = smart_asa_client.get_global_state()
    smart_asa_id = smart_asa.smart_asa_id
    ctrl_asa_opt_in = TransactionWithSigner(
        txn=algorand_client.transactions.asset_opt_in(
            AssetOptInParams(asset_id=smart_asa_id, sender=reserve.address)
        ),
        signer=reserve.signer,
    )
    smart_asa_client.opt_in_asset_opt_in(
        asset=smart_asa_id,
        ctrl_asa_opt_in=ctrl_asa_opt_in,
        transaction_parameters=OnCompleteCallParameters(
            signer=reserve.signer,
            sender=reserve.address,
        ),
    )
    sp = smart_asa_client.algod_client.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2
    smart_asa_client.asset_transfer(
        xfer_asset=smart_asa_id,
        asset_amount=smart_asa.total,
        asset_sender=smart_asa_client.app_address,
        asset_receiver=reserve.address,
        transaction_parameters=OnCompleteCallParameters(
            suggested_params=sp,
            signer=reserve.signer,
            sender=reserve.address,
        ),
    )
    return reserve


@pytest.fixture(scope="function")
def account_with_supply(
    reserve: AddressAndSigner,
    smart_asa_client: SmartAsaClient,
    opted_in_account_factory: Callable[..., AddressAndSigner],
) -> AddressAndSigner:
    account = opted_in_account_factory()
    smart_asa = smart_asa_client.get_global_state()
    sp = smart_asa_client.algod_client.suggested_params()
    sp.flat_fee = True
    sp.fee = sp.min_fee * 2
    smart_asa_client.asset_transfer(
        xfer_asset=smart_asa.smart_asa_id,
        asset_amount=smart_asa.total,
        asset_sender=smart_asa_client.app_address,
        asset_receiver=account.address,
        transaction_parameters=OnCompleteCallParameters(
            suggested_params=sp,
            signer=reserve.signer,
            sender=reserve.address,
        ),
    )
    return account
