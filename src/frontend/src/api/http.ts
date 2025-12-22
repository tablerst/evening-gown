type Json = Record<string, unknown> | unknown[] | string | number | boolean | null

const rawBase = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? ''
const API_BASE = rawBase.replace(/\/$/, '')

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

const safeJson = async (res: Response): Promise<unknown> => {
    const ct = res.headers.get('content-type') ?? ''
    if (!ct.includes('application/json')) {
        const text = await res.text().catch(() => '')
        return text
    }
    return res.json().catch(() => null)
}
