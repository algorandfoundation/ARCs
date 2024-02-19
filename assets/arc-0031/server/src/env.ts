import { createEnv } from '@t3-oss/env-core'
import dotenv from 'dotenv'
import { z } from 'zod'

dotenv.config()

export const env = createEnv({
  server: {
    SERVICE_NAME: z.string(),
    SERVICE_DESCRIPTION: z.string(),
    PORT: z.string().optional()
  },
  runtimeEnv: process.env,
  isServer: true
})
