from typing import Final
import beaker as bk
from pyteal import *
from beaker.lib.storage import Mapping
from typing import Literal

class Proposals(abi.NamedTuple):
    pr: abi.Field[abi.Uint64]
    commit: abi.Field[abi.StaticBytes[Literal[20]]]
    hash: abi.Field[abi.StaticBytes[Literal[32]]]


class Votes(abi.NamedTuple):
    current_vote: abi.Field[abi.Uint64]
    max_vote: abi.Field[abi.Uint64]


class xGov(bk.Application):
    
    max_vote: Final[bk.ApplicationStateValue] = bk.ApplicationStateValue(
        stack_type=TealType.uint64,
        default=Int(0),
        descr="max number of votes",
    )

    submit_check: Final[bk.ApplicationStateValue] = bk.ApplicationStateValue(
        stack_type=TealType.uint64,
        default=Int(0),
        descr="Bool to check submission",
    )

    vote_check: Final[bk.ApplicationStateValue] = bk.ApplicationStateValue(
        stack_type=TealType.uint64,
        default=Int(0),
        descr="Bool to check vote",
    )

    max_algo: Final[bk.ApplicationStateValue] = bk.ApplicationStateValue(
        stack_type=TealType.uint64,
        default=Int(0),
        descr="Algo available at the start of gov",
    )

    period: Final[bk.ApplicationStateValue] = bk.ApplicationStateValue(
        stack_type=TealType.uint64,
        default=Int(0),
        descr="Check if algo can be distributed",
    )
    proposal_vote = Mapping(Proposals, Votes)
    xgov_vote = Mapping(abi.Address, Votes)

    # @bk.external
    # def add(self, 
    # a: abi.StaticBytes[Literal[64]], 
    # b: abi.StaticBytes[Literal[64]], *, 
    # # output: abi.StaticBytes[Literal[128]]
    # output: abi.DynamicBytes
    # ):
    #     return output.set(BytesMul(a.get(), b.get()))
    #     return output.set(Bytes("https://pyteal.readthedocs.io/en/stable/api.html#pyteal.BytesAdd"))

    @bk.external(authorize=bk.Authorize.only(Global.creator_address()))
    def setperiod(self):#Check that every xgov is qualified #specify round
        return(
            Seq(
                self.period.set(Int(1))
            )
        )

    @bk.external(authorize=bk.Authorize.only(Global.creator_address()))
    def register(self, xgov_address: abi.Address, reward: abi.Uint64):
        return Seq(
            self.max_vote.set(self.max_vote + reward.get()),
            self.max_algo.set(self.max_algo + reward.get()),
            (z := abi.Uint64()).set(Int(0)),
            (v := Votes()).set(z, reward),
            self.xgov_vote[xgov_address].set(v) 
        )

    @bk.external(authorize=bk.Authorize.only(Global.creator_address()))
    def forfeit(self, xgov_address: abi.Address):
        return Seq(
            self.xgov_vote[xgov_address].store_into(v := Votes()),
            (c := abi.Uint64()).set(v[1]),
            self.max_vote.set(self.max_vote - c.get()),
            Pop(self.xgov_vote[xgov_address].delete())
        )

    @bk.external
    def distribute(self):
        return Seq(
            Assert(self.period == Int(1)),
            self.xgov_vote[Txn.sender()].store_into(votes_xgov := Votes()),
            (cur_vote_xgov := abi.Uint64()).set(votes_xgov[0]),
            (max_vote_xgov := abi.Uint64()).set(votes_xgov[1]),
            Assert(cur_vote_xgov.get() == max_vote_xgov.get()),
            (a := abi.Uint64()).set(max_vote_xgov.get() * self.max_algo / self.max_vote),
            InnerTxnBuilder.Execute(
            {
                TxnField.type_enum: TxnType.Payment,
                TxnField.amount: a.get(),
                TxnField.receiver: Txn.sender(),
            }),
            Pop(self.xgov_vote[Txn.sender()].delete())
        )

    @bk.external(authorize=bk.Authorize.only(Global.creator_address()))
    def startSubmit(self):
        return self.submit_check.set(Int(1))

    @bk.external(authorize=bk.Authorize.only(Global.creator_address()))
    def submitProposal(self,
    #    proposal: Proposals, 
    pr:abi.Uint64,
    commit: abi.StaticBytes[Literal[20]],
    hash: abi.StaticBytes[Literal[32]],
      max: abi.Uint64):
        return Seq(
            Assert(self.submit_check.get() == Int(1)),
            (z := abi.Uint64()).set(Int(0)),
            (v := Votes()).set(z, max),
            (p := Proposals()).set(pr, commit, hash),
            self.proposal_vote[p].set(v)
        )

    @bk.external(authorize=bk.Authorize.only(Global.creator_address()))
    def endSubmit(self):
        return self.submit_check.set(Int(0))

    @bk.external(authorize=bk.Authorize.only(Global.creator_address()))
    def removeProposal(self,    
    pr:abi.Uint64,
    commit: abi.StaticBytes[Literal[20]],
    hash: abi.StaticBytes[Literal[32]]):
        return Seq(
            Assert(self.submit_check.get() == Int(1)),
            (p := Proposals()).set(pr, commit, hash),
            Pop(self.proposal_vote[p].delete())
        )

    @bk.external(authorize=bk.Authorize.only(Global.creator_address()))
    def startVote(self):
        return self.vote_check.set(Int(1))
        
    @bk.external
    def vote(self,     
    pr:abi.Uint64,
    commit: abi.StaticBytes[Literal[20]],
    hash: abi.StaticBytes[Literal[32]]
    , vote: abi.Uint64): # Add Proposal
        return Seq(
            Assert(self.vote_check == Int(1)),
            (p := Proposals()).set(pr, commit, hash),
            self.proposal_vote[p].store_into(votes_prop := Votes()),
            (cur_vote_prop := abi.Uint64()).set(votes_prop[0]),
            (max_vote_prop := abi.Uint64()).set(votes_prop[1]),
            (new_cur_vote_prop := abi.Uint64()).set(cur_vote_prop.get() + vote.get()),
            (new_votes_prop := Votes()).set(new_cur_vote_prop, max_vote_prop),
            Assert(new_cur_vote_prop.get() <= max_vote_prop.get()),
            self.xgov_vote[Txn.sender()].store_into(votes_xgov := Votes()),
            (cur_vote_xgov := abi.Uint64()).set(votes_xgov[0]),
            (max_vote_xgov := abi.Uint64()).set(votes_xgov[1]),
            (new_cur_vote_xgov := abi.Uint64()).set(cur_vote_xgov.get() + vote.get()),
            Assert(new_cur_vote_xgov.get() <= max_vote_xgov.get()),
            self.proposal_vote[p].set(new_votes_prop),
            (newvote := Votes()).set(new_cur_vote_xgov, max_vote_xgov),
            self.xgov_vote[Txn.sender()].set(newvote),
        )

    @bk.external(authorize=bk.Authorize.only(Global.creator_address()))
    def endVote(self):
        return self.vote_check.set(Int(0))


