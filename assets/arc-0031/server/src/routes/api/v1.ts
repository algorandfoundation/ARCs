import algosdk from 'algosdk'
import type { AlgorandChainId, AuthMessage } from 'arc31'
import { AlgorandGenesisHash, encodeAuthMessage, verifyAuthMessage } from 'arc31'
import type { Request, Response } from 'express'
import express from 'express'
import { v4 as uuidv4 } from 'uuid'

import { env } from '../../env'
import users from '../../mock/users.db'

const v1 = express.Router()

v1.post('/signin/request', (req: Request<never, { authAcc: string, chainId: AlgorandChainId }, never, never>, res: Response) => {
  const { authAcc, chainId } = req.body
  if (!authAcc || !algosdk.isValidAddress(authAcc) || !chainId || !Object.keys(AlgorandGenesisHash).includes(chainId)) {
    return res.sendStatus(400)
  }

  const user = users.find(user => user.authAcc === authAcc)
  if (!user) {
    return res.sendStatus(401)
  }

  const challenge = uuidv4()
  const authMessage: AuthMessage = {
    domain: env.SERVICE_NAME,
    authAcc,
    challenge,
    chainId: AlgorandGenesisHash[chainId],
    desc: env.SERVICE_DESCRIPTION
  }
  user.challenge = challenge
  user.chainId = AlgorandGenesisHash[chainId]

  return res.send(encodeAuthMessage(authMessage))
})

v1.post('/signin/verify', (req: Request<never, { authAcc: string, signedMessageBase64: string }, never, never>, res: Response) => {
  const {
    authAcc,
    signedMessageBase64
  } = req.body
  try {
    if (!authAcc || !algosdk.isValidAddress(authAcc) || !signedMessageBase64) {
      return res.sendStatus(400)
    }
    const user = users.find(user => user.authAcc === authAcc)
    if (user) {
      const authMessage: AuthMessage = {
        domain: env.SERVICE_NAME,
        authAcc,
        challenge: user.challenge,
        chainId: user.chainId,
        desc: env.SERVICE_DESCRIPTION
      }
      if (verifyAuthMessage(authMessage, signedMessageBase64)) {
        const session = { authAcc, accessToken: uuidv4() }

        return res.status(200).send(session)
      }
    }
  } catch (error) {
    return res.sendStatus(400)
  }

  return res.sendStatus(401)
})

export default v1
