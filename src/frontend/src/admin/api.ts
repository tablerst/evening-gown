import { httpGet, httpPatch, httpPost } from '@/api/http'
import { getAdminToken } from '@/admin/auth'

const withAuth = (): Record<string, string> => {
    const token = getAdminToken()
    const headers: Record<string, string> = {}
    if (token) headers.Authorization = `Bearer ${token}`
    return headers
}

export const adminGet = async <T = unknown>(path: string) => {
    return httpGet<T>(path, { headers: withAuth() })
}

export const adminPost = async <T = unknown>(path: string, body?: unknown) => {
    return httpPost<T>(path, body as never, { headers: withAuth() })
}

export const adminPatch = async <T = unknown>(path: string, body?: unknown) => {
    return httpPatch<T>(path, body as never, { headers: withAuth() })
}
