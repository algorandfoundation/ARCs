from typing import Final
from pyteal import *
from beaker import *
from beaker.lib.storage import Mapping

algod_client = sandbox.get_algod_client()

# Use a box per member to denote membership parameters
class SR(abi.NamedTuple):
    address: abi.Field[abi.Address]
    round: abi.Field[abi.Uint64]

class Master(Application):
    public_key: Final[ApplicationStateValue] = ApplicationStateValue(
        stack_type=TealType.bytes,
        default=Bytes(""),
        descr="A Public Key use to encrypt the message",
    )
    inbox = Mapping(SR, abi.String)

    whitelist = Mapping(abi.Address, abi.String)

    @external(authorize=Authorize.only(Global.creator_address()))
    def set_public_key(self, public_key: abi.String):
        return self.public_key.set(public_key.get())

    @external(authorize=Authorize.only(Global.creator_address()))
    def authorize(self, address_to_add: abi.Address, info: abi.String):
        return Seq(
            (s := abi.Address()).set(address_to_add.get()),
            self.whitelist[s].set(info))  

    @external
    def write(self, text: abi.String):
        return Seq(
            (s := abi.Address()).set(Txn.sender()),
            Assert(self.whitelist[s].exists()),
            (r := abi.Uint64()).set(Global.round()),
            (v := SR()).set(s, r),
            self.inbox[v].set(text))  

    @delete(authorize=Authorize.only(Global.creator_address()))
    def delete(self):
        return Approve()   

    @external
    def remove(self, address: abi.Address, round: abi.Uint64):
        return Seq(
            (v := SR()).set(address, round),
            Pop(self.inbox[v].delete())
            )


if __name__ == "__main__":
    app = Master()
    app.generate_teal()
    app.dump("./artifacts")