import { encode, decode } from '@msgpack/msgpack'
import MyAlgoConnect from '@randlabs/myalgo-connect'
import { Buffer } from 'buffer/'
import type { AuthMessage, AuthStore } from '../../types'
import { NOTIFICATIONS } from './notifications'
import Arc31ApiClient from '~~/utils/Arc31ApiClient'
import { getCookie, deleteCookie } from '~~/utils/cookie'

export const useAuthStore = defineStore('auth', {
  state: (): AuthStore => {
    return {
      notification: null,
      message: '',
      authMessage: null,
      session: null
    }
  },
  actions: {
    loadSession() {
      const session = getCookie('session')
      if (session) {
        this.$patch({ session })
      }
    },
    async signIn() {
      try {
        const connector = new MyAlgoConnect()
        const accounts = await connector.connect()
        const authAcc = accounts[0].address
        const message = await new Arc31ApiClient().request(authAcc)
        if (!message) {
          throw new Error('message is empty')
        }
        const {
          'arc31:j': { ...authMessage }
        } = decode(Buffer.from(message.replace(/^AX/, ''), 'base64')) as { 'arc31:j': AuthMessage }
        this.$patch({ authMessage })
      } catch (error) {
        this.signInAbort()
      }
    },
    signInAbort() {
      this.$patch({ notification: NOTIFICATIONS.SIGN_IN_ABORT, authMessage: null })
    },
    async signInConfirm() {
      try {
        if (!this.authMessage) {
          throw new Error('authMessage is null')
        }
        const connector = new MyAlgoConnect()
        const message = `AX${Buffer.from(encode({ 'arc31:j': this.authMessage })).toString('base64')}`
        const messageBytes = Buffer.from(message, 'base64')
        const signedMessageBytes = await connector.signBytes(messageBytes, this.authMessage.authAcc)
        const signedMessageBase64 = Buffer.from(signedMessageBytes).toString('base64')
        const session = await new Arc31ApiClient().verify(signedMessageBase64, this.authMessage.authAcc)
        this.$patch({ notification: NOTIFICATIONS.SIGN_IN_COMPLETE, authMessage: null, session })
        const router = useRouter()
        router.push('/')
      } catch (error) {
        this.signInAbort()
      }
    },
    signOut() {
      deleteCookie('session', { path: '/', sameSite: 'lax' })
      this.$patch({ notification: NOTIFICATIONS.SIGN_OUT_COMPLETE, session: null })
      const router = useRouter()
      router.push('/signin')
    }
  }
})
