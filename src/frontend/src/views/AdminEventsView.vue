<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

import { adminDelete, adminGet } from '@/admin/api'
import EventsMetricsChart from '@/admin/components/EventsMetricsChart.vue'

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
const route = useRoute()
const { t } = useI18n()

const loading = ref(false)
const errorMsg = ref('')
const items = ref<EventItem[]>([])

const eventType = ref('')
const productId = ref('')

const fromLocal = ref('') // datetime-local
const toLocal = ref('')

const metricsRange = ref<'7d' | '30d' | '90d'>('7d')
const metricsLoading = ref(false)
const metricsError = ref('')
const metrics = ref<{
    tz: string
    series: Array<{ date: string; total: number; byType: Record<string, number> }>
    totals: { total: number; byType: Record<string, number> }
} | null>(null)

const tzName = computed(() => {
    if (typeof Intl === 'undefined') return 'UTC'
    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone
    return tz || 'UTC'
})

let updatingQuery = false
let initializing = true

const toRFC3339 = (localValue: string) => {
    const v = (localValue || '').trim()
    if (!v) return ''
    const d = new Date(v)
    if (Number.isNaN(d.getTime())) return ''
    return d.toISOString()
}

const fromRFC3339ToLocalInput = (iso: string) => {
    const v = (iso || '').trim()
    if (!v) return ''
    const d = new Date(v)
    if (Number.isNaN(d.getTime())) return ''
    // datetime-local expects local time without timezone.
    const local = new Date(d.getTime() - d.getTimezoneOffset() * 60_000)
    return local.toISOString().slice(0, 16)
}

const applyQueryToState = () => {
    const q = route.query
    const et = typeof q.event_type === 'string' ? q.event_type : Array.isArray(q.event_type) ? q.event_type[0] : ''
    const pid = typeof q.product_id === 'string' ? q.product_id : Array.isArray(q.product_id) ? q.product_id[0] : ''
    const from =
        typeof q.from === 'string' ? q.from : Array.isArray(q.from) && typeof q.from[0] === 'string' ? q.from[0] : ''
    const to = typeof q.to === 'string' ? q.to : Array.isArray(q.to) && typeof q.to[0] === 'string' ? q.to[0] : ''

    eventType.value = (et || '').trim()
    productId.value = (pid || '').trim()
    fromLocal.value = fromRFC3339ToLocalInput(from)
    toLocal.value = fromRFC3339ToLocalInput(to)
}

const syncStateToQuery = () => {
    const q: Record<string, any> = { ...route.query }
    const et = eventType.value.trim()
    const pid = productId.value.trim()
    const from = toRFC3339(fromLocal.value)
    const to = toRFC3339(toLocal.value)

    if (et) q.event_type = et
    else delete q.event_type

    if (pid) q.product_id = pid
    else delete q.product_id

    if (from) q.from = from
    else delete q.from

    if (to) q.to = to
    else delete q.to

    updatingQuery = true
    void router.replace({ query: q }).finally(() => {
        window.setTimeout(() => {
            updatingQuery = false
        }, 0)
    })
}

watch(
    () => [eventType.value, productId.value, fromLocal.value, toLocal.value],
    () => {
        if (initializing) return
        syncStateToQuery()
        void load()
        void loadMetrics()
    }
)

watch(
    () => route.fullPath,
    () => {
        if (updatingQuery) return
        initializing = true
        applyQueryToState()
        initializing = false
        void load()
        void loadMetrics()
    }
)

const load = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
        const qs = new URLSearchParams()
        qs.set('limit', '200')
        if (eventType.value.trim()) qs.set('event_type', eventType.value.trim())
        if (productId.value.trim()) qs.set('product_id', productId.value.trim())

        const from = toRFC3339(fromLocal.value)
        const to = toRFC3339(toLocal.value)
        if (from) qs.set('from', from)
        if (to) qs.set('to', to)

        const res = await adminGet<{ items: EventItem[] }>(`/api/v1/admin/events?${qs.toString()}`)
        items.value = res.items ?? []
    } catch (e) {
        errorMsg.value = t('admin.events.errors.load')
    } finally {
        loading.value = false
    }
}

const loadMetrics = async (force = false) => {
    metricsLoading.value = true
    metricsError.value = ''
    try {
        const qs = new URLSearchParams()
        qs.set('range', metricsRange.value)
        qs.set('tz', tzName.value)
        if (eventType.value.trim()) qs.set('event_type', eventType.value.trim())
        if (productId.value.trim()) qs.set('product_id', productId.value.trim())
        if (force) qs.set('force', 'true')

        const res = await adminGet<any>(`/api/v1/admin/events/metrics?${qs.toString()}`)
        metrics.value = {
            tz: res?.tz || tzName.value,
            series: Array.isArray(res?.series) ? res.series : [],
            totals: res?.totals || { total: 0, byType: {} },
        }
    } catch {
        metricsError.value = t('admin.events.metrics.errors.load')
        metrics.value = null
    } finally {
        metricsLoading.value = false
    }
}

const remove = async (id: number) => {
    if (!confirm(t('admin.events.confirmDelete', { id }))) return
    loading.value = true
    errorMsg.value = ''
    try {
        await adminDelete(`/api/v1/admin/events/${id}`)
        await load()
    } catch (e) {
        errorMsg.value = t('admin.events.errors.delete')
    } finally {
        loading.value = false
    }
}

