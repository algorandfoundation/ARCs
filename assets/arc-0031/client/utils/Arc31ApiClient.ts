import { ApiClient, type ApiClientOptions } from './ApiClient'
import { type Session } from '~~/types'

class Arc31ApiClient extends ApiClient {
  constructor(baseURL = '/api', options: Partial<ApiClientOptions> = { version: 'v1' }) {
    super(baseURL, options)
  }

  request = (authAcc: string) => this.post<string>('/signin/request', { authAcc })

  verify = (signedMessageBase64: string, authAcc: string) =>
    this.post<Session>('/signin/verify', { signedMessageBase64, authAcc })
}

export default Arc31ApiClient
