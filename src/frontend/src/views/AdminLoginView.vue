<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

import { HttpError } from '@/api/http'
import { adminLogin } from '@/admin/auth'
import { setLocale } from '@/i18n'

const router = useRouter()
const route = useRoute()
const { t, locale } = useI18n()

const localeToggleLabel = computed(() => (locale.value === 'zh' ? t('admin.language.toEn') : t('admin.language.toZh')))

const email = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')

const submit = async () => {
    errorMsg.value = ''
    loading.value = true
    try {
        await adminLogin(email.value, password.value)

        const ru = typeof route.query?.returnUrl === 'string' ? (route.query.returnUrl as string) : ''
        // Prevent open redirect: only allow internal admin paths.
        if (ru && ru.startsWith('/admin') && ru !== '/admin/login') {
            await router.replace(ru)
        } else {
            await router.replace({ name: 'admin-home' })
        }
    } catch (e) {
        if (e instanceof HttpError) {
            errorMsg.value = t('admin.login.errorInvalid')
        } else {
            errorMsg.value = t('admin.login.errorNetwork')
        }
    } finally {
        loading.value = false
    }
}

const toggleLocale = () => {
    setLocale(locale.value === 'zh' ? 'en' : 'zh')
}
</script>

<template>
    <main class="min-h-screen bg-white">
        <div class="max-w-md mx-auto px-6 py-16">
            <div class="flex items-start justify-between gap-3">
                <div>
                    <h1 class="font-display text-2xl uppercase tracking-wider">{{ t('admin.login.title') }}</h1>
                    <p class="mt-2 font-mono text-xs text-black/60">{{ t('admin.login.subtitle') }}</p>
                </div>
                <button @click="toggleLocale"
                    class="h-9 px-3 border border-border font-mono text-xs uppercase tracking-widest">
                    {{ localeToggleLabel }}
                </button>
            </div>

            <div class="mt-8 space-y-4">
                <label class="block">
                    <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{ t('admin.login.email')
                    }}</div>
                    <input v-model.trim="email" type="email" autocomplete="username"
                        class="mt-2 w-full h-10 px-3 border border-border focus:outline-none" />
                </label>
                <label class="block">
                    <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                        t('admin.login.password') }}</div>
                    <input v-model="password" type="password" autocomplete="current-password"
                        class="mt-2 w-full h-10 px-3 border border-border focus:outline-none" />
                </label>

                <button :disabled="loading" @click="submit"
                    class="w-full h-10 bg-brand text-white font-mono text-sm uppercase tracking-widest disabled:opacity-60">
                    {{ loading ? t('admin.login.submitting') : t('admin.login.submit') }}
                </button>

                <p v-if="errorMsg" class="font-mono text-xs text-red-600">{{ errorMsg }}</p>
            </div>
        </div>
    </main>
</template>