watch(metricsRange, () => {
    void loadMetrics()
})

onMounted(() => {
    applyQueryToState()
    initializing = false
    void load()
    void loadMetrics()
})
</script>

<template>
    <main class="min-h-screen bg-white">
        <div class="px-6 py-10 max-w-6xl mx-auto">
            <div class="flex items-center justify-between">
                <h1 class="font-display text-2xl uppercase tracking-wider">{{ t('admin.nav.events') }}</h1>
                <router-link :to="{ name: 'admin-home' }" class="font-mono text-xs uppercase tracking-[0.25em]">←
                    {{ t('admin.events.back') }}</router-link>
            </div>

            <div class="mt-6">
                <div class="flex items-center justify-between gap-3">
                    <div class="font-display text-xl uppercase tracking-wider">{{ t('admin.events.metrics.title') }}
                    </div>
                    <div class="flex items-center gap-2">
                        <select v-model="metricsRange" class="h-9 px-2 border border-border font-mono text-xs">
                            <option value="7d">{{ t('admin.events.metrics.ranges.last7d') }}</option>
                            <option value="30d">{{ t('admin.events.metrics.ranges.last30d') }}</option>
                            <option value="90d">{{ t('admin.events.metrics.ranges.last90d') }}</option>
                        </select>
                        <button @click="loadMetrics(true)" :disabled="metricsLoading"
                            class="h-9 px-3 border border-border font-mono text-xs uppercase tracking-[0.25em] disabled:opacity-60">
                            {{ t('admin.events.metrics.refresh') }}
                        </button>
                    </div>
                </div>

                <p v-if="metricsError" class="mt-3 font-mono text-xs text-red-600">{{ metricsError }}</p>
                <div class="mt-3">
                    <EventsMetricsChart :title="t('admin.events.metrics.chartTitle')" :tz="metrics?.tz || tzName"
                        :series="metrics?.series || []" :loading="metricsLoading" />
                </div>

                <div class="mt-3 font-mono text-xs text-black/60" v-if="metrics">
                    {{ t('admin.events.metrics.total', { count: metrics.totals?.total ?? 0 }) }}
                </div>
            </div>

            <div class="mt-6 grid md:grid-cols-5 gap-3 border border-border p-4">
                <label class="block">
                    <div class="font-mono text-xs text-black/60">{{ t('admin.events.filters.eventType') }}</div>
                    <input v-model.trim="eventType" class="mt-1 w-full h-10 px-3 border border-border"
                        :placeholder="t('admin.events.filters.placeholderEventType')" />
                </label>
                <label class="block">
                    <div class="font-mono text-xs text-black/60">{{ t('admin.events.filters.productId') }}</div>
                    <input v-model.trim="productId" class="mt-1 w-full h-10 px-3 border border-border"
                        :placeholder="t('admin.events.filters.placeholderProductId')" />
                </label>
                <label class="block">
                    <div class="font-mono text-xs text-black/60">{{ t('admin.events.filters.from') }}</div>
                    <input v-model="fromLocal" type="datetime-local"
                        class="mt-1 w-full h-10 px-3 border border-border" />
                </label>
                <label class="block">
                    <div class="font-mono text-xs text-black/60">{{ t('admin.events.filters.to') }}</div>
                    <input v-model="toLocal" type="datetime-local" class="mt-1 w-full h-10 px-3 border border-border" />
                </label>
                <div class="flex items-end">
                    <button @click="load" :disabled="loading"
                        class="h-10 px-4 bg-brand text-white font-mono text-xs uppercase tracking-[0.25em] disabled:opacity-60">
                        {{ t('admin.events.search') }}
                    </button>
                </div>
            </div>

            <p v-if="errorMsg" class="mt-4 font-mono text-xs text-red-600">{{ errorMsg }}</p>

            <div class="mt-6 overflow-x-auto border border-border">
                <table class="min-w-full text-left font-mono text-xs">
                    <thead class="bg-border/30">
                        <tr>
                            <th class="p-3">{{ t('admin.events.table.time') }}</th>
                            <th class="p-3">{{ t('admin.events.table.type') }}</th>
                            <th class="p-3">{{ t('admin.events.table.product') }}</th>
                            <th class="p-3">{{ t('admin.events.table.page') }}</th>
                            <th class="p-3">{{ t('admin.events.table.anon') }}</th>
                            <th class="p-3">{{ t('admin.events.table.actions') }}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="e in items" :key="e.id" class="border-t border-border align-top">
                            <td class="p-3 whitespace-nowrap">{{ e.occurredAt?.slice(0, 19).replace('T', ' ') }}</td>
                            <td class="p-3 whitespace-nowrap">{{ e.eventType }}</td>
                            <td class="p-3 whitespace-nowrap">{{ e.productId ?? '-' }}</td>
                            <td class="p-3 min-w-[360px] text-black/60 break-all">{{ e.pageUrl }}</td>
                            <td class="p-3 whitespace-nowrap text-black/60">{{ e.anonId || e.sessionId || '-' }}</td>
                            <td class="p-3">
                                <button :disabled="loading" @click="remove(e.id)"
                                    class="h-8 px-3 border border-red-300 bg-white text-red-700 hover:border-red-500 transition-none disabled:opacity-60">
                                    {{ t('admin.actions.delete') }}
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </main>
</template>
