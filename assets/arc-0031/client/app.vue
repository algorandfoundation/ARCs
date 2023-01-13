<script lang="ts" setup>
import { useDark } from '@vueuse/core'
import { darkTheme, type GlobalTheme, NConfigProvider, NGlobalStyle, NLayout, NMessageProvider } from 'naive-ui'
import { ref, onMounted } from 'vue'

useHead({
  title: 'Index',
  titleTemplate: '%s - arc-0031',
  meta: [
    { name: 'viewport', content: 'width=device-width, initial-scale=1' },
    {
      hid: 'description',
      name: 'description',
      content: 'Nuxt'
    }
  ],
  link: [{ rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }]
})

const theme = ref<GlobalTheme | null>(null)
onMounted(() => {
  const isDark = useDark()
  if (isDark.value) {
    theme.value = darkTheme
  }
})
</script>

<template>
  <Html>
    <Body>
      <NConfigProvider :theme="theme">
        <NMessageProvider>
          <NLayout>
            <NuxtLoadingIndicator :height="5" :duration="3000" />
            <NuxtPage style="width: 100vw; height: 100vh" />
          </NLayout>
        </NMessageProvider>
        <NGlobalStyle />
      </NConfigProvider>
    </Body>
  </Html>
</template>
