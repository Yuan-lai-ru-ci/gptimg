export function getBasePath() {
  return process.env.NEXT_PUBLIC_BASE_PATH || ''
}

export function withBasePath(path: string) {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  const basePath = getBasePath()
  return `${basePath}${normalizedPath}`
}
