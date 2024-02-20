import type { AlgorandChainId } from 'arc31'
import { AlgorandGenesisHash } from 'arc31'
import type { ZodObject, ZodType } from 'zod'
import { z } from 'zod'

const createEnv = <TClient extends Record<string, ZodType>, TServer extends Record<string, ZodType>>(runtimeEnv: {client: TClient, server: TServer}) => {
  const runtimeConfig = useRuntimeConfig()
  if (import.meta.client) {
    z.object(runtimeEnv.client).parse(runtimeConfig.public)
  } else {
    z.object(runtimeEnv.client).merge(z.object(runtimeEnv.server)).parse({ ...runtimeConfig, ...runtimeConfig.public })
  }

  return {
    client: Object.keys(runtimeEnv.client).reduce((client, key) => ({ ...client, [key]: (runtimeConfig.public as Record<string, any>)[key] }), {}) as z.infer<ZodObject<TClient>>,
    server: Object.keys(runtimeEnv.server).reduce((server, key) => ({ ...server, [key]: (runtimeConfig as Record<string, any>)[key] }), {}) as z.infer<ZodObject<TServer>>
  }
}

export const useEnv = () => createEnv({
  client: {
    algorandChainId: (() => {
      const [firstAlgorandChainId, ...otherAlgorandChainIds] = Object.keys(AlgorandGenesisHash) as AlgorandChainId[]

      return z.enum([firstAlgorandChainId, ...otherAlgorandChainIds])
    })()
  },
  server: {
    apiUrl: z.string().url()
  }
})
