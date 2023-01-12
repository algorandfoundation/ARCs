import Arc31ApiClient from '~~/utils/Arc31ApiClient'

export default defineEventHandler(async event => {
  const { authAcc } = await readBody(event)
  return new Arc31ApiClient(process.env.NUXT_APP_API_URL || 'http://localhost:3001/api').request(authAcc)
})
