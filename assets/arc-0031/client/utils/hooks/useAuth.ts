import type { AuthMessage } from 'arc31'
import { encodeAuthMessage, decodeAuthMessage } from 'arc31'

import { useNotifications } from './useNotifications'
import { usePeraWallet } from './usePeraWallet'

export interface Session {
  authAcc: string
  accessToken: string
}

export const useAuth = () => {
  const router = useRouter()

  const { address, connectWallet, disconnectWallet, signData } = usePeraWallet()
  const { showErrorNotification, showWarningNotification, showSuccessNotification } = useNotifications()

  const sessionCookie = useCookie('session', { sameSite: 'strict', maxAge: 60 * 60 * 24 * 365, secure: true })

  const authMessage = ref<AuthMessage | null>(null)
  const isConfirmingSignIn = ref(false)

  const clear = () => {
    disconnectWallet()
    sessionCookie.value = null
    authMessage.value = null
    isConfirmingSignIn.value = false
  }

  const signInAbort = () => {
    clear()
    showWarningNotification('Sign in aborted')
  }

  const signInError = (error: unknown) => {
    clear()
    showErrorNotification(normalizeError(error))
  }

  const signIn = async () => {
    try {
      await connectWallet()
      const message = await new Arc31ApiClient().request(address.value)
      authMessage.value = decodeAuthMessage(message)
    } catch (error) {
      if ((error as any).data.type === 'CONNECT_MODAL_CLOSED') {
        signInAbort()
      } else {
        signInError(error)
      }
    }
  }

  const signInConfirm = async () => {
    try {
      isConfirmingSignIn.value = true
      if (!authMessage.value) {
        throw new Error('authMessage is null')
      }
      const encodedAuthMessage = encodeAuthMessage(authMessage.value)
      // Convert the base64 message to bytes
      const encodedAuthMessageBytes = Buffer.from(encodedAuthMessage, 'base64')
      const signMessage = `You are going to login into ${authMessage.value.domain}. Please confirm that you are the owner of this wallet by signing this message.`
      // Sign the message
      const signedMessageBytes = await signData([{ data: encodedAuthMessageBytes, message: signMessage }], authMessage.value.authAcc)
      // Convert the signed message to base64
      const signedMessageBase64 = Buffer.from(signedMessageBytes[0]).toString('base64')
      // Verify the signed message and update the session cookie
      const session = await new Arc31ApiClient().verify(signedMessageBase64, authMessage.value.authAcc)
      sessionCookie.value = JSON.stringify(session)
      authMessage.value = null
      router.push('/')
      showSuccessNotification('Sign in completed')
    } catch (error) {
      if ((error as any).data.type === 'SIGN_TRANSACTIONS') {
        signInAbort()
      } else {
        signInError(error)
      }
    }
  }

  const signOut = () => {
    router.push('/signin')
    clear()
    showSuccessNotification('Sign out completed')
  }

  return {
    address,
    authMessage,
    isConfirmingSignIn,
    signInAbort,
    signIn,
    signInConfirm,
    signOut
  }
}
