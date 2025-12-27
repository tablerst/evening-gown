<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

import { adminDelete, adminGet, adminPatch } from '@/admin/api'

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
const route = useRoute()
const { t } = useI18n()

const loading = ref(false)
const errorMsg = ref('')
const items = ref<ContactLead[]>([])

const filterStatus = ref<'all' | 'new' | 'contacted' | 'closed'>('all')

let updatingQuery = false
let syncingFromRoute = false

const parseStatusQuery = (v: unknown) => {
    const raw = typeof v === 'string' ? v : Array.isArray(v) ? v[0] : ''
    const st = (raw || '').trim()
    if (st === 'new' || st === 'contacted' || st === 'closed') return st
    return 'all'
}

watch(
    () => route.query.status,
    (v) => {
        if (updatingQuery) return
        const next = parseStatusQuery(v)
        if (next !== filterStatus.value) {
            syncingFromRoute = true
            filterStatus.value = next
            syncingFromRoute = false
        }
    }
)

watch(filterStatus, (v) => {
    if (!syncingFromRoute) {
        const q: Record<string, any> = { ...route.query }
        if (v === 'all') delete q.status
        else q.status = v
        updatingQuery = true
        void router.replace({ query: q }).finally(() => {
            window.setTimeout(() => {
                updatingQuery = false
            }, 0)
        })
    }
    void load()
})

const dispatchContactsChanged = () => {
    if (typeof window === 'undefined') return
    window.dispatchEvent(new Event('admin:contacts:changed'))
}

const copied = ref<{ id: number; field: 'phone' | 'wechat' } | null>(null)
let copiedTimer: number | null = null

const copyText = async (id: number, field: 'phone' | 'wechat', text: string) => {
    const v = (text || '').trim()
    if (!v) return
    try {
        await navigator.clipboard.writeText(v)
        copied.value = { id, field }
        if (copiedTimer) window.clearTimeout(copiedTimer)
        copiedTimer = window.setTimeout(() => {
            copied.value = null
            copiedTimer = null
        }, 1500)
    } catch {
        // Ignore; clipboard may be unavailable depending on context.
    }
}

const load = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
        const qs = new URLSearchParams()
        qs.set('limit', '100')
        if (filterStatus.value !== 'all') qs.set('status', filterStatus.value)

        const res = await adminGet<{ items: ContactLead[] }>(`/api/v1/admin/contacts?${qs.toString()}`)
        items.value = res.items ?? []
    } catch {
        errorMsg.value = t('admin.contacts.errors.load')
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
        dispatchContactsChanged()
    } catch {
        errorMsg.value = t('admin.contacts.errors.update')
    } finally {
        loading.value = false
    }
}

const remove = async (id: number) => {
    if (!confirm(t('admin.contacts.confirmDelete', { id }))) return
    loading.value = true
    errorMsg.value = ''
    try {
        await adminDelete(`/api/v1/admin/contacts/${id}`)
        await load()
        dispatchContactsChanged()
    } catch {
        errorMsg.value = t('admin.contacts.errors.delete')
    } finally {
        loading.value = false
    }
}

onMounted(() => {
    const init = parseStatusQuery(route.query.status)
    syncingFromRoute = true
    filterStatus.value = init
    syncingFromRoute = false
    void load()
})

onBeforeUnmount(() => {
    if (copiedTimer) window.clearTimeout(copiedTimer)
    copiedTimer = null
})
</script>

