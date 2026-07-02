import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'
import { pool } from './pool.js'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

async function migrate() {
  const sqlPath = path.join(__dirname, '../../migrations/001_init.sql')
  const sql = fs.readFileSync(sqlPath, 'utf8')
  await pool.query(sql)
  console.log('caro-service migrations applied')
  await pool.end()
}

migrate().catch((err) => {
  console.error('migration failed:', err)
  process.exit(1)
})
