<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

import { NButton, NCard, NGrid, NGridItem, NSpace } from 'naive-ui'

import { adminGet } from '@/admin/api'

const router = useRouter()
const { t } = useI18n()

const loading = ref(false)
const newLeads = ref(0)
const events7d = ref(0)
const events30d = ref(0)
const errorMsg = ref('')

const tzName = computed(() => {
    if (typeof Intl === 'undefined') return 'UTC'
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
})

const toIso = (d: Date) => {
    if (Number.isNaN(d.getTime())) return ''
    return d.toISOString()
}

const makeWindowIso = (days: number) => {
    const now = new Date()
    const from = new Date(now)
    from.setDate(from.getDate() - days)
    from.setHours(0, 0, 0, 0)

    const to = new Date(now)
    // Use end-of-day so list filtering feels natural.
    to.setHours(23, 59, 59, 999)

    return {
        from: toIso(from),
        to: toIso(to),
    }
}

const loadOverview = async (force = false) => {
    loading.value = true
    errorMsg.value = ''
    try {
        const qs = force ? '?force=true' : ''
        const leadRes = await adminGet<{ count: number }>(`/api/v1/admin/contacts/unread-count${qs}`)
        newLeads.value = typeof leadRes?.count === 'number' ? leadRes.count : Number((leadRes as any)?.count) || 0

        const m7 = await adminGet<any>(
            `/api/v1/admin/events/metrics?range=7d&tz=${encodeURIComponent(tzName.value)}${force ? '&force=true' : ''}`
        )
        events7d.value = Number(m7?.totals?.total ?? 0) || 0

        const m30 = await adminGet<any>(
            `/api/v1/admin/events/metrics?range=30d&tz=${encodeURIComponent(tzName.value)}${force ? '&force=true' : ''}`
        )
        events30d.value = Number(m30?.totals?.total ?? 0) || 0
    } catch {
        errorMsg.value = t('admin.home.overview.error')
    } finally {
        loading.value = false
    }
}

const onContactsChanged = () => {
    void loadOverview(true)
}

const go = async (name: string) => {
    await router.push({ name })
}

const goNewLeads = async () => {
    await router.push({ name: 'admin-contacts', query: { status: 'new' } })
}

const goEventsWindow = async (days: 7 | 30) => {
    const w = makeWindowIso(days)
    await router.push({
        name: 'admin-events',
        query: {
            range: `${days}d`,
            ...(w.from ? { from: w.from } : {}),
            ...(w.to ? { to: w.to } : {}),
        },
    })
}

onMounted(() => {
    void loadOverview()
    if (typeof window !== 'undefined') {
        window.addEventListener('admin:contacts:changed', onContactsChanged as any)
    }
})

onBeforeUnmount(() => {
    if (typeof window !== 'undefined') {
        window.removeEventListener('admin:contacts:changed', onContactsChanged as any)
    }
})
</script>

<template>
    <div class="max-w-6xl mx-auto">
        <NCard size="large" :bordered="true">
            <div class="flex items-start justify-between gap-4">
                <div class="min-w-0">
                    <div class="font-display text-xl uppercase tracking-wider">{{ t('admin.home.overview.title') }}
                    </div>
                    <div class="mt-2 font-mono text-xs uppercase tracking-[0.25em] text-black/50">
                        {{ t('admin.home.overview.subtitle') }}
                    </div>
                    <div v-if="errorMsg" class="mt-3 font-mono text-xs text-red-600">{{ errorMsg }}</div>
                </div>

                <NSpace align="center" :size="10">
                    <div class="font-mono text-xs text-black/50">TZ: {{ tzName }}</div>
                    <NButton size="small" secondary :loading="loading" @click="loadOverview(true)">
                        {{ t('admin.actions.refresh') }}
                    </NButton>
                </NSpace>
            </div>

            <div class="mt-5">
                <NGrid :cols="3" :x-gap="12" :y-gap="12">
                    <NGridItem>
                        <button type="button" class="w-full text-left border border-border p-4 hover:border-black/30"
                            @click="goNewLeads">
                            <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                                t('admin.home.overview.cards.newLeads') }}</div>
                            <div class="mt-2 font-display text-3xl tracking-wider">{{ newLeads }}</div>
                        </button>
                    </NGridItem>
                    <NGridItem>
                        <button type="button" class="w-full text-left border border-border p-4 hover:border-black/30"
                            @click="goEventsWindow(7)">
                            <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                                t('admin.home.overview.cards.events7d') }}</div>
                            <div class="mt-2 font-display text-3xl tracking-wider">{{ events7d }}</div>
                        </button>
                    </NGridItem>
                    <NGridItem>
                        <button type="button" class="w-full text-left border border-border p-4 hover:border-black/30"
                            @click="goEventsWindow(30)">
                            <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                                t('admin.home.overview.cards.events30d') }}</div>
                            <div class="mt-2 font-display text-3xl tracking-wider">{{ events30d }}</div>
                        </button>
                    </NGridItem>
                </NGrid>
            </div>
        </NCard>

        <div class="mt-4"></div>

        <NCard size="large" :bordered="true">
            <div class="font-display text-xl uppercase tracking-wider">{{ t('admin.home.quickActions') }}</div>
            <div class="mt-2 font-mono text-xs uppercase tracking-[0.25em] text-black/50">
                {{ t('admin.home.subtitle') }}
            </div>
        </NCard>

        <div class="mt-4">
            <NGrid :cols="2" :x-gap="12" :y-gap="12">
                <NGridItem>
                    <NCard hoverable @click="go('admin-products')">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.home.sections.catalog') }}</div>
                        <div class="mt-2 font-display text-lg uppercase tracking-wider">{{ t('admin.nav.products') }}
                        </div>
                    </NCard>
                </NGridItem>
                <NGridItem>
                    <NCard hoverable @click="go('admin-updates')">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.home.sections.content') }}</div>
                        <div class="mt-2 font-display text-lg uppercase tracking-wider">{{ t('admin.nav.updates') }}
                        </div>
                    </NCard>
                </NGridItem>
                <NGridItem>
                    <NCard hoverable @click="go('admin-contacts')">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.home.sections.leads') }}</div>
                        <div class="mt-2 font-display text-lg uppercase tracking-wider">{{ t('admin.nav.contacts') }}
                        </div>
                    </NCard>
                </NGridItem>
                <NGridItem>
                    <NCard hoverable @click="go('admin-events')">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.home.sections.analytics') }}</div>
                        <div class="mt-2 font-display text-lg uppercase tracking-wider">{{ t('admin.nav.events') }}
                        </div>
                    </NCard>
                </NGridItem>
            </NGrid>
        </div>
    </div>
</template>
