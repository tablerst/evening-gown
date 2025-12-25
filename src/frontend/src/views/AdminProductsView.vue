<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

import { NButton, NCard, NForm, NFormItem, NInput, NInputNumber, NModal, NSpace, NSwitch } from 'naive-ui'

import { HttpError, resolveApiUrl } from '@/api/http'
import { adminDelete, adminGet, adminGetBlob, adminPatch, adminPost } from '@/admin/api'
import { appEnv } from '@/config/env'
import { compressImageToWebpUnderLimit, uploadAdminImage, type UploadKind } from '@/composables/useAdminImageUpload'
import { compareStyleNo, isValidStyleNo, normalizeStyleNo } from '@/utils/styleNo'

type Product = {
    id: number
    slug?: string
    styleNo: string
    season: string
    category: string
    availability: string
    isNew: boolean
    newRank: number
    coverImage: string
    coverImageKey?: string
    hoverImage: string
    hoverImageKey?: string
    detail?: any
    publishedAt?: string
    deletedAt?: string
}

const router = useRouter()
const { t } = useI18n()

const loading = ref(false)
const errorMsg = ref('')
const products = ref<Product[]>([])
const total = ref(0)

const PAGE_LIMIT = 100

const filterStatus = ref<'all' | 'draft' | 'published'>('all')
const filterSeason = ref<'all' | string>('all')
const filterCategory = ref<'all' | string>('all')
const filterIsNew = ref<'all' | 'true' | 'false'>('all')
const filterAvailability = ref<'all' | string>('all')
const keyword = ref('')
const sortBy = ref<'default' | 'style_asc' | 'style_desc'>('default')

const showCreateModal = ref(false)
const showEditModal = ref(false)

const editingId = ref<number | null>(null)
const editForm = ref({
    slug: '',
    styleNo: '',
    season: 'ss25',
    category: 'gown',
    availability: 'in_stock',
    isNew: false,
    newRank: 0,
    coverImage: '',
    coverImageKey: '',
    hoverImage: '',
    hoverImageKey: '',
    detailJson: '{"specs":[{"k":"件数","v":""},{"k":"交付时间","v":""}],"option_groups":[{"name":"颜色","options":[]},{"name":"尺码","options":[]}]}',
})

const form = ref({
    styleNo: '',
    season: 'ss25',
    category: 'gown',
    availability: 'in_stock',
    isNew: false,
    newRank: 0,
    coverImage: '',
    coverImageKey: '',
    hoverImage: '',
    hoverImageKey: '',
    detailJson: '{"specs":[{"k":"件数","v":""},{"k":"交付时间","v":""}],"option_groups":[{"name":"颜色","options":[]},{"name":"尺码","options":[]}]}',
})

type UploadSlotState = {
    uploading: boolean
    error: string
    previewUrl: string
}

const createUpload = ref<Record<UploadKind, UploadSlotState>>({
    cover: { uploading: false, error: '', previewUrl: '' },
    hover: { uploading: false, error: '', previewUrl: '' },
})

const editUpload = ref<Record<UploadKind, UploadSlotState>>({
    cover: { uploading: false, error: '', previewUrl: '' },
    hover: { uploading: false, error: '', previewUrl: '' },
})

const revokePreview = (slot: UploadSlotState) => {
    if (slot.previewUrl) URL.revokeObjectURL(slot.previewUrl)
    slot.previewUrl = ''
}

const maxUploadHint = computed(() => `${Math.max(1, Math.round((appEnv.maxImageUploadBytes ?? 1048576) / 1024 / 1024))}MB`)

