import { HttpError, httpDelete, httpGet, httpGetBlob, httpPatch, httpPost, httpPostForm } from '@/api/http'
import { adminRefresh, getAdminToken, setAdminRefreshToken, setAdminToken } from '@/admin/auth'

const handleAdminAuthExpired = () => {
    // Force logout: clear token and hard redirect to avoid router import cycles.
    setAdminToken('')
    setAdminRefreshToken('')
    if (typeof window !== 'undefined') {
        const returnUrl = `${window.location.pathname}${window.location.search}${window.location.hash}`
        const qs = returnUrl ? `?returnUrl=${encodeURIComponent(returnUrl)}` : ''
        window.location.replace(`/admin/login${qs}`)
    }
}

let refreshInFlight: Promise<void> | null = null
const ensureRefreshed = async () => {
    if (refreshInFlight) return refreshInFlight
    refreshInFlight = (async () => {
        await adminRefresh()
    })().finally(() => {
        refreshInFlight = null
    })
    return refreshInFlight
}

const wrap = async <T>(fn: () => Promise<T>): Promise<T> => {
    try {
        return await fn()
    } catch (e) {
        if (e instanceof HttpError) {
            // 401: likely access token expired -> try refresh once.
            if (e.status === 401) {
                try {
                    await ensureRefreshed()
                    return await fn()
                } catch {
                    handleAdminAuthExpired()
                }
            }
            // 403: role/status forbidden; refresh won't help.
            if (e.status === 403) {
                handleAdminAuthExpired()
            }
        }
        throw e
    }
}

const withAuth = (): Record<string, string> => {
    const token = getAdminToken()
    const headers: Record<string, string> = {}
    if (token) headers.Authorization = `Bearer ${token}`
    return headers
}

export const adminGet = async <T = unknown>(path: string) => {
    return wrap(() => httpGet<T>(path, { headers: withAuth() }))
}

export const adminGetBlob = async (path: string) => {
    return wrap(() => httpGetBlob(path, { headers: withAuth() }))
}

export const adminPost = async <T = unknown>(path: string, body?: unknown) => {
    return wrap(() => httpPost<T>(path, body as never, { headers: withAuth() }))
}

export const adminPostForm = async <T = unknown>(path: string, form: FormData) => {
    return wrap(() => httpPostForm<T>(path, form, { headers: withAuth() }))
}

export const adminPatch = async <T = unknown>(path: string, body?: unknown) => {
    return wrap(() => httpPatch<T>(path, body as never, { headers: withAuth() }))
}

export const adminDelete = async <T = unknown>(path: string) => {
    return wrap(() => httpDelete<T>(path, { headers: withAuth() }))
}
