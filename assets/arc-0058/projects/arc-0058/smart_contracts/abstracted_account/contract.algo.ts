import { Contract, GlobalState, BoxMap, assert, uint64, Account, TransactionType, Application, abimethod, gtxn, itxn, OnCompleteAction, Bytes, bytes, assertMatch, clone } from '@algorandfoundation/algorand-typescript'
import { abiCall, Address, methodSelector, Uint8 } from '@algorandfoundation/algorand-typescript/arc4';
import { btoi, Global, len, Txn } from '@algorandfoundation/algorand-typescript/op'
import { ERR_ADMIN_ONLY, ERR_ADMIN_PLUGINS_CANNOT_USE_ESCROWS, ERR_ALLOWANCE_ALREADY_EXISTS, ERR_ALLOWANCE_DOES_NOT_EXIST, ERR_ALLOWANCE_EXCEEDED, ERR_CANNOT_CALL_OTHER_APPS_DURING_REKEY, ERR_ESCROW_ALREADY_EXISTS, ERR_ESCROW_DOES_NOT_EXIST, ERR_ESCROW_LOCKED, ERR_ESCROW_NAME_REQUIRED, ERR_ESCROW_REQUIRED_TO_BE_SET_AS_DEFAULT, ERR_EXECUTION_EXPIRED, ERR_EXECUTION_KEY_DOES_NOT_EXIST, ERR_EXECUTION_KEY_NOT_FOUND, ERR_EXECUTION_KEY_UPDATE_MUST_MATCH_FIRST_VALID, ERR_EXECUTION_KEY_UPDATE_MUST_MATCH_LAST_VALID, ERR_EXECUTION_NOT_READY, ERR_GROUP_NOT_FOUND, ERR_INVALID_METHOD_SIGNATURE_LENGTH, ERR_INVALID_ONCOMPLETE, ERR_INVALID_SENDER_ARG, ERR_INVALID_SENDER_VALUE, ERR_MALFORMED_OFFSETS, ERR_METHOD_ON_COOLDOWN, ERR_MISSING_REKEY_BACK, ERR_ONLY_ADMIN_CAN_CHANGE_ADMIN, ERR_PLUGIN_DOES_NOT_EXIST, ERR_PLUGIN_EXPIRED, ERR_PLUGIN_ON_COOLDOWN, ERR_SENDER_MUST_BE_ADMIN_OR_CONTROLLED_ADDRESS, ERR_SENDER_MUST_BE_ADMIN_PLUGIN, ERR_USING_EXECUTION_KEY_REQUIRES_GLOBAL, ERR_ZERO_ADDRESS_DELEGATION_TYPE } from './errors';
import { AbstractAccountBoxMBRData, AddAllowanceInfo, AllowanceInfo, AllowanceKey, DelegationTypeSelf, EscrowInfo, EscrowReclaim, ExecutionInfo, FundsRequest, MethodInfo, MethodRestriction, MethodValidation, PluginInfo, PluginKey, PluginValidation, SpendAllowanceTypeDrip, SpendAllowanceTypeFlat, SpendAllowanceTypeWindow } from './types';
import { EscrowFactory } from '../escrow/factory.algo';
import { AbstractAccountBoxPrefixAllowances, AbstractAccountBoxPrefixEscrows, AbstractAccountBoxPrefixExecutions, AbstractAccountBoxPrefixNamedPlugins, AbstractAccountBoxPrefixPlugins, AbstractAccountGlobalStateKeysAdmin, AbstractAccountGlobalStateKeysControlledAddress, AbstractAccountGlobalStateKeysCurrentPlugin, AbstractAccountGlobalStateKeysEscrowFactory, AbstractAccountGlobalStateKeysLastChange, AbstractAccountGlobalStateKeysLastUserInteraction, AbstractAccountGlobalStateKeysRekeyIndex, AbstractAccountGlobalStateKeysSpendingAddress, BoxCostPerByte, MethodRestrictionByteLength, MinAllowanceMBR, MinEscrowsMBR, MinExecutionsMBR, MinNamedPluginMBR, MinPluginMBR } from './constants';
import { ERR_INVALID_PAYMENT } from '../utils/errors';
import { ERR_FORBIDDEN } from '../escrow/errors';
import { ARC58WalletIDsByAccountsMbr, NewCostForARC58 } from '../escrow/constants';
import { emptyAllowanceInfo, emptyEscrowInfo, emptyExecutionInfo, emptyPluginInfo } from './utils';

export class AbstractedAccount extends Contract {

  /** The admin of the abstracted account. This address can add plugins and initiate rekeys */
  admin = GlobalState<Account>({ key: AbstractAccountGlobalStateKeysAdmin })
  /** The address this app controls */
  controlledAddress = GlobalState<Account>({ key: AbstractAccountGlobalStateKeysControlledAddress });
  /** The last time the contract was interacted with in unix time */
  lastUserInteraction = GlobalState<uint64>({ key: AbstractAccountGlobalStateKeysLastUserInteraction })
  /** The last time state has changed on the abstracted account (not including lastCalled for cooldowns) in unix time */
  lastChange = GlobalState<uint64>({ key: AbstractAccountGlobalStateKeysLastChange })
  /** the escrow account factory to use for allowances */
  escrowFactory = GlobalState<Application>({ key: AbstractAccountGlobalStateKeysEscrowFactory })
  /** [TEMPORARY STATE FIELD] The spending address for the currently active plugin */
  spendingAddress = GlobalState<Account>({ key: AbstractAccountGlobalStateKeysSpendingAddress })
  /** [TEMPORARY STATE FIELD] The current plugin key being used */
  currentPlugin = GlobalState<PluginKey>({ key: AbstractAccountGlobalStateKeysCurrentPlugin })
  /** [TEMPORARY STATE FIELD] The index of the transaction that created the rekey sandwich */
  rekeyIndex = GlobalState<uint64>({ initialValue: 0, key: AbstractAccountGlobalStateKeysRekeyIndex })

