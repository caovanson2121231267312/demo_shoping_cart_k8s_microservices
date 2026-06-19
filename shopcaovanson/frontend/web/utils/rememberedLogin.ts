const STORAGE_KEY = 'shop_remembered_logins'
const DEVICE_KEY = 'shop_login_device_key'
const REMEMBER_PREF_KEY = 'shop_remember_login_pref'

export interface RememberedAccount {
  email: string
  iv: string
  passwordEnc: string
  savedAt: number
}

interface RememberedStore {
  version: 1
  lastEmail: string | null
  accounts: RememberedAccount[]
}

function assertClient() {
  if (!import.meta.client) {
    throw new Error('rememberedLogin is client-only')
  }
}

function toBase64(bytes: Uint8Array): string {
  return btoa(String.fromCharCode(...bytes))
}

function fromBase64(value: string): Uint8Array {
  return Uint8Array.from(atob(value), (c) => c.charCodeAt(0))
}

async function getDeviceKey(): Promise<CryptoKey> {
  assertClient()

  const existing = localStorage.getItem(DEVICE_KEY)
  if (existing) {
    return crypto.subtle.importKey(
      'raw',
      fromBase64(existing),
      { name: 'AES-GCM', length: 256 },
      false,
      ['encrypt', 'decrypt'],
    )
  }

  const key = await crypto.subtle.generateKey(
    { name: 'AES-GCM', length: 256 },
    true,
    ['encrypt', 'decrypt'],
  )
  const raw = await crypto.subtle.exportKey('raw', key)
  localStorage.setItem(DEVICE_KEY, toBase64(new Uint8Array(raw)))
  return key
}

async function encryptPassword(password: string): Promise<{ iv: string; passwordEnc: string }> {
  const key = await getDeviceKey()
  const iv = crypto.getRandomValues(new Uint8Array(12))
  const cipher = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    key,
    new TextEncoder().encode(password),
  )
  return {
    iv: toBase64(iv),
    passwordEnc: toBase64(new Uint8Array(cipher)),
  }
}

async function decryptPassword(iv: string, passwordEnc: string): Promise<string | null> {
  try {
    const key = await getDeviceKey()
    const plain = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv: fromBase64(iv) },
      key,
      fromBase64(passwordEnc),
    )
    return new TextDecoder().decode(plain)
  } catch {
    return null
  }
}

function readStore(): RememberedStore {
  assertClient()
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return { version: 1, lastEmail: null, accounts: [] }
    }
    const parsed = JSON.parse(raw) as RememberedStore
    if (parsed.version !== 1 || !Array.isArray(parsed.accounts)) {
      return { version: 1, lastEmail: null, accounts: [] }
    }
    return parsed
  } catch {
    return { version: 1, lastEmail: null, accounts: [] }
  }
}

function writeStore(store: RememberedStore) {
  assertClient()
  localStorage.setItem(STORAGE_KEY, JSON.stringify(store))
}

export function getRememberPreference(): boolean {
  if (!import.meta.client) {
    return true
  }
  return localStorage.getItem(REMEMBER_PREF_KEY) !== '0'
}

export function setRememberPreference(enabled: boolean) {
  assertClient()
  localStorage.setItem(REMEMBER_PREF_KEY, enabled ? '1' : '0')
}

export function listSavedEmails(): string[] {
  const store = readStore()
  return store.accounts
    .slice()
    .sort((a, b) => b.savedAt - a.savedAt)
    .map((a) => a.email)
}

export function getLastSavedEmail(): string | null {
  return readStore().lastEmail
}

export async function getSavedPassword(email: string): Promise<string | null> {
  const normalized = email.trim().toLowerCase()
  const account = readStore().accounts.find((a) => a.email === normalized)
  if (!account) {
    return null
  }
  return decryptPassword(account.iv, account.passwordEnc)
}

export async function saveRememberedAccount(email: string, password: string) {
  const normalized = email.trim().toLowerCase()
  const { iv, passwordEnc } = await encryptPassword(password)
  const store = readStore()
  const next: RememberedAccount = {
    email: normalized,
    iv,
    passwordEnc,
    savedAt: Date.now(),
  }
  store.accounts = [next, ...store.accounts.filter((a) => a.email !== normalized)]
  store.lastEmail = normalized
  writeStore(store)
}

export function removeRememberedAccount(email: string) {
  const normalized = email.trim().toLowerCase()
  const store = readStore()
  store.accounts = store.accounts.filter((a) => a.email !== normalized)
  if (store.lastEmail === normalized) {
    store.lastEmail = store.accounts[0]?.email ?? null
  }
  writeStore(store)
}

export async function loadLastRememberedCredentials(): Promise<{ email: string; password: string } | null> {
  if (!getRememberPreference()) {
    return null
  }
  const email = getLastSavedEmail()
  if (!email) {
    return null
  }
  const password = await getSavedPassword(email)
  if (!password) {
    return null
  }
  return { email, password }
}
