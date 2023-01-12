import { CookieOptions } from 'nuxt/dist/app/composables'

export const setCookie = (key: string, value: any, options?: CookieOptions) => {
  const cookie = useCookie(key, options)
  cookie.value = value
}

export const getCookie = (key: string): any => {
  const cookie = useCookie(key)
  return cookie.value
}

export const deleteCookie = (key: string, options?: CookieOptions) => {
  setCookie(key, null, options)
}