const onPickImage = async (scope: 'create' | 'edit', kind: UploadKind, e: Event) => {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    // reset input so picking the same file again still triggers change
    input.value = ''
    if (!file) return

    const targetForm = scope === 'create' ? form.value : editForm.value
    const state = scope === 'create' ? createUpload.value : editUpload.value
    const slot = state[kind]

    slot.error = ''
    slot.uploading = true
    try {
        if (!isValidStyleNo(targetForm.styleNo)) {
            slot.error = t('admin.products.upload.needStyleNo')
            return
        }

        // Normalize once so uploads/keys match backend expectations.
        const normalizedStyleNo = normalizeStyleNo(targetForm.styleNo)
        targetForm.styleNo = normalizedStyleNo

        const webp = await compressImageToWebpUnderLimit(file, {
            maxBytes: appEnv.maxImageUploadBytes ?? 1048576,
        })

        revokePreview(slot)
        slot.previewUrl = URL.createObjectURL(webp)

        const res = await uploadAdminImage(kind, normalizedStyleNo, webp)

        if (kind === 'cover') {
            targetForm.coverImage = res.url
            targetForm.coverImageKey = res.objectKey
        } else {
            targetForm.hoverImage = res.url
            targetForm.hoverImageKey = res.objectKey
        }
    } catch (err) {
        if (err instanceof HttpError) {
            slot.error = t('admin.products.upload.failHttp', { status: err.status })
        } else if (err instanceof Error) {
            slot.error = err.message
        } else {
            slot.error = t('admin.products.upload.fail')
        }
    } finally {
        slot.uploading = false
    }
}

const canSubmit = computed(() => isValidStyleNo(form.value.styleNo))

const buildListQuery = (offset: number) => {
    const qs = new URLSearchParams()
    qs.set('limit', String(PAGE_LIMIT))
    qs.set('offset', String(Math.max(0, offset)))

    if (filterStatus.value !== 'all') qs.set('status', filterStatus.value)
    if (filterSeason.value !== 'all') qs.set('season', filterSeason.value)
    if (filterCategory.value !== 'all') qs.set('category', filterCategory.value)
    if (filterIsNew.value !== 'all') qs.set('is_new', filterIsNew.value)

    return qs
}

const load = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
        const qs = buildListQuery(0)
        const res = await adminGet<{ items: Product[]; total: number }>(`/api/v1/admin/products?${qs.toString()}`)
        products.value = res.items ?? []
        total.value = Number(res.total ?? 0)
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = t('admin.products.errors.load')
    } finally {
        loading.value = false
    }
}

const loadMore = async () => {
    if (loading.value) return
    if (products.value.length >= total.value) return
    loading.value = true
    errorMsg.value = ''
    try {
        const qs = buildListQuery(products.value.length)
        const res = await adminGet<{ items: Product[]; total: number }>(`/api/v1/admin/products?${qs.toString()}`)
        const incoming = Array.isArray(res.items) ? res.items : []
        const existingIds = new Set(products.value.map((p) => p.id))
        for (const item of incoming) {
            if (!existingIds.has(item.id)) products.value.push(item)
        }
        total.value = Number(res.total ?? total.value)
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = t('admin.products.errors.load')
    } finally {
        loading.value = false
    }
}

const resetFilters = async () => {
    filterStatus.value = 'all'
    filterSeason.value = 'all'
    filterCategory.value = 'all'
    filterIsNew.value = 'all'
    filterAvailability.value = 'all'
    keyword.value = ''
    sortBy.value = 'default'
    await load()
}

const normalizeText = (v: unknown) => String(v ?? '').trim().toLowerCase()

const seasonLabel = (season: string) => {
    const s = String(season ?? '').trim().toLowerCase()
    if (s === 'ss25') return t('admin.products.filters.season.ss25')
    if (s === 'fw25') return t('admin.products.filters.season.fw25')
    return season
}

const categoryLabel = (category: string) => {
    const c = String(category ?? '').trim().toLowerCase()
    if (c === 'gown') return t('admin.products.filters.category.gown')
    if (c === 'couture') return t('admin.products.filters.category.couture')
    if (c === 'bridal') return t('admin.products.filters.category.bridal')
    return category
}

const availabilityLabel = (availability: string) => {
    const a = String(availability ?? '').trim().toLowerCase()
    if (a === 'in_stock') return t('admin.products.filters.availability.in_stock')
    if (a === 'preorder') return t('admin.products.filters.availability.preorder')
    if (a === 'archived') return t('admin.products.filters.availability.archived')
    return availability
}

