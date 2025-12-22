<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { HttpError } from '@/api/http'
import { adminDelete, adminGet, adminPatch, adminPost } from '@/admin/api'

type UpdatePost = {
    id: number
    type: string
    status: string
    tag: string
    title: string
    body: string
    refCode: string
    summary?: string
    pinnedRank?: number
    publishedAt?: string
    deletedAt?: string
}

const router = useRouter()
const loading = ref(false)
const errorMsg = ref('')
const items = ref<UpdatePost[]>([])

const filterType = ref<'all' | 'company' | 'industry'>('all')
const filterStatus = ref<'all' | 'draft' | 'published' | 'archived'>('all')

const editingId = ref<number | null>(null)
const editForm = ref({
    type: 'company' as 'company' | 'industry',
    status: 'draft' as 'draft' | 'published' | 'archived',
    tag: '新品',
    title: '',
    summary: '',
    body: '',
    ref: '',
    pinnedRank: 0,
})

const form = ref({
    tag: '新品',
    title: '',
    body: '',
    ref: '',
    status: 'published' as 'draft' | 'published',
})

const load = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
        const qs = new URLSearchParams()
        qs.set('limit', '50')
        if (filterType.value !== 'all') qs.set('type', filterType.value)
        if (filterStatus.value !== 'all') qs.set('status', filterStatus.value)
        const res = await adminGet<{ items: UpdatePost[] }>(`/api/v1/admin/updates?${qs.toString()}`)
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
            status: form.value.status,
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

const startEdit = async (id: number) => {
    loading.value = true
    errorMsg.value = ''
    try {
        const u = await adminGet<UpdatePost>(`/api/v1/admin/updates/${id}`)
        editingId.value = id
        editForm.value = {
            type: (u.type === 'industry' ? 'industry' : 'company') as any,
            status: (u.status === 'archived' ? 'archived' : u.status === 'published' ? 'published' : 'draft') as any,
            tag: u.tag ?? '',
            title: u.title ?? '',
            summary: u.summary ?? '',
            body: u.body ?? '',
            ref: u.refCode ?? '',
            pinnedRank: (u.pinnedRank ?? 0) as number,
        }
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = '加载详情失败'
    } finally {
        loading.value = false
    }
}

const cancelEdit = () => {
    editingId.value = null
}

const saveEdit = async () => {
    if (!editingId.value) return
    if (!editForm.value.title.trim()) {
        errorMsg.value = 'Title 不能为空'
        return
    }
    loading.value = true
    errorMsg.value = ''
    try {
        await adminPatch(`/api/v1/admin/updates/${editingId.value}`, {
            type: editForm.value.type,
            status: editForm.value.status,
            tag: editForm.value.tag,
            title: editForm.value.title,
            summary: editForm.value.summary,
            body: editForm.value.body,
            ref: editForm.value.ref,
            pinnedRank: editForm.value.pinnedRank,
        })
        editingId.value = null
        await load()
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = '保存失败'
    } finally {
        loading.value = false
    }
}

const publish = async (id: number) => {
    loading.value = true
    errorMsg.value = ''
    try {
        await adminPost(`/api/v1/admin/updates/${id}/publish`)
        await load()
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = '发布失败'
    } finally {
        loading.value = false
    }
}

const unpublish = async (id: number) => {
    loading.value = true
    errorMsg.value = ''
    try {
        await adminPost(`/api/v1/admin/updates/${id}/unpublish`)
        await load()
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = '取消发布失败'
    } finally {
        loading.value = false
    }
}

