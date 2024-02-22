export interface Session {
  authAcc: string
  accessToken: string
}

export const useSessionCookie = () => {
  const session = useCookie<Session | null>('session', { sameSite: 'strict', maxAge: 60 * 60 * 24 * 365, secure: true })

  const clear = () => {
    session.value = null
  }

  const set = (value: Session) => {
    session.value = value
  }

  return {
    session,
    clear,
    set
  }
}
