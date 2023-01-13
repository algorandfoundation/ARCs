interface ApiClientOptions {
  version?: string
}

class ApiClient {
  baseURL: string
  options: ApiClientOptions = {}

  constructor(baseURL = '/api', options?: Partial<ApiClientOptions>) {
    this.options = { ...this.options, ...options }
    this.baseURL = baseURL
    if (this.options.version) {
      this.baseURL += `/${this.options.version}`
    }
  }

  post = <T>(path: string, body = {}) => $fetch<T>(`${this.baseURL}${path}`, { method: 'POST', body })
}

export { ApiClientOptions, ApiClient }
