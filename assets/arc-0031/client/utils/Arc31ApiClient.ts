import type { Session } from '@/utils/hooks/useAuth'

export class Arc31ApiClient {
  private apiFetch

  constructor (baseURL: string = '/api/v1') {
    this.apiFetch = $fetch.create({ baseURL })
  }

  public request = (authAcc: string) => this.apiFetch<string>('/signin/request', {
    method: 'POST',
    body: {
      authAcc
    }
  })

  public verify = (signedMessageBase64: string, authAcc: string) =>
    this.apiFetch<Session>('/signin/verify', {
      method: 'POST',
      body: {
        signedMessageBase64,
        authAcc
      }
    })
}

export { ApiClientOptions, ApiClient }
