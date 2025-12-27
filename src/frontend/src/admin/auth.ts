import { httpPost } from '@/api/http'

const TOKEN_KEY = 'admin_token'
const REFRESH_TOKEN_KEY = 'admin_refresh_token'

export const getAdminToken = () => {
    if (typeof localStorage === 'undefined') return ''
    return localStorage.getItem(TOKEN_KEY) ?? ''
}

export const setAdminToken = (token: string) => {
    if (typeof localStorage === 'undefined') return
    if (!token) localStorage.removeItem(TOKEN_KEY)
    else localStorage.setItem(TOKEN_KEY, token)
}

export const getAdminRefreshToken = () => {
    if (typeof localStorage === 'undefined') return ''
    return localStorage.getItem(REFRESH_TOKEN_KEY) ?? ''
}

export const setAdminRefreshToken = (token: string) => {
    if (typeof localStorage === 'undefined') return
    if (!token) localStorage.removeItem(REFRESH_TOKEN_KEY)
    else localStorage.setItem(REFRESH_TOKEN_KEY, token)
}

export const adminLogin = async (email: string, password: string) => {
    const res = await httpPost<{ token: string; expires_at: string; refresh_token: string; refresh_expires_at: string }>('/api/v1/admin/auth/login', {
        email,
        password,
    })
    setAdminToken(res.token)
    setAdminRefreshToken(res.refresh_token)
    return res
}

export const adminRefresh = async () => {
    const refreshToken = getAdminRefreshToken()
    const res = await httpPost<{ token: string; expires_at: string; refresh_token: string; refresh_expires_at: string }>(
        '/api/v1/admin/auth/refresh',
        { refresh_token: refreshToken },
    )
    setAdminToken(res.token)
    setAdminRefreshToken(res.refresh_token)
    return res
}
