import { ss } from '@/utils/storage'

const LOCAL_NAME = 'SECRET_TOKEN'

export function getToken() {
  return ss.get(LOCAL_NAME)
}

export function setToken(token: string) {
  return ss.set(LOCAL_NAME, token)
}

export function removeToken() {
  return ss.remove(LOCAL_NAME)
}

const EXPIRE_LOCAL_NAME = 'expiresIn'

export function getExpiresIn() {
  return ss.get(EXPIRE_LOCAL_NAME)
}

export function setExpiresIn(expiresIn: number) {
  return ss.set(EXPIRE_LOCAL_NAME, expiresIn)
}

export function removeExpiresIn() {
  return ss.remove(EXPIRE_LOCAL_NAME)
}
