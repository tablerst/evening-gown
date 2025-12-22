<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { HttpError } from '@/api/http'
import { adminGet, adminPatch } from '@/admin/api'

type ContactLead = {
    id: number
    name: string
    phone: string
    wechat: string
    message: string
    sourcePage: string
    utmSource: string
    utmMedium: string
    utmCampaign: string
    utmContent: string
    utmTerm: string
    status: 'new' | 'contacted' | 'closed'
    createdAt: string
}

const router = useRouter()

const loading = ref(false)
const errorMsg = ref('')
const items = ref<ContactLead[]>([])

const filterStatus = ref<'all' | 'new' | 'contacted' | 'closed'>('all')

const load = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
        const qs = new URLSearchParams()
        qs.set('limit', '100')
        if (filterStatus.value !== 'all') qs.set('status', filterStatus.value)

        const res = await adminGet<{ items: ContactLead[] }>(`/api/v1/admin/contacts?${qs.toString()}`)
        items.value = res.items ?? []
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = '加载失败'
    } finally {
        loading.value = false
    }
}

const setStatus = async (id: number, status: ContactLead['status']) => {
    loading.value = true
    errorMsg.value = ''
    try {
        await adminPatch(`/api/v1/admin/contacts/${id}`, { status })
        await load()
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = '更新失败'
    } finally {
        loading.value = false
    }
}

onMounted(load)
</script>

<template>
    <main class="min-h-screen bg-white">
        <div class="px-6 py-10 max-w-6xl mx-auto">
            <div class="flex items-center justify-between">
                <h1 class="font-display text-2xl uppercase tracking-wider">Contacts</h1>
                <router-link :to="{ name: 'admin-home' }" class="font-mono text-xs uppercase tracking-[0.25em]">←
                    Back</router-link>
            </div>

            <div class="mt-6 flex items-center justify-between gap-4">
                <div class="flex items-center gap-3">
                    <span class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">Status</span>
                    <select v-model="filterStatus" class="h-9 px-2 border border-border font-mono text-xs">
                        <option value="all">all</option>
                        <option value="new">new</option>
                        <option value="contacted">contacted</option>
                        <option value="closed">closed</option>
                    </select>
                    <button @click="load"
                        class="h-9 px-3 border border-border font-mono text-xs uppercase tracking-[0.25em]">Refresh</button>
                </div>
                <div class="font-mono text-xs text-black/60">{{ items.length }} items</div>
            </div>

            <p v-if="errorMsg" class="mt-4 font-mono text-xs text-red-600">{{ errorMsg }}</p>

            <div class="mt-6 overflow-x-auto border border-border">
                <table class="min-w-full text-left font-mono text-xs">
                    <thead class="bg-border/30">
                        <tr>
                            <th class="p-3">ID</th>
                            <th class="p-3">Created</th>
                            <th class="p-3">Name</th>
                            <th class="p-3">Phone</th>
                            <th class="p-3">Wechat</th>
                            <th class="p-3">Message</th>
                            <th class="p-3">Source</th>
                            <th class="p-3">Status</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="c in items" :key="c.id" class="border-t border-border align-top">
                            <td class="p-3">{{ c.id }}</td>
                            <td class="p-3 whitespace-nowrap">{{ c.createdAt?.slice(0, 19).replace('T', ' ') }}</td>
                            <td class="p-3 whitespace-nowrap">{{ c.name }}</td>
                            <td class="p-3 whitespace-nowrap">{{ c.phone }}</td>
                            <td class="p-3 whitespace-nowrap">{{ c.wechat }}</td>
                            <td class="p-3 min-w-[240px] text-black/70">{{ c.message }}</td>
                            <td class="p-3 min-w-[200px] text-black/60">{{ c.sourcePage }}</td>
                            <td class="p-3">
                                <select :disabled="loading" :value="c.status"
                                    @change="setStatus(c.id, ($event.target as HTMLSelectElement).value as any)"
                                    class="h-9 px-2 border border-border font-mono text-xs">
                                    <option value="new">new</option>
                                    <option value="contacted">contacted</option>
                                    <option value="closed">closed</option>
                                </select>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </main>
</template>
