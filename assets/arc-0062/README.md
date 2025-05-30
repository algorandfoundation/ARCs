# ARC-62 Reference Implementation

This is the reference implementation of ASA Circulating Supply App based on the
[ARC-62 specification](../../ARCs/arc-0062.md).

## Adoption

Explorers supporting ARC-62:

- Pera Explorer (<a href="https://testnet.explorer.perawallet.app/asset/740315456/">example</a>)

ASA adopting ARC-62:

- Midas US Treasury Bill Token (<a href="https://explorer.perawallet.app/asset/2849506951/">mTBILL</a>)

## Demo

1. Download the [AppSpec JSON file](./smart_contracts/artifacts/circulating_supply/CirculatingSupply.arc56.json)
2. Open the <a href="https://lora.algokit.io/testnet/app-lab/create">Lora App Lab</a> in TestNet
3. Enter the App ID `740315445` and click on _"Use existing"_
4. Upload the AppSpec JSON and click _"Next"_ and then _"Create"_
5. Open the <a href="https://lora.algokit.io/testnet/application/740315445">App</a>
6. Connect a TestNet Wallet
7. Select the `arc62_get_circulating_supply` method and click _"Call"_
8. Enter `740315456` as `asset_id` argument and click _"Add"_
9. Click on _"Populate Resources"_
10. Click on _"Simulate"_ and inspect the result (answer should be `1`)

## Test

Install the project Python dependencies:

`poetry install`

Run the test:

```shell
poetry run pytest -s -v tests/test_get_circulating_supply.py::test_pass_get_circulating_supply
``` 
