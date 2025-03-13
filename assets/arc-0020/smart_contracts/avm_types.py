from algopy import arc4


class AssetConfig(arc4.Struct, kw_only=True):
    """Smart ASA Configuration"""

    total: arc4.UInt64
    decimals: arc4.UInt32
    default_frozen: arc4.Bool
    unit_name: arc4.String
    name: arc4.String
    url: arc4.String
    metadata_hash: arc4.DynamicBytes
    manager_addr: arc4.Address
    reserve_addr: arc4.Address
    freeze_addr: arc4.Address
    clawback_addr: arc4.Address