const filteredProducts = computed(() => {
    const kw = normalizeText(keyword.value)
    const avail = filterAvailability.value

    let items = products.value.slice()

    // Client-only filtering (keyword + availability)
    if (avail !== 'all') {
        items = items.filter((p) => p.availability === avail)
    }
    if (kw) {
        items = items.filter((p) => {
            const hay = [
                p.id,
                p.styleNo,
                p.slug ?? '',
                p.season,
                p.category,
                p.availability,
            ]
                .map(normalizeText)
                .join(' ')
            return hay.includes(kw)
        })
    }

    // Client-only sort
    if (sortBy.value === 'style_asc') {
        items.sort((a, b) => compareStyleNo(a.styleNo, b.styleNo))
    } else if (sortBy.value === 'style_desc') {
        items.sort((a, b) => compareStyleNo(b.styleNo, a.styleNo))
    }

    return items
})

const shownCount = computed(() => filteredProducts.value.length)

let autoReloadTimer: number | null = null
let pendingAutoReload = false
const scheduleAutoReload = () => {
    pendingAutoReload = true
    if (typeof window === 'undefined') return
    if (autoReloadTimer) window.clearTimeout(autoReloadTimer)
    autoReloadTimer = window.setTimeout(() => {
        autoReloadTimer = null
        if (loading.value) {
            // Still busy (e.g. loadMore). Keep one pending refresh queued.
            scheduleAutoReload()
            return
        }
        pendingAutoReload = false
        load()
    }, 250)
}

watch([filterStatus, filterSeason, filterCategory, filterIsNew], () => {
    scheduleAutoReload()
})

onBeforeUnmount(() => {
    if (autoReloadTimer) window.clearTimeout(autoReloadTimer)
})

const create = async () => {
    if (!canSubmit.value) return
    loading.value = true
    errorMsg.value = ''
    try {
        const normalizedStyleNo = normalizeStyleNo(form.value.styleNo)
        if (!isValidStyleNo(normalizedStyleNo)) {
            errorMsg.value = t('admin.products.upload.needStyleNo')
            return
        }
        form.value.styleNo = normalizedStyleNo

        const detail = JSON.parse(form.value.detailJson)
        await adminPost('/api/v1/admin/products', {
            styleNo: normalizedStyleNo,
            season: form.value.season,
            category: form.value.category,
            availability: form.value.availability,
            isNew: form.value.isNew,
            newRank: form.value.newRank,
            coverImage: form.value.coverImage,
            coverImageKey: form.value.coverImageKey,
            hoverImage: form.value.hoverImage,
            hoverImageKey: form.value.hoverImageKey,
            detail,
        })
        await load()

        showCreateModal.value = false
    } catch {
        errorMsg.value = t('admin.products.errors.create')
    } finally {
        loading.value = false
    }
}

const loadDraftAssetPreview = async (kind: UploadKind, objectKey: string) => {
    const key = String(objectKey ?? '').replace(/^\/+/, '')
    if (!key) return

    const slot = editUpload.value[kind]
    // If a local preview already exists (e.g. user picked a new file), don't override it.
    if (slot.previewUrl) return

    slot.error = ''
    slot.uploading = true
    try {
        const blob = await adminGetBlob(`/api/v1/admin/assets/${key}`)
        revokePreview(slot)
        slot.previewUrl = URL.createObjectURL(blob)
    } catch (err) {
        if (err instanceof HttpError) {
            slot.error = t('admin.products.upload.failHttp', { status: err.status })
        } else if (err instanceof Error) {
            slot.error = err.message
        } else {
            slot.error = t('admin.products.upload.fail')
        }
    } finally {
        slot.uploading = false
    }
}

