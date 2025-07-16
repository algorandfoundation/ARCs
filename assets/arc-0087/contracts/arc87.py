from algopy import ARC4Contract, BoxMap,BoxRef, String, log, Txn, Global, TemplateVar, UInt64, itxn
from algopy.arc4 import abimethod, baremethod

# This contract is a simple key-value store that allows you to set and get values


class ARC87(ARC4Contract):
    """Simple Contract for storing key-value pairs.

    Attributes
    ----------
    public : BoxMap
        Boxes that store the public object
    """
    def __init__(self) -> None:
        self.public = BoxMap(String, String, key_prefix=b"o_")

    @baremethod(allow_actions=["DeleteApplication"])
    def on_delete(self) -> None:
        assert Txn.sender == Global.creator_address
        assert TemplateVar[UInt64]("DELETABLE")

    @abimethod()
    def set(self, path: String, value: String) -> None:
        assert Txn.sender == Global.creator_address
        self.public[path] = value

    @abimethod()
    def get(self, path: String) -> String:
        assert Txn.sender == Global.creator_address
        assert path in self.public
        return self.public[path]

    @abimethod()
    def remove(self, path: String) -> None:
        assert Txn.sender == Global.creator_address
        assert path in self.public
        ref = BoxRef(key=path)
        assert ref.delete()

    @abimethod()
    def concat(self, path: String, value: String) -> None:
        assert Txn.sender == Global.creator_address
        log(b"Adding {value} to {path}")
        assert path in self.public
        self.public[path] = self.public[path] + value

    @abimethod()
    def reclaim(self, amount: UInt64) -> None:
        assert Txn.sender == Global.creator_address
        log(b"Reclaiming MBR")
        itxn.Payment(sender=Global.current_application_address, receiver=Global.creator_address, amount=amount, fee=0).submit()
