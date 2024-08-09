# ARC-62 Reference Implementation

This is the reference implementation of ASA Circulating Supply App based on the
[ARC-62 specification](../../ARCs/arc-0062.md).

## Example

Install the project Python dependencies:

`poetry install`

Run the test:

```shell
poetry run pytest -s -v tests/test_get_circulating_supply.py::test_pass_get_circulating_supply
``` 
