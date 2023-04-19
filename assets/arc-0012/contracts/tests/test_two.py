from fixtures import *
import pytest


class TestTwo(ARC12TestClass):
    def test_two_vault_asa_balance(
        self,
        second_verify_axfer,
    ):
        asa_info = self.algod.account_asset_info(
            self.vault_client.app_addr, self.asa_id
        )

        second_asa_info = self.algod.account_asset_info(
            self.vault_client.app_addr, self.second_asa_id
        )

        assert asa_info["asset-holding"]["amount"] == 1
        assert second_asa_info["asset-holding"]["amount"] == 1

    def test_two_vault_balance(
        self,
        second_verify_axfer,
    ):
        info = self.algod.account_info(self.vault_client.app_addr)
        assert info["amount"] == info["min-balance"]

    def test_two_claim_receiver_balance(
        self,
        second_claim,
    ):
        info = self.algod.account_asset_info(self.receiver.address, self.second_asa_id)
        assert info["asset-holding"]["amount"] == 1

    def test_two_claim_vault_balance(
        self,
        second_claim,
    ):
        info = self.algod.account_info(self.vault_client.app_addr)
        assert info["amount"] == info["min-balance"]
