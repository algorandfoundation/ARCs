import algosdk from 'algosdk'

import type { AuthMessage } from './types'

export const prefix = 'arc0031'

export const encodeAuthMessage = (authMessage: AuthMessage): string =>
  `${prefix}${JSON.stringify(authMessage)}`

export const decodeAuthMessage = (encodedAuthMessage: string): AuthMessage => {
  if (!(encodedAuthMessage.substring(0, prefix.length) === prefix)) {
    throw new Error('Invalid prefix')
  }

  return JSON.parse(encodedAuthMessage.slice(prefix.length))
}

export const verifyAuthMessage = (authMessage: AuthMessage, signedAuthMessage: string): boolean =>
  algosdk.verifyBytes(
    Buffer.from(encodeAuthMessage(authMessage), 'utf-8'),
    Buffer.from(signedAuthMessage, 'base64'),
    authMessage.authAcc)

export * from './types'
