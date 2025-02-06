import logging
import sys
from pathlib import Path

from dotenv import load_dotenv

from smart_contracts._helpers.build import build
from smart_contracts._helpers.config import contracts
from smart_contracts._helpers.deploy import deploy

# Uncomment the following lines to enable auto generation of AVM Debugger compliant sourcemap and simulation trace file.
# Learn more about using AlgoKit AVM Debugger to debug your TEAL source codes and inspect various kinds of
# Algorand transactions in atomic groups -> https://github.com/algorandfoundation/algokit-avm-vscode-debugger
# from algokit_utils.config import config
# config.configure(debug=True, trace_all=True)
logging.basicConfig(
    level=logging.DEBUG, format="%(asctime)s %(levelname)-10s: %(message)s"
)
logger = logging.getLogger(__name__)
logger.info("Loading .env")
# For manual script execution (bypassing `algokit project deploy`) with a custom .env,
# modify `load_dotenv()` accordingly. For example, `load_dotenv('.env.localnet')`.
load_dotenv()
root_path = Path(__file__).parent


def main(action: str, contract_name: str | None = None) -> None:
    artifact_path = root_path / "artifacts"

    # Filter contracts if a specific contract name is provided
    filtered_contracts = [
        c for c in contracts if contract_name is None or c.name == contract_name
    ]

    match action:
        case "build":
            for contract in filtered_contracts:
                logger.info(f"Building app at {contract.path}")
                build(artifact_path / contract.name, contract.path)
        case "deploy":
            for contract in filtered_contracts:
                output_dir = artifact_path / contract.name
                app_spec_file_name = next(
                    (
                        file.name
                        for file in output_dir.iterdir()
                        if file.is_file() and file.suffixes == [".arc32", ".json"]
                    ),
                    None,
                )
                if app_spec_file_name is None:
                    raise Exception("Could not deploy app, .arc32.json file not found")
                app_spec_path = output_dir / app_spec_file_name
                if contract.deploy:
                    logger.info(f"Deploying app {contract.name}")
                    deploy(app_spec_path, contract.deploy)
        case "all":
            for contract in filtered_contracts:
                logger.info(f"Building app at {contract.path}")
                app_spec_path = build(artifact_path / contract.name, contract.path)
                if contract.deploy:
                    logger.info(f"Deploying {contract.path.name}")
                    deploy(app_spec_path, contract.deploy)


if __name__ == "__main__":
    if len(sys.argv) > 2:
        main(sys.argv[1], sys.argv[2])
    elif len(sys.argv) > 1:
        main(sys.argv[1])
    else:
        main("all")
