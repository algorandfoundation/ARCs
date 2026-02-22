<script setup lang="ts">
import { useAuth } from '@/utils/hooks/useAuth'

useHead({ title: 'Home' })

const { address, signOut } = useAuth()
</script>

<template>
  <div class="flex content-center p-4 justify-center h-screen">
    <div class="self-center space-y-4 text-center">
      <div>
        <p>www.servicedomain.com</p>
        <p class="text-primary break-all">
          {{ address }}
        </p>
      </div>
      <UButton
        icon="i-heroicons-arrow-left-on-rectangle"
        trailing
        @click="signOut"
      >
        Sign Out
      </UButton>
    </div>
  </div>
</template>
