import { HttpError, httpDelete, httpGet, httpPatch, httpPost, httpPostForm } from '@/api/http'
import { getAdminToken, setAdminToken } from '@/admin/auth'

const handleAdminAuthExpired = () => {
    // Force logout: clear token and hard redirect to avoid router import cycles.
    setAdminToken('')
    if (typeof window !== 'undefined') {
        window.location.replace('/admin/login')
    }
}

const wrap = async <T>(fn: () => Promise<T>): Promise<T> => {
    try {
        return await fn()
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            handleAdminAuthExpired()
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
