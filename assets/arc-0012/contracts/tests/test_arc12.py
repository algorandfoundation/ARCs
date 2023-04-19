import pytest
from fixtures import *
from algosdk.error import AlgodHTTPError


class TestARC12(ARC12TestClass):
    @pytest.mark.create_master
    def test_create_master(self, create_master):
        pass
    @pytest.mark.create_vault
    def test_create_vault_id(self, create_vault):
        assert self.vault_client.app_id > 0

    @pytest.mark.create_vault
    def test_create_vault_box_name(self, create_vault):
        box_names = self.master_client.get_box_names()
        assert len(box_names) == 1
        assert box_names[0] == decode_address(self.receiver.address)

    @pytest.mark.create_vault
    def test_create_vault_box_value(self, create_vault):
        box_value = int.from_bytes(
            self.master_client.get_box_contents(decode_address(self.receiver.address)),
            byteorder="big",
        )
        assert box_value == self.vault_client.app_id

    @pytest.mark.create_vault
    def test_create_vault_creator(self, create_vault):
        assert decode_address(self.creator.address) == bytes.fromhex(
            self.vault_client.get_application_state()["creator"]
        )

    @pytest.mark.create_vault
    def test_create_vault_receiver(self, create_vault):
        assert decode_address(self.receiver.address) == bytes.fromhex(
            self.vault_client.get_application_state()["receiver"]
        )

    @pytest.mark.create_vault
    def test_create_vault_mbr(self, create_vault):
        info = self.algod.account_info(self.master_client.app_addr)
        assert info["amount"] == info["min-balance"]

    @pytest.mark.opt_in
    def test_opt_in_box_name(self, opt_in):
        box_names = self.vault_client.get_box_names()
        assert len(box_names) == 1
        assert box_names[0] == self.asa_id.to_bytes(8, "big")

    @pytest.mark.opt_in
    def test_opt_in_box_value(self, opt_in):
        box_value = self.vault_client.get_box_contents(self.asa_id.to_bytes(8, "big"))
        assert box_value == decode_address(self.creator.address)

    @pytest.mark.opt_in
    def test_opt_in_mbr(self, opt_in):
        info = self.algod.account_info(self.vault_client.app_addr)
        assert info["amount"] == info["min-balance"]

    @pytest.mark.verify_axfer
    def test_verify_axfer(self, verify_axfer):
        asa_info = self.algod.account_asset_info(
            self.vault_client.app_addr, self.asa_id
        )
        assert asa_info["asset-holding"]["amount"] == 1

    @pytest.mark.claim
    def test_claim_axfer(self, claim):
        asa_info = self.algod.account_asset_info(self.receiver.address, self.asa_id)
        assert asa_info["asset-holding"]["amount"] == 1

    @pytest.mark.claim
    def test_claim_delete_vault(self, claim):
        with pytest.raises(AlgodHTTPError) as e:
            self.algod.application_info(self.vault_client.app_id)
        assert e.match("application does not exist")

    @pytest.mark.claim
    def test_claim_vault_balance(self, claim):
        info = self.algod.account_info(self.vault_client.app_addr)
        assert info["amount"] == 0

    @pytest.mark.claim
    def test_claim_master_balance(self, claim):
        info = self.algod.account_info(self.master_client.app_addr)
        assert info["amount"] == info["min-balance"]

    @pytest.mark.claim
    def test_claim_creator_balance(self, claim):
        amt = self.algod.account_info(self.creator.address)["amount"]
        expected_amt = self.creator_pre_vault_balance - 1_000 * 10
        assert amt == expected_amt

    @pytest.mark.claim
    def test_claim_receiver_balance(self, claim):
        amt = self.algod.account_info(self.receiver.address)["amount"]
        expected_amt = self.receiver_pre_vault_balance - 1_000 * 8
        assert amt == expected_amt
