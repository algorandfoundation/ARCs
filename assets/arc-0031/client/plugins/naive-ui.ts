import { setup } from '@css-render/vue3-ssr'
import { defineNuxtPlugin } from '#app'

export default defineNuxtPlugin(nuxtApp => {
  if (process.server) {
    const { collect } = setup(nuxtApp.vueApp)
    const originalRenderMeta = nuxtApp.ssrContext?.renderMeta
    if (nuxtApp?.ssrContext) {
      nuxtApp.ssrContext = nuxtApp.ssrContext || {}
      nuxtApp.ssrContext.renderMeta = () => {
        if (!originalRenderMeta) {
          return {
            headTags: collect()
          }
        }
        const originalMeta = originalRenderMeta()
        if ('then' in originalMeta) {
          return originalMeta.then(resolvedOriginalMeta => {
            return {
              ...resolvedOriginalMeta,
              headTags: resolvedOriginalMeta.headTags + collect()
            }
          })
        } else {
          return {
            ...originalMeta,
            headTags: originalMeta.headTags + collect()
          }
        }
      }
    }
  }
})