class IARC72Metadata(bk.Application):
    @bk.external
    def tokenURI(self, tokenId: abi.Uint64, *, output:abi.StaticBytes[Literal[32]]): 
        return output.set(Bytes("https://www.algorand.foundation/"))

class ARC73(bk.Application): 
    @bk.external
    def supportsInterface(self, interfaceID: abi.StaticBytes[Literal[4]], *, output:abi.Bool): # 0x4e22a3ba
        return output.set(interfaceID.get() == Txn.application_args[0])

class IARC72(bk.Application):
    @bk.external(read_only=True)# check why not working
    def ownerOf(self, tokenId: abi.Uint64, *, output:abi.Address): 
        return output.set(
            Addr("3C5IFPAZLET3FLGGFK5AXN7NISVD3OCOMEZJESCXUNUHDOIPMVKYB4DILM")
            )
    @bk.external
    def balanceOf(self, address: abi.Address, *, output:abi.Uint64): 
        return output.set(Eq(
                Addr("3C5IFPAZLET3FLGGFK5AXN7NISVD3OCOMEZJESCXUNUHDOIPMVKYB4DILM"),
                address.get()
            ))
    @bk.external
    def transferFrom(self, source: abi.Address, destination: abi.Address, tokenId: abi.Uint64): # 0x52e43cad
        return Log(Concat(
            MethodSignature("transferFrom(address,address,uint64)"),
            source.get(),
            destination.get()
            ))

    @bk.external
    def Swapped(self, a: abi.Uint64 ,b: abi.Uint64): # 0xe24b28eb
        # return Log(Bytes("base64","HMvZJQAAAAAAAAAqAAAAAAAAAGQ="))
        return Log(Concat(
            MethodSignature("Swapped(uint64,uint64)"),
            Itob(a.get()),
            Itob(b.get()),
        ))

class ARC72(IARC72Metadata, IARC72, ARC73): 
    @bk.external
    def supportsInterface(self, interfaceID: abi.StaticBytes[Literal[4]], *, output:abi.Bool): # 0x4e22a3ba
        return output.set(Or(
            interfaceID.get() == MethodSignature("supportsInterface(byte[4])bool")
            ))
            # interfaceID.get() == Bytes("base16", "4e22a3ba")
        # interfaceID.get() == Txn.application_args[0], #



        # return output.set(Bytes("base16", "0xb5fde7d6") == Txn.application_args[0])
    # def supportsInterface(self, interfaceID: abi.Uint64, *, output:abi.Uint64): # Teal: 529
    #     return output.set(interfaceID.get() == Int(73))
    # def supportsInterface(self, interfaceID: abi.Uint64, *, output:abi.Uint64): # Teal: 610
    #     return output.set(Or(
    #         interfaceID.get() == Int(73),
    #         interfaceID.get() == Int(72),
    #         interfaceID.get() == Int(71),
    #         interfaceID.get() == Int(70),
    #     ))
    # def supportsInterface(self, interfaceID: abi.Uint64, *, output:abi.Uint64): ##Allows Number up to 255 #Teal: 721
    #     array= Bytes("base16","49484746")
    #     i = ScratchVar(TealType.uint64)
    #     return For(i.store(Int(1)), i.load() < Len(array), i.store(i.load() + Int(1))).Do(
    #         output.set(Or(output.get(), interfaceID.get() == GetByte(array, i.load()))),
    #     )
    # def supportsInterface(self, interfaceID: abi.Uint64, *, output:abi.Uint64): ##Allows Number up to 65535 #Teal: 764
    #     array= Bytes("base16","004900480047ffff")
    #     i = ScratchVar(TealType.uint64)
    #     return For(i.store(Int(1)), i.load() < Len(array), i.store(i.load() + Int(2))).Do(
    #         output.set(Or(output.get(), interfaceID.get() == Add(
    #                 GetByte(array, i.load()), 
    #                 GetByte(array, i.load()- Int(1)) * Int(256)
    #                 ))),
    #     )

if __name__ == "__main__":
    xGov().dump("./artifacts")