  /** Plugins that add functionality to the controlledAddress and the account that has permission to use it. */
  plugins = BoxMap<PluginKey, PluginInfo>({ keyPrefix: AbstractAccountBoxPrefixPlugins });
  /** Plugins that have been given a name for discoverability */
  namedPlugins = BoxMap<string, PluginKey>({ keyPrefix: AbstractAccountBoxPrefixNamedPlugins });
  /** the escrows that this wallet has created for specific callers with allowances */
  escrows = BoxMap<string, EscrowInfo>({ keyPrefix: AbstractAccountBoxPrefixEscrows })
  /** The Allowances for plugins installed on the smart contract with useAllowance set to true */
  allowances = BoxMap<AllowanceKey, AllowanceInfo>({ keyPrefix: AbstractAccountBoxPrefixAllowances }) // 38_500
  /** execution keys */
  executions = BoxMap<bytes<32>, ExecutionInfo>({ keyPrefix: AbstractAccountBoxPrefixExecutions })

  private updateLastUserInteraction() {
    this.lastUserInteraction.value = Global.latestTimestamp
  }

  private updateLastChange() {
    this.lastChange.value = Global.latestTimestamp
  }

  private pluginsMbr(escrow: string, methodCount: uint64): uint64 {
    return MinPluginMBR + (
      BoxCostPerByte * ((MethodRestrictionByteLength * methodCount) + Bytes(escrow).length)
    );
  }

  private namedPluginsMbr(name: string): uint64 {
    return MinNamedPluginMBR + (BoxCostPerByte * Bytes(name).length);
  }

  private escrowsMbr(escrow: string): uint64 {
    return MinEscrowsMBR + (BoxCostPerByte * Bytes(escrow).length);
  }

  private allowancesMbr(escrow: string): uint64 {
    return MinAllowanceMBR + (BoxCostPerByte * Bytes(escrow).length);
  }

  private executionsMbr(groups: uint64): uint64 {
    return MinExecutionsMBR + (BoxCostPerByte * (groups * 32));
  }

  private maybeNewEscrow(escrow: string): uint64 {
    if (escrow === '') {
      return 0;
    }

    return this.escrows(escrow).exists
      ? this.escrows(escrow).value.id
      : this.newEscrow(escrow);
  }

