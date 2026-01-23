<script setup lang="ts">
import { usePeraWallet } from '@/utils/hooks/usePeraWallet'

if (import.meta.client) {
  const { reconnectWallet } = usePeraWallet()
  await reconnectWallet()
}

useHead({
  titleTemplate: '%s | ARC31',
  meta: [
    { name: 'viewport', content: 'width=device-width, initial-scale=1' },
    {
      hid: 'description',
      name: 'description',
      content:
        'ARC31 Reference Implementation'
    }
  ],
  link: [{ rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }]
})
</script>

<template>
  <Body>
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <UNotifications />
  </Body>
</template>
