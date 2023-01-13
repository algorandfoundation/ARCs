import { setCookie } from 'h3'
import Arc31ApiClient from '~~/utils/Arc31ApiClient'

export default defineEventHandler(async event => {
  const { signedMessageBase64, authAcc } = await readBody(event)
  const session = await new Arc31ApiClient(process.env.NUXT_APP_API_URL || 'http://localhost:3001/api').verify(
    signedMessageBase64,
    authAcc
  )
  setCookie(event, 'session', JSON.stringify(session), { path: '/', sameSite: 'lax' })

  return session
})
