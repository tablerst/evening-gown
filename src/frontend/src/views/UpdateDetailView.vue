<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { HttpError, httpGet } from '@/api/http'

type UpdateDetail = {
    id: number
    type: string
    date: string
    tag: string
    title: string
    body: string
    ref?: string
}

const route = useRoute()
const router = useRouter()

const id = computed(() => String(route.params.id ?? ''))

const loading = ref(false)
const errorMsg = ref('')
const item = ref<UpdateDetail | null>(null)

const load = async () => {
    if (!id.value) return
    loading.value = true
    errorMsg.value = ''
    try {
        const res = await httpGet<UpdateDetail>(`/api/v1/updates/${id.value}`)
        item.value = res
    } catch (e) {
        if (e instanceof HttpError && e.status === 404) {
            errorMsg.value = '未找到'
        } else {
            errorMsg.value = '加载失败'
        }
    } finally {
        loading.value = false
    }
}

onMounted(load)
</script>

<template>
    <main class="min-h-screen bg-white">
        <div class="px-6 py-10 max-w-3xl mx-auto">
            <div class="flex items-center justify-between">
                <h1 class="font-display text-2xl uppercase tracking-wider">Update</h1>
                <button @click="router.back()" class="font-mono text-xs uppercase tracking-[0.25em]">← Back</button>
            </div>

            <p v-if="errorMsg" class="mt-6 font-mono text-xs text-red-600">{{ errorMsg }}</p>

            <article v-if="item" class="mt-6 border border-border p-6">
                <div class="flex items-start justify-between gap-4">
                    <div class="font-mono text-xs text-black/60">#{{ item.id }} · {{ item.date }} · {{ item.tag }}</div>
                    <div class="font-mono text-xs text-black/60">{{ item.ref }}</div>
                </div>
                <h2 class="mt-3 font-sans font-semibold uppercase tracking-[0.22em] text-base">{{ item.title }}</h2>
                <p class="mt-4 text-sm leading-relaxed text-black/80 whitespace-pre-wrap">{{ item.body }}</p>
            </article>
        </div>
    </main>
</template>
