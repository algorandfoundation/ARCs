import pytest
from algokit_utils import AlgorandClient
from algokit_utils.config import config

# Uncomment if you want to load network specific or generic .env file
# @pytest.fixture(autouse=True, scope="session")
# def environment_fixture() -> None:
#     env_path = Path(__file__).parent.parent / ".env"
#     load_dotenv(env_path)

config.configure(
    debug=True,
    # trace_all=True, # uncomment to trace all transactions
)


@pytest.fixture(scope="session")
def algorand_client() -> AlgorandClient:
    # by default we are using localnet algod
    return AlgorandClient.from_environment()
