<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

import type { MenuOption } from 'naive-ui'
import { NButton, NCard, NConfigProvider, NDivider, NForm, NFormItem, NInput, NLayout, NLayoutContent, NLayoutHeader, NLayoutSider, NMenu, NModal, NSpace, createDiscreteApi, dateEnUS, dateZhCN, enUS, zhCN } from 'naive-ui'

import { setAdminRefreshToken, setAdminToken } from '@/admin/auth'
import { adminGet, adminPatch } from '@/admin/api'
import { adminThemeOverrides } from '@/admin/theme'
import { HttpError } from '@/api/http'
import { setLocale } from '@/i18n'

const router = useRouter()
const route = useRoute()

const { t, locale } = useI18n()

const me = ref<{ id: number; email: string; role: string } | null>(null)

const menuOptions = computed<MenuOption[]>(() => {
    // ensure reactivity when locale changes
    void locale.value
    return [
        { key: 'admin-home', label: t('admin.nav.dashboard') },
        { key: 'admin-products', label: t('admin.nav.products') },
        { key: 'admin-updates', label: t('admin.nav.updates') },
        { key: 'admin-contacts', label: t('admin.nav.contacts') },
        { key: 'admin-events', label: t('admin.nav.events') },
    ]
})

const activeMenuKey = computed(() => {
    const n = route.name
    return typeof n === 'string' ? n : null
})

const pageTitle = computed(() => {
    void locale.value
    const n = activeMenuKey.value
    switch (n) {
        case 'admin-home':
            return t('admin.nav.dashboard')
        case 'admin-products':
            return t('admin.nav.products')
        case 'admin-updates':
            return t('admin.nav.updates')
        case 'admin-contacts':
            return t('admin.nav.contacts')
        case 'admin-events':
            return t('admin.nav.events')
        default:
            return t('admin.layout.brand')
    }
})

const naiveLocale = computed(() => (locale.value === 'zh' ? zhCN : enUS))
const naiveDateLocale = computed(() => (locale.value === 'zh' ? dateZhCN : dateEnUS))

const toggleLocale = () => {
    setLocale(locale.value === 'zh' ? 'en' : 'zh')
}

const localeToggleLabel = computed(() => (locale.value === 'zh' ? t('admin.language.toEn') : t('admin.language.toZh')))

const onMenuUpdate = async (key: string) => {
    if (!key) return
    if (key === route.name) return
    await router.push({ name: key })
}

const forceLogout = async () => {
    setAdminToken('')
    setAdminRefreshToken('')
    await router.replace({ name: 'admin-login' })
}

const { message, dialog, unmount } = createDiscreteApi(['message', 'dialog'], {
    configProviderProps: {
        themeOverrides: adminThemeOverrides,
    },
})

onBeforeUnmount(() => {
    // Unmount discrete APIs to avoid leaking DOM between route changes.
    unmount?.()
})

onMounted(async () => {
    try {
        me.value = await adminGet<{ id: number; email: string; role: string }>('/api/v1/admin/me')
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await forceLogout()
            return
        }
        // Non-auth errors: keep layout but show minimal info.
        me.value = null
    }
})

const showChangePassword = ref(false)
const changing = ref(false)
const pwForm = ref({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
})

const openChangePassword = () => {
    pwForm.value = { oldPassword: '', newPassword: '', confirmPassword: '' }
    showChangePassword.value = true
}

const submitChangePassword = async () => {
    const oldPassword = pwForm.value.oldPassword.trim()
    const newPassword = pwForm.value.newPassword.trim()
    const confirmPassword = pwForm.value.confirmPassword.trim()

    if (!oldPassword || !newPassword) {
        message.error(t('admin.password.errorMissing'))
        return
    }
    if (newPassword.length < 10) {
        message.error(t('admin.password.errorMin'))
        return
    }
    if (newPassword !== confirmPassword) {
        message.error(t('admin.password.errorMismatch'))
        return
    }

    changing.value = true
    try {
        await adminPatch<{ ok: boolean }>('/api/v1/admin/me/password', {
            oldPassword,
            newPassword,
        })
        showChangePassword.value = false
        message.success(t('admin.password.success'))
        await forceLogout()
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await forceLogout()
            return
        }
        const payload = (e instanceof HttpError ? e.payload : null) as any
        const msg = payload && typeof payload === 'object' && typeof payload.error === 'string' ? payload.error : t('admin.password.fail')
        message.error(msg)
    } finally {
        changing.value = false
    }
}

