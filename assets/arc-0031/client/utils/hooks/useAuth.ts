import type { AuthMessage } from 'arc31'
import { encodeAuthMessage, decodeAuthMessage } from 'arc31'

import { useEnv } from './useEnv'
import { useNotifications } from './useNotifications'
import { usePeraWallet } from './usePeraWallet'
import { useSessionCookie } from './useSessionCookie'

export const useAuth = () => {
  const router = useRouter()

  const env = useEnv()

  const { address, connectWallet, disconnectWallet, signData } = usePeraWallet()
  const { showErrorNotification, showWarningNotification, showSuccessNotification } = useNotifications()

  const sessionCookie = useSessionCookie()

  const authMessage = ref<AuthMessage | null>(null)
  const isConfirmingSignIn = ref(false)

  const clear = () => {
    disconnectWallet()
    sessionCookie.clear()
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
      const message = await new Arc31ApiClient().request(address.value, env.client.algorandChainId)
      authMessage.value = decodeAuthMessage(message)
    } catch (error) {
      if ((error as any).data?.type === 'CONNECT_MODAL_CLOSED') {
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
      // Encode the authMessage
      const message = encodeAuthMessage(authMessage.value)
      // Convert the message string to bytes
      const messageBytes = Buffer.from(message, 'utf-8')
      const signMessage = `You are going to login into ${authMessage.value.domain}. Please confirm that you are the owner of this wallet by signing this message.`
      // Sign the message
      const signedMessageBytes = await signData([{ data: messageBytes, message: signMessage }], authMessage.value.authAcc)
      // Convert the signed message to string
      const signedMessageBase64 = Buffer.from(signedMessageBytes[0]).toString('base64')
      // Verify the signed message and update the session cookie
      const session = await new Arc31ApiClient().verify(signedMessageBase64, authMessage.value.authAcc)
      sessionCookie.set(session)
      authMessage.value = null
      router.push('/')
      showSuccessNotification('Sign in completed')
    } catch (error) {
      if ((error as any).data?.type === 'SIGN_TRANSACTIONS') {
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
    session: sessionCookie.session,
    authMessage,
    isConfirmingSignIn,
    signInAbort,
    signIn,
    signInConfirm,
    signOut
  }
}
