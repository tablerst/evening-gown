import { httpGet, httpPost } from '@/api/http'

const TOKEN_KEY = 'admin_token'

export const getAdminToken = () => {
    if (typeof localStorage === 'undefined') return ''
    return localStorage.getItem(TOKEN_KEY) ?? ''
}

export const setAdminToken = (token: string) => {
    if (typeof localStorage === 'undefined') return
    if (!token) localStorage.removeItem(TOKEN_KEY)
    else localStorage.setItem(TOKEN_KEY, token)
}

export const adminLogin = async (email: string, password: string) => {
    const res = await httpPost<{ token: string; expires_at: string }>('/api/v1/admin/auth/login', {
        email,
        password,
    })
    setAdminToken(res.token)
    return res
}

export const adminMe = async () => {
    const token = getAdminToken()
    return httpGet<{ id: number; email: string; role: string }>('/api/v1/admin/me', {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
}
