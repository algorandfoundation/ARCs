import type { AlgorandGenesisHash } from 'arc31'

export interface User {
  authAcc: string
  challenge: string
  chainId: AlgorandGenesisHash
}
