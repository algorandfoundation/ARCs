const NOTIFICATIONS = {
  ERROR: {
    icon: 'i-heroicons-x-circle',
    color: 'red' as const
  },
  WARNING: {
    icon: 'i-heroicons-exclamation-circle',
    color: 'orange' as const
  },
  SUCCESS: {
    icon: 'i-heroicons-check-circle',
    color: 'primary' as const
  }
}

export const useNotifications = (timeout = 3000) => {
  const toast = useToast()

  const showErrorNotification = (error: Error) => {
    toast.add({ ...NOTIFICATIONS.ERROR, title: 'Error', description: (error as Error)?.message || undefined, timeout })
  }

  const showWarningNotification = (title: string, description?: string) => {
    toast.add({ ...NOTIFICATIONS.WARNING, title, description, timeout })
  }

  const showSuccessNotification = (title: string, description?: string) => {
    toast.add({ ...NOTIFICATIONS.SUCCESS, title, description, timeout })
  }

  return {
    showErrorNotification,
    showWarningNotification,
    showSuccessNotification
  }
}
