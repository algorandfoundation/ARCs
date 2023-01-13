declare global {
  namespace NodeJS {
    interface ProcessEnv {
      SERVICE_NAME: string
      SERVICE_DESC?: string
      PORT?: string
    }
  }
}

export {}
