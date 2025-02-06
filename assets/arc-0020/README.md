# ARC-20 Reference Implementation

This is the reference implementation of Smart ASA based on the [ARC-20 specification](../../ARCs/arc-0020.md).

## Setup

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

## How to contribute

Some test cases have been left as `TODO`.

Community contributions to complete them are welcome.
