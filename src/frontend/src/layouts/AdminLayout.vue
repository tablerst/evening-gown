<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { MenuOption } from 'naive-ui'
import { NButton, NCard, NConfigProvider, NDivider, NForm, NFormItem, NInput, NLayout, NLayoutContent, NLayoutHeader, NLayoutSider, NMenu, NModal, NSpace, createDiscreteApi } from 'naive-ui'

import { adminMe, adminChangePassword, setAdminToken } from '@/admin/auth'
import { adminThemeOverrides } from '@/admin/theme'
import { HttpError } from '@/api/http'

const router = useRouter()
const route = useRoute()

const me = ref<{ id: number; email: string; role: string } | null>(null)

const menuOptions: MenuOption[] = [
    { key: 'admin-home', label: 'Dashboard' },
    { key: 'admin-products', label: 'Products' },
    { key: 'admin-updates', label: 'Updates' },
    { key: 'admin-contacts', label: 'Contacts' },
    { key: 'admin-events', label: 'Events' },
]

const activeMenuKey = computed(() => {
    const n = route.name
    return typeof n === 'string' ? n : null
})

const pageTitle = computed(() => {
    const n = activeMenuKey.value
    switch (n) {
        case 'admin-home':
            return 'Admin'
        case 'admin-products':
            return 'Products'
        case 'admin-updates':
            return 'Updates'
        case 'admin-contacts':
            return 'Contacts'
        case 'admin-events':
            return 'Events'
        default:
            return 'Admin'
    }
})

const onMenuUpdate = async (key: string) => {
    if (!key) return
    if (key === route.name) return
    await router.push({ name: key })
}

const forceLogout = async () => {
    setAdminToken('')
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
        me.value = await adminMe()
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
        message.error('请输入旧密码与新密码')
        return
    }
    if (newPassword.length < 10) {
        message.error('新密码至少 10 位')
        return
    }
    if (newPassword !== confirmPassword) {
        message.error('两次输入的新密码不一致')
        return
    }

    changing.value = true
    try {
        await adminChangePassword(oldPassword, newPassword)
        showChangePassword.value = false
        message.success('密码已更新，请重新登录')
        await forceLogout()
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await forceLogout()
            return
        }
        const payload = (e instanceof HttpError ? e.payload : null) as any
        const msg = payload && typeof payload === 'object' && typeof payload.error === 'string' ? payload.error : '修改失败'
        message.error(msg)
    } finally {
        changing.value = false
    }
}

const confirmLogout = () => {
    dialog.warning({
        title: '确认退出',
        content: '你将退出后台登录状态。',
        positiveText: '退出',
        negativeText: '取消',
        onPositiveClick: () => void forceLogout(),
    })
}
</script>

<template>
    <NConfigProvider :theme-overrides="adminThemeOverrides">
        <div class="min-h-screen bg-white text-black">
            <NLayout has-sider class="min-h-screen">
                <NLayoutSider bordered collapse-mode="width" :collapsed-width="64" :width="220">
                    <div class="px-4 py-4">
                        <div class="font-display text-lg uppercase tracking-wider">Admin</div>
                        <div class="mt-1 font-mono text-[11px] uppercase tracking-[0.25em] text-black/50">
                            White Phantom
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
                                    Backoffice
                                </div>
                                <div class="font-display text-lg uppercase tracking-wider truncate">{{ pageTitle }}
                                </div>
                            </div>

                            <NSpace align="center" :size="12">
                                <div class="hidden md:block font-mono text-xs text-black/60" v-if="me">
                                    {{ me.email }} · {{ me.role }}
                                </div>
                                <NButton size="small" secondary @click="openChangePassword">修改密码</NButton>
                                <NButton size="small" @click="confirmLogout">退出</NButton>
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
                    <div class="font-display text-lg uppercase tracking-wider">修改密码</div>
                </template>

                <NCard :bordered="false" style="padding: 0">
                    <NForm :show-feedback="false" :show-label="true" label-placement="top">
                        <NFormItem label="旧密码">
                            <NInput v-model:value="pwForm.oldPassword" type="password"
                                autocomplete="current-password" />
                        </NFormItem>
                        <NFormItem label="新密码（至少 10 位）">
                            <NInput v-model:value="pwForm.newPassword" type="password" autocomplete="new-password" />
                        </NFormItem>
                        <NFormItem label="确认新密码">
                            <NInput v-model:value="pwForm.confirmPassword" type="password"
                                autocomplete="new-password" />
                        </NFormItem>

                        <NSpace justify="end" :size="12">
                            <NButton secondary :disabled="changing" @click="showChangePassword = false">取消</NButton>
                            <NButton type="primary" :loading="changing" @click="submitChangePassword">确认修改</NButton>
                        </NSpace>
                    </NForm>
                </NCard>
            </NModal>
        </div>
    </NConfigProvider>
</template>
