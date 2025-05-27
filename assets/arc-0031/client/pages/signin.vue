<script setup lang="ts">
import { useAuth } from '@/utils/hooks/useAuth'

useHead({ title: 'Sign In' })

const { authMessage, isConfirmingSignIn, signIn, signInAbort, signInConfirm } = useAuth()

</script>

<template>
  <div class="flex content-center justify-center p-4 h-screen">
    <UButton
      class="self-center"
      icon="i-heroicons-arrow-right-on-rectangle"
      trailing
      @click="signIn"
    >
      Sign In w/ Pera
    </UButton>
    <UModal
      v-if="!!authMessage"
      :model-value="!!authMessage"
      :ui="{ container: 'items-center' }"
      :overlay="false"
      @close="signInAbort"
    >
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3>
              Sign In Confirmation
            </h3>
            <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark" @click="signInAbort" />
          </div>
        </template>
        <div class="m-0 p-4">
          <p class="break-all">
            You are going to sign in to
            <span class="text-primary">{{ authMessage.domain }}</span>
            as
            <span class="text-primary">{{ authMessage.authAcc }}</span>
          </p>
        </div>
        <template #footer>
          <div class="flex justify-end space-x-4 m-0 px-4">
            <UButton variant="outline" :disabled="isConfirmingSignIn" @click="signInAbort">
              Abort
            </UButton>
            <UButton :loading="isConfirmingSignIn" @click="signInConfirm">
              Confirm
            </UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </div>
</template>
