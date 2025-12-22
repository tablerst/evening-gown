<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { HttpError } from '@/api/http'
import { adminLogin } from '@/admin/auth'

const router = useRouter()

const email = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')

const submit = async () => {
    errorMsg.value = ''
    loading.value = true
    try {
        await adminLogin(email.value, password.value)
        await router.replace({ name: 'admin-home' })
    } catch (e) {
        if (e instanceof HttpError) {
            errorMsg.value = typeof e.payload === 'string' ? e.payload : '登录失败（请检查账号或密码）'
        } else {
            errorMsg.value = '登录失败（网络或服务器错误）'
        }
    } finally {
        loading.value = false
    }
}
</script>

<template>
    <main class="min-h-screen bg-white">
        <div class="max-w-md mx-auto px-6 py-16">
            <h1 class="font-display text-2xl uppercase tracking-wider">Admin Login</h1>
            <p class="mt-2 font-mono text-xs text-black/60">单一超级管理员后台（无多租户）</p>

            <div class="mt-8 space-y-4">
                <label class="block">
                    <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">Email</div>
                    <input v-model.trim="email" type="email" autocomplete="username"
                        class="mt-2 w-full h-10 px-3 border border-border focus:outline-none" />
                </label>
                <label class="block">
                    <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">Password</div>
                    <input v-model="password" type="password" autocomplete="current-password"
                        class="mt-2 w-full h-10 px-3 border border-border focus:outline-none" />
                </label>

                <button :disabled="loading" @click="submit"
                    class="w-full h-10 bg-brand text-white font-mono text-sm uppercase tracking-widest disabled:opacity-60">
                    {{ loading ? 'Signing in…' : 'Sign in' }}
                </button>

                <p v-if="errorMsg" class="font-mono text-xs text-red-600">{{ errorMsg }}</p>
            </div>
        </div>
    </main>
</template>
