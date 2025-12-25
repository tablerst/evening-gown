type Json = Record<string, unknown> | unknown[] | string | number | boolean | null

const rawBase = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? ''
// Scheme A (dev same-origin): in development we intentionally DO NOT prefix API paths with
// VITE_API_BASE_URL. Instead, requests stay as /api/... and are forwarded by Vite server.proxy.
// This ensures <img> and canvas poster generation are same-origin (no CORS/canvas tainting).
const API_BASE = import.meta.env.DEV ? '' : rawBase.replace(/\/$/, '')

export class HttpError extends Error {
    status: number
    payload: unknown

    constructor(status: number, payload: unknown) {
        super(`HTTP ${status}`)
        this.status = status
        this.payload = payload
    }
}

const buildUrl = (path: string) => {
    if (!API_BASE) return path
    return `${API_BASE}${path.startsWith('/') ? '' : '/'}${path}`
}

// resolveApiUrl converts a backend-relative API path (e.g. /api/v1/assets/...) into an
// absolute URL using VITE_API_BASE_URL. This is mainly for <img src>, which bypasses
// our fetch wrapper.
export const resolveApiUrl = (maybeUrl: string) => {
    const raw = String(maybeUrl ?? '')
    if (!raw) return ''
    // keep absolute / special scheme URLs unchanged
    if (/^https?:\/\//i.test(raw) || raw.startsWith('data:') || raw.startsWith('blob:')) return raw
    // only prefix API paths; avoid breaking normal app-relative assets
    if (raw.startsWith('/api/')) return buildUrl(raw)
    return raw
}

export const httpGet = async <T = unknown>(path: string, init?: RequestInit): Promise<T> => {
    const res = await fetch(buildUrl(path), {
        ...init,
        method: 'GET',
        headers: {
            Accept: 'application/json',
            ...(init?.headers ?? {}),
        },
    })

    const payload = await safeJson(res)
    if (!res.ok) throw new HttpError(res.status, payload)
    return payload as T
}

export const httpGetBlob = async (path: string, init?: RequestInit): Promise<Blob> => {
    const res = await fetch(buildUrl(path), {
        ...init,
        method: 'GET',
        headers: {
            ...(init?.headers ?? {}),
        },
    })

    if (!res.ok) {
        const payload = await safeJson(res)
        throw new HttpError(res.status, payload)
    }

    return res.blob()
}

export const httpPost = async <T = unknown>(path: string, body?: Json, init?: RequestInit): Promise<T> => {
    const res = await fetch(buildUrl(path), {
        ...init,
        method: 'POST',
        headers: {
            Accept: 'application/json',
            'Content-Type': 'application/json',
            ...(init?.headers ?? {}),
        },
        body: body === undefined ? undefined : JSON.stringify(body),
    })

    const payload = await safeJson(res)
    if (!res.ok) throw new HttpError(res.status, payload)
    return payload as T
}

export const httpPostForm = async <T = unknown>(path: string, form: FormData, init?: RequestInit): Promise<T> => {
    const res = await fetch(buildUrl(path), {
        ...init,
        method: 'POST',
        headers: {
            Accept: 'application/json',
            // NOTE: do NOT set Content-Type here; the browser will set it with boundary.
            ...(init?.headers ?? {}),
        },
        body: form,
    })

    const payload = await safeJson(res)
    if (!res.ok) throw new HttpError(res.status, payload)
    return payload as T
}

export const httpPatch = async <T = unknown>(path: string, body?: Json, init?: RequestInit): Promise<T> => {
    const res = await fetch(buildUrl(path), {
        ...init,
        method: 'PATCH',
        headers: {
            Accept: 'application/json',
            'Content-Type': 'application/json',
            ...(init?.headers ?? {}),
        },
        body: body === undefined ? undefined : JSON.stringify(body),
    })

    const payload = await safeJson(res)
    if (!res.ok) throw new HttpError(res.status, payload)
    return payload as T
}

export const httpDelete = async <T = unknown>(path: string, init?: RequestInit): Promise<T> => {
    const res = await fetch(buildUrl(path), {
        ...init,
        method: 'DELETE',
        headers: {
            Accept: 'application/json',
            ...(init?.headers ?? {}),
        },
    })

    // 204 No Content
    if (res.status === 204) return undefined as T

    const payload = await safeJson(res)
    if (!res.ok) throw new HttpError(res.status, payload)
    return payload as T
}

const safeJson = async (res: Response): Promise<unknown> => {
    const ct = res.headers.get('content-type') ?? ''
    if (!ct.includes('application/json')) {
        const text = await res.text().catch(() => '')
        return text
    }
    return res.json().catch(() => null)
}
