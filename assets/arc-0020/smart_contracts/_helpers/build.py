import logging
import subprocess
from pathlib import Path
from shutil import rmtree

logger = logging.getLogger(__name__)
deployment_extension = "py"


def _get_output_path(output_dir: Path, deployment_extension: str) -> Path:
    return output_dir / Path(
        "{contract_name}"
        + ("_client" if deployment_extension == "py" else "Client")
        + f".{deployment_extension}"
    )


def build(output_dir: Path, contract_path: Path) -> Path:
    output_dir = output_dir.resolve()
    if output_dir.exists():
        rmtree(output_dir)
    output_dir.mkdir(exist_ok=True, parents=True)
    logger.info(f"Exporting {contract_path} to {output_dir}")

    build_result = subprocess.run(
        [
            "algokit",
            "--no-color",
            "compile",
            "python",
            contract_path.absolute(),
            f"--out-dir={output_dir}",
            "--output-arc32",
            "--output-source-map",
        ],
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
    )
    if build_result.returncode:
        raise Exception(f"Could not build contract:\n{build_result.stdout}")

    app_spec_file_names = [file.name for file in output_dir.glob("*.arc32.json")]
    app_spec_file_name = None
    for app_spec_file_name in app_spec_file_names:
        if app_spec_file_name is None:
            logger.warning(
                "No '*.arc32.json' file found (likely a logic signature being compiled). Skipping client generation."
            )
            continue
        print(app_spec_file_name)
        generate_result = subprocess.run(
            [
                "algokit",
                "generate",
                "client",
                output_dir,
                "--output",
                _get_output_path(output_dir, deployment_extension),
            ],
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
        )
        if generate_result.returncode:
            if "No such command" in generate_result.stdout:
                raise Exception(
                    "Could not generate typed client, requires AlgoKit 2.0.0 or "
                    "later. Please update AlgoKit"
                )
            else:
                raise Exception(
                    f"Could not generate typed client:\n{generate_result.stdout}"
                )

    return output_dir / app_spec_file_name if app_spec_file_name else output_dir
