/* eslint-disable @typescript-eslint/no-shadow */
import { encode } from '@msgpack/msgpack'
import algosdk from 'algosdk'
import express, { Request, Response } from 'express'
import { v4 as uuidv4 } from 'uuid'
import users from '../../mock/users.json'
import { type AuthMessage } from '../../types'

const v1 = express.Router()

v1.post('/signin/request', (req: Request, res: Response) => {
  const { authAcc }: { authAcc: string } = req.body
  if (!authAcc) {
    return res.sendStatus(400)
  }

  const user = users.find(user => user.authAcc === authAcc)
  if (!user) {
    return res.sendStatus(401)
  }

  const nonce = uuidv4()
  const authMessage: AuthMessage = {
    domain: process.env.SERVICE_NAME,
    authAcc,
    nonce,
    desc: process.env.SERVICE_DESCRIPTION
  }
  user.nonce = nonce
  const message = `AX${Buffer.from(encode({ 'arc31:j': authMessage })).toString('base64')}`

  return res.send(message)
})

v1.post('/signin/verify', async (req: Request, res: Response) => {
  const {
    signedMessageBase64,
    authAcc
  }: {
    signedMessageBase64: string
    authAcc: string
  } = req.body
  try {
    const user = users.find(user => user.authAcc === authAcc)
    if (user) {
      const authMessage: AuthMessage = {
        domain: process.env.SERVICE_NAME,
        authAcc,
        nonce: user.nonce,
        desc: process.env.SERVICE_DESCRIPTION
      }
      const message = `AX${Buffer.from(encode({ 'arc31:j': authMessage })).toString('base64')}`
      const messageBytes = Buffer.from(message, 'base64')
      const signedMessageBytes = Buffer.from(signedMessageBase64, 'base64')
      if (await algosdk.verifyBytes(messageBytes, signedMessageBytes, authAcc)) {
        const session = { authAcc, accessToken: uuidv4() }

        return res.status(200).send(session)
      }
    }
  } catch (error) {
    console.error(error)
    return res.sendStatus(400)
  }

  return res.sendStatus(401)
})

export default v1