const startEdit = async (id: number) => {
    loading.value = true
    errorMsg.value = ''
    try {
        const p = await adminGet<Product>(`/api/v1/admin/products/${id}`)
        editingId.value = id
        editForm.value = {
            slug: p.slug ?? '',
            styleNo: normalizeStyleNo(p.styleNo),
            season: p.season,
            category: p.category,
            availability: p.availability,
            isNew: !!p.isNew,
            newRank: p.newRank ?? 0,
            coverImage: p.coverImage ?? '',
            coverImageKey: p.coverImageKey ?? '',
            hoverImage: p.hoverImage ?? '',
            hoverImageKey: p.hoverImageKey ?? '',
            detailJson: JSON.stringify(p.detail ?? { specs: [], option_groups: [] }),
        }

        showEditModal.value = true

        // Draft images are not publicly readable; load via protected admin assets endpoint.
        const isPublished = !!p.publishedAt
        if (!isPublished) {
            if (p.coverImageKey) await loadDraftAssetPreview('cover', p.coverImageKey)
            if (p.hoverImageKey) await loadDraftAssetPreview('hover', p.hoverImageKey)
        }
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = t('admin.products.errors.loadDetail')
    } finally {
        loading.value = false
    }
}

const cancelEdit = () => {
    editingId.value = null
    showEditModal.value = false

    revokePreview(editUpload.value.cover)
    revokePreview(editUpload.value.hover)
    editUpload.value.cover.error = ''
    editUpload.value.hover.error = ''
}

const saveEdit = async () => {
    if (!editingId.value) return
    loading.value = true
    errorMsg.value = ''
    try {
        const normalizedStyleNo = normalizeStyleNo(editForm.value.styleNo)
        if (!isValidStyleNo(normalizedStyleNo)) {
            errorMsg.value = t('admin.products.upload.needStyleNo')
            return
        }
        editForm.value.styleNo = normalizedStyleNo

        let detail: any = undefined
        try {
            detail = JSON.parse(editForm.value.detailJson)
        } catch {
            errorMsg.value = t('admin.products.errors.detailJson')
            return
        }

        await adminPatch(`/api/v1/admin/products/${editingId.value}`, {
            slug: editForm.value.slug,
            styleNo: normalizedStyleNo,
            season: editForm.value.season,
            category: editForm.value.category,
            availability: editForm.value.availability,
            isNew: editForm.value.isNew,
            newRank: editForm.value.newRank,
            coverImage: editForm.value.coverImage,
            coverImageKey: editForm.value.coverImageKey,
            hoverImage: editForm.value.hoverImage,
            hoverImageKey: editForm.value.hoverImageKey,
            detail,
        })

        editingId.value = null
        showEditModal.value = false
        await load()
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = t('admin.products.errors.save')
    } finally {
        loading.value = false
    }
}

const remove = async (id: number) => {
    if (!confirm(t('admin.products.confirmDelete', { id }))) return
    loading.value = true
    errorMsg.value = ''
    try {
        await adminDelete(`/api/v1/admin/products/${id}`)
        if (editingId.value === id) editingId.value = null
        await load()
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = t('admin.products.errors.delete')
    } finally {
        loading.value = false
    }
}

const togglePublish = async (id: number, next: 'publish' | 'unpublish') => {
    loading.value = true
    errorMsg.value = ''
    try {
        await adminPost(`/api/v1/admin/products/${id}/${next}`)
        await load()
    } catch (e) {
        if (e instanceof HttpError && (e.status === 401 || e.status === 403)) {
            await router.replace({ name: 'admin-login' })
            return
        }
        errorMsg.value = next === 'publish' ? t('admin.products.errors.publish') : t('admin.products.errors.unpublish')
    } finally {
        loading.value = false
    }
}

onMounted(load)
</script>

