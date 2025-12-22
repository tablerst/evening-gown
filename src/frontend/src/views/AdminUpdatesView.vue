<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { HttpError } from '@/api/http'
import { adminGet, adminPost } from '@/admin/api'

type UpdatePost = {
    id: number
    type: string
    status: string
    tag: string
    title: string
    body: string
    refCode: string
    publishedAt?: string
}

const router = useRouter()
const loading = ref(false)
const errorMsg = ref('')
const items = ref<UpdatePost[]>([])

const form = ref({
    tag: '新品',
    title: '',
    body: '',
    ref: '',
})

const load = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
        const res = await adminGet<{ items: UpdatePost[] }>('/api/v1/admin/updates?limit=50')
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

const create = async () => {
    if (!form.value.title.trim()) return
    loading.value = true
    errorMsg.value = ''
    try {
        await adminPost('/api/v1/admin/updates', {
            type: 'company',
            status: 'published',
            tag: form.value.tag,
            title: form.value.title,
            body: form.value.body,
            ref: form.value.ref,
        })
        form.value.title = ''
        form.value.body = ''
        form.value.ref = ''
        await load()
    } catch {
        errorMsg.value = '创建失败'
    } finally {
        loading.value = false
    }
}

onMounted(load)
</script>

<template>
    <main class="min-h-screen bg-white">
        <div class="px-6 py-10 max-w-5xl mx-auto">
            <div class="flex items-center justify-between">
                <h1 class="font-display text-2xl uppercase tracking-wider">Updates</h1>
                <router-link :to="{ name: 'admin-home' }" class="font-mono text-xs uppercase tracking-[0.25em]">←
                    Back</router-link>
            </div>

            <p v-if="errorMsg" class="mt-4 font-mono text-xs text-red-600">{{ errorMsg }}</p>

            <section class="mt-8 border border-border p-6">
                <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">Create (company only)</h2>
                <div class="mt-4 grid md:grid-cols-2 gap-3">
                    <label class="block">
                        <div class="font-mono text-xs text-black/60">Tag</div>
                        <input v-model.trim="form.tag" class="mt-1 w-full h-10 px-3 border border-border" />
                    </label>
                    <label class="block">
                        <div class="font-mono text-xs text-black/60">Ref</div>
                        <input v-model.trim="form.ref" class="mt-1 w-full h-10 px-3 border border-border" />
                    </label>
                </div>
                <label class="block mt-3">
                    <div class="font-mono text-xs text-black/60">Title</div>
                    <input v-model.trim="form.title" class="mt-1 w-full h-10 px-3 border border-border" />
                </label>
                <label class="block mt-3">
                    <div class="font-mono text-xs text-black/60">Body</div>
                    <textarea v-model="form.body" rows="4"
                        class="mt-1 w-full px-3 py-2 border border-border"></textarea>
                </label>
                <button :disabled="loading || !form.title.trim()" @click="create"
                    class="mt-4 h-10 px-4 bg-brand text-white font-mono text-xs uppercase tracking-[0.25em] disabled:opacity-60">
                    Publish
                </button>
            </section>

            <section class="mt-8">
                <div class="flex items-center justify-between">
                    <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">List</h2>
                    <button @click="load"
                        class="h-9 px-3 border border-border font-mono text-xs uppercase tracking-[0.25em]">Refresh</button>
                </div>
                <div class="mt-4 space-y-3">
                    <article v-for="u in items" :key="u.id" class="border border-border p-4">
                        <div class="flex justify-between gap-4">
                            <div class="font-mono text-xs text-black/60">#{{ u.id }} · {{ u.tag }} · {{ u.status }}
                            </div>
                            <div class="font-mono text-xs text-black/60">{{ u.refCode }}</div>
                        </div>
                        <h3 class="mt-2 font-sans font-semibold uppercase tracking-[0.22em] text-sm">{{ u.title }}</h3>
                        <p class="mt-2 text-sm text-black/70">{{ u.body }}</p>
                    </article>
                </div>
            </section>
        </div>
    </main>
</template>