const confirmLogout = () => {
    dialog.warning({
        title: t('admin.logoutConfirm.title'),
        content: t('admin.logoutConfirm.content'),
        positiveText: t('admin.logoutConfirm.positive'),
        negativeText: t('admin.logoutConfirm.negative'),
        onPositiveClick: () => void forceLogout(),
    })
}
</script>

<template>
    <NConfigProvider :theme-overrides="adminThemeOverrides" :locale="naiveLocale" :date-locale="naiveDateLocale">
        <div class="min-h-screen bg-white text-black">
            <NLayout has-sider class="min-h-screen">
                <NLayoutSider bordered collapse-mode="width" :collapsed-width="64" :width="220">
                    <div class="px-4 py-4">
                        <div class="font-display text-lg uppercase tracking-wider">{{ t('admin.layout.brand') }}</div>
                        <div class="mt-1 font-mono text-[11px] uppercase tracking-[0.25em] text-black/50">
                            {{ t('admin.layout.productName') }}
                        </div>
                    </div>
                    <NDivider style="margin: 0" />
                    <div class="py-2">
                        <NMenu :value="activeMenuKey" :options="menuOptions" @update:value="onMenuUpdate" />
                    </div>
                </NLayoutSider>

                <NLayout>
                    <NLayoutHeader bordered style="height: 56px">
                        <div class="h-full px-6 flex items-center justify-between gap-4">
                            <div class="min-w-0">
                                <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">
                                    {{ t('admin.layout.backoffice') }}
                                </div>
                                <div class="font-display text-lg uppercase tracking-wider truncate">{{ pageTitle }}
                                </div>
                            </div>

                            <NSpace align="center" :size="12">
                                <div class="hidden md:block font-mono text-xs text-black/60" v-if="me">
                                    {{ me.email }} · {{ me.role }}
                                </div>
                                <NButton size="small" secondary @click="toggleLocale">{{ localeToggleLabel }}</NButton>
                                <NButton size="small" secondary @click="openChangePassword">{{
                                    t('admin.actions.changePassword') }}</NButton>
                                <NButton size="small" @click="confirmLogout">{{ t('admin.actions.logout') }}</NButton>
                            </NSpace>
                        </div>
                    </NLayoutHeader>

                    <NLayoutContent>
                        <div class="px-6 py-6">
                            <slot />
                        </div>
                    </NLayoutContent>
                </NLayout>
            </NLayout>

            <NModal v-model:show="showChangePassword" preset="card" style="width: min(560px, calc(100vw - 32px))">
                <template #header>
                    <div class="font-display text-lg uppercase tracking-wider">{{ t('admin.password.title') }}</div>
                </template>

                <NCard :bordered="false" style="padding: 0">
                    <NForm :show-feedback="false" :show-label="true" label-placement="top">
                        <NFormItem :label="t('admin.password.old')">
                            <NInput v-model:value="pwForm.oldPassword" type="password"
                                autocomplete="current-password" />
                        </NFormItem>
                        <NFormItem :label="t('admin.password.new')">
                            <NInput v-model:value="pwForm.newPassword" type="password" autocomplete="new-password" />
                        </NFormItem>
                        <NFormItem :label="t('admin.password.confirm')">
                            <NInput v-model:value="pwForm.confirmPassword" type="password"
                                autocomplete="new-password" />
                        </NFormItem>

                        <NSpace justify="end" :size="12">
                            <NButton secondary :disabled="changing" @click="showChangePassword = false">{{
                                t('admin.actions.cancel') }}
                            </NButton>
                            <NButton type="primary" :loading="changing" @click="submitChangePassword">{{
                                t('admin.actions.confirm') }}
                            </NButton>
                        </NSpace>
                    </NForm>
                </NCard>
            </NModal>
        </div>
    </NConfigProvider>
</template>
