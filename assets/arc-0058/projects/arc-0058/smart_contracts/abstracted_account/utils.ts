import { Uint8 } from "@algorandfoundation/algorand-typescript/arc4";
import { AllowanceInfo, EscrowInfo, ExecutionInfo, PluginInfo } from "./types";

export function emptyPluginInfo(): PluginInfo {
  return {
    escrow: 0,
    delegationType: new Uint8(0),
    lastValid: 0,
    cooldown: 0,
    methods: [],
    admin: false,
    useRounds: false,
    useExecutionKey: false,
    lastCalled: 0,
    start: 0,
  };
}

export function emptyEscrowInfo(): EscrowInfo {
  return {
    id: 0,
    locked: false
  };
}

export function emptyAllowanceInfo(): AllowanceInfo {
  return {
    type: new Uint8(0),
    max: 0,
    amount: 0,
    spent: 0,
    interval: 0,
    last: 0,
    start: 0,
    useRounds: false
  };
}

export function emptyExecutionInfo(): ExecutionInfo {
  return {
    groups: [],
    firstValid: 0,
    lastValid: 0
  };
}