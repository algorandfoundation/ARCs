import { Contract, GlobalState, BoxMap, assert, uint64, Account, TransactionType, Application, abimethod, gtxn, itxn, OnCompleteAction, Bytes, bytes, assertMatch } from '@algorandfoundation/algorand-typescript'
import { abiCall, Address, Bool, decodeArc4, DynamicArray, methodSelector, UintN64, UintN8 } from '@algorandfoundation/algorand-typescript/arc4';
import { btoi, Global, len, Txn } from '@algorandfoundation/algorand-typescript/op'
import { ERR_ADMIN_ONLY, ERR_ALLOWANCE_ALREADY_EXISTS, ERR_ALLOWANCE_DOES_NOT_EXIST, ERR_ALLOWANCE_EXCEEDED, ERR_CANNOT_CALL_OTHER_APPS_DURING_REKEY, ERR_ESCROW_ALREADY_EXISTS, ERR_ESCROW_DOES_NOT_EXIST, ERR_INVALID_METHOD_SIGNATURE_LENGTH, ERR_INVALID_ONCOMPLETE, ERR_INVALID_PLUGIN_CALL, ERR_INVALID_SENDER_ARG, ERR_INVALID_SENDER_VALUE, ERR_MALFORMED_OFFSETS, ERR_METHOD_ON_COOLDOWN, ERR_MISSING_REKEY_BACK, ERR_NOT_USING_ALLOWANCE, ERR_NOT_USING_ESCROW, ERR_ONLY_ADMIN_CAN_CHANGE_ADMIN, ERR_PLUGIN_DOES_NOT_EXIST, ERR_PLUGIN_EXPIRED, ERR_PLUGIN_ON_COOLDOWN, ERR_SENDER_MUST_BE_ADMIN_OR_CONTROLLED_ADDRESS, ERR_SENDER_MUST_BE_ADMIN_PLUGIN, ERR_ZERO_ADDRESS_DELEGATION_TYPE } from './errors';
import { AbstractAccountBoxMBRData, AddAllowanceInfo, AllowanceInfo, AllowanceKey, arc4MethodInfo, arc4MethodRestriction, arc4PluginInfo, DelegationTypeSelf, EscrowReclaim, FullPluginValidation, FundsRequest, MethodRestriction, MethodValidation, PluginInfo, PluginKey, PluginValidation, SpendAllowanceType, SpendAllowanceTypeDrip, SpendAllowanceTypeFlat, SpendAllowanceTypeWindow } from './types';
import { EscrowFactory } from '../escrow/factory.algo';
import { AbstractAccountBoxPrefixAllowances, AbstractAccountBoxPrefixEscrows, AbstractAccountBoxPrefixNamedPlugins, AbstractAccountBoxPrefixPlugins, AbstractAccountGlobalStateKeysAdmin, AbstractAccountGlobalStateKeysControlledAddress, AbstractAccountGlobalStateKeysEscrowFactory, AbstractAccountGlobalStateKeysLastChange, AbstractAccountGlobalStateKeysLastUserInteraction, AbstractAccountGlobalStateKeysSpendingAddress, AllowanceMBR, BoxCostPerByte, DynamicLength, DynamicOffset, DynamicOffsetAndLength, MethodRestrictionByteLength, MinEscrowsMBR, MinNamedPluginMBR, MinPluginMBR } from './constants';
import { fee } from "../utils/constants";
import { ERR_INVALID_PAYMENT } from '../utils/errors';
import { ERR_FORBIDDEN } from '../escrow/errors';
import { MinPages, NewCostForARC58 } from '../escrow/constants';

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

  /**
   * TEMPORARY STATE FIELDS
   * 
   * These are global state fields that are used for sharing metadata about usage
   * of the smart wallet and cleared before the end of the usage of the plugin.
   * by doing this we avoid sending application calls to fetch box data
   * & save ourselves from adding more arg requirements for plugins to adhere to.
  */

  /** The spending address for the currently active plugin */
  spendingAddress = GlobalState<Account>({ key: AbstractAccountGlobalStateKeysSpendingAddress })

  /** Plugins that add functionality to the controlledAddress and the account that has permission to use it. */
  plugins = BoxMap<PluginKey, arc4PluginInfo>({ keyPrefix: AbstractAccountBoxPrefixPlugins });
  /** Plugins that have been given a name for discoverability */
  namedPlugins = BoxMap<string, PluginKey>({ keyPrefix: AbstractAccountBoxPrefixNamedPlugins });
  /** the escrows that this wallet has created for specific callers with allowances */
  escrows = BoxMap<string, uint64>({ keyPrefix: AbstractAccountBoxPrefixEscrows })
  /** The Allowances for plugins installed on the smart contract with useAllowance set to true */
  allowances = BoxMap<AllowanceKey, AllowanceInfo>({ keyPrefix: AbstractAccountBoxPrefixAllowances }) // 38_500

  private updateLastUserInteraction() {
    this.lastUserInteraction.value = Global.latestTimestamp
  }

  private updateLastChange() {
    this.lastChange.value = Global.latestTimestamp
  }

  private pluginsMbr(methodCount: uint64): uint64 {
    return MinPluginMBR + (
      BoxCostPerByte * (
        (MethodRestrictionByteLength * methodCount)
        + DynamicOffsetAndLength
      )
    );
  }

  private namedPluginsMbr(name: string): uint64 {
    return MinNamedPluginMBR + (BoxCostPerByte * Bytes(name).length);
  }

  private escrowsMbr(name: string): uint64 {
    return MinEscrowsMBR + (BoxCostPerByte * Bytes(name).length);
  }

  private allowancesMbr(): uint64 {
    return AllowanceMBR;
  }

  private maybeNewEscrow(escrow: string): uint64 {
    if (escrow === '') {
      return 0;
    }

    return this.escrows(escrow).exists
      ? this.escrows(escrow).value
      : this.newEscrow(escrow);
  }

  private newEscrow(escrow: string): uint64 {
    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          sender: this.controlledAddress.value,
          receiver: Global.currentApplicationAddress,
          amount: this.escrowsMbr(escrow),
          fee,
        })
        .submit()
    }

    const escrowID = abiCall(
      EscrowFactory.prototype.new,
      {
        sender: this.controlledAddress.value,
        appId: this.escrowFactory.value,
        args: [
          itxn.payment({
            sender: this.controlledAddress.value,
            amount: NewCostForARC58 + Global.minBalance,
            receiver: this.escrowFactory.value.address,
            fee,
          }),
        ],
        fee,
      }
    ).returnValue

    this.escrows(escrow).value = escrowID;

    return escrowID;
  }

  private pluginCallAllowed(application: uint64, allowedCaller: Account, method: bytes<4>): boolean {
    const key: PluginKey = { application, allowedCaller }

    if (!this.plugins(key).exists) {
      return false;
    }

    const methods = this.plugins(key).value.methods.copy();
    let methodAllowed = methods.length > 0 ? false : true;
    for (let i: uint64 = 0; i < methods.length; i += 1) {
      if (methods[i].selector.native === method) {
        methodAllowed = true;
        break;
      }
    }

    const p = decodeArc4<PluginInfo>(this.plugins(key).value.copy().bytes);
    const epochRef = p.useRounds ? Global.round : Global.latestTimestamp;

    return (
      p.lastCalled >= epochRef &&
      (epochRef - p.lastCalled) >= p.cooldown &&
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
      && txn.appArgs(0) === methodSelector('arc58_verifyAuthAddr()void')
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
        hasCooldown: true,
        onCooldown: true,
        hasMethodRestrictions: false,
        valid: false
      }
    }

    const pluginInfo = decodeArc4<PluginInfo>(this.plugins(key).value.copy().bytes);
    const epochRef = pluginInfo.useRounds ? Global.round : Global.latestTimestamp;

    const expired = epochRef > pluginInfo.lastValid;
    const hasCooldown = pluginInfo.cooldown > 0;
    const onCooldown = (epochRef - pluginInfo.lastCalled) < pluginInfo.cooldown;
    const hasMethodRestrictions = pluginInfo.methods.length > 0;

    const valid = exists && !expired && !onCooldown;

    return {
      exists,
      expired,
      hasCooldown,
      onCooldown,
      hasMethodRestrictions,
      valid
    }
  }

  private fullPluginCheck(
    key: PluginKey,
    txn: gtxn.ApplicationCallTxn,
    methodOffsets: uint64[],
    methodIndex: uint64
  ): FullPluginValidation {

    const check = this.pluginCheck(key);

    if (!check.valid) {
      return {
        ...check,
        methodAllowed: false,
        methodHasCooldown: true,
        methodOnCooldown: true
      }
    }

    let mCheck: MethodValidation = {
      methodAllowed: !check.hasMethodRestrictions,
      methodHasCooldown: false,
      methodOnCooldown: false
    }

    if (check.hasMethodRestrictions) {
      assert(methodIndex < methodOffsets.length, ERR_MALFORMED_OFFSETS);
      mCheck = this.methodCheck(key, txn, methodOffsets[methodIndex]);
    }

    return {
      ...check,
      ...mCheck,
      valid: check.valid && mCheck.methodAllowed
    }
  }

  /**
   * Guarantee that our txn group is valid in a single loop over all txns in the group
   * 
   * @param key the box key for the plugin were checking
   * @param methodOffsets the indices of the methods being used in the group
  */
  private assertValidGroup(key: PluginKey, methodOffsets: uint64[]): void {

    const epochRef = this.plugins(key).value.useRounds.native
      ? Global.round
      : Global.latestTimestamp;

    const initialCheck = this.pluginCheck(key);

    assert(initialCheck.exists, ERR_PLUGIN_DOES_NOT_EXIST);
    assert(!initialCheck.expired, ERR_PLUGIN_EXPIRED);
    assert(!initialCheck.onCooldown, ERR_PLUGIN_ON_COOLDOWN);

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

      assert(txn.appId.id === key.application, ERR_CANNOT_CALL_OTHER_APPS_DURING_REKEY);
      assert(txn.onCompletion === OnCompleteAction.NoOp, ERR_INVALID_ONCOMPLETE);
      // ensure the first arg to a method call is the app id itself
      // index 1 is used because arg[0] is the method selector
      assert(txn.numAppArgs > 1, ERR_INVALID_SENDER_ARG);
      assert(Application(btoi(txn.appArgs(1))) === Global.currentApplicationId, ERR_INVALID_SENDER_VALUE);

      const check = this.fullPluginCheck(key, txn, methodOffsets, methodIndex);

      assert(!check.methodOnCooldown, ERR_METHOD_ON_COOLDOWN);
      assert(check.valid, ERR_INVALID_PLUGIN_CALL);

      if (initialCheck.hasCooldown) {
        this.plugins(key).value.lastCalled = new UintN64(epochRef);
      }

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

    assert(len(txn.appArgs(0)) === 4, ERR_INVALID_METHOD_SIGNATURE_LENGTH);
    const selectorArg = txn.appArgs(0).toFixed({ length: 4 });

    const methods = this.plugins(key).value.methods.copy()
    const allowedMethod = methods[offset].copy();

    const hasCooldown = allowedMethod.cooldown.native > 0;

    const useRounds = this.plugins(key).value.useRounds.native

    const epochRef = useRounds ? Global.round : Global.latestTimestamp;
    const onCooldown = (epochRef - allowedMethod.lastCalled.native) < allowedMethod.cooldown.native;

    if (allowedMethod.selector.native === selectorArg && (!hasCooldown || !onCooldown)) {
      // update the last called round for the method
      if (hasCooldown) {
        const lastCalled = useRounds
          ? Global.round
          : Global.latestTimestamp;

        methods[offset].lastCalled = new UintN64(lastCalled);

        this.plugins(key).value = new arc4PluginInfo({
          ...this.plugins(key).value,
          methods: methods.copy()
        });
      }

      return {
        methodAllowed: true,
        methodHasCooldown: hasCooldown,
        methodOnCooldown: onCooldown
      }
    }

    return {
      methodAllowed: false,
      methodHasCooldown: true,
      methodOnCooldown: true
    }
  }

  private transferFunds(key: PluginKey, fundsRequests: FundsRequest[]): void {
    for (let i: uint64 = 0; i < fundsRequests.length; i += 1) {
      
      const pluginInfo = decodeArc4<PluginInfo>(this.plugins(key).value.bytes);

      const allowanceKey: AllowanceKey = {
        escrow: pluginInfo.escrow,
        asset: fundsRequests[i].asset
      }

      this.verifyAllowance(allowanceKey, fundsRequests[i]);

      if (fundsRequests[i].asset !== 0) {
        itxn
          .assetTransfer({
            sender: this.controlledAddress.value,
            assetReceiver: this.spendingAddress.value,
            assetAmount: fundsRequests[i].amount,
            xferAsset: fundsRequests[i].asset,
            fee,
          })
          .submit();
      } else {
        itxn
          .payment({
            sender: this.controlledAddress.value,
            receiver: this.spendingAddress.value,
            amount: fundsRequests[i].amount,
            fee,
          })
          .submit();
      }
    }
  }

  private verifyAllowance(key: AllowanceKey, fundRequest: FundsRequest): void {
    assert(this.allowances(key).exists, ERR_ALLOWANCE_DOES_NOT_EXIST);
    const { type, spent, allowed, last, max, interval, start, useRounds } = this.allowances(key).value
    const newLast = useRounds ? Global.round : Global.latestTimestamp;

    if (type === SpendAllowanceTypeFlat) {
      const leftover: uint64 = allowed - spent;

      assert(leftover >= fundRequest.amount, ERR_ALLOWANCE_EXCEEDED);

      this.allowances(key).value = {
        ...this.allowances(key).value,
        spent: (spent + fundRequest.amount)
      }
    } else if (type === SpendAllowanceTypeWindow) {
      const currentWindowStart = this.getLatestWindowStart(useRounds, start, interval)

      if (currentWindowStart > last) {
        assert(allowed >= fundRequest.amount, ERR_ALLOWANCE_EXCEEDED);

        this.allowances(key).value = {
          ...this.allowances(key).value,
          spent: fundRequest.amount,
          last: newLast
        }
      } else {
        // calc the remaining amount available in the current window
        const leftover: uint64 = allowed - spent;
        assert(leftover >= fundRequest.amount, ERR_ALLOWANCE_EXCEEDED);

        this.allowances(key).value = {
          ...this.allowances(key).value,
          spent: (spent + fundRequest.amount),
          last: newLast
        }
      }

    } else if (type === SpendAllowanceTypeDrip) {
      const epochRef = useRounds ? Global.round : Global.latestTimestamp;

      const amount = fundRequest.amount
      const accrualRate = allowed
      const lastLeftover = spent

      const passed: uint64 = epochRef - last
      const accrued: uint64 = lastLeftover + ((passed / interval) * accrualRate)

      const available: uint64 = accrued > max ? max : accrued

      assert(available >= amount, ERR_ALLOWANCE_EXCEEDED);

      this.allowances(key).value = {
        ...this.allowances(key).value,
        spent: (available - amount),
        last: newLast
      }
    }
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
  private getAuthAddr(): Account {
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
  arc58_pluginChangeAdmin(plugin: uint64, allowedCaller: Address, newAdmin: Address): void {
    assert(Txn.sender === Application(plugin).address, ERR_SENDER_MUST_BE_ADMIN_PLUGIN);
    assert(
      this.controlledAddress.value.authAddress === Application(plugin).address,
      'This plugin is not in control of the account'
    );

    const key = { application: plugin, allowedCaller: allowedCaller.native };

    assert(
      this.plugins(key).exists && this.plugins(key).value.admin.native,
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
  arc58_verifyAuthAddr(): void {
    assert(this.spendingAddress.value.authAddress === this.getAuthAddr());
    this.spendingAddress.value = Global.zeroAddress
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
        note: 'rekeying abstracted account',
        fee,
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
    method: bytes<4>
  ): boolean {
    if (global) {
      this.pluginCallAllowed(plugin, Global.zeroAddress, method);
    }
    return this.pluginCallAllowed(plugin, address.native, method);
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
    methodOffsets: uint64[],
    fundsRequest: FundsRequest[]
  ): void {
    const pluginApp = Application(plugin)
    const caller = global ? Global.zeroAddress : Txn.sender
    const key = { application: plugin, allowedCaller: caller }

    assert(this.plugins(key).exists, ERR_PLUGIN_DOES_NOT_EXIST);

    this.assertValidGroup(key, methodOffsets);

    if (this.plugins(key).value.escrow.native !== 0) {
      const spendingApp = Application(this.plugins(key).value.escrow.native)
      this.spendingAddress.value = spendingApp.address;
      this.transferFunds(key, fundsRequest);
    } else {
      this.spendingAddress.value = this.controlledAddress.value;
    }

    itxn
      .payment({
        sender: this.spendingAddress.value,
        receiver: this.spendingAddress.value,
        rekeyTo: pluginApp.address,
        note: 'rekeying to plugin app',
        fee,
      })
      .submit();

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
  arc58_rekeyToNamedPlugin(name: string, global: boolean, methodOffsets: uint64[], fundsRequest: FundsRequest[]): void {
    this.arc58_rekeyToPlugin(
      this.namedPlugins(name).value.application,
      global,
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
    app: uint64,
    allowedCaller: Address,
    admin: boolean,
    delegationType: UintN8,
    escrow: string,
    lastValid: uint64,
    cooldown: uint64,
    methods: MethodRestriction[],
    useRounds: boolean,
  ): void {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);
    const badDelegationCombo = (
      delegationType === DelegationTypeSelf &&
      allowedCaller.native === Global.zeroAddress
    )
    assert(!badDelegationCombo, ERR_ZERO_ADDRESS_DELEGATION_TYPE)
    const key: PluginKey = { application: app, allowedCaller: allowedCaller.native }

    let methodInfos = new DynamicArray<arc4MethodInfo>();
    for (let i: uint64 = 0; i < methods.length; i += 1) {
      methodInfos.push(
        new arc4MethodInfo({
          selector: methods[i].selector,
          cooldown: new UintN64(methods[i].cooldown),
          lastCalled: new UintN64(),
        })
      );
    }

    const epochRef = useRounds ? Global.round : Global.latestTimestamp;

    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          sender: this.controlledAddress.value,
          receiver: Global.currentApplicationAddress,
          amount: this.pluginsMbr(methodInfos.length),
          fee,
        })
        .submit()
    }

    const escrowID = this.maybeNewEscrow(escrow);

    this.plugins(key).value = new arc4PluginInfo({
      admin: new Bool(admin),
      delegationType,
      escrow: new UintN64(escrowID),
      lastValid: new UintN64(lastValid),
      cooldown: new UintN64(cooldown),
      methods: methodInfos.copy(),
      useRounds: new Bool(useRounds),
      lastCalled: new UintN64(0),
      start: new UintN64(epochRef),
    });

    this.updateLastUserInteraction();
    this.updateLastChange();
  }

  /**
   * Remove an app from the list of approved plugins
   *
   * @param app The app to remove
   * @param allowedCaller The address that's allowed to call the app
  */
  arc58_removePlugin(app: uint64, allowedCaller: Address): void {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);

    const key: PluginKey = { application: app, allowedCaller: allowedCaller.native };
    assert(this.plugins(key).exists, ERR_PLUGIN_DOES_NOT_EXIST);

    const methods = this.plugins(key).value.methods.copy();

    this.plugins(key).delete();

    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          receiver: this.controlledAddress.value,
          amount: this.pluginsMbr(methods.length),
          fee,
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
    app: uint64,
    allowedCaller: Address,
    admin: boolean,
    delegationType: UintN8,
    escrow: string,
    lastValid: uint64,
    cooldown: uint64,
    methods: MethodRestriction[],
    useRounds: boolean,
  ): void {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);
    assert(!this.namedPlugins(name).exists);

    const key: PluginKey = { application: app, allowedCaller: allowedCaller.native };
    this.namedPlugins(name).value = key;

    let methodInfos = new DynamicArray<arc4MethodInfo>()
    for (let i: uint64 = 0; i < methods.length; i += 1) {
      methodInfos.push(
        new arc4MethodInfo({
          selector: methods[i].selector,
          cooldown: new UintN64(methods[i].cooldown),
          lastCalled: new UintN64(),
        })
      )
    }

    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          sender: this.controlledAddress.value,
          receiver: Global.currentApplicationAddress,
          amount: this.pluginsMbr(methodInfos.length) + this.namedPluginsMbr(name),
          fee,
        })
        .submit()
    }

    const escrowID = this.maybeNewEscrow(escrow);

    const epochRef = useRounds ? Global.round : Global.latestTimestamp;

    this.plugins(key).value = new arc4PluginInfo({
      admin: new Bool(admin),
      delegationType,
      escrow: new UintN64(escrowID),
      lastValid: new UintN64(lastValid),
      cooldown: new UintN64(cooldown),
      methods: methodInfos.copy(),
      useRounds: new Bool(useRounds),
      lastCalled: new UintN64(0),
      start: new UintN64(epochRef)
    })

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
    const app = this.namedPlugins(name).value
    assert(this.plugins(app).exists, ERR_PLUGIN_DOES_NOT_EXIST);

    const methods = this.plugins(app).value.methods.copy();

    this.namedPlugins(name).delete();
    this.plugins(app).delete();

    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          receiver: this.controlledAddress.value,
          amount: this.namedPluginsMbr(name) + this.pluginsMbr(methods.length),
          fee,
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
  arc58_newEscrow(escrow: string): void {
    assert(Txn.sender === this.admin.value, ERR_ADMIN_ONLY);
    assert(!this.escrows(escrow).exists, ERR_ESCROW_ALREADY_EXISTS);
    this.newEscrow(escrow);
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
    const sender = Application(this.escrows(escrow).value).address

    for (let i: uint64 = 0; i < reclaims.length; i += 1) {
      if (reclaims[i].asset === 0) {
        const pmt = itxn.payment({
          sender,
          receiver: this.controlledAddress.value,
          amount: reclaims[i].amount,
          fee,
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
          xferAsset: reclaims[i].asset,
          fee,
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
    assert(this.escrows(escrow).exists, ERR_ESCROW_DOES_NOT_EXIST);
    const escrowAddress = Application(this.escrows(escrow).value).address

    itxn
      .payment({
        sender: this.controlledAddress.value,
        receiver: escrowAddress,
        amount: Global.assetOptInMinBalance * assets.length,
        fee,
      })
      .submit();

    for (let i: uint64 = 0; i < assets.length; i += 1) {
      assert(
        this.allowances({ escrow: this.escrows(escrow).value, asset: assets[i] }).exists,
        ERR_ALLOWANCE_DOES_NOT_EXIST
      );

      itxn
        .assetTransfer({
          sender: escrowAddress,
          assetReceiver: escrowAddress,
          assetAmount: 0,
          xferAsset: assets[i],
          fee,
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
    app: uint64,
    allowedCaller: Address,
    assets: uint64[],
    mbrPayment: gtxn.PaymentTxn
  ): void {
    const key: PluginKey = { application: app, allowedCaller: allowedCaller.native };

    assert(this.plugins(key).exists, ERR_PLUGIN_DOES_NOT_EXIST);
    const pluginInfo = decodeArc4<PluginInfo>(this.plugins(key).value.copy().bytes);
    assert(pluginInfo.escrow !== 0, ERR_NOT_USING_ESCROW);
    assert(
      Txn.sender === Application(app).address ||
      Txn.sender === allowedCaller.native ||
      allowedCaller.native === Global.zeroAddress,
      ERR_FORBIDDEN
    );

    const escrowAddress = Application(pluginInfo.escrow).address

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
        amount: Global.assetOptInMinBalance * assets.length,
        fee,
      })
      .submit();

    for (let i: uint64 = 0; i < assets.length; i += 1) {
      assert(
        this.allowances({ escrow: pluginInfo.escrow, asset: assets[i] }).exists,
        ERR_ALLOWANCE_DOES_NOT_EXIST
      );

      itxn
        .assetTransfer({
          sender: escrowAddress,
          assetReceiver: escrowAddress,
          assetAmount: 0,
          xferAsset: assets[i],
          fee,
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

    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          sender: this.controlledAddress.value,
          receiver: Global.currentApplicationAddress,
          amount: this.allowancesMbr() * allowances.length,
          fee,
        })
        .submit()
    }

    for (let i: uint64 = 0; i < allowances.length; i += 1) {
      const { asset, type, allowed, max, interval, useRounds } = allowances[i];
      const key: AllowanceKey = { escrow: this.escrows(escrow).value, asset }
      assert(!this.allowances(key).exists, ERR_ALLOWANCE_ALREADY_EXISTS);
      const start = useRounds ? Global.round : Global.latestTimestamp;

      this.allowances(key).value = {
        type,
        spent: 0,
        allowed,
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

    if (this.controlledAddress.value !== Global.currentApplicationAddress) {
      itxn
        .payment({
          receiver: this.controlledAddress.value,
          amount: this.allowancesMbr() * assets.length,
          fee,
        })
        .submit()
    }

    for (let i: uint64 = 0; i < assets.length; i += 1) {
      const key: AllowanceKey = {
        escrow: this.escrows(escrow).value,
        asset: assets[i]
      }
      assert(this.allowances(key).exists, ERR_ALLOWANCE_DOES_NOT_EXIST);
      this.allowances(key).delete();
    }

    this.updateLastUserInteraction();
    this.updateLastChange();
  }

  @abimethod({ readonly: true })
  mbr(
    methodCount: uint64,
    pluginName: string,
    escrowName: string
  ): AbstractAccountBoxMBRData {
    return {
      plugins: this.pluginsMbr(methodCount),
      namedPlugins: this.namedPluginsMbr(pluginName),
      escrows: this.escrowsMbr(escrowName),
      allowances: this.allowancesMbr(),
    }
  }
}
