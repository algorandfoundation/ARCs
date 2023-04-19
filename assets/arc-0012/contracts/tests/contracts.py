from beaker import *
from pyteal import *


class Master(Application):
    @external
    def create(self):
        return Reject()

    @external
    def createVault(
        self,
        receiver: abi.Account,
        mbr_payment: abi.PaymentTransaction,
        *,
        output: abi.Uint64,
    ):
        return Reject()

    @external
    def verifyAxfer(
        self,
        receiver: abi.Account,
        vault_axfer: abi.AssetTransferTransaction,
        vault: abi.Application,
    ):
        return Reject()

    @external
    def getVaultID(self, receiver: abi.Account, *, output: abi.Uint64):
        return Reject()

    @external
    def getVaultAddr(self, receiver: abi.Account, *, output: abi.Address):
        return Reject()

    @external
    def deleteVault(self, vault: abi.Application, creator: abi.Account):
        return Reject()


class Vault(Application):
    @external
    def optIn(self, asa: abi.Asset, mbr_payment: abi.PaymentTransaction):
        return Reject()

    @external
    def claim(
        self,
        asa: abi.Asset,
        creator: abi.Account,
        asa_mbr_funder: abi.Account,
    ):
        return Reject()

    @external
    def reject(
        self,
        asa_creator: abi.Account,
        fee_sink: abi.Account,
        asa: abi.Asset,
        vault_creator: abi.Account,
    ):
        return Reject()
