import { uint64 } from "@algorandfoundation/algorand-typescript"

export const AbstractAccountGlobalStateKeysAdmin = 'admin'
export const AbstractAccountGlobalStateKeysControlledAddress = 'controlled_address'
export const AbstractAccountGlobalStateKeysLastUserInteraction = 'last_user_interaction'
export const AbstractAccountGlobalStateKeysLastChange = 'last_change'
export const AbstractAccountGlobalStateKeysEscrowFactory = 'escrow_factory'
export const AbstractAccountGlobalStateKeysSpendingAddress = 'spending_address'

export const AbstractAccountBoxPrefixPlugins = 'p'
export const AbstractAccountBoxPrefixNamedPlugins = 'n'
export const AbstractAccountBoxPrefixEscrows = 'e'
export const AbstractAccountBoxPrefixAllowances = 'a'

export const BoxCostPerBox: uint64 = 2_500
export const BoxCostPerByte: uint64 = 400

export const MethodRestrictionByteLength: uint64 = 20
export const DynamicOffset: uint64 = 2
export const DynamicLength: uint64 = 2

export const MinPluginMBR: uint64 = 36_100
export const MinNamedPluginMBR: uint64 = 18_900
export const MinEscrowsMBR: uint64 = 6_100
export const AllowanceMBR: uint64 = 29_300