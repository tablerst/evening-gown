import { createI18n } from 'vue-i18n'

import { appEnv } from '@/config/env'
import type { SupportedLocale } from './messages'

import en from './locales/en.json'
import zh from './locales/zh.json'

const messages = {
    en,
    zh,
} satisfies Record<SupportedLocale, Record<string, unknown>>

const SUPPORTED_LOCALES = ['zh', 'en'] as const satisfies SupportedLocale[]
const STORAGE_KEY = 'evening-gowm-locale'

type LocaleInput = SupportedLocale | string | null | undefined

const normalizeLocale = (value: LocaleInput): SupportedLocale => {
    const normalized = String(value ?? '').trim().toLowerCase()

    if (SUPPORTED_LOCALES.includes(normalized as SupportedLocale)) {
        return normalized as SupportedLocale
    }

    return 'zh'
}

const readStoredLocale = (): SupportedLocale | null => {
    if (typeof localStorage === 'undefined') return null

    const stored = localStorage.getItem(STORAGE_KEY)
    return stored ? normalizeLocale(stored) : null
}

const initialLocale = readStoredLocale() ?? normalizeLocale(appEnv.locale)

export const i18n = createI18n({
    legacy: false,
    locale: initialLocale,
    fallbackLocale: 'en',
    messages,
})

export const setLocale = (next: LocaleInput): SupportedLocale => {
    const locale = normalizeLocale(next)

    i18n.global.locale.value = locale

    if (typeof document !== 'undefined') {
        document.documentElement.setAttribute('lang', locale)
    }

    if (typeof localStorage !== 'undefined') {
        localStorage.setItem(STORAGE_KEY, locale)
    }

    return locale
}

if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('lang', initialLocale)
}

export { SUPPORTED_LOCALES }
