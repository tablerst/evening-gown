<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { HttpError, httpGet } from '@/api/http'

type UpdateItem = {
    id: number
    date: string
    tag: string
    title: string
    body: string
    ref?: string
}

const router = useRouter()

const loading = ref(false)
const errorMsg = ref('')
const items = ref<UpdateItem[]>([])

const load = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
        const res = await httpGet<{ items: UpdateItem[] }>('/api/v1/updates?limit=50')
        items.value = res.items ?? []
    } catch (e) {
        if (e instanceof HttpError) {
            errorMsg.value = '加载失败'
        } else {
            errorMsg.value = '加载失败'
        }
    } finally {
        loading.value = false
    }
}

const open = async (id: number) => {
    await router.push({ name: 'update-detail', params: { id: String(id) } })
}

onMounted(load)
</script>

<template>
    <main class="min-h-screen bg-white">
        <div class="px-6 py-10 max-w-5xl mx-auto">
            <div class="flex items-center justify-between">
                <h1 class="font-display text-2xl uppercase tracking-wider">Updates</h1>
                <router-link :to="{ name: 'home' }" class="font-mono text-xs uppercase tracking-[0.25em]">←
                    Back</router-link>
            </div>

            <div class="mt-6 flex items-center justify-between gap-4">
                <button @click="load" :disabled="loading"
                    class="h-9 px-3 border border-border font-mono text-xs uppercase tracking-[0.25em] disabled:opacity-60">Refresh</button>
                <div class="font-mono text-xs text-black/60">{{ items.length }} items</div>
            </div>

            <p v-if="errorMsg" class="mt-4 font-mono text-xs text-red-600">{{ errorMsg }}</p>

            <div class="mt-6 space-y-3">
                <article v-for="u in items" :key="u.id" class="border border-border p-4">
                    <div class="flex items-start justify-between gap-4">
                        <div class="font-mono text-xs text-black/60">#{{ u.id }} · {{ u.date }} · {{ u.tag }}</div>
                        <div class="font-mono text-xs text-black/60">{{ u.ref }}</div>
                    </div>
                    <h2 class="mt-2 font-sans font-semibold uppercase tracking-[0.22em] text-sm">{{ u.title }}</h2>
                    <p class="mt-2 text-sm text-black/70">{{ u.body }}</p>
                    <div class="mt-3">
                        <button @click="open(u.id)"
                            class="h-9 px-3 border border-black bg-white hover:bg-brand hover:text-white transition-none">
                            Read
                        </button>
                    </div>
                </article>
            </div>
        </div>
    </main>
</template>
