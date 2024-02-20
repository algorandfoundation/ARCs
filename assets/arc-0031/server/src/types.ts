import type { AlgorandChainId } from 'arc31'

export interface User {
  authAcc: string
  challenge: string
  chainId: AlgorandChainId
}
