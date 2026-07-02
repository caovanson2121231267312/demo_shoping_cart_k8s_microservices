import dotenv from 'dotenv'

dotenv.config()

function env(key, fallback = '') {
  return process.env[key] ?? fallback
}

export const config = {
  port: parseInt(env('PORT', '8087'), 10),
  databaseUrl: env('DATABASE_URL'),
  jwtPublicKey: env('JWT_PUBLIC_KEY').replace(/\\n/g, '\n'),
  corsOrigins: env('CORS_ORIGINS', '*').split(',').map((s) => s.trim()).filter(Boolean),
  turnUrl: env('TURN_URL', ''),
  turnUsername: env('TURN_USERNAME', ''),
  turnCredential: env('TURN_CREDENTIAL', ''),
}