  private newEscrow(escrow: string): uint64 {
    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          sender: this.controlledAddress.value,
          receiver: Global.currentApplicationAddress,
          amount: this.escrowsMbr(escrow)
        })
        .submit()
    }

    const id = abiCall<typeof EscrowFactory.prototype.new>({
      sender: this.controlledAddress.value,
      appId: this.escrowFactory.value,
      args: [
        itxn.payment({
          sender: this.controlledAddress.value,
          amount: NewCostForARC58 + Global.minBalance,
          receiver: this.escrowFactory.value.address
        }),
      ]
    }).returnValue

    this.escrows(escrow).value = { id, locked: false }

    return id;
  }

  private pluginCallAllowed(plugin: uint64, caller: Account, escrow: string, method: bytes<4>): boolean {
    const key: PluginKey = { plugin, caller, escrow }

    if (!this.plugins(key).exists) {
      return false;
    }

    const { methods, useRounds, lastCalled, cooldown, useExecutionKey } = this.plugins(key).value as Readonly<PluginInfo>

    if (useExecutionKey) {
      return false
    }

    let methodAllowed = methods.length > 0 ? false : true;
    for (let i: uint64 = 0; i < methods.length; i += 1) {
      if (methods[i].selector === method) {
        methodAllowed = true;
        break;
      }
    }

    const epochRef = useRounds ? Global.round : Global.latestTimestamp;

    return (
      lastCalled >= epochRef &&
      (epochRef - lastCalled) >= cooldown &&
      methodAllowed
    )
  }

  private txnRekeysBack(txn: gtxn.Transaction): boolean {
    // this check is for manual rekeyTo calls, it only ever uses the controlled address so its okay to hardcode it here
    if (
      txn.sender === this.controlledAddress.value &&
      txn.rekeyTo === Global.currentApplicationAddress
    ) {
      return true;
    }

    return (
      txn.type === TransactionType.ApplicationCall
      && txn.appId === Global.currentApplicationId
      && txn.numAppArgs === 1
      && txn.onCompletion === OnCompleteAction.NoOp
      && txn.appArgs(0) === methodSelector('arc58_verifyAuthAddress()void')
    )
  }

  private assertRekeysBack(): void {
    let rekeysBack = false;
    for (let i: uint64 = (Txn.groupIndex + 1); i < Global.groupSize; i += 1) {
      const txn = gtxn.Transaction(i)

      if (this.txnRekeysBack(txn)) {
        rekeysBack = true;
        break;
      }
    }

    assert(rekeysBack, ERR_MISSING_REKEY_BACK);
  }

  private pluginCheck(key: PluginKey): PluginValidation {

    const exists = this.plugins(key).exists;
    if (!exists) {
      return {
        exists: false,
        expired: true,
        onCooldown: true,
        hasMethodRestrictions: false,
      }
    }

    const { useRounds, lastValid, cooldown, lastCalled, methods } = this.plugins(key).value as Readonly<PluginInfo>
    const epochRef = useRounds ? Global.round : Global.latestTimestamp;

    return {
      exists,
      expired: epochRef > lastValid,
      onCooldown: (epochRef - lastCalled) < cooldown,
      hasMethodRestrictions: methods.length > 0,
    }
  }

  /**
   * Guarantee that our txn group is valid in a single loop over all txns in the group
   * 
   * @param key the box key for the plugin were checking
   * @param methodOffsets the indices of the methods being used in the group
  */
  private assertValidGroup(key: PluginKey, methodOffsets: uint64[]): void {

    const { useRounds, useExecutionKey } = this.plugins(key).value

    if (useExecutionKey && !(Txn.sender === this.admin.value)) {
      assert(this.executions(Txn.lease).exists, ERR_EXECUTION_KEY_NOT_FOUND);
      assert(this.executions(Txn.lease).value.firstValid <= Global.round, ERR_EXECUTION_NOT_READY);
      assert(this.executions(Txn.lease).value.lastValid >= Global.round, ERR_EXECUTION_EXPIRED);

      const groups = this.executions(Txn.lease).value.groups as Readonly<bytes<32>[]>;

      let foundGroup = false;
      for (let i: uint64 = 0; i < groups.length; i += 1) {
        if (groups[i] === Global.groupId) {
          foundGroup = true;
        }
      }

      assert(foundGroup, ERR_GROUP_NOT_FOUND);
      this.executions(Txn.lease).delete();
    }

    const initialCheck = this.pluginCheck(key);

    assert(initialCheck.exists, ERR_PLUGIN_DOES_NOT_EXIST);
    assert(!initialCheck.expired, ERR_PLUGIN_EXPIRED);
    assert(!initialCheck.onCooldown, ERR_PLUGIN_ON_COOLDOWN);

    const epochRef = useRounds
      ? Global.round
      : Global.latestTimestamp;

    let rekeysBack = false;
    let methodIndex: uint64 = 0;

    for (let i: uint64 = (Txn.groupIndex + 1); i < Global.groupSize; i += 1) {
      const txn = gtxn.Transaction(i)

      if (this.txnRekeysBack(txn)) {
        rekeysBack = true;
        break;
      }

      if (txn.type !== TransactionType.ApplicationCall) {
        continue;
      }

      assert(txn.appId.id === key.plugin, ERR_CANNOT_CALL_OTHER_APPS_DURING_REKEY);
      assert(txn.onCompletion === OnCompleteAction.NoOp, ERR_INVALID_ONCOMPLETE);
      // ensure the first arg to a method call is the app id itself
      // index 1 is used because arg[0] is the method selector
      assert(txn.numAppArgs > 1, ERR_INVALID_SENDER_ARG);
      assert(Application(btoi(txn.appArgs(1))) === Global.currentApplicationId, ERR_INVALID_SENDER_VALUE);

      const { expired, onCooldown, hasMethodRestrictions } = this.pluginCheck(key);

      assert(!expired, ERR_PLUGIN_EXPIRED);
      assert(!onCooldown, ERR_PLUGIN_ON_COOLDOWN);

      if (hasMethodRestrictions) {
        assert(methodIndex < methodOffsets.length, ERR_MALFORMED_OFFSETS);
        const { methodAllowed, methodOnCooldown } = this.methodCheck(key, txn, methodOffsets[methodIndex]);
        assert(methodAllowed && !methodOnCooldown, ERR_METHOD_ON_COOLDOWN);
      }

      this.plugins(key).value.lastCalled = epochRef
      methodIndex += 1;
    }

    assert(rekeysBack, ERR_MISSING_REKEY_BACK);
  }

  /**
   * Checks if the method call is allowed
   * 
   * @param key the box key for the plugin were checking
   * @param caller the address that triggered the plugin or global address
   * @param offset the index of the method being used
   * @returns whether the method call is allowed
  */
  private methodCheck(key: PluginKey, txn: gtxn.ApplicationCallTxn, offset: uint64): MethodValidation {

    assert(len(txn.appArgs(0)) === 4, ERR_INVALID_METHOD_SIGNATURE_LENGTH)
    const selectorArg = txn.appArgs(0).toFixed({ length: 4 })

    const { useRounds } = this.plugins(key).value
    const { selector, cooldown, lastCalled } = this.plugins(key).value.methods[offset]

    const hasCooldown = cooldown > 0

    const epochRef = useRounds ? Global.round : Global.latestTimestamp
    const methodOnCooldown = (epochRef - lastCalled) < cooldown

    if (selector === selectorArg && (!hasCooldown || !methodOnCooldown)) {
      // update the last called round for the method
      if (hasCooldown) {
        const lastCalled = useRounds ? Global.round : Global.latestTimestamp;
        this.plugins(key).value.methods[offset].lastCalled = lastCalled
      }

      return {
        methodAllowed: true,
        methodOnCooldown
      }
    }

    return {
      methodAllowed: false,
      methodOnCooldown: true
    }
  }

  private transferFunds(escrow: string, fundsRequests: FundsRequest[]): void {
    const escrowID = this.escrows(escrow).value.id;
    const escrowAddress = Application(escrowID).address;

    for (let i: uint64 = 0; i < fundsRequests.length; i += 1) {

      const allowanceKey: AllowanceKey = {
        escrow,
        asset: fundsRequests[i].asset
      }

      this.verifyAllowance(allowanceKey, fundsRequests[i]);

      if (fundsRequests[i].asset !== 0) {
        itxn
          .assetTransfer({
            sender: this.controlledAddress.value,
            assetReceiver: escrowAddress,
            assetAmount: fundsRequests[i].amount,
            xferAsset: fundsRequests[i].asset
          })
          .submit();
      } else {
        itxn
          .payment({
            sender: this.controlledAddress.value,
            receiver: escrowAddress,
            amount: fundsRequests[i].amount
          })
          .submit();
      }
    }
  }

  private verifyAllowance(key: AllowanceKey, fundRequest: FundsRequest): void {
    assert(this.allowances(key).exists, ERR_ALLOWANCE_DOES_NOT_EXIST);
    const { type, spent, amount, last, max, interval, start, useRounds } = this.allowances(key).value
    const newLast = useRounds ? Global.round : Global.latestTimestamp;

    if (type === SpendAllowanceTypeFlat) {
      const leftover: uint64 = amount - spent;
      assert(leftover >= fundRequest.amount, ERR_ALLOWANCE_EXCEEDED);
      this.allowances(key).value.spent += fundRequest.amount
    } else if (type === SpendAllowanceTypeWindow) {
      const currentWindowStart = this.getLatestWindowStart(useRounds, start, interval)

      if (currentWindowStart > last) {
        assert(amount >= fundRequest.amount, ERR_ALLOWANCE_EXCEEDED);
        this.allowances(key).value.spent = fundRequest.amount
      } else {
        // calc the remaining amount available in the current window
        const leftover: uint64 = amount - spent;
        assert(leftover >= fundRequest.amount, ERR_ALLOWANCE_EXCEEDED);
        this.allowances(key).value.spent += fundRequest.amount
      }
    } else if (type === SpendAllowanceTypeDrip) {
      const epochRef = useRounds ? Global.round : Global.latestTimestamp;
      const passed: uint64 = epochRef - last
      // in this context:
      // amount represents our accrual rate
      // spent represents the last leftover amount available
      const accrued: uint64 = spent + ((passed / interval) * amount)
      const available: uint64 = accrued > max ? max : accrued

      assert(available >= fundRequest.amount, ERR_ALLOWANCE_EXCEEDED);
      this.allowances(key).value.spent = (available - fundRequest.amount)
    }
    this.allowances(key).value.last = newLast
  }

  private getLatestWindowStart(useRounds: boolean, start: uint64, interval: uint64): uint64 {
    if (useRounds) {
      return Global.round - ((Global.round - start) % interval)
    }
    return Global.latestTimestamp - ((Global.latestTimestamp - start) % interval)
  }

  /**
   * What the value of this.address.value.authAddr should be when this.controlledAddress
   * is able to be controlled by this app. It will either be this.app.address or zeroAddress
  */
  private getAuthAddress(): Account {
    return (
      this.spendingAddress.value === this.controlledAddress.value
      && this.controlledAddress.value === Global.currentApplicationAddress
    ) ? Global.zeroAddress : Global.currentApplicationAddress
  }

  /**
   * Create an abstracted account application.
   * This is not part of ARC58 and implementation specific.
   *
   * @param controlledAddress The address of the abstracted account. If zeroAddress, then the address of the contract account will be used
   * @param admin The admin for this app
  */
  @abimethod({ onCreate: 'require' })
  createApplication(controlledAddress: Address, admin: Address, escrowFactory: Application): void {
    assert(
      Txn.sender === controlledAddress.native
      || Txn.sender === admin.native,
      ERR_SENDER_MUST_BE_ADMIN_OR_CONTROLLED_ADDRESS
    );
    assert(admin !== controlledAddress);

    this.admin.value = admin.native;
    this.controlledAddress.value = controlledAddress.native === Global.zeroAddress ? Global.currentApplicationAddress : controlledAddress.native;
    this.escrowFactory.value = escrowFactory;
    this.spendingAddress.value = Global.zeroAddress;
    this.updateLastUserInteraction()
    this.updateLastChange()
  }

  /**
   * Register the abstracted account with the escrow factory.
   * This allows apps to correlate the account with the app without needing
   * it to be explicitly provided.
   */
  register(escrow: string): void {
    let app: uint64 = 0
    if (escrow !== '') {
      assert(this.escrows(escrow).exists, ERR_ESCROW_DOES_NOT_EXIST)
      app = this.escrows(escrow).value.id
    }

    abiCall<typeof EscrowFactory.prototype.register>({
      appId: this.escrowFactory.value,
      args: [
        itxn.payment({
          receiver: this.escrowFactory.value.address,
          amount: ARC58WalletIDsByAccountsMbr
        }),
        app
      ]
    })
  }

  /**
   * Attempt to change the admin for this app. Some implementations MAY not support this.
   *
   * @param newAdmin The new admin
  */
  arc58_changeAdmin(newAdmin: Address): void {
    assert(Txn.sender === this.admin.value, ERR_ONLY_ADMIN_CAN_CHANGE_ADMIN);
    this.admin.value = newAdmin.native;
    this.updateLastUserInteraction()
    this.updateLastChange()
  }

  /**
   * Attempt to change the admin via plugin.
   *
   * @param plugin The app calling the plugin
   * @param allowedCaller The address that triggered the plugin
   * @param newAdmin The new admin
   *
  */
  arc58_pluginChangeAdmin(newAdmin: Address): void {
    const key = clone(this.currentPlugin.value)
    const { plugin, escrow } = key

    assert(escrow === '', ERR_ADMIN_PLUGINS_CANNOT_USE_ESCROWS);
    assert(Txn.sender === Application(plugin).address, ERR_SENDER_MUST_BE_ADMIN_PLUGIN);
    assert(
      this.controlledAddress.value.authAddress === Application(plugin).address,
      'This plugin is not in control of the account'
    );

    assert(
      this.plugins(key).exists && this.plugins(key).value.admin,
      'This plugin does not have admin privileges'
    );

    this.admin.value = newAdmin.native;
    if (this.plugins(key).value.delegationType === DelegationTypeSelf) {
      this.updateLastUserInteraction();
    }
    this.updateLastChange()
  }

  /**
   * Get the admin of this app. This method SHOULD always be used rather than reading directly from state
   * because different implementations may have different ways of determining the admin.
  */
  @abimethod({ readonly: true })
  arc58_getAdmin(): Address {
    return new Address(this.admin.value);
  }

  /**
   * Verify the abstracted account is rekeyed to this app
  */
  arc58_verifyAuthAddress(): void {
    assert(this.spendingAddress.value.authAddress === this.getAuthAddress());
    this.spendingAddress.value = Global.zeroAddress
    this.currentPlugin.value = { plugin: 0, caller: Global.currentApplicationAddress, escrow: '' }
    this.rekeyIndex.value = 0
  }

  /**
   * Rekey the abstracted account to another address. Primarily useful for rekeying to an EOA.
   *
   * @param address The address to rekey to
   * @param flash Whether or not this should be a flash rekey. If true, the rekey back to the app address must done in the same txn group as this call
  */
  arc58_rekeyTo(address: Address, flash: boolean): void {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);

    itxn
      .payment({
        sender: this.controlledAddress.value,
        receiver: address.native,
        rekeyTo: address.native,
        note: 'rekeying abstracted account'
      })
      .submit();

    if (flash) this.assertRekeysBack();

    this.updateLastUserInteraction();
  }

  /**
   * check whether the plugin can be used
   *
   * @param plugin the plugin to be rekeyed to
   * @param global whether this is callable globally
   * @param address the address that will trigger the plugin
   * @param method the method being called on the plugin, if applicable
   * @returns whether the plugin can be called with these parameters
  */
  @abimethod({ readonly: true })
  arc58_canCall(
    plugin: uint64,
    global: boolean,
    address: Address,
    escrow: string,
    method: bytes<4>
  ): boolean {
    if (global) {
      this.pluginCallAllowed(plugin, Global.zeroAddress, escrow, method);
    }
    return this.pluginCallAllowed(plugin, address.native, escrow, method);
  }

  /**
   * Temporarily rekey to an approved plugin app address
   *
   * @param plugin The app to rekey to
   * @param global Whether the plugin is callable globally
   * @param methodOffsets The indices of the methods being used in the group if the plugin has method restrictions these indices are required to match the methods used on each subsequent call to the plugin within the group
   * @param fundsRequest If the plugin is using an escrow, this is the list of funds to transfer to the escrow for the plugin to be able to use during execution
   * 
  */
  arc58_rekeyToPlugin(
    plugin: uint64,
    global: boolean,
    escrow: string,
    methodOffsets: uint64[],
    fundsRequest: FundsRequest[]
  ): void {
    const pluginApp = Application(plugin)
    const caller = global ? Global.zeroAddress : Txn.sender
    const key: PluginKey = { plugin, caller, escrow }

    assert(this.plugins(key).exists, ERR_PLUGIN_DOES_NOT_EXIST)
    this.currentPlugin.value = clone(key)

    if (escrow !== '') {
      assert(this.escrows(escrow).exists, ERR_ESCROW_DOES_NOT_EXIST)
      const escrowID = this.escrows(escrow).value.id
      const spendingApp = Application(escrowID)
      this.spendingAddress.value = spendingApp.address
      this.transferFunds(escrow, fundsRequest)
    } else {
      this.spendingAddress.value = this.controlledAddress.value
    }

    this.assertValidGroup(key, methodOffsets)

    itxn
      .payment({
        sender: this.spendingAddress.value,
        receiver: this.spendingAddress.value,
        rekeyTo: pluginApp.address,
        note: 'rekeying to plugin app'
      })
      .submit();

    /** track the index of the transaction that triggered the rekey */
    this.rekeyIndex.value = Txn.groupIndex

    if (this.plugins(key).value.delegationType === DelegationTypeSelf) {
      this.updateLastUserInteraction();
    }
  }

  /**
   * Temporarily rekey to a named plugin app address
   *
   * @param name The name of the plugin to rekey to
   * @param global Whether the plugin is callable globally
   * @param methodOffsets The indices of the methods being used in the group if the plugin has method restrictions these indices are required to match the methods used on each subsequent call to the plugin within the group
   * @param fundsRequest If the plugin is using an escrow, this is the list of funds to transfer to the escrow for the plugin to be able to use during execution
   *
  */
  arc58_rekeyToNamedPlugin(
    name: string,
    global: boolean,
    escrow: string,
    methodOffsets: uint64[],
    fundsRequest: FundsRequest[]): void {
    this.arc58_rekeyToPlugin(
      this.namedPlugins(name).value.plugin,
      global,
      escrow,
      methodOffsets,
      fundsRequest
    );
  }

  /**
   * Add an app to the list of approved plugins
   *
   * @param app The app to add
   * @param allowedCaller The address of that's allowed to call the app
   * or the global zero address for any address
   * @param admin Whether the plugin has permissions to change the admin account
   * @param delegationType the ownership of the delegation for last_interval updates
   * @param escrow The escrow account to use for the plugin, if any. If empty, no escrow will be used, if the named escrow does not exist, it will be created
   * @param lastValid The timestamp or round when the permission expires
   * @param cooldown The number of seconds or rounds that must pass before the plugin can be called again
   * @param methods The methods that are allowed to be called for the plugin by the address
   * @param useRounds Whether the plugin uses rounds for cooldowns and lastValid, defaults to timestamp
  */
  arc58_addPlugin(
    plugin: uint64,
    caller: Address,
    escrow: string,
    admin: boolean,
    delegationType: Uint8,
    lastValid: uint64,
    cooldown: uint64,
    methods: MethodRestriction[],
    useRounds: boolean,
    useExecutionKey: boolean,
    defaultToEscrow: boolean
  ): void {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);
    assert(
      !(
        delegationType === DelegationTypeSelf &&
        caller.native === Global.zeroAddress
      ),
      ERR_ZERO_ADDRESS_DELEGATION_TYPE
    )
    assert(
      !(
        useExecutionKey &&
        caller.native !== Global.zeroAddress
      ),
      ERR_USING_EXECUTION_KEY_REQUIRES_GLOBAL
    )

    let escrowKey: string = escrow
    if (defaultToEscrow) {
      assert(escrow !== '', ERR_ESCROW_REQUIRED_TO_BE_SET_AS_DEFAULT)
      escrowKey = ''
    }

    const key: PluginKey = { plugin, caller: caller.native, escrow: escrowKey }

    let methodInfos: MethodInfo[] = []
    for (let i: uint64 = 0; i < methods.length; i += 1) {
      methodInfos.push({ ...methods[i], lastCalled: 0 })
    }

    const epochRef = useRounds ? Global.round : Global.latestTimestamp;

    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          sender: this.controlledAddress.value,
          receiver: Global.currentApplicationAddress,
          amount: this.pluginsMbr(escrowKey, methodInfos.length)
        })
        .submit()
    }

    const escrowID = this.maybeNewEscrow(escrow);

    this.plugins(key).value = {
      escrow: escrowID,
      admin,
      delegationType,
      lastValid,
      cooldown,
      methods: clone(methodInfos),
      useRounds,
      useExecutionKey,
      lastCalled: 0,
      start: epochRef,
    }

    this.updateLastUserInteraction();
    this.updateLastChange();
  }

  /**
   * Remove an app from the list of approved plugins
   *
   * @param app The app to remove
   * @param allowedCaller The address that's allowed to call the app
  */
  arc58_removePlugin(plugin: uint64, caller: Address, escrow: string): void {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);

    const key: PluginKey = { plugin, caller: caller.native, escrow }
    assert(this.plugins(key).exists, ERR_PLUGIN_DOES_NOT_EXIST)

    const methodsLength: uint64 = this.plugins(key).value.methods.length

    this.plugins(key).delete();

    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          receiver: this.controlledAddress.value,
          amount: this.pluginsMbr(escrow, methodsLength)
        })
        .submit()
    }

    this.updateLastUserInteraction();
    this.updateLastChange();
  }

  /**
   * Add a named plugin
   *
   * @param name The plugin name
   * @param app The app to add
   * @param allowedCaller The address that's allowed to call the app
   * or the global zero address for any address
   * @param admin Whether the plugin has permissions to change the admin account
   * @param delegationType the ownership of the delegation for last_interval updates
   * @param escrow The escrow account to use for the plugin, if any. If empty, no escrow will be used, if the named escrow does not exist, it will be created
   * @param lastValid The timestamp or round when the permission expires
   * @param cooldown The number of seconds or rounds that must pass before the plugin can be called again
   * @param methods The methods that are allowed to be called for the plugin by the address
   * @param useRounds Whether the plugin uses rounds for cooldowns and lastValid, defaults to timestamp
  */
  arc58_addNamedPlugin(
    name: string,
    plugin: uint64,
    caller: Address,
    escrow: string,
    admin: boolean,
    delegationType: Uint8,
    lastValid: uint64,
    cooldown: uint64,
    methods: MethodRestriction[],
    useRounds: boolean,
    useExecutionKey: boolean,
    defaultToEscrow: boolean
  ): void {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);
    assert(!this.namedPlugins(name).exists);
    assert(
      !(
        delegationType === DelegationTypeSelf &&
        caller.native === Global.zeroAddress
      ),
      ERR_ZERO_ADDRESS_DELEGATION_TYPE
    )
    assert(
      !(
        useExecutionKey &&
        caller.native !== Global.zeroAddress
      ),
      ERR_USING_EXECUTION_KEY_REQUIRES_GLOBAL
    )

    let escrowKey: string = escrow
    if (defaultToEscrow) {
      assert(escrow !== '', ERR_ESCROW_REQUIRED_TO_BE_SET_AS_DEFAULT)
      escrowKey = ''
    }

    const key: PluginKey = { plugin, caller: caller.native, escrow: escrowKey }
    this.namedPlugins(name).value = clone(key)

    let methodInfos: MethodInfo[] = []
    for (let i: uint64 = 0; i < methods.length; i += 1) {
      methodInfos.push({ ...methods[i], lastCalled: 0 })
    }

    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          sender: this.controlledAddress.value,
          receiver: Global.currentApplicationAddress,
          amount: this.pluginsMbr(escrowKey, methodInfos.length) + this.namedPluginsMbr(name)
        })
        .submit()
    }

    const escrowID = this.maybeNewEscrow(escrow);

    const epochRef = useRounds ? Global.round : Global.latestTimestamp;

    this.plugins(key).value = {
      escrow: escrowID,
      admin,
      delegationType,
      lastValid,
      cooldown,
      methods: clone(methodInfos),
      useRounds,
      useExecutionKey,
      lastCalled: 0,
      start: epochRef
    }

    this.updateLastUserInteraction();
    this.updateLastChange();
  }

  /**
   * Remove a named plugin
   *
   * @param name The plugin name
  */
  arc58_removeNamedPlugin(name: string): void {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);
    assert(this.namedPlugins(name).exists, ERR_PLUGIN_DOES_NOT_EXIST);
    const app = clone(this.namedPlugins(name).value)
    assert(this.plugins(app).exists, ERR_PLUGIN_DOES_NOT_EXIST);

    const methodsLength: uint64 = this.plugins(app).value.methods.length

    this.namedPlugins(name).delete();
    this.plugins(app).delete();

    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          receiver: this.controlledAddress.value,
          amount: this.namedPluginsMbr(name) + this.pluginsMbr(app.escrow, methodsLength)
        })
        .submit()
    }

    this.updateLastUserInteraction();
    this.updateLastChange();
  }

  /**
   * Create a new escrow for the controlled address
   *
   * @param escrow The name of the escrow to create
  */
  arc58_newEscrow(escrow: string): uint64 {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);
    assert(!this.escrows(escrow).exists, ERR_ESCROW_ALREADY_EXISTS);
    assert(escrow !== '', ERR_ESCROW_NAME_REQUIRED);
    return this.newEscrow(escrow);
  }

  /**
   * Lock or Unlock an escrow account
   *
   * @param escrow The escrow to lock or unlock
  */
  arc58_toggleEscrowLock(escrow: string): EscrowInfo {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);
    assert(this.escrows(escrow).exists, ERR_ESCROW_DOES_NOT_EXIST);

    this.escrows(escrow).value.locked = !this.escrows(escrow).value.locked;

    this.updateLastUserInteraction();
    this.updateLastChange();

    return this.escrows(escrow).value;
  }

  /**
   * Transfer funds from an escrow back to the controlled address.
   * 
   * @param escrow The escrow to reclaim funds from
   * @param reclaims The list of reclaims to make from the escrow
  */
  arc58_reclaim(escrow: string, reclaims: EscrowReclaim[]): void {
    assert(Txn.sender === this.admin.value, ERR_FORBIDDEN);
    assert(this.escrows(escrow).exists, ERR_ESCROW_DOES_NOT_EXIST);
    const sender = Application(this.escrows(escrow).value.id).address

    for (let i: uint64 = 0; i < reclaims.length; i += 1) {
      if (reclaims[i].asset === 0) {
        const pmt = itxn.payment({
          sender,
          receiver: this.controlledAddress.value,
          amount: reclaims[i].amount
        })

        if (reclaims[i].closeOut) {
          pmt.set({ closeRemainderTo: this.controlledAddress.value });
        }

        pmt.submit();
      } else {
        const xfer = itxn.assetTransfer({
          sender,
          assetReceiver: this.controlledAddress.value,
          assetAmount: reclaims[i].amount,
          xferAsset: reclaims[i].asset
        })

        if (reclaims[i].closeOut) {
          xfer.set({ assetCloseTo: this.controlledAddress.value });
        }

        xfer.submit();
      }
    }
  }

  /**
   * Opt-in an escrow account to assets
   *
   * @param escrow The escrow to opt-in to
   * @param assets The list of assets to opt-in to
  */
  arc58_optinEscrow(escrow: string, assets: uint64[]): void {
    assert(Txn.sender === this.admin.value, ERR_FORBIDDEN);
    assert(this.escrows(escrow).exists, ERR_ESCROW_DOES_NOT_EXIST)
    const escrowID = this.escrows(escrow).value.id
    const escrowAddress = Application(escrowID).address
    assert(!this.escrows(escrow).value.locked, ERR_ESCROW_LOCKED)

    itxn
      .payment({
        sender: this.controlledAddress.value,
        receiver: escrowAddress,
        amount: Global.assetOptInMinBalance * assets.length
      })
      .submit();

    for (let i: uint64 = 0; i < assets.length; i += 1) {
      assert(
        this.allowances({ escrow, asset: assets[i] }).exists,
        ERR_ALLOWANCE_DOES_NOT_EXIST
      );

      itxn
        .assetTransfer({
          sender: escrowAddress,
          assetReceiver: escrowAddress,
          assetAmount: 0,
          xferAsset: assets[i]
        })
        .submit();
    }
  }

  /**
   * Opt-in an escrow account to assets via a plugin / allowed caller
   *
   * @param app The app related to the escrow optin
   * @param allowedCaller The address allowed to call the plugin related to the escrow optin
   * @param assets The list of assets to opt-in to
   * @param mbrPayment The payment txn that is used to pay for the asset opt-in
  */
  arc58_pluginOptinEscrow(
    plugin: uint64,
    caller: Address,
    escrow: string,
    assets: uint64[],
    mbrPayment: gtxn.PaymentTxn
  ): void {
    const key: PluginKey = { plugin, caller: caller.native, escrow }

    assert(this.plugins(key).exists, ERR_PLUGIN_DOES_NOT_EXIST)
    assert(this.escrows(escrow).exists, ERR_ESCROW_DOES_NOT_EXIST)
    assert(!this.escrows(escrow).value.locked, ERR_ESCROW_LOCKED)

    const escrowID = this.escrows(escrow).value.id

    assert(
      Txn.sender === Application(plugin).address ||
      Txn.sender === caller.native ||
      caller.native === Global.zeroAddress,
      ERR_FORBIDDEN
    )

    const escrowAddress = Application(escrowID).address

    assertMatch(
      mbrPayment,
      {
        receiver: this.controlledAddress.value,
        amount: Global.assetOptInMinBalance * assets.length
      },
      ERR_INVALID_PAYMENT
    )

    itxn
      .payment({
        sender: this.controlledAddress.value,
        receiver: escrowAddress,
        amount: Global.assetOptInMinBalance * assets.length
      })
      .submit();

    for (let i: uint64 = 0; i < assets.length; i += 1) {
      assert(
        this.allowances({ escrow, asset: assets[i] }).exists,
        ERR_ALLOWANCE_DOES_NOT_EXIST
      );

      itxn
        .assetTransfer({
          sender: escrowAddress,
          assetReceiver: escrowAddress,
          assetAmount: 0,
          xferAsset: assets[i]
        })
        .submit();
    }
  }

  /**
   * Add an allowance for an escrow account
   *
   * @param escrow The escrow to add the allowance for
   * @param allowances The list of allowances to add
  */
  arc58_addAllowances(escrow: string, allowances: AddAllowanceInfo[]): void {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);
    assert(this.escrows(escrow).exists, ERR_ESCROW_DOES_NOT_EXIST);
    assert(!this.escrows(escrow).value.locked, ERR_ESCROW_LOCKED);

    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          sender: this.controlledAddress.value,
          receiver: Global.currentApplicationAddress,
          amount: this.allowancesMbr(escrow) * allowances.length
        })
        .submit()
    }

    for (let i: uint64 = 0; i < allowances.length; i += 1) {
      const { asset, type, amount, max, interval, useRounds } = allowances[i];
      const key: AllowanceKey = { escrow, asset }
      assert(!this.allowances(key).exists, ERR_ALLOWANCE_ALREADY_EXISTS);
      const start = useRounds ? Global.round : Global.latestTimestamp;

      this.allowances(key).value = {
        type,
        spent: 0,
        amount,
        last: 0,
        max,
        interval,
        start,
        useRounds
      }
    }

    this.updateLastUserInteraction();
    this.updateLastChange();
  }

  /**
   * Remove an allowances for an escrow account
   *
   * @param escrow The escrow to remove the allowance for
   * @param assets The list of assets to remove the allowance for
  */
  arc58_removeAllowances(escrow: string, assets: uint64[]): void {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);
    assert(this.escrows(escrow).exists, ERR_ESCROW_DOES_NOT_EXIST);
    assert(!this.escrows(escrow).value.locked, ERR_ESCROW_LOCKED);

    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          receiver: this.controlledAddress.value,
          amount: this.allowancesMbr(escrow) * assets.length
        })
        .submit()
    }

    for (let i: uint64 = 0; i < assets.length; i += 1) {
      const key: AllowanceKey = {
        escrow,
        asset: assets[i]
      }
      assert(this.allowances(key).exists, ERR_ALLOWANCE_DOES_NOT_EXIST)
      this.allowances(key).delete()
    }

    this.updateLastUserInteraction()
    this.updateLastChange()
  }

  arc58_addExecutionKey(lease: bytes<32>, groups: bytes<32>[], firstValid: uint64, lastValid: uint64): void {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY)
    if (!this.executions(lease).exists) {
      this.executions(lease).value = {
        groups: clone(groups),
        firstValid,
        lastValid
      }
    } else {
      assert(this.executions(lease).value.firstValid === firstValid, ERR_EXECUTION_KEY_UPDATE_MUST_MATCH_FIRST_VALID)
      assert(this.executions(lease).value.lastValid === lastValid, ERR_EXECUTION_KEY_UPDATE_MUST_MATCH_LAST_VALID)

      this.executions(lease).value.groups = [...clone(this.executions(lease).value.groups), ...clone(groups)]
    }

    this.updateLastUserInteraction()
    this.updateLastChange()
  }

  arc58_removeExecutionKey(lease: bytes<32>): void {
    assert(this.executions(lease).exists, ERR_EXECUTION_KEY_DOES_NOT_EXIST)
    assert(Txn.sender === this.admin.value || this.executions(lease).value.lastValid < Global.round, ERR_ADMIN_ONLY)

    this.executions(lease).delete()

    this.updateLastUserInteraction()
    this.updateLastChange()
  }

  @abimethod({ readonly: true })
  arc58_getPlugins(keys: PluginKey[]): PluginInfo[] {
    let plugins: PluginInfo[] = []
    for (let i: uint64 = 0; i < keys.length; i += 1) {
      if (this.plugins(keys[i]).exists) {
        plugins.push(this.plugins(keys[i]).value)
        continue
      }
      plugins.push(emptyPluginInfo())
    }
    return plugins
  }

  @abimethod({ readonly: true })
  arc58_getNamedPlugins(names: string[]): PluginInfo[] {
    let plugins: PluginInfo[] = []
    for (let i: uint64 = 0; i < names.length; i += 1) {
      if (this.namedPlugins(names[i]).exists) {
        const nameKey = clone(this.namedPlugins(names[i]).value)
        if (this.plugins(nameKey).exists) {
          plugins.push(this.plugins(nameKey).value)
          continue
        }
        plugins.push(emptyPluginInfo())
        continue
      }
      plugins.push(emptyPluginInfo())
    }
    return plugins
  }

  @abimethod({ readonly: true })
  arc58_getEscrows(escrows: string[]): EscrowInfo[] {
    let result: EscrowInfo[] = []
    for (let i: uint64 = 0; i < escrows.length; i += 1) {
      if (this.escrows(escrows[i]).exists) {
        result.push(this.escrows(escrows[i]).value)
        continue
      }
      result.push(emptyEscrowInfo())
    }
    return result
  }

  @abimethod({ readonly: true })
  arc58_getAllowances(escrow: string, assets: uint64[]): AllowanceInfo[] {
    let result: AllowanceInfo[] = []
    for (let i: uint64 = 0; i < assets.length; i += 1) {
      const key: AllowanceKey = { escrow, asset: assets[i] }
      if (this.allowances(key).exists) {
        result.push(this.allowances(key).value)
        continue
      }
      result.push(emptyAllowanceInfo())
    }
    return result
  }

  @abimethod({ readonly: true })
  arc58_getExecutions(leases: bytes<32>[]): ExecutionInfo[] {
    let result: ExecutionInfo[] = []
    for (let i: uint64 = 0; i < leases.length; i += 1) {
      if (this.executions(leases[i]).exists) {
        result.push(this.executions(leases[i]).value)
        continue
      }
      result.push(emptyExecutionInfo())
    }
    return result
  }

  @abimethod({ readonly: true })
  mbr(
    escrow: string,
    methodCount: uint64,
    plugin: string,
    groups: uint64,
  ): AbstractAccountBoxMBRData {
    const escrows = this.escrowsMbr(escrow)

    return {
      plugins: this.pluginsMbr(escrow, methodCount),
      namedPlugins: this.namedPluginsMbr(plugin),
      escrows,
      allowances: this.allowancesMbr(escrow),
      executions: this.executionsMbr(groups),
      escrowExists: this.escrows(escrow).exists,
      newEscrowMintCost: (
        NewCostForARC58 +
        Global.minBalance +
        ARC58WalletIDsByAccountsMbr +
        escrows
      )
    }
  }
}
