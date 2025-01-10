from algopy import (
    Account,
    ARC4Contract,
    Asset,
    Bytes,
    Global,
    LocalState,
    OnCompleteAction,
    StateTotals,
    String,
    TransactionType,
    Txn,
    UInt64,
    arc4,
    gtxn,
    itxn,
    op,
    subroutine,
)

from smart_contracts import errors as err
from smart_contracts.types import AssetConfig

from . import config as cfg


class SmartAsa(
    ARC4Contract,
    state_totals=StateTotals(
        global_bytes=cfg.GLOBAL_BYTES,
        global_uints=cfg.GLOBAL_UINTS,
        local_bytes=cfg.LOCAL_BYTES,
        local_uints=cfg.LOCAL_UINTS,
    ),
):
    """
    ARC-0020 (Smart ASA) - Reference Implementation
    """

    def __init__(self) -> None:
        # Preconditions
        assert Txn.global_num_byte_slice == cfg.GLOBAL_BYTES, err.WRONG_GLOBAL_BYTES
        assert Txn.global_num_uint == cfg.GLOBAL_UINTS, err.WRONG_GLOBAL_UINTS
        assert Txn.local_num_byte_slice == cfg.LOCAL_BYTES, err.WRONG_LOCAL_BYTES
        assert Txn.local_num_uint == cfg.LOCAL_UINTS, err.WRONG_LOCAL_UINTS

        # GLOBAL STATE
        # ASA Fields
        self.total = UInt64()
        self.decimals = UInt64()
        self.default_frozen = False
        self.unit_name = String()
        self.name = String()
        self.url = String()
        self.metadata_hash = Bytes()
        self.manager_addr = Account()
        self.reserve_addr = Account()
        self.freeze_addr = Account()
        self.clawback_addr = Account()
        # Smart ASA Fields
        self.smart_asa_id = UInt64()
        self.global_frozen = False

        # LOCAL STATE
        # Smart ASA Fields
        self.account_smart_asa_id = LocalState(UInt64)
        self.account_frozen = LocalState(bool)

    @subroutine
    def itoa(self, n: UInt64) -> Bytes:
        digits = Bytes(b"0123456789")
        acc = Bytes()
        while n > 0:
            acc = digits[n % 10] + acc
            n //= 10
        return acc or Bytes(b"0")

    @subroutine
    def circulating_supply(self, ctrl_asset: Asset) -> UInt64:
        return cfg.TOTAL - ctrl_asset.balance(Global.current_application_address)

    @subroutine
    def assert_common_preconditions(self, asset_id: UInt64) -> None:
        assert self.smart_asa_id, err.MISSING_CTRL_ASA
        assert self.smart_asa_id == asset_id, err.INVALID_CTRL_ASA

    @subroutine
    def assert_minting_preconditions(
        self, *, asset_receiver: Account, asset_amount: UInt64
    ) -> None:
        # Mint permission restricted to Reserve.
        assert Txn.sender == self.reserve_addr, err.UNAUTHORIZED_RESERVE
        # Forbidden self-mint (to Creator) and over-mint (> total).
        assert asset_receiver != Global.current_application_address, err.SELF_MINT
        assert (
            asset_amount + self.circulating_supply(Asset(self.smart_asa_id))
            <= self.total
        ), err.OVER_MINT
        # In the case of Controlled ASA destroyed and re-created, the Smart ADA ID in Local State could be outdated.
        assert (
            self.account_smart_asa_id[asset_receiver] == self.smart_asa_id
        ), err.INVALID_CTRL_ASA
        assert not self.global_frozen, err.GLOBAL_FROZEN
        if self.reserve_addr != self.clawback_addr:
            assert not self.account_frozen[asset_receiver], err.RECEIVER_FROZEN

    @subroutine
    def assert_burning_preconditions(self, *, asset_sender: Account) -> None:
        # Burn permission restricted to Reserve.
        assert Txn.sender == self.reserve_addr, err.UNAUTHORIZED_RESERVE
        # In case of Controlled ASA destroyed and re-created the Smart ADA ID in Local State could be outdated.
        assert (
            self.account_smart_asa_id[asset_sender] == self.smart_asa_id
        ), err.INVALID_CTRL_ASA
        assert not self.global_frozen, err.GLOBAL_FROZEN
        if self.reserve_addr != self.clawback_addr:
            assert not self.account_frozen[asset_sender], err.SENDER_FROZEN
            # Forbidden clawback through burning (burned amount not from Reserve).
            assert asset_sender == self.reserve_addr, err.CLAWBACK_BURN

    @subroutine
    def assert_clawback_preconditions(
        self, *, asset_sender: Account, asset_receiver: Account
    ) -> None:
        # In the case of Controlled ASA destroyed and re-created, the Smart ADA ID in Local State could be outdated.
        assert (
            self.account_smart_asa_id[asset_sender] == self.smart_asa_id
        ), err.INVALID_CTRL_ASA
        assert (
            self.account_smart_asa_id[asset_receiver] == self.smart_asa_id
        ), err.INVALID_CTRL_ASA

    @subroutine
    def assert_regular_transfer_preconditions(
        self, *, asset_sender: Account, asset_receiver: Account
    ) -> None:
        assert Txn.sender == asset_sender, err.UNAUTHORIZED_CLAWBACK
        # In the case of Controlled ASA destroyed and re-created, the Smart ADA ID in Local State could be outdated.
        assert (
            self.account_smart_asa_id[asset_sender] == self.smart_asa_id
        ), err.INVALID_CTRL_ASA
        assert (
            self.account_smart_asa_id[asset_receiver] == self.smart_asa_id
        ), err.INVALID_CTRL_ASA
        assert not self.global_frozen, err.GLOBAL_FROZEN
        assert not self.account_frozen[asset_sender], err.SENDER_FROZEN
        assert not self.account_frozen[asset_receiver], err.RECEIVER_FROZEN

    @subroutine
    def assert_close_out_preconditions(self, close_asset: Asset) -> None:
        asa_close_out_relative_idx = Txn.group_index + 1
        asa_close_out_txn = gtxn.AssetTransferTransaction(asa_close_out_relative_idx)
        assert Txn.on_completion == OnCompleteAction.CloseOut, err.WRONG_ON_COMPLETE
        assert (
            self.account_smart_asa_id[Txn.sender] == close_asset.id
        ), err.INVALID_CTRL_ASA
        assert (
            Global.group_size > asa_close_out_relative_idx
        ), err.INVALID_CLOSE_OUT_GROUP_SIZE
        assert (
            asa_close_out_txn.type == TransactionType.AssetTransfer
        ), err.CLOSE_OUT_WRONG_TYPE  # TODO: Redundant?
        assert (
            asa_close_out_txn.xfer_asset.id == close_asset.id
        ), err.CLOSE_OUT_WRONG_ASA
        assert asa_close_out_txn.sender == Txn.sender, err.CLOSE_OUT_WRONG_SENDER
        assert asa_close_out_txn.asset_amount == UInt64(0), err.CLOSE_OUT_WRONG_AMOUNT
        assert (
            asa_close_out_txn.asset_close_to != Global.zero_address
        ), err.CLOSE_OUT_WRONG_CLOSE_TO

    @subroutine
    def assert_close_out_not_destroyed_preconditions(
        self, close_asset: Asset, asset_creator: Account
    ) -> None:
        asa_close_out_relative_idx = Txn.group_index + 1
        asa_close_out_txn = gtxn.AssetTransferTransaction(asa_close_out_relative_idx)
        assert (
            asa_close_out_txn.asset_close_to == asset_creator
        ), err.CLOSE_OUT_WRONG_CLOSE_TO
        assert close_asset.id == self.smart_asa_id, err.INVALID_CTRL_ASA

    @subroutine
    def assert_close_out_not_to_creator(self, close_to: Account) -> None:
        assert not self.global_frozen, err.GLOBAL_FROZEN
        assert not self.account_frozen[Txn.sender], err.SENDER_FROZEN
        assert not self.account_frozen[close_to], err.CLOSE_TO_FROZEN

    @subroutine
    def inner_asset_config(self) -> UInt64:
        return (
            itxn.AssetConfig(
                fee=0,
                total=cfg.TOTAL,
                decimals=cfg.DECIMALS,
                default_frozen=cfg.DEFAULT_FROZEN,
                unit_name=cfg.UNIT_NAME,
                asset_name=cfg.NAME,
                url=cfg.APP_BINDING + self.itoa(Global.current_application_id.id),
                manager=Global.current_application_address,
                reserve=Global.current_application_address,
                freeze=Global.current_application_address,
                clawback=Global.current_application_address,
            )
            .submit()
            .created_asset.id
        )

    @subroutine
    def inner_asset_transfer(
        self,
        *,
        xfer_asset: Asset,
        asset_amount: UInt64,
        asset_sender: Account,
        asset_receiver: Account
    ) -> None:
        itxn.AssetTransfer(
            fee=0,
            xfer_asset=xfer_asset.id,
            asset_amount=asset_amount,
            asset_sender=asset_sender,
            asset_receiver=asset_receiver,
            sender=Global.current_application_address,
        ).submit()

    @subroutine
    def inner_asset_destroy(self, *, destroy_asset: Asset) -> None:
        itxn.AssetConfig(
            fee=0,
            config_asset=destroy_asset,
            sender=Global.current_application_address,
        ).submit()

    @arc4.abimethod
    def asset_create(
        self,
        total: arc4.UInt64,
        decimals: arc4.UInt32,
        default_frozen: arc4.Bool,
        unit_name: arc4.String,
        name: arc4.String,
        url: arc4.String,
        metadata_hash: arc4.DynamicBytes,
        manager_addr: arc4.Address,
        reserve_addr: arc4.Address,
        freeze_addr: arc4.Address,
        clawback_addr: arc4.Address,
    ) -> arc4.UInt64:
        """
        Create the Controlled ASA

        Args:
            total: The total number of base units of the Smart ASA to create
            decimals: The number of digits to use after the decimal point when displaying the Smart ASA
            default_frozen: Smart ASA default frozen (True to freeze holdings by default)
            unit_name: The name of a unit of Smart ASA
            name: The name of the Smart ASA
            url: Smart ASA external URL
            metadata_hash: Smart ASA metadata hash
            manager_addr: Account that can manage the configuration of the Smart ASA and destroy it
            reserve_addr: Account that holds the reserve (non-minted) units of Smart ASA and can mint or burn it
            freeze_addr: Account that can freeze/unfreeze holdings of the Smart ASA globally or locally
            clawback_addr: Account that can clawback holdings of the Smart ASA

        Returns:
            Controlled ASA ID
        """
        # Preconditions
        assert Txn.sender == Global.creator_address, err.UNAUTHORIZED
        assert not self.smart_asa_id, err.EXISTING_CTRL_ASA

        # Effects
        self.smart_asa_id = self.inner_asset_config()
        self.total = total.native
        self.decimals = decimals.native
        self.default_frozen = default_frozen.native
        self.unit_name = unit_name.native
        self.name = name.native
        self.url = url.native
        self.metadata_hash = metadata_hash.native
        self.manager_addr = manager_addr.native
        self.reserve_addr = reserve_addr.native
        self.freeze_addr = freeze_addr.native
        self.clawback_addr = clawback_addr.native
        return arc4.UInt64(self.smart_asa_id)

    @arc4.abimethod(allow_actions=["OptIn"])
    def asset_opt_in(
        self, asset: Asset, ctrl_asa_opt_in: gtxn.AssetTransferTransaction
    ) -> None:
        """
        Smart ASA opt in (App and Controlled ASA)

        Args:
            asset: Smart ASA ID
            ctrl_asa_opt_in: Controlled ASA opt in transaction
        """
        # Preconditions
        self.assert_common_preconditions(asset.id)
        assert (
            ctrl_asa_opt_in.type == TransactionType.AssetTransfer
        ), err.OPT_IN_WRONG_TYPE  # Pedant
        assert ctrl_asa_opt_in.xfer_asset.id == self.smart_asa_id, err.OPT_IN_WRONG_ASA
        assert ctrl_asa_opt_in.sender == Txn.sender, err.OPT_IN_WRONG_SENDER
        assert ctrl_asa_opt_in.asset_receiver == Txn.sender, err.OPT_IN_WRONG_RECEIVER
        assert (
            ctrl_asa_opt_in.asset_amount == 0
        ), err.OPT_IN_WRONG_AMOUNT  # Pedant: Controlled ASA is default frozen
        assert (
            ctrl_asa_opt_in.asset_close_to == Global.zero_address
        ), err.OPT_IN_WRONG_CLOSE_TO
        assert Txn.on_completion == OnCompleteAction.OptIn, err.WRONG_ON_COMPLETE
        assert Txn.sender.is_opted_in(
            asset
        ), err.MISSING_CTRL_ASA  # Pedant: ctrl_asa_opt_in is checked properly

        # Local State Init
        self.account_smart_asa_id[Txn.sender] = self.smart_asa_id
        self.account_frozen[Txn.sender] = False

        # Effects
        if (
            self.default_frozen or asset.balance(Txn.sender) > 0
        ):  # Prevent close-out circumventing account frozen state
            self.account_frozen[Txn.sender] = True

    @arc4.abimethod
    def asset_config(
        self,
        config_asset: Asset,
        total: arc4.UInt64,
        decimals: arc4.UInt32,
        default_frozen: arc4.Bool,
        unit_name: arc4.String,
        name: arc4.String,
        url: arc4.String,
        metadata_hash: arc4.DynamicBytes,
        manager_addr: arc4.Address,
        reserve_addr: arc4.Address,
        freeze_addr: arc4.Address,
        clawback_addr: arc4.Address,
    ) -> None:
        """
        Configure Smart ASA (for unchanged parameters use existing value - no optional args on AVM)

        Args:
            config_asset: Smart ASA ID to configure
            total: Total number of base units if the Smart ASA. It can not be less than current circulating supply
            decimals: The number of digits to use after the decimal point when displaying the Smart ASA
            default_frozen: Smart ASA default frozen (True to freeze holdings by default)
            unit_name: The name of a unit of Smart ASA
            name: The name of the Smart ASA
            url: Smart ASA external URL
            metadata_hash: Smart ASA metadata hash
            manager_addr: Account that can manage the configuration of the Smart ASA and destroy it
            reserve_addr: Account that holds the reserve (non-minted) units of Smart ASA and can mint or burn it
            freeze_addr: Account that can freeze/unfreeze holdings of the Smart ASA globally or locally
            clawback_addr: Account that can clawback holdings of the Smart ASA
        """
        # Preconditions
        self.assert_common_preconditions(config_asset.id)
        assert Txn.sender == self.manager_addr, err.UNAUTHORIZED_MANAGER
        if reserve_addr != self.reserve_addr:
            assert self.reserve_addr != Global.zero_address, err.DISABLED_RESERVE
        if freeze_addr != self.freeze_addr:
            assert self.freeze_addr != Global.zero_address, err.DISABLED_FREEZE
        if clawback_addr != self.clawback_addr:
            assert self.clawback_addr != Global.zero_address, err.DISABLED_CLAWBACK
        assert total >= self.circulating_supply(config_asset), err.INVALID_TOTAL

        # Effects
        self.total = total.native
        self.decimals = decimals.native
        self.default_frozen = default_frozen.native
        self.unit_name = unit_name.native
        self.name = name.native
        self.url = url.native
        self.metadata_hash = metadata_hash.native
        self.manager_addr = manager_addr.native
        self.reserve_addr = reserve_addr.native
        self.freeze_addr = freeze_addr.native
        self.clawback_addr = clawback_addr.native

    @arc4.abimethod
    def asset_transfer(
        self,
        xfer_asset: Asset,
        asset_amount: arc4.UInt64,
        asset_sender: Account,
        asset_receiver: Account,
    ) -> None:
        """
        Smart ASA transfers: regular, clawback, mint, burn

        Args:
            xfer_asset: Smart ASA ID to transfer
            asset_amount: Amount to transfer
            asset_sender: Smart ASA sender
            asset_receiver: Smart ASA receiver
        """
        # Preconditions
        self.assert_common_preconditions(xfer_asset.id)
        if asset_sender == Global.current_application_address:
            self.assert_minting_preconditions(
                asset_receiver=asset_receiver, asset_amount=asset_amount.native
            )
        elif asset_receiver == Global.current_application_address:
            self.assert_burning_preconditions(asset_sender=asset_sender)
        elif Txn.sender == self.clawback_addr:
            self.assert_clawback_preconditions(
                asset_sender=asset_sender, asset_receiver=asset_receiver
            )
        else:
            self.assert_regular_transfer_preconditions(
                asset_sender=asset_sender, asset_receiver=asset_receiver
            )

        # Effects
        self.inner_asset_transfer(
            xfer_asset=xfer_asset,
            asset_amount=asset_amount.native,
            asset_sender=asset_sender,
            asset_receiver=asset_receiver,
        )

    @arc4.abimethod
    def asset_freeze(self, freeze_asset: Asset, asset_frozen: arc4.Bool) -> None:
        """
        Smart ASA global freeze (all accounts)

        Args:
            freeze_asset: Smart ASA ID to globally freeze/unfreeze
            asset_frozen: Smart ASA frozen status
        """
        # Preconditions
        self.assert_common_preconditions(freeze_asset.id)
        assert Txn.sender == self.freeze_addr, err.UNAUTHORIZED_FREEZE

        # Effects
        self.global_frozen = asset_frozen.native

    @arc4.abimethod
    def account_freeze(
        self, freeze_asset: Asset, freeze_account: Account, asset_frozen: arc4.Bool
    ) -> None:
        """
        Smart ASA local freeze (account specific)

        Args:
            freeze_asset: Smart ASA ID to locally freeze/unfreeze
            freeze_account: Account to freeze/unfreeze
            asset_frozen: Smart ASA frozen status
        """
        # Preconditions
        self.assert_common_preconditions(freeze_asset.id)
        assert (
            self.account_smart_asa_id[freeze_account] == self.smart_asa_id
        ), err.INVALID_CTRL_ASA
        assert Txn.sender == self.freeze_addr, err.UNAUTHORIZED_FREEZE

        # Effects
        self.account_frozen[freeze_account] = asset_frozen.native

    @arc4.abimethod(allow_actions=["CloseOut"])
    def asset_close_out(self, close_asset: Asset, close_to: Account) -> None:
        """
        Smart ASA close out (App and Controlled ASA)

        Args:
            close_asset: Smart ASA ID to close out
            close_to: Account to send all the Smart ASA remainder to.
        """
        # Preconditions
        self.assert_close_out_preconditions(close_asset)
        (creator, exists) = op.AssetParamsGet.asset_creator(close_asset.id)
        if exists:  # Smart ASA has not been destroyed
            self.assert_close_out_not_destroyed_preconditions(close_asset, creator)
            if (
                close_to != creator
            ):  # If close-out target is not the Creator, then close-out target MUST be opted-in
                assert (
                    self.account_smart_asa_id[close_to] == self.smart_asa_id
                ), err.INVALID_CTRL_ASA
                self.assert_close_out_not_to_creator(close_to)

            # Effects
            self.inner_asset_transfer(
                xfer_asset=close_asset,
                asset_amount=close_asset.balance(Txn.sender),
                asset_sender=Txn.sender,
                asset_receiver=close_to,
            )

    @arc4.abimethod
    def asset_destroy(self, destroy_asset: Asset) -> None:
        """
        Destroy the Controlled ASA

        Args:
            destroy_asset: Smart ASA ID to destroy
        """
        # Preconditions
        self.assert_common_preconditions(destroy_asset.id)
        assert Txn.sender == self.manager_addr, err.UNAUTHORIZED_MANAGER

        # Effects
        self.inner_asset_destroy(destroy_asset=destroy_asset)
        self.total = UInt64()
        self.decimals = UInt64()
        self.default_frozen = False
        self.unit_name = String()
        self.name = String()
        self.url = String()
        self.metadata_hash = Bytes()
        self.manager_addr = Account()
        self.reserve_addr = Account()
        self.freeze_addr = Account()
        self.clawback_addr = Account()
        self.smart_asa_id = UInt64()
        self.global_frozen = False

    @arc4.abimethod(readonly=True)
    def get_asset_config(self, asset: Asset) -> AssetConfig:
        """
        Get Smart ASA configuration

        Args:
            asset: Smart ASA ID

        Returns:
            Smart ASA configuration parameters
        """
        # Preconditions
        self.assert_common_preconditions(asset.id)

        # Effects
        return AssetConfig(
            total=arc4.UInt64(self.total),
            decimals=arc4.UInt32(self.decimals),
            default_frozen=arc4.Bool(self.default_frozen),
            unit_name=arc4.String(self.unit_name),
            name=arc4.String(self.name),
            url=arc4.String(self.url),
            metadata_hash=arc4.DynamicBytes(self.metadata_hash),
            manager_addr=arc4.Address(self.manager_addr),
            reserve_addr=arc4.Address(self.reserve_addr),
            freeze_addr=arc4.Address(self.freeze_addr),
            clawback_addr=arc4.Address(self.clawback_addr),
        )

    @arc4.abimethod(readonly=True)
    def get_asset_is_frozen(self, freeze_asset: Asset) -> arc4.Bool:
        """
        Get Smart ASA global frozen status

        Args:
            freeze_asset: Smart ASA ID

        Returns:
            Smart ASA global frozen status
        """
        # Preconditions
        self.assert_common_preconditions(freeze_asset.id)

        # Effects
        return arc4.Bool(self.global_frozen)

    @arc4.abimethod(readonly=True)
    def get_account_is_frozen(
        self, freeze_asset: Asset, freeze_account: Account
    ) -> arc4.Bool:
        """
        Get Smart ASA account frozen status

        Args:
            freeze_asset: Smart ASA ID
            freeze_account: Account to check

        Returns:
            Smart ASA account frozen status
        """
        # Preconditions
        self.assert_common_preconditions(freeze_asset.id)

        # Effects
        return arc4.Bool(self.account_frozen[freeze_account])

    @arc4.abimethod(readonly=True)
    def get_circulating_supply(self, asset: Asset) -> arc4.UInt64:
        """
        Get Smart ASA circulating supply

        Args:
            asset: Smart ASA ID

        Returns:
            Smart ASA circulating supply
        """
        # Preconditions
        self.assert_common_preconditions(asset.id)

        # Effects
        return arc4.UInt64(self.circulating_supply(asset))
