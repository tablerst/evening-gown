<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { NButton, NCard, NForm, NFormItem, NInput, NInputNumber, NModal, NSelect, NSpace } from 'naive-ui'

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

const showCreateModal = ref(false)
const showEditModal = ref(false)

const typeOptions = [
    { label: 'all', value: 'all' },
    { label: 'company', value: 'company' },
    { label: 'industry', value: 'industry' },
] as const

const statusFilterOptions = [
    { label: 'all', value: 'all' },
    { label: 'draft', value: 'draft' },
    { label: 'published', value: 'published' },
    { label: 'archived', value: 'archived' },
] as const

const createStatusOptions = [
    { label: 'draft', value: 'draft' },
    { label: 'published', value: 'published' },
] as const

const editTypeOptions = [
    { label: 'company', value: 'company' },
    { label: 'industry', value: 'industry' },
] as const

const editStatusOptions = [
    { label: 'draft', value: 'draft' },
    { label: 'published', value: 'published' },
    { label: 'archived', value: 'archived' },
] as const

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

        showCreateModal.value = false
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

        showEditModal.value = true
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
    showEditModal.value = false
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
        showEditModal.value = false
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
    <div class="max-w-5xl mx-auto">
        <NCard size="large">
            <NSpace justify="space-between" align="center" :wrap="true">
                <NSpace align="center" :size="12" :wrap="true">
                    <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">Filter</div>
                    <NSelect v-model:value="filterType" :options="typeOptions as any" style="width: 160px"
                        size="small" />
                    <NSelect v-model:value="filterStatus" :options="statusFilterOptions as any" style="width: 160px"
                        size="small" />
                    <NButton size="small" secondary :loading="loading" @click="load">Refresh</NButton>
                </NSpace>
                <NButton size="small" type="primary" @click="showCreateModal = true">New Update</NButton>
            </NSpace>

            <p v-if="errorMsg" class="mt-3 font-mono text-xs text-red-600">{{ errorMsg }}</p>

            <div class="mt-4 space-y-3">
                <NCard v-for="u in items" :key="u.id" size="small">
                    <div class="flex justify-between gap-4">
                        <div class="font-mono text-xs text-black/60">#{{ u.id }} · {{ u.tag }} · {{ u.status }}</div>
                        <div class="font-mono text-xs text-black/60">{{ u.refCode }}</div>
                    </div>
                    <div class="mt-2 font-sans font-semibold uppercase tracking-[0.22em] text-sm">{{ u.title }}</div>
                    <p class="mt-2 text-sm text-black/70 whitespace-pre-wrap">{{ u.body }}</p>
                    <div class="mt-3">
                        <NSpace :size="8" :wrap="true">
                            <NButton size="tiny" secondary :disabled="loading" @click="startEdit(u.id)">Edit</NButton>
                            <NButton v-if="u.status !== 'published'" size="tiny" :disabled="loading"
                                @click="publish(u.id)">Publish</NButton>
                            <NButton v-else size="tiny" secondary :disabled="loading" @click="unpublish(u.id)">Unpublish
                            </NButton>
                            <NButton size="tiny" type="error" secondary :disabled="loading" @click="remove(u.id)">Delete
                            </NButton>
                        </NSpace>
                    </div>
                </NCard>
            </div>
        </NCard>

        <NModal v-model:show="showCreateModal" preset="card" style="width: min(860px, calc(100vw - 32px))">
            <template #header>
                <div class="font-display text-lg uppercase tracking-wider">New Update (company)</div>
            </template>
            <NForm :show-feedback="false" label-placement="top">
                <div class="grid md:grid-cols-2 gap-3">
                    <NFormItem label="Tag">
                        <NInput v-model:value="form.tag" />
                    </NFormItem>
                    <NFormItem label="Ref">
                        <NInput v-model:value="form.ref" />
                    </NFormItem>
                    <NFormItem label="Status">
                        <NSelect v-model:value="form.status" :options="createStatusOptions as any" />
                    </NFormItem>
                </div>
                <NFormItem label="Title">
                    <NInput v-model:value="form.title" />
                </NFormItem>
                <NFormItem label="Body">
                    <NInput v-model:value="form.body" type="textarea" :autosize="{ minRows: 4, maxRows: 12 }" />
                </NFormItem>
                <NSpace justify="end" :size="12">
                    <NButton secondary :disabled="loading" @click="showCreateModal = false">Cancel</NButton>
                    <NButton type="primary" :loading="loading" :disabled="!form.title.trim()" @click="create">Create
                    </NButton>
                </NSpace>
            </NForm>
        </NModal>

        <NModal v-model:show="showEditModal" preset="card" style="width: min(860px, calc(100vw - 32px))">
            <template #header>
                <div class="font-display text-lg uppercase tracking-wider">Edit #{{ editingId }}</div>
            </template>
            <NForm :show-feedback="false" label-placement="top">
                <div class="grid md:grid-cols-2 gap-3">
                    <NFormItem label="Type">
                        <NSelect v-model:value="editForm.type" :options="editTypeOptions as any" />
                    </NFormItem>
                    <NFormItem label="Status">
                        <NSelect v-model:value="editForm.status" :options="editStatusOptions as any" />
                    </NFormItem>
                    <NFormItem label="Pinned Rank">
                        <NInputNumber v-model:value="editForm.pinnedRank" :min="0" />
                    </NFormItem>
                    <NFormItem label="Tag">
                        <NInput v-model:value="editForm.tag" />
                    </NFormItem>
                </div>

                <div class="grid md:grid-cols-2 gap-3">
                    <NFormItem label="Ref">
                        <NInput v-model:value="editForm.ref" />
                    </NFormItem>
                    <NFormItem label="Title">
                        <NInput v-model:value="editForm.title" />
                    </NFormItem>
                </div>

                <NFormItem label="Summary">
                    <NInput v-model:value="editForm.summary" type="textarea" :autosize="{ minRows: 2, maxRows: 6 }" />
                </NFormItem>
                <NFormItem label="Body">
                    <NInput v-model:value="editForm.body" type="textarea" :autosize="{ minRows: 4, maxRows: 12 }" />
                </NFormItem>

                <NSpace justify="end" :size="12">
                    <NButton secondary :disabled="loading" @click="cancelEdit">Cancel</NButton>
                    <NButton type="primary" :loading="loading" :disabled="!editingId" @click="saveEdit">Save</NButton>
                </NSpace>
            </NForm>
        </NModal>
    </div>
</template>
