import { AlgorandGenesisHash } from 'arc31'

import type { User } from '../types'

const users: User[] = [
  {
    authAcc: 'REPLACE_WITH_ALGORAND_USER_ADDRESS',
    challenge: '',
    chainId: AlgorandGenesisHash.TestNet
  }
]

export default users
