<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { HttpError } from '@/api/http'
import { adminMe, setAdminToken } from '@/admin/auth'

const router = useRouter()
const me = ref<{ id: number; email: string; role: string } | null>(null)
const errorMsg = ref('')

const logout = async () => {
    setAdminToken('')
    await router.replace({ name: 'admin-login' })
}

onMounted(async () => {
    try {
        me.value = await adminMe()
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = '无法加载管理员信息'
    }
})
</script>

<template>
    <main class="min-h-screen bg-white">
        <div class="px-6 py-10 max-w-5xl mx-auto">
            <div class="flex items-start justify-between gap-6">
                <div>
                    <h1 class="font-display text-2xl uppercase tracking-wider">Admin</h1>
                    <p class="mt-2 font-mono text-xs text-black/60" v-if="me">{{ me.email }} · {{ me.role }}</p>
                    <p class="mt-2 font-mono text-xs text-red-600" v-if="errorMsg">{{ errorMsg }}</p>
                </div>
                <button @click="logout"
                    class="h-10 px-4 border border-border font-mono text-xs uppercase tracking-[0.25em]">
                    Logout
                </button>
            </div>

            <div class="mt-10 grid md:grid-cols-2 gap-4">
                <router-link :to="{ name: 'admin-products' }" class="p-6 border border-border hover:border-black">
                    <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">Catalog</div>
                    <div class="mt-2 font-display text-xl uppercase tracking-wider">Products</div>
                </router-link>
                <router-link :to="{ name: 'admin-updates' }" class="p-6 border border-border hover:border-black">
                    <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">Content</div>
                    <div class="mt-2 font-display text-xl uppercase tracking-wider">Updates</div>
                </router-link>
                <router-link :to="{ name: 'admin-contacts' }" class="p-6 border border-border hover:border-black">
                    <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">Leads</div>
                    <div class="mt-2 font-display text-xl uppercase tracking-wider">Contacts</div>
                </router-link>
                <router-link :to="{ name: 'admin-events' }" class="p-6 border border-border hover:border-black">
                    <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">Analytics</div>
                    <div class="mt-2 font-display text-xl uppercase tracking-wider">Events</div>
                </router-link>
            </div>
        </div>
    </main>
</template>
