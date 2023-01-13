<script setup lang="ts">
import { LogInOutline as SignInIcon } from '@vicons/ionicons5'
import { NButton, NModal, NSpace, NText, useMessage } from 'naive-ui'
import { h, toRef, watch } from 'vue'
import { useAuthStore } from '~~/store/auth'

const authStore = useAuthStore()
const { signIn, signInAbort, signInConfirm } = authStore
const message = useMessage()
watch(toRef(authStore, 'notification'), notification => {
  if (notification) {
    message[notification.type](notification.content)
    authStore.$patch({ notification: null })
  }
})
</script>

<template>
  <NSpace align="center" justify="center">
    <NButton type="success" :render-icon="() => h(SignInIcon)" icon-placement="right" @click="signIn"
      >Sign In w/ MyAlgoConnect</NButton
    >
    <NModal
      :show="!!authStore.authMessage"
      preset="dialog"
      title="Confirm"
      negative-text="Abort"
      positive-text="Confirm"
      :show-icon="false"
      @close="signInAbort"
      @negative-click="signInAbort"
      @positive-click="signInConfirm"
    >
      <NText
        >You are going to sign in to <NText type="success" strong>{{ authStore.authMessage?.domain }}</NText> as
        <NText type="success" strong>{{ authStore.authMessage?.authAcc }}</NText>
      </NText>
    </NModal>
  </NSpace>
</template>
