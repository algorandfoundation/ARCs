from algopy import (
    ARC4Contract,
    BoxMap,
    Bytes,
    Global,
    String,
    Txn,
    UInt64,
    itxn,
    op,
    subroutine,
)
from algopy.arc4 import Address, abimethod


class Recovery(ARC4Contract):
    def __init__(self) -> None:
        self.rekey_records = BoxMap(Address, Address, key_prefix=b"records:")
        self.rekey_deadlines = BoxMap(Address, UInt64, key_prefix=b"deadline:")
        self.rekey_cancelled = BoxMap(Address, bool, key_prefix=b"cancelled:")
        self.rekey_notice_periods = BoxMap(Address, UInt64, key_prefix=b"notice:")
        self.db = BoxMap(Address, String, key_prefix=b"pubsig:")
        self.authorized_callers = BoxMap(Address, Address, key_prefix=b"caller:")

    @subroutine()
    def _build_notice_message(self, rounds: UInt64) -> Bytes:
        prefix = Bytes(b"someone wants to rekey your wallet, you have ")
        suffix = Bytes(b" rounds to prevent the operation")
        rounds_bytes = op.itob(rounds)
        return op.concat(prefix, op.concat(rounds_bytes, suffix))

    @abimethod()
    def set_public_sig(self, account: Address, sig: String) -> bool:
        self.db[account] = sig
        return self.db[account] == sig

    @abimethod(readonly=True)
    def get_public_sig(self, account: Address) -> String:
        return self.db[account]

    @abimethod()
    def hello(self, name: String) -> String:
        return "Hello, " + name

    @abimethod()
    def set_rekey_record(
        self,
        wallet: Address,
        allowed_sender: Address,
        rekey_target: Address,
        notification_window: UInt64,
    ) -> None:
        assert Txn.sender == Global.creator_address
        self.rekey_records[wallet] = rekey_target
        self.rekey_deadlines[wallet] = UInt64(0)
        self.rekey_cancelled[wallet] = False
        self.rekey_notice_periods[wallet] = notification_window
        self.authorized_callers[wallet] = allowed_sender

    @abimethod()
    def rekey_wallet(self, wallet: Address, new_address: Address) -> None:
        stored_address, exists = self.rekey_records.maybe(wallet)
        assert exists
        assert stored_address == new_address
        authorized_sender, has_authorized = self.authorized_callers.maybe(wallet)
        assert has_authorized

        cancelled, cancel_exists = self.rekey_cancelled.maybe(wallet)
        assert cancel_exists
        assert not cancelled

        deadline, deadline_exists = self.rekey_deadlines.maybe(wallet)
        assert deadline_exists
        notice_period, notice_exists = self.rekey_notice_periods.maybe(wallet)
        assert notice_exists
        current_round = Global.round
        if deadline == UInt64(0):
            assert Txn.sender == authorized_sender.native
            self.rekey_deadlines[wallet] = current_round + notice_period
            note_bytes = self._build_notice_message(notice_period)
            itxn.Payment(
                receiver=Global.creator_address,
                amount=0,
                note=note_bytes,
                fee=2 * Global.min_txn_fee,
            ).submit()

        else:
            assert Txn.sender == Global.creator_address
            assert current_round >= deadline
            self.rekey_deadlines[wallet] = UInt64(0)
            self.rekey_cancelled[wallet] = False
        return

    @abimethod()
    def cancel_rekey(self, wallet: Address) -> None:
        assert Txn.sender == Global.creator_address
        self.rekey_cancelled[wallet] = True
        self.rekey_deadlines[wallet] = UInt64(0)
