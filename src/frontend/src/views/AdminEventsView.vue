<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { HttpError } from '@/api/http'
import { adminGet } from '@/admin/api'

type EventItem = {
    id: number
    eventType: string
    occurredAt: string
    sessionId: string
    anonId: string
    userId?: number
    productId?: number
    pageUrl: string
    referrer: string
    utmSource: string
    utmMedium: string
    utmCampaign: string
    utmContent: string
    utmTerm: string
    payload: any
}

const router = useRouter()

const loading = ref(false)
const errorMsg = ref('')
const items = ref<EventItem[]>([])

const eventType = ref('')
const productId = ref('')

const load = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
        const qs = new URLSearchParams()
        qs.set('limit', '200')
        if (eventType.value.trim()) qs.set('event_type', eventType.value.trim())
        if (productId.value.trim()) qs.set('product_id', productId.value.trim())

        const res = await adminGet<{ items: EventItem[] }>(`/api/v1/admin/events?${qs.toString()}`)
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

onMounted(load)
</script>

<template>
    <main class="min-h-screen bg-white">
        <div class="px-6 py-10 max-w-6xl mx-auto">
            <div class="flex items-center justify-between">
                <h1 class="font-display text-2xl uppercase tracking-wider">Events</h1>
                <router-link :to="{ name: 'admin-home' }" class="font-mono text-xs uppercase tracking-[0.25em]">←
                    Back</router-link>
            </div>

            <div class="mt-6 grid md:grid-cols-3 gap-3 border border-border p-4">
                <label class="block">
                    <div class="font-mono text-xs text-black/60">event_type</div>
                    <input v-model.trim="eventType" class="mt-1 w-full h-10 px-3 border border-border"
                        placeholder="poster_generated" />
                </label>
                <label class="block">
                    <div class="font-mono text-xs text-black/60">product_id</div>
                    <input v-model.trim="productId" class="mt-1 w-full h-10 px-3 border border-border"
                        placeholder="123" />
                </label>
                <div class="flex items-end">
                    <button @click="load" :disabled="loading"
                        class="h-10 px-4 bg-brand text-white font-mono text-xs uppercase tracking-[0.25em] disabled:opacity-60">
                        Search
                    </button>
                </div>
            </div>

            <p v-if="errorMsg" class="mt-4 font-mono text-xs text-red-600">{{ errorMsg }}</p>

            <div class="mt-6 overflow-x-auto border border-border">
                <table class="min-w-full text-left font-mono text-xs">
                    <thead class="bg-border/30">
                        <tr>
                            <th class="p-3">Time</th>
                            <th class="p-3">Type</th>
                            <th class="p-3">Product</th>
                            <th class="p-3">Page</th>
                            <th class="p-3">Anon</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="e in items" :key="e.id" class="border-t border-border align-top">
                            <td class="p-3 whitespace-nowrap">{{ e.occurredAt?.slice(0, 19).replace('T', ' ') }}</td>
                            <td class="p-3 whitespace-nowrap">{{ e.eventType }}</td>
                            <td class="p-3 whitespace-nowrap">{{ e.productId ?? '-' }}</td>
                            <td class="p-3 min-w-[360px] text-black/60 break-all">{{ e.pageUrl }}</td>
                            <td class="p-3 whitespace-nowrap text-black/60">{{ e.anonId || e.sessionId || '-' }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </main>
</template>
