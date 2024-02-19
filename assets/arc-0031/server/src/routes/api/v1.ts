import algosdk from 'algosdk'
import type { AuthMessage } from 'arc31'
import { encodeAuthMessage } from 'arc31'
import type { Request, Response } from 'express'
import express from 'express'
import { v4 as uuidv4 } from 'uuid'

import { env } from '../../env'
import users from '../../mock/users.db'

const v1 = express.Router()

v1.post('/signin/request', (req: Request<never, { authAcc: string }, never, never>, res: Response) => {
  const { authAcc } = req.body
  if (!authAcc || !algosdk.isValidAddress(authAcc)) {
    return res.sendStatus(400)
  }

  const user = users.find(user => user.authAcc === authAcc)
  if (!user) {
    return res.sendStatus(401)
  }

  const nonce = uuidv4()
  const authMessage: AuthMessage = {
    domain: env.SERVICE_NAME,
    authAcc,
    nonce,
    desc: env.SERVICE_DESCRIPTION
  }
  user.nonce = nonce

  return res.send(encodeAuthMessage(authMessage))
})

v1.post('/signin/verify', async (req: Request<never, { authAcc: string, signedMessageBase64: string }, never, never>, res: Response) => {
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
        nonce: user.nonce,
        desc: env.SERVICE_DESCRIPTION
      }
      const encodedAuthMessageBytes = Buffer.from(encodeAuthMessage(authMessage), 'base64')
      const signedMessageBytes = Buffer.from(signedMessageBase64, 'base64')
      if (await algosdk.verifyBytes(encodedAuthMessageBytes, signedMessageBytes, authAcc)) {
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
