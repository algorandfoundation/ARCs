import dataclasses
import importlib
from collections.abc import Callable
from pathlib import Path

from algokit_utils import Account, ApplicationSpecification
from algosdk.v2client.algod import AlgodClient
from algosdk.v2client.indexer import IndexerClient


@dataclasses.dataclass
class SmartContract:
    path: Path
    name: str
    deploy: (
        Callable[[AlgodClient, IndexerClient, ApplicationSpecification, Account], None]
        | None
    ) = None


def import_contract(folder: Path) -> Path:
    """Imports the contract from a folder if it exists."""
    contract_path = folder / "contract.py"
    if contract_path.exists():
        return contract_path
    else:
        raise Exception(f"Contract not found in {folder}")


def import_deploy_if_exists(
    folder: Path,
) -> (
    Callable[[AlgodClient, IndexerClient, ApplicationSpecification, Account], None]
    | None
):
    """Imports the deploy function from a folder if it exists."""
    try:
        deploy_module = importlib.import_module(
            f"{folder.parent.name}.{folder.name}.deploy_config"
        )
        return deploy_module.deploy  # type: ignore
    except ImportError:
        return None


def has_contract_file(directory: Path) -> bool:
    """Checks whether the directory contains contract.py file."""
    return (directory / "contract.py").exists()


# define contracts to build and/or deploy
base_dir = Path("smart_contracts")
contracts = [
    SmartContract(
        path=import_contract(folder),
        name=folder.name,
        deploy=import_deploy_if_exists(folder),
    )
    for folder in base_dir.iterdir()
    if folder.is_dir() and has_contract_file(folder)
]
