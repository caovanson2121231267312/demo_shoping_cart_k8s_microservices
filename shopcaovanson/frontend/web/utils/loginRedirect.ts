/** Safe internal path for post-login redirect */
export function resolveLoginRedirect(redirect: unknown, fallback = '/'): string {
  if (typeof redirect !== 'string') {
    return fallback
  }
  const path = redirect.trim()
  if (!path.startsWith('/') || path.startsWith('//')) {
    return fallback
  }
  return path
}

export function loginRedirectQuery(fullPath: string): { redirect: string } | undefined {
  if (!fullPath || fullPath.startsWith('/auth/')) {
    return undefined
  }
  return { redirect: fullPath }
}
