const NOTIFICATIONS = {
  SIGN_IN_ABORT: { type: 'warning' as const, content: 'Sign In aborted' },
  SIGN_IN_COMPLETE: { type: 'success' as const, content: 'Successfully Signed In' },
  SIGN_OUT_COMPLETE: { type: 'success' as const, content: 'Successfully Signed Out' }
}

export { NOTIFICATIONS }
