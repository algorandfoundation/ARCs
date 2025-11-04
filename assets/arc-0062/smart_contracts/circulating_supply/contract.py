# ruff: noqa: B011

from algopy import (
    Account,
    ARC4Contract,
    Asset,
    Global,
    GlobalState,
    String,
    Txn,
    UInt64,
)
from algopy.arc4 import Address, abimethod

import smart_contracts.errors.std_errors as err

from . import config as cfg


class CirculatingSupply(ARC4Contract):
    """ARC-62 Reference Implementation"""

    def __init__(self) -> None:
        # Global State
        self.asset_id = UInt64()
        self.not_circulating_label_1 = GlobalState(
            Address(), key=cfg.NOT_CIRCULATING_LABEL_1
        )
        self.not_circulating_label_2 = GlobalState(
            Address(), key=cfg.NOT_CIRCULATING_LABEL_2
        )
        self.not_circulating_label_3 = GlobalState(
            Address(), key=cfg.NOT_CIRCULATING_LABEL_3
        )

    @abimethod()
    def set_asset(self, asset_id: UInt64) -> None:
        """
        Set the ASA ID for the circulating supply - Authorization: ASA Manager Address.

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
        Set non-circulating supply addresses - Authorization: ASA Manager Address.

        Args:
            address: Address to assign to the label to
            label: Not-circulating label selector
        """
        asset = Asset(self.asset_id)
        # Preconditions
        assert Txn.sender == asset.manager, err.UNAUTHORIZED
        assert Account(address.bytes).is_opted_in(asset), err.NOT_OPTED_IN
        # Effects
        match label:
            case cfg.NOT_CIRCULATING_LABEL_1:
                self.not_circulating_label_1.value = address
            case cfg.NOT_CIRCULATING_LABEL_2:
                self.not_circulating_label_2.value = address
            case cfg.NOT_CIRCULATING_LABEL_3:
                self.not_circulating_label_3.value = address
            case _:
                assert False, err.INVALID_LABEL

    @abimethod(readonly=True)
    def arc62_get_circulating_supply(self, asset_id: UInt64) -> UInt64:
        """
        Get ASA circulating supply.

        Args:
            asset_id: ASA ID of the circulating supply

        Returns:
            ASA circulating supply
        """
        asset = Asset(asset_id)
        not_circulating_1 = Account(self.not_circulating_label_1.value.bytes)
        not_circulating_2 = Account(self.not_circulating_label_2.value.bytes)
        not_circulating_3 = Account(self.not_circulating_label_3.value.bytes)
        # Preconditions
        assert asset_id == self.asset_id, err.INVALID_ASSET_ID
        # Effects
        reserve_balance = (
            UInt64(0)
            if asset.reserve == Global.zero_address
            or not asset.reserve.is_opted_in(asset)
            else asset.balance(asset.reserve)
        )
        not_circulating_balance_1 = (
            UInt64(0)
            if not_circulating_1 == Global.zero_address
            or not not_circulating_1.is_opted_in(asset)
            else asset.balance(not_circulating_1)
        )
        not_circulating_balance_2 = (
            UInt64(0)
            if not_circulating_2 == Global.zero_address
            or not not_circulating_2.is_opted_in(asset)
            else asset.balance(not_circulating_2)
        )
        not_circulating_balance_3 = (
            UInt64(0)
            if not_circulating_3 == Global.zero_address
            or not not_circulating_3.is_opted_in(asset)
            else asset.balance(not_circulating_3)
        )
        return (
            asset.total
            - reserve_balance
            - not_circulating_balance_1
            - not_circulating_balance_2
            - not_circulating_balance_3
        )
