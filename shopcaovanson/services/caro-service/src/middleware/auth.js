import jwt from 'jsonwebtoken'
import { config } from '../config.js'

export function verifyToken(token) {
  if (!token) throw new Error('missing token')
  const payload = jwt.verify(token, config.jwtPublicKey, { algorithms: ['RS256'] })
  return {
    userId: payload.sub,
    email: payload.email || '',
    role: payload.role || 'customer',
    name: payload.full_name || payload.email || 'Player',
  }
}

export function authMiddleware(req, res, next) {
  const userId = req.headers['x-user-id']
  const role = req.headers['x-user-role']
  const email = req.headers['x-user-email']

  if (userId) {
    req.user = {
      userId,
      role: role || 'customer',
      email: email || '',
      name: email?.split('@')[0] || 'Player',
    }
    return next()
  }

  const header = req.headers.authorization || ''
  const parts = header.split(' ')
  if (parts.length !== 2 || parts[0].toLowerCase() !== 'bearer') {
    return res.status(401).json({ error: 'authentication required' })
  }

  try {
    req.user = verifyToken(parts[1])
    return next()
  } catch {
    return res.status(401).json({ error: 'invalid token' })
  }
}

export function staffMiddleware(req, res, next) {
  const staffRoles = ['support', 'manager', 'admin', 'super_admin']
  if (!staffRoles.includes(req.user?.role)) {
    return res.status(403).json({ error: 'staff access required' })
  }
  return next()
}
