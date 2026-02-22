import { useEnv } from '@/utils/hooks/useEnv'

enum AlgorandChainId {
  MainNet = 416001,
  TestNet = 416002,
  BetaNet = 416003
}

export default defineNuxtPlugin(async () => {
  const { PeraWalletConnect } = await import('@perawallet/connect')
  const env = useEnv()
  const peraWalletClient = new PeraWalletConnect({ chainId: AlgorandChainId[env.client.algorandChainId] })

  return {
    provide: {
      peraWalletClient
    }
  }
})
