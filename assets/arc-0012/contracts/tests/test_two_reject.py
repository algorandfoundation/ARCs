from fixtures import *
from algosdk.error import AlgodHTTPError
import pytest


class TestTwoReject(ARC12TestClass):
    def test_two_reject_vault_asa_balance(self, second_reject):
        with pytest.raises(AlgodHTTPError) as e:
            self.algod.account_asset_info(self.receiver.address, self.second_asa_id)
        assert e.match("account asset info not found")

    def test_two_claim_vault_balance(
        self,
        second_reject,
    ):
        info = self.algod.account_info(self.vault_client.app_addr)
        assert info["amount"] == info["min-balance"]