const remove = async (id: number) => {
    if (!confirm(`确认删除 Update #${id}？（软删除，前台将不可见）`)) return
    loading.value = true
    errorMsg.value = ''
    try {
        await adminDelete(`/api/v1/admin/updates/${id}`)
        if (editingId.value === id) editingId.value = null
        await load()
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = '删除失败'
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
                    <label class="block">
                        <div class="font-mono text-xs text-black/60">Status</div>
                        <select v-model="form.status"
                            class="mt-1 w-full h-10 px-3 border border-border font-mono text-xs">
                            <option value="draft">draft</option>
                            <option value="published">published</option>
                        </select>
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
                    Create
                </button>
            </section>

            <section v-if="editingId" class="mt-8 border border-border p-6">
                <div class="flex items-center justify-between gap-4">
                    <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">Edit #{{ editingId }}</h2>
                    <button :disabled="loading" @click="cancelEdit"
                        class="h-9 px-3 border border-border font-mono text-xs uppercase tracking-[0.25em] disabled:opacity-60">Cancel</button>
                </div>
                <div class="mt-4 grid md:grid-cols-2 gap-3">
                    <label class="block">
                        <div class="font-mono text-xs text-black/60">Type</div>
                        <select v-model="editForm.type"
                            class="mt-1 w-full h-10 px-3 border border-border font-mono text-xs">
                            <option value="company">company</option>
                            <option value="industry">industry</option>
                        </select>
                    </label>
                    <label class="block">
                        <div class="font-mono text-xs text-black/60">Status</div>
                        <select v-model="editForm.status"
                            class="mt-1 w-full h-10 px-3 border border-border font-mono text-xs">
                            <option value="draft">draft</option>
                            <option value="published">published</option>
                            <option value="archived">archived</option>
                        </select>
                    </label>
                    <label class="block">
                        <div class="font-mono text-xs text-black/60">Pinned Rank</div>
                        <input v-model.number="editForm.pinnedRank" type="number"
                            class="mt-1 w-full h-10 px-3 border border-border" />
                    </label>
                    <label class="block">
                        <div class="font-mono text-xs text-black/60">Tag</div>
                        <input v-model.trim="editForm.tag" class="mt-1 w-full h-10 px-3 border border-border" />
                    </label>
                </div>
                <label class="block mt-3">
                    <div class="font-mono text-xs text-black/60">Ref</div>
                    <input v-model.trim="editForm.ref" class="mt-1 w-full h-10 px-3 border border-border" />
                </label>
                <label class="block mt-3">
                    <div class="font-mono text-xs text-black/60">Title</div>
                    <input v-model.trim="editForm.title" class="mt-1 w-full h-10 px-3 border border-border" />
                </label>
                <label class="block mt-3">
                    <div class="font-mono text-xs text-black/60">Summary</div>
                    <textarea v-model="editForm.summary" rows="2"
                        class="mt-1 w-full px-3 py-2 border border-border"></textarea>
                </label>
                <label class="block mt-3">
                    <div class="font-mono text-xs text-black/60">Body</div>
                    <textarea v-model="editForm.body" rows="4"
                        class="mt-1 w-full px-3 py-2 border border-border"></textarea>
                </label>
                <button :disabled="loading" @click="saveEdit"
                    class="mt-4 h-10 px-4 bg-brand text-white font-mono text-xs uppercase tracking-[0.25em] disabled:opacity-60">
                    Save
                </button>
            </section>

            <section class="mt-8">
                <div class="flex items-center justify-between">
                    <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">List</h2>
                    <div class="flex items-center gap-3">
                        <select v-model="filterType" class="h-9 px-2 border border-border font-mono text-xs">
                            <option value="all">all</option>
                            <option value="company">company</option>
                            <option value="industry">industry</option>
                        </select>
                        <select v-model="filterStatus" class="h-9 px-2 border border-border font-mono text-xs">
                            <option value="all">all</option>
                            <option value="draft">draft</option>
                            <option value="published">published</option>
                            <option value="archived">archived</option>
                        </select>
                        <button @click="load"
                            class="h-9 px-3 border border-border font-mono text-xs uppercase tracking-[0.25em]">Refresh</button>
                    </div>
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
                        <div class="mt-3 flex flex-wrap gap-2">
                            <button :disabled="loading" @click="startEdit(u.id)"
                                class="h-8 px-3 border border-border bg-white hover:border-black transition-none disabled:opacity-60">Edit</button>
                            <button v-if="u.status !== 'published'" :disabled="loading" @click="publish(u.id)"
                                class="h-8 px-3 border border-black bg-white hover:bg-brand hover:text-white transition-none disabled:opacity-60">Publish</button>
                            <button v-else :disabled="loading" @click="unpublish(u.id)"
                                class="h-8 px-3 border border-border bg-white hover:border-black transition-none disabled:opacity-60">Unpublish</button>
                            <button :disabled="loading" @click="remove(u.id)"
                                class="h-8 px-3 border border-red-300 bg-white text-red-700 hover:border-red-500 transition-none disabled:opacity-60">Delete</button>
                        </div>
                    </article>
                </div>
            </section>
        </div>
    </main>
</template>
