// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  telemetry: false,
  modules: ['@nuxt/ui'],
  modulesDir: ['../node_modules'],
  runtimeConfig: {
    apiUrl: process.env.NUXT_API_URL,
    public: {
      algorandChainId: process.env.NUXT_PUBLIC_ALGORAND_CHAIN_ID
    }
  },
  devtools: { enabled: true },
  devServer: {
    port: 3001
  },
  css: ['normalize.css'],
  vite: {
    define: {
      'global.WebSocket': 'globalThis.WebSocket'
    },
    esbuild: {
      legalComments: 'none'
    }
  }
})
