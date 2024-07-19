from algopy import ARC4Contract, Account, UInt64, itxn, BoxMap
from algopy.arc4 import abimethod, String


class SmartApp(ARC4Contract):
    def __init__(self) -> None:
        self.db = BoxMap(Account, String, key_prefix="")

    @abimethod()
    def opt_in(self, id: UInt64, account: Account) -> None:
        itxn.AssetTransfer(
            asset_amount=0,
            xfer_asset=id,
            sender=account,
            asset_receiver=account,
            fee=1000,
        ).submit()

    @abimethod()
    def set_public_sig(self, account: Account, sig: String) -> bool:
        self.db[account] = sig
        return self.db[account] == sig

    @abimethod(readonly=True)
    def get_public_sig(self, account: Account) -> String:
        return self.db[account]