<template>
    <div class="max-w-6xl mx-auto">
        <NCard size="large">
            <NSpace justify="space-between" align="center" :wrap="true">
                <div class="flex flex-col gap-2">
                    <NSpace align="center" :size="12" :wrap="true">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.products.filters.status') }}</div>
                        <select v-model="filterStatus" class="h-9 px-2 border border-border font-mono text-xs">
                            <option value="all">{{ t('admin.products.filters.all') }}</option>
                            <option value="draft">{{ t('admin.products.filters.draft') }}</option>
                            <option value="published">{{ t('admin.products.filters.published') }}</option>
                        </select>

                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.products.filters.season.label') }}</div>
                        <select v-model="filterSeason" class="h-9 px-2 border border-border font-mono text-xs">
                            <option value="all">{{ t('admin.products.filters.season.all') }}</option>
                            <option value="ss25">{{ t('admin.products.filters.season.ss25') }}</option>
                            <option value="fw25">{{ t('admin.products.filters.season.fw25') }}</option>
                        </select>

                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.products.filters.category.label') }}</div>
                        <select v-model="filterCategory" class="h-9 px-2 border border-border font-mono text-xs">
                            <option value="all">{{ t('admin.products.filters.category.all') }}</option>
                            <option value="gown">{{ t('admin.products.filters.category.gown') }}</option>
                            <option value="couture">{{ t('admin.products.filters.category.couture') }}</option>
                            <option value="bridal">{{ t('admin.products.filters.category.bridal') }}</option>
                        </select>

                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.products.filters.isNew.label') }}</div>
                        <select v-model="filterIsNew" class="h-9 px-2 border border-border font-mono text-xs">
                            <option value="all">{{ t('admin.products.filters.isNew.all') }}</option>
                            <option value="true">{{ t('admin.products.filters.isNew.true') }}</option>
                            <option value="false">{{ t('admin.products.filters.isNew.false') }}</option>
                        </select>

                        <NButton size="small" :loading="loading" secondary @click="load">{{ t('admin.actions.refresh')
                        }}
                        </NButton>
                        <NButton size="small" secondary :disabled="loading" @click="resetFilters">{{
                            t('admin.products.filters.reset') }}</NButton>
                    </NSpace>

                    <NSpace align="center" :size="12" :wrap="true">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.products.filters.keyword') }}</div>
                        <NInput v-model:value="keyword" size="small" clearable
                            :placeholder="t('admin.products.filters.keywordPlaceholder')" class="w-[260px]" />

                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.products.filters.availability.label') }}</div>
                        <select v-model="filterAvailability" class="h-9 px-2 border border-border font-mono text-xs">
                            <option value="all">{{ t('admin.products.filters.availability.all') }}</option>
                            <option value="in_stock">{{ t('admin.products.filters.availability.in_stock') }}</option>
                            <option value="preorder">{{ t('admin.products.filters.availability.preorder') }}</option>
                            <option value="archived">{{ t('admin.products.filters.availability.archived') }}</option>
                        </select>

                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.products.filters.sort.label') }}</div>
                        <select v-model="sortBy" class="h-9 px-2 border border-border font-mono text-xs">
                            <option value="default">{{ t('admin.products.filters.sort.default') }}</option>
                            <option value="style_asc">{{ t('admin.products.filters.sort.styleAsc') }}</option>
                            <option value="style_desc">{{ t('admin.products.filters.sort.styleDesc') }}</option>
                        </select>

                        <div class="font-mono text-xs text-black/50">
                            {{ t('admin.products.filters.count', { shown: shownCount, total }) }}
                        </div>
                    </NSpace>
                </div>
                <NButton size="small" type="primary" @click="showCreateModal = true">{{ t('admin.products.new') }}
                </NButton>
            </NSpace>

            <p v-if="errorMsg" class="mt-3 font-mono text-xs text-red-600">{{ errorMsg }}</p>

            <div class="mt-4">
                <div class="[column-gap:16px] columns-1 sm:columns-2 xl:columns-3">
                    <div v-for="p in filteredProducts" :key="p.id"
                        class="mb-4 break-inside-avoid border border-border bg-white">
                        <div class="relative">
                            <div class="aspect-[4/5] bg-border/30 overflow-hidden">
                                <img v-if="p.coverImage" :src="resolveApiUrl(p.coverImage)" :alt="String(p.styleNo)"
                                    class="h-full w-full object-cover" loading="lazy" />
                                <div v-else class="h-full w-full flex items-center justify-center">
                                    <div class="font-mono text-xs text-black/40">{{ t('admin.products.card.noCover') }}
                                    </div>
                                </div>
                            </div>

                            <div class="absolute left-2 top-2 flex gap-2">
                                <span v-if="p.isNew"
                                    class="px-2 py-1 bg-black text-white font-mono text-[10px] tracking-[0.25em] uppercase">
                                    {{ t('admin.products.card.badgeNew') }}
                                </span>
                                <span v-if="p.publishedAt"
                                    class="px-2 py-1 bg-white/90 border border-border font-mono text-[10px] tracking-[0.25em] uppercase">
                                    {{ t('admin.products.card.badgePublished') }}
                                </span>
                                <span v-else
                                    class="px-2 py-1 bg-white/90 border border-border font-mono text-[10px] tracking-[0.25em] uppercase">
                                    {{ t('admin.products.card.badgeDraft') }}
                                </span>
                            </div>
                        </div>

                        <div class="p-3">
                            <div class="flex items-start justify-between gap-3">
                                <div>
                                    <div class="font-display text-sm uppercase tracking-wider">
                                        {{ t('admin.products.card.style', { styleNo: p.styleNo }) }}
                                    </div>
                                    <div class="mt-1 font-mono text-xs text-black/60">
                                        <span class="mr-3">ID: {{ p.id }}</span>
                                        <span class="mr-3">{{ seasonLabel(p.season) }}</span>
                                        <span class="mr-3">{{ categoryLabel(p.category) }}</span>
                                        <span>{{ availabilityLabel(p.availability) }}</span>
                                    </div>
                                </div>

                                <NSpace :size="8" align="center">
                                    <NButton size="tiny" secondary :disabled="loading" @click="startEdit(p.id)">{{
                                        t('admin.actions.edit') }}</NButton>
                                    <NButton v-if="!p.publishedAt" size="tiny" :disabled="loading"
                                        @click="togglePublish(p.id, 'publish')">{{ t('admin.actions.publish') }}
                                    </NButton>
                                    <NButton v-else size="tiny" secondary :disabled="loading"
                                        @click="togglePublish(p.id, 'unpublish')">{{ t('admin.actions.unpublish') }}
                                    </NButton>
                                    <NButton size="tiny" type="error" secondary :disabled="loading"
                                        @click="remove(p.id)">{{ t('admin.actions.delete') }}</NButton>
                                </NSpace>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="mt-4 flex items-center justify-center gap-3">
                    <NButton size="small" secondary :disabled="loading || products.length >= total" @click="loadMore">
                        {{ t('admin.products.actions.loadMore') }}
                    </NButton>
                </div>
            </div>
        </NCard>

        <NModal v-model:show="showCreateModal" preset="card" style="width: min(860px, calc(100vw - 32px))">
            <template #header>
                <div class="font-display text-lg uppercase tracking-wider">{{ t('admin.products.modal.createTitle') }}
                </div>
            </template>
            <NForm :show-feedback="false" label-placement="top">
                <div class="grid md:grid-cols-2 gap-3">
                    <NFormItem :label="t('admin.products.fields.styleNo')">
                        <NInput v-model:value="form.styleNo" placeholder="EG-1001" :maxlength="64" />
                    </NFormItem>
                    <NFormItem :label="t('admin.products.fields.season')">
                        <NInput v-model:value="form.season" />
                    </NFormItem>
                    <NFormItem :label="t('admin.products.fields.category')">
                        <NInput v-model:value="form.category" />
                    </NFormItem>
                    <NFormItem :label="t('admin.products.fields.availability')">
                        <NInput v-model:value="form.availability" />
                    </NFormItem>
                    <NFormItem :label="t('admin.products.fields.isNew')">
                        <NSwitch v-model:value="form.isNew" />
                    </NFormItem>
                    <NFormItem :label="t('admin.products.fields.newRank')">
                        <NInputNumber v-model:value="form.newRank" :min="0" />
                    </NFormItem>
                </div>

                <div class="grid md:grid-cols-2 gap-3">
                    <div>
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.products.fields.coverImage') }}</div>
                        <div class="mt-2">
                            <NInput v-model:value="form.coverImage" @input="form.coverImageKey = ''"
                                :placeholder="t('admin.products.upload.url')" />
                        </div>
                        <div class="mt-2 flex items-center gap-3">
                            <input type="file" accept="image/*" :disabled="loading || createUpload.cover.uploading"
                                @change="(e) => onPickImage('create', 'cover', e)" class="block w-full text-xs" />
                            <span v-if="createUpload.cover.uploading" class="font-mono text-xs text-black/60">{{
                                t('admin.products.upload.uploading') }}</span>
                        </div>
                        <p class="mt-1 font-mono text-xs text-black/40">{{ t('admin.products.upload.hint', {
                            size:
                                maxUploadHint
                        })
                        }}</p>
                        <p v-if="createUpload.cover.error" class="mt-1 font-mono text-xs text-red-600">{{
                            createUpload.cover.error
                        }}</p>
                        <div v-if="createUpload.cover.previewUrl || form.coverImage" class="mt-2">
                            <img :src="createUpload.cover.previewUrl || resolveApiUrl(form.coverImage)"
                                class="h-14 w-14 object-cover border border-border" />
                        </div>
                    </div>
                    <div>
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.products.fields.hoverImage') }}</div>
                        <div class="mt-2">
                            <NInput v-model:value="form.hoverImage" @input="form.hoverImageKey = ''"
                                :placeholder="t('admin.products.upload.url')" />
                        </div>
                        <div class="mt-2 flex items-center gap-3">
                            <input type="file" accept="image/*" :disabled="loading || createUpload.hover.uploading"
                                @change="(e) => onPickImage('create', 'hover', e)" class="block w-full text-xs" />
                            <span v-if="createUpload.hover.uploading" class="font-mono text-xs text-black/60">{{
                                t('admin.products.upload.uploading') }}</span>
                        </div>
                        <p class="mt-1 font-mono text-xs text-black/40">{{ t('admin.products.upload.hint', {
                            size:
                                maxUploadHint
                        })
                        }}</p>
                        <p v-if="createUpload.hover.error" class="mt-1 font-mono text-xs text-red-600">{{
                            createUpload.hover.error
                        }}</p>
                        <div v-if="createUpload.hover.previewUrl || form.hoverImage" class="mt-2">
                            <img :src="createUpload.hover.previewUrl || resolveApiUrl(form.hoverImage)"
                                class="h-14 w-14 object-cover border border-border" />
                        </div>
                    </div>
                </div>

                <NFormItem :label="t('admin.products.fields.detailJson')" class="mt-3">
                    <NInput v-model:value="form.detailJson" type="textarea" :autosize="{ minRows: 6, maxRows: 14 }"
                        class="font-mono text-xs" />
                </NFormItem>

                <NSpace justify="end" :size="12">
                    <NButton secondary :disabled="loading" @click="showCreateModal = false">{{ t('admin.actions.cancel')
                        }}
                    </NButton>
                    <NButton type="primary" :loading="loading" :disabled="!canSubmit" @click="create">{{
                        t('admin.actions.create')
                        }}</NButton>
                </NSpace>
            </NForm>
        </NModal>

        <NModal v-model:show="showEditModal" preset="card" style="width: min(860px, calc(100vw - 32px))">
            <template #header>
                <div class="font-display text-lg uppercase tracking-wider">{{ t('admin.products.modal.editTitle', {
                    id:
                        editingId
                }) }}</div>
            </template>
            <NForm :show-feedback="false" label-placement="top">
                <div class="grid md:grid-cols-2 gap-3">
                    <NFormItem :label="t('admin.products.fields.slug')">
                        <NInput v-model:value="editForm.slug" />
                    </NFormItem>
                    <NFormItem :label="t('admin.products.fields.styleNo')">
                        <NInput v-model:value="editForm.styleNo" placeholder="EG-1001" :maxlength="64" />
                    </NFormItem>
                    <NFormItem :label="t('admin.products.fields.season')">
                        <NInput v-model:value="editForm.season" />
                    </NFormItem>
                    <NFormItem :label="t('admin.products.fields.category')">
                        <NInput v-model:value="editForm.category" />
                    </NFormItem>
                    <NFormItem :label="t('admin.products.fields.availability')">
                        <NInput v-model:value="editForm.availability" />
                    </NFormItem>
                    <NFormItem :label="t('admin.products.fields.isNew')">
                        <NSwitch v-model:value="editForm.isNew" />
                    </NFormItem>
                    <NFormItem :label="t('admin.products.fields.newRank')">
                        <NInputNumber v-model:value="editForm.newRank" :min="0" />
                    </NFormItem>
                </div>

                <div class="grid md:grid-cols-2 gap-3">
                    <div>
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.products.fields.coverImage') }}</div>
                        <div class="mt-2">
                            <NInput v-model:value="editForm.coverImage" @input="editForm.coverImageKey = ''"
                                :placeholder="t('admin.products.upload.url')" />
                        </div>
                        <div class="mt-2 flex items-center gap-3">
                            <input type="file" accept="image/*" :disabled="loading || editUpload.cover.uploading"
                                @change="(e) => onPickImage('edit', 'cover', e)" class="block w-full text-xs" />
                            <span v-if="editUpload.cover.uploading" class="font-mono text-xs text-black/60">{{
                                t('admin.products.upload.uploading') }}</span>
                        </div>
                        <p class="mt-1 font-mono text-xs text-black/40">{{ t('admin.products.upload.hint', {
                            size:
                                maxUploadHint
                        }) }}</p>
                        <p v-if="editUpload.cover.error" class="mt-1 font-mono text-xs text-red-600">{{
                            editUpload.cover.error
                        }}</p>
                        <div v-if="editUpload.cover.previewUrl || editForm.coverImage" class="mt-2">
                            <img :src="editUpload.cover.previewUrl || resolveApiUrl(editForm.coverImage)"
                                class="h-14 w-14 object-cover border border-border" />
                        </div>
                    </div>
                    <div>
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{
                            t('admin.products.fields.hoverImage') }}</div>
                        <div class="mt-2">
                            <NInput v-model:value="editForm.hoverImage" @input="editForm.hoverImageKey = ''"
                                :placeholder="t('admin.products.upload.url')" />
                        </div>
                        <div class="mt-2 flex items-center gap-3">
                            <input type="file" accept="image/*" :disabled="loading || editUpload.hover.uploading"
                                @change="(e) => onPickImage('edit', 'hover', e)" class="block w-full text-xs" />
                            <span v-if="editUpload.hover.uploading" class="font-mono text-xs text-black/60">{{
                                t('admin.products.upload.uploading') }}</span>
                        </div>
                        <p class="mt-1 font-mono text-xs text-black/40">{{ t('admin.products.upload.hint', {
                            size:
                                maxUploadHint
                        }) }}</p>
                        <p v-if="editUpload.hover.error" class="mt-1 font-mono text-xs text-red-600">{{
                            editUpload.hover.error
                        }}</p>
                        <div v-if="editUpload.hover.previewUrl || editForm.hoverImage" class="mt-2">
                            <img :src="editUpload.hover.previewUrl || resolveApiUrl(editForm.hoverImage)"
                                class="h-14 w-14 object-cover border border-border" />
                        </div>
                    </div>
                </div>

                <NFormItem :label="t('admin.products.fields.detailJson')" class="mt-3">
                    <NInput v-model:value="editForm.detailJson" type="textarea" :autosize="{ minRows: 6, maxRows: 14 }"
                        class="font-mono text-xs" />
                </NFormItem>

                <NSpace justify="end" :size="12">
                    <NButton secondary :disabled="loading" @click="cancelEdit">{{ t('admin.actions.cancel') }}</NButton>
                    <NButton type="primary" :loading="loading" :disabled="!editingId" @click="saveEdit">{{
                        t('admin.actions.save') }}</NButton>
                </NSpace>
            </NForm>
        </NModal>
    </div>
</template>
