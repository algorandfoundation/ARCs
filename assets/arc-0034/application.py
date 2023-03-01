from typing import Final
import beaker as bk
from pyteal import *
from beaker.lib.storage import Mapping
from typing import Literal

class Proposals(abi.NamedTuple):
    pr: abi.Field[abi.Uint64] #Proposals number
    commit: abi.Field[abi.StaticBytes[Literal[20]]] #Version
    hash: abi.Field[abi.StaticBytes[Literal[32]]] #Hash of the proposal


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

    # On the Distribute App
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

    @bk.external(authorize=bk.Authorize.only(Global.creator_address()))
    def setperiod(self):
        return(
            Seq(
                Assert(Global.latest_timestamp() > Int(0)), #Adjust based on Period
                self.period.set(Int(1))
            )
        )

    @bk.external(authorize=bk.Authorize.only(Global.creator_address()))
    def register(self, xgov_address: abi.Address, reward: abi.Uint64): #Send a payment instead?
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

    @bk.external # On the Distribute App
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
    , vote: abi.Uint64):
        return Seq(
            Assert(self.vote_check == Int(1)),
            (p := Proposals()).set(pr, commit, hash),
            self.proposal_vote[p].store_into(votes_prop := Votes()),
            (cur_vote_prop := abi.Uint64()).set(votes_prop[0]),
            (max_vote_prop := abi.Uint64()).set(votes_prop[1]),
            (new_cur_vote_prop := abi.Uint64()).set(cur_vote_prop.get() + vote.get()),
            If(new_cur_vote_prop.get() > max_vote_prop.get())
            .Then(
                (real_vote := abi.Uint64()).set(max_vote_prop.get() + vote.get() - new_cur_vote_prop.get()),
                (new_cur_vote_prop := abi.Uint64()).set(cur_vote_prop.get() + real_vote.get()),
                (new_votes_prop := Votes()).set(new_cur_vote_prop, max_vote_prop),

                self.xgov_vote[Txn.sender()].store_into(votes_xgov := Votes()),
                (cur_vote_xgov := abi.Uint64()).set(votes_xgov[0]),
                (max_vote_xgov := abi.Uint64()).set(votes_xgov[1]),
                (new_cur_vote_xgov := abi.Uint64()).set(cur_vote_xgov.get() + real_vote.get()),
                Assert(new_cur_vote_xgov.get() <= max_vote_xgov.get()),
                self.proposal_vote[p].set(new_votes_prop),
                (newvote := Votes()).set(new_cur_vote_xgov, max_vote_xgov),
                self.xgov_vote[Txn.sender()].set(newvote),
            ).Else(
                (real_vote := abi.Uint64()).set(vote.get()),
                (new_votes_prop := Votes()).set(new_cur_vote_prop, max_vote_prop),

                self.xgov_vote[Txn.sender()].store_into(votes_xgov := Votes()),
                (cur_vote_xgov := abi.Uint64()).set(votes_xgov[0]),
                (max_vote_xgov := abi.Uint64()).set(votes_xgov[1]),
                (new_cur_vote_xgov := abi.Uint64()).set(cur_vote_xgov.get() + real_vote.get()),
                Assert(new_cur_vote_xgov.get() <= max_vote_xgov.get()),
                self.proposal_vote[p].set(new_votes_prop),
                (newvote := Votes()).set(new_cur_vote_xgov, max_vote_xgov),
                self.xgov_vote[Txn.sender()].set(newvote),
            ),
        )

    @bk.external(authorize=bk.Authorize.only(Global.creator_address()))
    def endVote(self):
        return self.vote_check.set(Int(0))

if __name__ == "__main__":
    xGov().dump("./artifacts")