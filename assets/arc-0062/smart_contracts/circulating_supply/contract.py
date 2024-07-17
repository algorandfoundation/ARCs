# ruff: noqa: B011

from algopy import Account, ARC4Contract, Asset, Global, String, Txn, UInt64
from algopy.arc4 import Address, abimethod

import smart_contracts.errors.std_errors as err

from . import config as cfg


class CirculatingSupply(ARC4Contract):
    """ARC-62 Reference Implementation"""

    def __init__(self) -> None:
        # Global State
        self.asset_id = UInt64()
        self.burned = Address()
        self.locked = Address()
        self.generic = Address()

    @abimethod()
    def set_asset(self, asset_id: UInt64) -> None:
        """
        Set the ASA ID for the circulating supply - Authorization: ASA Manager Address

        Args:
            asset_id: ASA ID of the circulating supply
        """
        asset = Asset(asset_id)
        # Preconditions
        assert Txn.sender == asset.manager and not self.asset_id, err.UNAUTHORIZED
        # Effects
        self.asset_id = asset_id

    @abimethod()
    def set_not_circulating_address(self, address: Address, label: String) -> None:
        """
        Set non-circulating supply addresses - Authorization: ASA Manager Address

        Args:
            address: Address to assign to the label to
            label: Label selector ("burned", "locked", "generic")
        """
        asset = Asset(self.asset_id)
        # Preconditions
        assert Txn.sender == asset.manager, err.UNAUTHORIZED
        assert Account(address.bytes).is_opted_in(asset), err.NOT_OPTED_IN
        # Effects
        match label:
            case cfg.BURNED:
                self.burned = address
            case cfg.LOCKED:
                self.locked = address
            case cfg.GENERIC:
                self.generic = address
            case _:
                assert False, err.INVALID_LABEL

    @abimethod(readonly=True)
    def arc62_get_circulating_supply(self, asset_id: UInt64) -> UInt64:
        """
        Get ASA circulating supply

        Args:
            asset_id: ASA ID of the circulating supply

        Returns:
            ASA circulating supply
        """
        asset = Asset(asset_id)
        burned = Account(self.burned.bytes)
        locked = Account(self.locked.bytes)
        generic = Account(self.generic.bytes)
        # Preconditions
        assert asset_id == self.asset_id, err.INVALID_ASSET_ID
        # Effects
        reserve_balance = (
            UInt64(0)
            if asset.reserve == Global.zero_address
            or not asset.reserve.is_opted_in(asset)
            else asset.balance(asset.reserve)
        )
        burned_balance = (
            UInt64(0)
            if burned == Global.zero_address or not burned.is_opted_in(asset)
            else asset.balance(burned)
        )
        locked_balance = (
            UInt64(0)
            if locked == Global.zero_address or not locked.is_opted_in(asset)
            else asset.balance(locked)
        )
        generic_balance = (
            UInt64(0)
            if generic == Global.zero_address or not generic.is_opted_in(asset)
            else asset.balance(generic)
        )
        return (
            asset.total
            - reserve_balance
            - burned_balance
            - locked_balance
            - generic_balance
        )
