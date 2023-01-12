<script setup lang="ts">
import { LogOutOutline as SignOutIcon } from '@vicons/ionicons5'
import { h, toRef, watch } from 'vue'
import { NButton, NSpace, NText, useMessage } from 'naive-ui'
import { useAuthStore } from '~~/store/auth'

const authStore = useAuthStore()
const { signOut } = authStore
const message = useMessage()
watch(toRef(authStore, 'notification'), notification => {
  if (notification) {
    message[notification.type](notification.content)
    authStore.$patch({ notification: null })
  }
})
</script>

<template>
  <NSpace vertical align="center" justify="center">
    <NText>www.servicedomain.com</NText>
    <NText tag="p" style="margin: 0 0 24px 0" type="success" strong>{{ authStore.session?.authAcc }}</NText>
    <NButton type="success" :render-icon="() => h(SignOutIcon)" icon-placement="right" @click="signOut"
      >Sign Out</NButton
    >
  </NSpace>
</template>