<template>
    <main class="min-h-screen bg-white">
        <div class="px-6 py-10 max-w-6xl mx-auto">
            <div class="flex items-center justify-between">
                <h1 class="font-display text-2xl uppercase tracking-wider">{{ t('admin.nav.contacts') }}</h1>
                <router-link :to="{ name: 'admin-home' }" class="font-mono text-xs uppercase tracking-[0.25em]">←
                    {{ t('admin.contacts.back') }}</router-link>
            </div>

            <div class="mt-6 flex items-center justify-between gap-4">
                <div class="flex items-center gap-3">
                    <span class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                        t('admin.contacts.filters.status') }}</span>
                    <select v-model="filterStatus" class="h-9 px-2 border border-border font-mono text-xs">
                        <option value="all">{{ t('admin.contacts.filters.all') }}</option>
                        <option value="new">{{ t('admin.contacts.filters.new') }}</option>
                        <option value="contacted">{{ t('admin.contacts.filters.contacted') }}</option>
                        <option value="closed">{{ t('admin.contacts.filters.closed') }}</option>
                    </select>
                    <button @click="load"
                        class="h-9 px-3 border border-border font-mono text-xs uppercase tracking-[0.25em]">{{
                            t('admin.actions.refresh') }}</button>
                </div>
                <div class="font-mono text-xs text-black/60">{{ t('admin.contacts.items', { count: items.length }) }}
                </div>
            </div>

            <p v-if="errorMsg" class="mt-4 font-mono text-xs text-red-600">{{ errorMsg }}</p>

            <div class="mt-6 overflow-x-auto border border-border">
                <table class="min-w-full text-left font-mono text-xs">
                    <thead class="bg-border/30">
                        <tr>
                            <th class="p-3">{{ t('admin.contacts.table.id') }}</th>
                            <th class="p-3">{{ t('admin.contacts.table.created') }}</th>
                            <th class="p-3">{{ t('admin.contacts.table.name') }}</th>
                            <th class="p-3">{{ t('admin.contacts.table.phone') }}</th>
                            <th class="p-3">{{ t('admin.contacts.table.wechat') }}</th>
                            <th class="p-3">{{ t('admin.contacts.table.message') }}</th>
                            <th class="p-3">{{ t('admin.contacts.table.source') }}</th>
                            <th class="p-3 whitespace-nowrap">{{ t('admin.contacts.table.status') }}</th>
                            <th class="p-3 whitespace-nowrap min-w-[260px]">{{ t('admin.contacts.table.actions') }}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="c in items" :key="c.id" class="border-t border-border">
                            <td class="p-3">{{ c.id }}</td>
                            <td class="p-3 whitespace-nowrap">{{ c.createdAt?.slice(0, 19).replace('T', ' ') }}</td>
                            <td class="p-3 whitespace-nowrap">{{ c.name }}</td>

                            <td class="p-3 whitespace-nowrap">
                                <div class="flex items-center gap-2">
                                    <span>{{ c.phone || '-' }}</span>
                                    <button v-if="c.phone"
                                        class="h-7 px-2 border border-border text-black/70 whitespace-nowrap"
                                        @click="copyText(c.id, 'phone', c.phone)">
                                        {{ t('admin.contacts.actions.copy') }}
                                    </button>
                                    <span v-if="copied?.id === c.id && copied?.field === 'phone'"
                                        class="text-[11px] text-green-700">
                                        {{ t('admin.contacts.actions.copied') }}
                                    </span>
                                </div>
                            </td>

                            <td class="p-3 whitespace-nowrap">
                                <div class="flex items-center gap-2">
                                    <span>{{ c.wechat || '-' }}</span>
                                    <button v-if="c.wechat"
                                        class="h-7 px-2 border border-border text-black/70 whitespace-nowrap"
                                        @click="copyText(c.id, 'wechat', c.wechat)">
                                        {{ t('admin.contacts.actions.copy') }}
                                    </button>
                                    <span v-if="copied?.id === c.id && copied?.field === 'wechat'"
                                        class="text-[11px] text-green-700">
                                        {{ t('admin.contacts.actions.copied') }}
                                    </span>
                                </div>
                            </td>

                            <td class="p-3 min-w-[240px] text-black/70 align-top">{{ c.message }}</td>
                            <td class="p-3 min-w-[200px] text-black/60 align-top">{{ c.sourcePage }}</td>

                            <td class="p-3 whitespace-nowrap">
                                <select :disabled="loading" :value="c.status"
                                    @change="setStatus(c.id, ($event.target as HTMLSelectElement).value as any)"
                                    class="h-9 px-2 border border-border font-mono text-xs">
                                    <option value="new">{{ t('admin.contacts.filters.new') }}</option>
                                    <option value="contacted">{{ t('admin.contacts.filters.contacted') }}</option>
                                    <option value="closed">{{ t('admin.contacts.filters.closed') }}</option>
                                </select>
                            </td>

                            <td class="p-3 whitespace-nowrap min-w-[260px]">
                                <div class="flex items-center gap-2 flex-nowrap whitespace-nowrap">
                                    <button :disabled="loading" @click="remove(c.id)"
                                        class="h-8 px-3 border border-red-300 bg-white text-red-700 hover:border-red-500 transition-none disabled:opacity-60 whitespace-nowrap">
                                        {{ t('admin.actions.delete') }}
                                    </button>
                                    <button v-if="c.status !== 'contacted'" :disabled="loading"
                                        @click="setStatus(c.id, 'contacted')"
                                        class="h-8 px-3 border border-border bg-white text-black/70 hover:border-black/40 transition-none disabled:opacity-60 whitespace-nowrap">
                                        {{ t('admin.contacts.actions.markContacted') }}
                                    </button>
                                    <button v-if="c.status !== 'closed'" :disabled="loading"
                                        @click="setStatus(c.id, 'closed')"
                                        class="h-8 px-3 border border-border bg-white text-black/70 hover:border-black/40 transition-none disabled:opacity-60 whitespace-nowrap">
                                        {{ t('admin.contacts.actions.close') }}
                                    </button>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <div v-if="!loading && items.length === 0" class="mt-6 border border-border p-6 text-center">
                <div class="font-display text-lg uppercase tracking-wider">{{ t('admin.contacts.emptyTitle') }}</div>
                <div class="mt-2 font-mono text-xs text-black/60">{{ t('admin.contacts.emptyBody') }}</div>
                <button @click="load"
                    class="mt-4 h-9 px-3 border border-border font-mono text-xs uppercase tracking-[0.25em]">
                    {{ t('admin.actions.refresh') }}
                </button>
            </div>
        </div>
    </main>
</template>
