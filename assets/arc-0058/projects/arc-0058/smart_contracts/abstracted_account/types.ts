import { Account, uint64 } from "@algorandfoundation/algorand-typescript";
import { Address, Bool, DynamicArray, StaticBytes, Struct, UintN64, UintN8 } from "@algorandfoundation/algorand-typescript/arc4";

export type AllowanceKey = {
  escrow: uint64;
  asset: uint64;
}

export type SpendAllowanceType = UintN8

export const SpendAllowanceTypeFlat: SpendAllowanceType = new UintN8(1)
export const SpendAllowanceTypeWindow: SpendAllowanceType = new UintN8(2)
export const SpendAllowanceTypeDrip: SpendAllowanceType = new UintN8(3)

export type AllowanceInfo = {
  /** the type of allowance to use */
  type: SpendAllowanceType
  /** the maximum size of the bucket if using drip */
  max: uint64
  /** the amount of the asset the plugin is allowed to access or per window */
  allowed: uint64
  /** the amount spent during the current or last interacted window */
  spent: uint64
  /** the rate the allowance should be expanded */
  interval: uint64
  /** the amount leftover when the bucket was last accessed */
  last: uint64
  /** the timestamp or round the allowance was added */
  start: uint64
  /** whether to use round number or unix timestamp when evaluating this allowance */
  useRounds: boolean
}

export type AddAllowanceInfo = {
  asset: uint64;
  type: SpendAllowanceType;
  allowed: uint64;
  max: uint64;
  interval: uint64;
  useRounds: boolean;
}

export type FundsRequest = {
  asset: uint64;
  amount: uint64;
}

export type PluginKey = {
  application: uint64;
  allowedCaller: Account;
}

export const DelegationTypeSelf = new UintN8(1)
export const DelegationTypeAgent = new UintN8(2)
export const DelegationTypeOther = new UintN8(3)

export class arc4PluginInfo extends Struct<{
  /** Whether the plugin has permissions to change the admin account */
  admin: Bool;
  /** the type of delegation the plugin is using */
  delegationType: UintN8;
  /** the escrow account to use for the plugin */
  escrow: UintN64;
  /** The last round or unix time at which this plugin can be called */
  lastValid: UintN64;
  /** The number of rounds or seconds that must pass before the plugin can be called again */
  cooldown: UintN64;
  /** The methods that are allowed to be called for the plugin by the address */
  methods: DynamicArray<arc4MethodInfo>;
  /** Whether to use unix timestamps or round for lastValid and cooldown */
  useRounds: Bool;
  /** The last round or unix time the plugin was called */
  lastCalled: UintN64;
  /** The round or unix time the plugin was installed */
  start: UintN64;
}> { }

export type PluginInfo = {
  admin: boolean;
  delegationType: UintN8;
  escrow: uint64;
  lastValid: uint64;
  cooldown: uint64;
  methods: DynamicArray<arc4MethodInfo>;
  useRounds: boolean;
  lastCalled: uint64;
  start: uint64;
}

export class arc4MethodRestriction extends Struct<{
  /** The method signature */
  selector: StaticBytes<4>;
  /** The number of rounds that must pass before the method can be called again */
  cooldown: UintN64;
}> { }

export type MethodRestriction = {
  selector: StaticBytes<4>;
  cooldown: uint64;
}

export class arc4MethodInfo extends Struct<{
  /** The method signature */
  selector: StaticBytes<4>;
  /** The number of rounds that must pass before the method can be called again */
  cooldown: UintN64;
  /** The last round the method was called */
  lastCalled: UintN64;
}> { }

export type MethodInfo = {
  selector: StaticBytes<4>;
  cooldown: uint64;
  lastCalled: uint64;
}

export type PluginValidation = {
  exists: boolean;
  expired: boolean;
  hasCooldown: boolean;
  onCooldown: boolean;
  hasMethodRestrictions: boolean;
  valid: boolean;
}

export type MethodValidation = {
  methodAllowed: boolean;
  methodHasCooldown: boolean;
  methodOnCooldown: boolean;
}

export type FullPluginValidation = {
  exists: boolean;
  expired: boolean;
  hasCooldown: boolean;
  onCooldown: boolean;
  hasMethodRestrictions: boolean;
  methodAllowed: boolean;
  methodHasCooldown: boolean;
  methodOnCooldown: boolean;
  valid: boolean;
}

export type EscrowReclaim = {
  asset: uint64;
  amount: uint64;
  closeOut: boolean;
}

export type AbstractAccountBoxMBRData = {
  plugins: uint64;
  namedPlugins: uint64;
  escrows: uint64;
  allowances: uint64;
}