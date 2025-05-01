import logging
from typing import Final

import algokit_utils
from algokit_utils.config import config

logger = logging.getLogger(__name__)

ASA_NAME: Final[str] = "ARC-62 Test ASA"
ASA_UNIT_NAME: Final[str] = "ARC-62"
ASA_DECIMALS: Final[int] = 0
ASA_TOTAL: Final[int] = 42
APP_URI: Final[str] = "algorand://app/"


# define deployment behaviour based on supplied app spec
def deploy() -> None:
    from smart_contracts.artifacts.circulating_supply.circulating_supply_client import (
        CirculatingSupplyFactory,
        SetAssetArgs,
    )

    config.configure(
        debug=False,
        populate_app_call_resources=True,
    )

    algorand = algokit_utils.AlgorandClient.from_environment()
    deployer = algorand.account.from_environment("DEPLOYER")

    factory = algorand.client.get_typed_app_factory(
        CirculatingSupplyFactory, default_sender=deployer.address
    )

    app_client, _ = factory.deploy(
        on_schema_break=algokit_utils.OnSchemaBreak.AppendApp,
        on_update=algokit_utils.OnUpdate.AppendApp,
    )

    if not app_client.state.global_state.asset_id:
        asset_id = algorand.send.asset_create(
            algokit_utils.AssetCreateParams(
                sender=deployer.address,
                signer=deployer.signer,
                asset_name=ASA_NAME,
                unit_name=ASA_UNIT_NAME,
                total=ASA_TOTAL,
                decimals=ASA_DECIMALS,
                manager=deployer.address,
                url=APP_URI + str(app_client.app_id),
            )
        ).asset_id

        app_client.send.set_asset(args=SetAssetArgs(asset_id=asset_id))
