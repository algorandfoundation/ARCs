const NOTIFICATIONS = {
  SIGN_IN_ABORT: { type: 'warning' as const, title: 'Sign In aborted' },
  SIGN_IN_COMPLETE: { type: 'success' as const, title: 'Successfully Signed In' },
  SIGN_OUT_COMPLETE: { type: 'success' as const, title: 'Successfully Signed Out' }
}

export const useNotifications = (timeout = 3000) => {
  const toast = useToast()

  const showNotification = (notificationType: keyof typeof NOTIFICATIONS) => {
    const notification = NOTIFICATIONS[notificationType]
    toast.add({ ...notification, timeout })
  }

  return {
    showNotification
  }
}
