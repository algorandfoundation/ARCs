from typing import Final
import beaker as bk
import pyteal as pt
from beaker.lib.storage import Mapping
from typing import Literal
algod_client = bk.sandbox.get_algod_client()

# Use a box per member to denote membership parameters
class SR(pt.abi.NamedTuple):
    address: pt.abi.Field[pt.abi.Address]
    round: pt.abi.Field[pt.abi.Uint64]

class Master(bk.Application):
    public_key: Final[bk.ApplicationStateValue] = bk.ApplicationStateValue(
        stack_type=pt.TealType.bytes,
        default=pt.Bytes(""),
        descr="A Public Key use to encrypt the message",
    )
    arc: Final[bk.ApplicationStateValue] = bk.ApplicationStateValue(
        stack_type=pt.TealType.bytes,
        default=pt.Bytes(""),
        descr="Format use to encrypt the public key",
    )
    inbox = Mapping(SR, pt.abi.DynamicBytes)

    whitelist = Mapping(pt.abi.Address, pt.abi.String)

    @bk.external(authorize=bk.Authorize.only(pt.Global.creator_address()))
    def set_public_key(self, key_encryption: pt.abi.String, public_key: pt.abi.StaticBytes[Literal[32]]):
        return pt.Seq(
            self.public_key.set(public_key.get()),
            self.arc.set(key_encryption.get()),
        )

    @bk.external(authorize=bk.Authorize.only(pt.Global.creator_address()))
    def authorize(self, address_to_add: pt.abi.Address, info: pt.abi.String):
        return pt.Seq(
            (s := pt.abi.Address()).set(address_to_add.get()),
            self.whitelist[s].set(info))  

    @bk.external
    def write(self, text: pt.abi.DynamicBytes):
        return pt.Seq(
            (s := pt.abi.Address()).set(pt.Txn.sender()),
            pt.Assert(self.whitelist[s].exists()),
            (r := pt.abi.Uint64()).set(pt.Global.round()),
            (v := SR()).set(s, r),
            self.inbox[v].set(text))  

    @bk.delete(authorize=bk.Authorize.only(pt.Global.creator_address()))
    def delete(self):
        return pt.Approve()   

    @bk.external
    def remove(self, address: pt.abi.Address, round: pt.abi.Uint64):
        return pt.Seq(
            (v := SR()).set(address, round),
            pt.Pop(self.inbox[v].delete())
            )

if __name__ == "__main__":
    app = Master()
    app.generate_teal()
    app.dump("./artifacts")