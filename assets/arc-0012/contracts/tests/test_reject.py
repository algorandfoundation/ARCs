from fixtures import *
from algosdk.error import AlgodHTTPError


class TestReject(ARC12TestClass):
    def test_reject_axfer(self, reject):
        asa_info = self.algod.account_asset_info(self.creator.address, self.asa_id)
        assert asa_info["asset-holding"]["amount"] == 1

    def test_reject_delete_vault(self, reject):
        with pytest.raises(AlgodHTTPError) as e:
            self.algod.application_info(self.vault_client.app_id)
        assert e.match("application does not exist")

    def test_reject_vault_balance(self, reject):
        info = self.algod.account_info(self.vault_client.app_addr)
        assert info["amount"] == 0

    def test_reject_master_balance(self, reject):
        info = self.algod.account_info(self.master_client.app_addr)
        assert info["amount"] == info["min-balance"]

    def test_reject_creator_balance(self, reject):
        assert (
            self.algod.account_info(self.creator.address)["amount"]
            == self.creator_pre_reject_balance
            + 347_000  # 347_000 is the amount refunded for the vault
        )

    def test_reject_receiver_balance(self, reject):
        assert (
            self.algod.account_info(self.receiver.address)["amount"]
            == self.receiver_pre_reject_balance
        )
