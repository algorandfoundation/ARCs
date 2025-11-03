# ARC-20 Reference Implementation

This is the reference implementation of Smart ASA based on the [ARC-20 specification](../../ARCs/arc-0020.md).

## Deployments

Smart ASA examples deployed on TestNet:

| App ID                                                             | App Spec                                                                                                                               |
|--------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| [749063457](https://lora.algokit.io/testnet/application/749063457) | [ARC-56](https://github.com/algorandfoundation/ARCs/blob/main/assets/arc-0020/smart_contracts/artifacts/smart_asa/SmartAsa.arc56.json) |

1. Download the App Spec JSON file;
1. Navigate to the [Lora App Lab](https://lora.algokit.io/testnet/app-lab);
1. Create the App Interface using the existing App ID and App Spec JSON;
1. Explore the D-ASA interface.

## Local Setup and Tests

This reference implementation is developed with [AlgoKit](https://algorand.co/algokit).

- Install AlgoKit
- Setup your virtual environment (managed with [Poetry](https://python-poetry.org/))

```shell
algokit bootstrap all
```

- Start your Algorand LocalNet (requires [Docker](https://www.docker.com/get-started/))

```shell
algokit localnet start
```

- Run tests (managed with PyTest)

```shell
algokit project run test
```

or, for verbose results:

```shell
poetry run pytest -s -v tests/<test_case>.py
```
