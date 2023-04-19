import jwt_decode from 'jwt-decode'

export function isAdmin(token: string): boolean {
  if (token) {
    const decoded: { role: string } = jwt_decode(token)
    if (decoded && decoded.role === 'admin')
      return true
  }
  return false
}
