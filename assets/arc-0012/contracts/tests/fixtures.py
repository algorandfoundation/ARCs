from algosdk.v2client.algod import AlgodClient
from beaker import *
import beaker as bkr
from pyteal import *
from pathlib import Path
from algosdk.future import transaction
from algosdk.encoding import decode_address
from algosdk.atomic_transaction_composer import (
    TransactionWithSigner,
    AtomicTransactionComposer,
)
import pytest
from contracts import Vault, Master

ARTIFACTS = Path.joinpath(Path(__file__).parent.parent)


class ARC12TestClass:
    creator: sandbox.SandboxAccount
    receiver: sandbox.SandboxAccount
    random_acct: sandbox.SandboxAccount
    algod: AlgodClient
    master_client: client.ApplicationClient
    creator_pre_vault_balance: int
    receiver_pre_vault_balance: int
    vault_client: client.ApplicationClient
    asa_id: int
    creator_pre_reject_balance: int
    receiver_pre_reject_balance: int
    second_asa_id: int

    @pytest.fixture(scope="class")
    def setup(self, request):
        accounts = sorted(
            sandbox.get_accounts(),
            key=lambda a: sandbox.clients.get_algod_client().account_info(a.address)[
                "amount"
            ],
        )

        request.cls.creator = accounts.pop()
        request.cls.receiver = accounts.pop()
        request.cls.random_acct = accounts.pop()
        request.cls.algod = sandbox.get_algod_client()

    @pytest.fixture(scope="class")
    def create_master(self, request, setup):
        master_app = Master(version=8)
        master_app.approval_program = Path.joinpath(
            ARTIFACTS, "Master.teal"
        ).read_text()

        master_client = client.ApplicationClient(
            client=sandbox.get_algod_client(),
            app=master_app,
            signer=request.cls.creator.signer,
        )

        master_client.create(args=[get_method_spec(Master.create).get_selector()])
        master_client.fund(100_000)

        request.cls.master_client = master_client

    @pytest.fixture(scope="class")
    def create_vault(self, request, create_master):
        request.cls.creator_pre_vault_balance = request.cls.algod.account_info(
            request.cls.creator.address
        )["amount"]
        request.cls.receiver_pre_vault_balance = request.cls.algod.account_info(
            request.cls.receiver.address
        )["amount"]

        sp = request.cls.algod.suggested_params()
        sp.fee = sp.min_fee * 3
        sp.flat_fee = True

        pay_txn = TransactionWithSigner(
            txn=transaction.PaymentTxn(
                sender=request.cls.creator.address,
                receiver=request.cls.master_client.app_addr,
                amt=347_000,
                sp=sp,
            ),
            signer=request.cls.creator.signer,
        )

        vault_id = request.cls.master_client.call(
            method=Master.createVault,
            receiver=request.cls.receiver.address,
            mbr_payment=pay_txn,
            boxes=[
                (
                    request.cls.master_client.app_id,
                    decode_address(request.cls.receiver.address),
                )
            ],
        ).return_value

        vault_app = Vault(version=8)
        vault_app.approval_program = Path.joinpath(ARTIFACTS, "Vault.teal").read_text()

        request.cls.vault_client = client.ApplicationClient(
            client=sandbox.get_algod_client(),
            app=vault_app,
            signer=request.cls.creator.signer,
            app_id=vault_id,
        )

    @pytest.fixture(scope="class")
    def opt_in(self, request, create_vault):
        txn = transaction.AssetConfigTxn(
            sender=request.cls.creator.address,
            sp=request.cls.master_client.get_suggested_params(),
            total=1,
            default_frozen=False,
            unit_name="LATINUM",
            asset_name="latinum",
            decimals=0,
            strict_empty_address_check=False,
        )

        stxn = txn.sign(request.cls.creator.private_key)
        txid = request.cls.algod.send_transaction(stxn)
        confirmed_txn = transaction.wait_for_confirmation(request.cls.algod, txid, 4)

        request.cls.asa_id = confirmed_txn["asset-index"]

        sp = request.cls.vault_client.get_suggested_params()
        sp.fee = sp.min_fee * 2
        sp.flat_fee = True

        pay_txn = TransactionWithSigner(
            txn=transaction.PaymentTxn(
                sender=request.cls.creator.address,
                receiver=request.cls.vault_client.app_addr,
                amt=118_500,
                sp=sp,
            ),
            signer=request.cls.creator.signer,
        )

        request.cls.vault_client.call(
            method=Vault.optIn,
            asa=request.cls.asa_id,
            mbr_payment=pay_txn,
            boxes=[
                (request.cls.vault_client.app_id, request.cls.asa_id.to_bytes(8, "big"))
            ],
        )

    @pytest.fixture(scope="class")
    def verify_axfer(self, request, opt_in):
        axfer = TransactionWithSigner(
            txn=transaction.AssetTransferTxn(
                sender=request.cls.creator.address,
                receiver=request.cls.vault_client.app_addr,
                amt=1,
                sp=request.cls.vault_client.get_suggested_params(),
                index=request.cls.asa_id,
            ),
            signer=request.cls.creator.signer,
        )

        request.cls.master_client.call(
            method=Master.verifyAxfer,
            receiver=request.cls.receiver.address,
            vault_axfer=axfer,
            vault=request.cls.vault_client.app_id,
            boxes=[
                (
                    request.cls.master_client.app_id,
                    decode_address(request.cls.receiver.address),
                )
            ],
        )

    def _claim(self, claimer):
        atc = AtomicTransactionComposer()
        claim_sp = self.algod.suggested_params()
        claim_sp.fee = claim_sp.min_fee * 7
        claim_sp.flat_fee = True

        del_sp = self.algod.suggested_params()
        del_sp.fee = 0
        del_sp.flat_fee = True

        atc.add_transaction(
            TransactionWithSigner(
                txn=transaction.AssetOptInTxn(
                    sender=claimer.address,
                    sp=self.algod.suggested_params(),
                    index=self.asa_id,
                ),
                signer=claimer.signer,
            )
        )

        atc.add_method_call(
            app_id=self.vault_client.app_id,
            sender=claimer.address,
            signer=claimer.signer,
            sp=claim_sp,
            method=application.get_method_spec(Vault.claim),
            method_args=[self.asa_id, self.creator.address, self.creator.address],
            boxes=[(self.vault_client.app_id, self.asa_id.to_bytes(8, "big"))],
        )

        atc.add_method_call(
            app_id=self.master_client.app_id,
            sender=claimer.address,
            signer=claimer.signer,
            sp=del_sp,
            method=application.get_method_spec(Master.deleteVault),
            method_args=[self.vault_client.app_id, self.creator.address],
            boxes=[
                (
                    self.master_client.app_id,
                    decode_address(claimer.address),
                )
            ],
        )

        atc.execute(self.algod, 3)

    @pytest.fixture(scope="class")
    def claim(self, request, verify_axfer):
        self._claim(request.cls.receiver)

    @pytest.fixture(scope="class")
    def reject(self, request, verify_axfer):
        request.cls.creator_pre_reject_balance = request.cls.algod.account_info(
            request.cls.creator.address
        )["amount"]

        request.cls.receiver_pre_reject_balance = request.cls.algod.account_info(
            request.cls.receiver.address
        )["amount"]

        atc = AtomicTransactionComposer()
        reject_sp = request.cls.algod.suggested_params()
        reject_sp.fee = reject_sp.min_fee * 8
        reject_sp.flat_fee = True

        del_sp = request.cls.algod.suggested_params()
        del_sp.fee = 0
        del_sp.flat_fee = True

        atc.add_method_call(
            app_id=request.cls.vault_client.app_id,
            sender=request.cls.receiver.address,
            signer=request.cls.receiver.signer,
            sp=reject_sp,
            method=application.get_method_spec(Vault.reject),
            method_args=[
                request.cls.creator.address,
                "Y76M3MSY6DKBRHBL7C3NNDXGS5IIMQVQVUAB6MP4XEMMGVF2QWNPL226CA",
                request.cls.asa_id,
                request.cls.creator.address,
            ],
            boxes=[
                (request.cls.vault_client.app_id, request.cls.asa_id.to_bytes(8, "big"))
            ],
        )

        atc.add_method_call(
            app_id=request.cls.master_client.app_id,
            sender=request.cls.receiver.address,
            signer=request.cls.receiver.signer,
            sp=del_sp,
            method=application.get_method_spec(Master.deleteVault),
            method_args=[request.cls.vault_client.app_id, request.cls.creator.address],
            boxes=[
                (
                    request.cls.master_client.app_id,
                    decode_address(request.cls.receiver.address),
                )
            ],
        )

        atc.execute(request.cls.algod, 3)

    @pytest.fixture(scope="class")
    def second_opt_in(self, request, verify_axfer):
        txn = transaction.AssetConfigTxn(
            sender=request.cls.creator.address,
            sp=request.cls.master_client.get_suggested_params(),
            total=1,
            default_frozen=False,
            unit_name="LATINUM",
            asset_name="latinum",
            decimals=0,
            strict_empty_address_check=False,
        )

        stxn = txn.sign(request.cls.creator.private_key)
        txid = request.cls.algod.send_transaction(stxn)
        confirmed_txn = transaction.wait_for_confirmation(request.cls.algod, txid, 4)

        request.cls.second_asa_id = confirmed_txn["asset-index"]

        sp = request.cls.vault_client.get_suggested_params()
        sp.fee = sp.min_fee * 2
        sp.flat_fee = True

        pay_txn = TransactionWithSigner(
            txn=transaction.PaymentTxn(
                sender=request.cls.creator.address,
                receiver=request.cls.vault_client.app_addr,
                amt=118_500,
                sp=sp,
            ),
            signer=request.cls.creator.signer,
        )

        request.cls.vault_client.call(
            method=Vault.optIn,
            asa=request.cls.second_asa_id,
            mbr_payment=pay_txn,
            boxes=[
                (
                    request.cls.vault_client.app_id,
                    request.cls.second_asa_id.to_bytes(8, "big"),
                )
            ],
        )

    @pytest.fixture(scope="class")
    def second_verify_axfer(
        self,
        request,
        second_opt_in,
    ):
        axfer = TransactionWithSigner(
            txn=transaction.AssetTransferTxn(
                sender=request.cls.creator.address,
                receiver=request.cls.vault_client.app_addr,
                amt=1,
                sp=request.cls.vault_client.get_suggested_params(),
                index=request.cls.second_asa_id,
            ),
            signer=request.cls.creator.signer,
        )

        request.cls.master_client.call(
            method=Master.verifyAxfer,
            receiver=request.cls.receiver.address,
            vault_axfer=axfer,
            vault=request.cls.vault_client.app_id,
            boxes=[
                (
                    request.cls.master_client.app_id,
                    decode_address(request.cls.receiver.address),
                )
            ],
        )

    @pytest.fixture(scope="class")
    def second_claim(
        self,
        request,
        second_opt_in,
    ):
        atc = AtomicTransactionComposer()
        claim_sp = request.cls.algod.suggested_params()
        claim_sp.fee = claim_sp.min_fee * 7
        claim_sp.flat_fee = True

        atc.add_transaction(
            TransactionWithSigner(
                txn=transaction.AssetOptInTxn(
                    sender=request.cls.receiver.address,
                    sp=request.cls.algod.suggested_params(),
                    index=request.cls.second_asa_id,
                ),
                signer=request.cls.receiver.signer,
            )
        )

        atc.add_method_call(
            app_id=request.cls.vault_client.app_id,
            sender=request.cls.receiver.address,
            signer=request.cls.receiver.signer,
            sp=claim_sp,
            method=application.get_method_spec(Vault.claim),
            method_args=[
                request.cls.second_asa_id,
                request.cls.creator.address,
                request.cls.creator.address,
            ],
            boxes=[
                (
                    request.cls.vault_client.app_id,
                    request.cls.second_asa_id.to_bytes(8, "big"),
                )
            ],
        )

        atc.execute(request.cls.algod, 3)

    @pytest.fixture(scope="class")
    def second_reject(
        self,
        request,
        second_opt_in,
    ):
        request.cls.creator_pre_reject_balance = request.cls.algod.account_info(
            request.cls.creator.address
        )["amount"]

        request.cls.receiver_pre_reject_balance = request.cls.algod.account_info(
            request.cls.receiver.address
        )["amount"]

        atc = AtomicTransactionComposer()
        reject_sp = request.cls.algod.suggested_params()
        reject_sp.fee = reject_sp.min_fee * 8
        reject_sp.flat_fee = True

        atc.add_method_call(
            app_id=request.cls.vault_client.app_id,
            sender=request.cls.receiver.address,
            signer=request.cls.receiver.signer,
            sp=reject_sp,
            method=application.get_method_spec(Vault.reject),
            method_args=[
                request.cls.creator.address,
                "Y76M3MSY6DKBRHBL7C3NNDXGS5IIMQVQVUAB6MP4XEMMGVF2QWNPL226CA",
                request.cls.second_asa_id,
                request.cls.creator.address,
            ],
            boxes=[
                (
                    request.cls.vault_client.app_id,
                    request.cls.second_asa_id.to_bytes(8, "big"),
                )
            ],
        )

        atc.execute(request.cls.algod, 3)
