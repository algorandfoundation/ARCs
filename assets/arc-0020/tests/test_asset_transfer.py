import pytest
from algokit_utils import LogicError, OnCompleteCallParameters
from algokit_utils.beta.account_manager import AddressAndSigner

import smart_contracts.errors as err
from smart_contracts.artifacts.smart_asa.smart_asa_client import SmartAsaClient

from . import utils


class TestMint:
    @pytest.mark.parametrize("asa_config", [False], indirect=True)
    def test_pass_as_reserve(
        self,
        reserve: AddressAndSigner,
        smart_asa_client: SmartAsaClient,
        receiver: AddressAndSigner,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        assert (
            smart_asa_client.get_circulating_supply(
                asset=smart_asa.smart_asa_id
            ).return_value
            == 0
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client, receiver.address, smart_asa.smart_asa_id
            )
            == 0
        )
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        smart_asa_client.asset_transfer(
            xfer_asset=smart_asa.smart_asa_id,
            asset_amount=smart_asa.total,
            asset_sender=smart_asa_client.app_address,
            asset_receiver=receiver.address,
            transaction_parameters=OnCompleteCallParameters(
                suggested_params=sp,
                signer=reserve.signer,
                sender=reserve.address,
            ),
        )
        assert (
            smart_asa_client.get_circulating_supply(
                asset=smart_asa.smart_asa_id
            ).return_value
            == smart_asa.total
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client, receiver.address, smart_asa.smart_asa_id
            )
            == smart_asa.total
        )

    def test_pass_as_reserve_and_clawback(
        self,
        smart_asa_client: SmartAsaClient,
        reserve_and_clawback: AddressAndSigner,
        receiver: AddressAndSigner,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        assert (
            smart_asa_client.get_circulating_supply(
                asset=smart_asa.smart_asa_id
            ).return_value
            == 0
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client, receiver.address, smart_asa.smart_asa_id
            )
            == 0
        )
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        smart_asa_client.asset_transfer(
            xfer_asset=smart_asa.smart_asa_id,
            asset_amount=smart_asa.total,
            asset_sender=smart_asa_client.app_address,
            asset_receiver=receiver.address,
            transaction_parameters=OnCompleteCallParameters(
                suggested_params=sp,
                signer=reserve_and_clawback.signer,
                sender=reserve_and_clawback.address,
            ),
        )
        assert (
            smart_asa_client.get_circulating_supply(
                asset=smart_asa.smart_asa_id
            ).return_value
            == smart_asa.total
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client, receiver.address, smart_asa.smart_asa_id
            )
            == smart_asa.total
        )

    def test_fail_unauthorized(
        self,
        eve: AddressAndSigner,
        smart_asa_client: SmartAsaClient,
        receiver: AddressAndSigner,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        with pytest.raises(LogicError, match=err.UNAUTHORIZED_RESERVE):
            smart_asa_client.asset_transfer(
                xfer_asset=smart_asa.smart_asa_id,
                asset_amount=smart_asa.total,
                asset_sender=smart_asa_client.app_address,
                asset_receiver=receiver.address,
                transaction_parameters=OnCompleteCallParameters(
                    suggested_params=sp,
                    signer=eve.signer,
                    sender=eve.address,
                ),
            )

    def test_fail_as_clawback(
        self,
        clawback: AddressAndSigner,
        smart_asa_client: SmartAsaClient,
        receiver: AddressAndSigner,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        with pytest.raises(LogicError, match=err.UNAUTHORIZED_RESERVE):
            smart_asa_client.asset_transfer(
                xfer_asset=smart_asa.smart_asa_id,
                asset_amount=smart_asa.total,
                asset_sender=smart_asa_client.app_address,
                asset_receiver=receiver.address,
                transaction_parameters=OnCompleteCallParameters(
                    suggested_params=sp,
                    signer=clawback.signer,
                    sender=clawback.address,
                ),
            )

    def test_fail_global_frozen(self) -> None:
        pass  # TODO: Once asset_freeze is available

    @pytest.mark.parametrize("asa_config", [True], indirect=True)
    def test_fail_frozen_receiver(
        self,
        reserve: AddressAndSigner,
        smart_asa_client: SmartAsaClient,
        receiver: AddressAndSigner,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        with pytest.raises(LogicError, match=err.RECEIVER_FROZEN):
            smart_asa_client.asset_transfer(
                xfer_asset=smart_asa.smart_asa_id,
                asset_amount=smart_asa.total,
                asset_sender=smart_asa_client.app_address,
                asset_receiver=receiver.address,
                transaction_parameters=OnCompleteCallParameters(
                    suggested_params=sp,
                    signer=reserve.signer,
                    sender=reserve.address,
                ),
            )

    @pytest.mark.parametrize("asa_config", [False], indirect=True)
    def test_fail_self_minting(
        self,
        reserve: AddressAndSigner,
        smart_asa_client: SmartAsaClient,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        with pytest.raises(LogicError, match=err.SELF_MINT):
            smart_asa_client.asset_transfer(
                xfer_asset=smart_asa.smart_asa_id,
                asset_amount=smart_asa.total,
                asset_sender=smart_asa_client.app_address,
                asset_receiver=smart_asa_client.app_address,
                transaction_parameters=OnCompleteCallParameters(
                    suggested_params=sp,
                    signer=reserve.signer,
                    sender=reserve.address,
                ),
            )

    def test_fail_over_minting(
        self,
        reserve: AddressAndSigner,
        smart_asa_client: SmartAsaClient,
        receiver: AddressAndSigner,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        with pytest.raises(LogicError, match=err.OVER_MINT):
            smart_asa_client.asset_transfer(
                xfer_asset=smart_asa.smart_asa_id,
                asset_amount=smart_asa.total + 1,
                asset_sender=smart_asa_client.app_address,
                asset_receiver=receiver.address,
                transaction_parameters=OnCompleteCallParameters(
                    suggested_params=sp,
                    signer=reserve.signer,
                    sender=reserve.address,
                ),
            )


class TestBurn:
    @pytest.mark.parametrize("asa_config", [False], indirect=True)
    def test_pass_as_reserve(
        self, smart_asa_client: SmartAsaClient, reserve_with_supply: AddressAndSigner
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        assert (
            smart_asa_client.get_circulating_supply(
                asset=smart_asa.smart_asa_id
            ).return_value
            == smart_asa.total
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client,
                reserve_with_supply.address,
                smart_asa.smart_asa_id,
            )
            == smart_asa.total
        )
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        smart_asa_client.asset_transfer(
            xfer_asset=smart_asa.smart_asa_id,
            asset_amount=smart_asa.total,
            asset_sender=reserve_with_supply.address,
            asset_receiver=smart_asa_client.app_address,
            transaction_parameters=OnCompleteCallParameters(
                suggested_params=sp,
                signer=reserve_with_supply.signer,
                sender=reserve_with_supply.address,
            ),
        )
        assert (
            smart_asa_client.get_circulating_supply(
                asset=smart_asa.smart_asa_id
            ).return_value
            == 0
        )

        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client,
                reserve_with_supply.address,
                smart_asa.smart_asa_id,
            )
            == 0
        )

    def test_pass_as_reserve_and_clawback(
        self,
        smart_asa_client: SmartAsaClient,
        reserve_and_clawback: AddressAndSigner,
        account_with_supply: AddressAndSigner,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        assert (
            smart_asa_client.get_circulating_supply(
                asset=smart_asa.smart_asa_id
            ).return_value
            == smart_asa.total
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client,
                account_with_supply.address,
                smart_asa.smart_asa_id,
            )
            == smart_asa.total
        )
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        smart_asa_client.asset_transfer(
            xfer_asset=smart_asa.smart_asa_id,
            asset_amount=smart_asa.total,
            asset_sender=account_with_supply.address,
            asset_receiver=smart_asa_client.app_address,
            transaction_parameters=OnCompleteCallParameters(
                suggested_params=sp,
                signer=reserve_and_clawback.signer,
                sender=reserve_and_clawback.address,
            ),
        )
        assert (
            smart_asa_client.get_circulating_supply(
                asset=smart_asa.smart_asa_id
            ).return_value
            == 0
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client,
                account_with_supply.address,
                smart_asa.smart_asa_id,
            )
            == 0
        )

    @pytest.mark.parametrize("asa_config", [False], indirect=True)
    def test_fail_unauthorized(
        self, smart_asa_client: SmartAsaClient, account_with_supply: AddressAndSigner
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        with pytest.raises(LogicError, match=err.UNAUTHORIZED_RESERVE):
            smart_asa_client.asset_transfer(
                xfer_asset=smart_asa.smart_asa_id,
                asset_amount=smart_asa.total,
                asset_sender=account_with_supply.address,
                asset_receiver=smart_asa_client.app_address,
                transaction_parameters=OnCompleteCallParameters(
                    suggested_params=sp,
                    signer=account_with_supply.signer,
                    sender=account_with_supply.address,
                ),
            )

    @pytest.mark.parametrize("asa_config", [False], indirect=True)
    def test_fail_as_clawback(
        self,
        smart_asa_client: SmartAsaClient,
        clawback: AddressAndSigner,
        account_with_supply: AddressAndSigner,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        with pytest.raises(LogicError, match=err.UNAUTHORIZED_RESERVE):
            smart_asa_client.asset_transfer(
                xfer_asset=smart_asa.smart_asa_id,
                asset_amount=smart_asa.total,
                asset_sender=account_with_supply.address,
                asset_receiver=smart_asa_client.app_address,
                transaction_parameters=OnCompleteCallParameters(
                    suggested_params=sp,
                    signer=clawback.signer,
                    sender=clawback.address,
                ),
            )

    def test_fail_global_frozen(self) -> None:
        pass  # TODO: Once asset_freeze is available

    def test_fail_frozen_reserve(self) -> None:
        pass  # TODO: Once asset_freeze is available

    @pytest.mark.parametrize("asa_config", [False], indirect=True)
    def test_fail_clawback_burn(
        self,
        smart_asa_client: SmartAsaClient,
        reserve: AddressAndSigner,
        account_with_supply: AddressAndSigner,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        with pytest.raises(LogicError, match=err.CLAWBACK_BURN):
            smart_asa_client.asset_transfer(
                xfer_asset=smart_asa.smart_asa_id,
                asset_amount=smart_asa.total,
                asset_sender=account_with_supply.address,
                asset_receiver=smart_asa_client.app_address,
                transaction_parameters=OnCompleteCallParameters(
                    suggested_params=sp,
                    signer=reserve.signer,
                    sender=reserve.address,
                ),
            )


class TestClawback:
    @pytest.mark.parametrize("asa_config", [False], indirect=True)
    def test_pass(
        self,
        smart_asa_client: SmartAsaClient,
        clawback: AddressAndSigner,
        account_with_supply: AddressAndSigner,
        receiver: AddressAndSigner,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client,
                account_with_supply.address,
                smart_asa.smart_asa_id,
            )
            == smart_asa.total
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client, receiver.address, smart_asa.smart_asa_id
            )
            == 0
        )
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        smart_asa_client.asset_transfer(
            xfer_asset=smart_asa.smart_asa_id,
            asset_amount=smart_asa.total,
            asset_sender=account_with_supply.address,
            asset_receiver=receiver.address,
            transaction_parameters=OnCompleteCallParameters(
                suggested_params=sp,
                signer=clawback.signer,
                sender=clawback.address,
            ),
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client,
                account_with_supply.address,
                smart_asa.smart_asa_id,
            )
            == 0
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client, receiver.address, smart_asa.smart_asa_id
            )
            == smart_asa.total
        )

    def test_pass_global_frozen(self) -> None:
        pass  # TODO: Once asset_freeze is available

    def test_pass_frozen_accounts(self) -> None:
        pass  # TODO: Once asset_freeze is available

    @pytest.mark.parametrize("asa_config", [False], indirect=True)
    def test_fail_unauthorized(
        self,
        smart_asa_client: SmartAsaClient,
        eve: AddressAndSigner,
        account_with_supply: AddressAndSigner,
        receiver: AddressAndSigner,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        with pytest.raises(LogicError, match=err.UNAUTHORIZED_CLAWBACK):
            smart_asa_client.asset_transfer(
                xfer_asset=smart_asa.smart_asa_id,
                asset_amount=smart_asa.total,
                asset_sender=account_with_supply.address,
                asset_receiver=receiver.address,
                transaction_parameters=OnCompleteCallParameters(
                    suggested_params=sp,
                    signer=eve.signer,
                    sender=eve.address,
                ),
            )


class TestRegularTransfer:
    @pytest.mark.parametrize("asa_config", [False], indirect=True)
    def test_pass_transfer(
        self,
        smart_asa_client: SmartAsaClient,
        account_with_supply: AddressAndSigner,
        receiver: AddressAndSigner,
    ) -> None:
        smart_asa = smart_asa_client.get_global_state()
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client,
                account_with_supply.address,
                smart_asa.smart_asa_id,
            )
            == smart_asa.total
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client, receiver.address, smart_asa.smart_asa_id
            )
            == 0
        )
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        smart_asa_client.asset_transfer(
            xfer_asset=smart_asa.smart_asa_id,
            asset_amount=smart_asa.total,
            asset_sender=account_with_supply.address,
            asset_receiver=receiver.address,
            transaction_parameters=OnCompleteCallParameters(
                suggested_params=sp,
                signer=account_with_supply.signer,
                sender=account_with_supply.address,
            ),
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client,
                account_with_supply.address,
                smart_asa.smart_asa_id,
            )
            == 0
        )
        assert (
            utils.get_account_asset_balance(
                smart_asa_client.algod_client, receiver.address, smart_asa.smart_asa_id
            )
            == smart_asa.total
        )

    def test_fail_missing_ctrl_asa(self) -> None:
        pass  # TODO

    def test_fail_invalid_ctrl_asa(self) -> None:
        pass  # TODO

    def test_fail_no_opted_in_sender(self) -> None:
        pass  # TODO

    def test_fail_no_opted_in_receiver(self) -> None:
        pass  # TODO

    def test_fail_invalid_current_ctrl_asa(self) -> None:
        pass  # TODO

    @pytest.mark.parametrize("asa_config", [False], indirect=True)
    def test_fail_global_frozen(
        self,
        smart_asa_client: SmartAsaClient,
        freeze: AddressAndSigner,
        account_with_supply: AddressAndSigner,
        receiver: AddressAndSigner,
    ) -> None:
        smart_asa_client.asset_freeze(
            freeze_asset=smart_asa_client.get_global_state().smart_asa_id,
            asset_frozen=True,
            transaction_parameters=OnCompleteCallParameters(
                signer=freeze.signer,
                sender=freeze.address,
            ),
        )
        smart_asa = smart_asa_client.get_global_state()
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        with pytest.raises(LogicError, match=err.GLOBAL_FROZEN):
            smart_asa_client.asset_transfer(
                xfer_asset=smart_asa.smart_asa_id,
                asset_amount=smart_asa.total,
                asset_sender=account_with_supply.address,
                asset_receiver=receiver.address,
                transaction_parameters=OnCompleteCallParameters(
                    suggested_params=sp,
                    signer=account_with_supply.signer,
                    sender=account_with_supply.address,
                ),
            )

    @pytest.mark.parametrize("asa_config", [False], indirect=True)
    def test_fail_frozen_sender(
        self,
        smart_asa_client: SmartAsaClient,
        freeze: AddressAndSigner,
        account_with_supply: AddressAndSigner,
        receiver: AddressAndSigner,
    ) -> None:
        smart_asa_client.account_freeze(
            freeze_asset=smart_asa_client.get_global_state().smart_asa_id,
            freeze_account=account_with_supply.address,
            asset_frozen=True,
            transaction_parameters=OnCompleteCallParameters(
                signer=freeze.signer,
                sender=freeze.address,
            ),
        )
        smart_asa = smart_asa_client.get_global_state()
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        with pytest.raises(LogicError, match=err.SENDER_FROZEN):
            smart_asa_client.asset_transfer(
                xfer_asset=smart_asa.smart_asa_id,
                asset_amount=smart_asa.total,
                asset_sender=account_with_supply.address,
                asset_receiver=receiver.address,
                transaction_parameters=OnCompleteCallParameters(
                    suggested_params=sp,
                    signer=account_with_supply.signer,
                    sender=account_with_supply.address,
                ),
            )

    @pytest.mark.parametrize("asa_config", [False], indirect=True)
    def test_fail_frozen_receiver(
        self,
        smart_asa_client: SmartAsaClient,
        freeze: AddressAndSigner,
        account_with_supply: AddressAndSigner,
        receiver: AddressAndSigner,
    ) -> None:
        smart_asa_client.account_freeze(
            freeze_asset=smart_asa_client.get_global_state().smart_asa_id,
            freeze_account=receiver.address,
            asset_frozen=True,
            transaction_parameters=OnCompleteCallParameters(
                signer=freeze.signer,
                sender=freeze.address,
            ),
        )
        smart_asa = smart_asa_client.get_global_state()
        sp = smart_asa_client.algod_client.suggested_params()
        sp.flat_fee = True
        sp.fee = sp.min_fee * 2
        with pytest.raises(LogicError, match=err.RECEIVER_FROZEN):
            smart_asa_client.asset_transfer(
                xfer_asset=smart_asa.smart_asa_id,
                asset_amount=smart_asa.total,
                asset_sender=account_with_supply.address,
                asset_receiver=receiver.address,
                transaction_parameters=OnCompleteCallParameters(
                    suggested_params=sp,
                    signer=account_with_supply.signer,
                    sender=account_with_supply.address,
                ),
            )
