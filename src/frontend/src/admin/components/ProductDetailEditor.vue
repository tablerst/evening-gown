<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import {
    NAlert,
    NButton,
    NCard,
    NDivider,
    NInput,
    NSelect,
    NSpace,
    NTabPane,
    NTabs,
} from 'naive-ui'

// eslint-disable-next-line import/no-named-as-default
import draggable from 'vuedraggable'

import { HttpError, resolveApiUrl } from '@/api/http'
import { adminGetBlob } from '@/admin/api'
import { compressImageToWebpUnderLimit, uploadAdminImage } from '@/composables/useAdminImageUpload'
import { isValidStyleNo, normalizeStyleNo } from '@/utils/styleNo'

import { ensureDetailV2, safeParseJsonObject, stringifyPretty } from '@/admin/productDetail/detail'
import type { DetailSection, ProductDetailV2 } from '@/admin/productDetail/types'

const props = defineProps<{ json: string; styleNo: string; disabled?: boolean }>()
const emit = defineEmits<{ (e: 'update:json', v: string): void }>()

const { t } = useI18n()

const activeTab = ref<'visual' | 'json'>('visual')
const parseError = ref('')

const detail = ref<ProductDetailV2>(ensureDetailV2({}))

// Local-only gallery previews (do NOT persist into detail JSON).
// Keyed by gallery item id.
const galleryPreviewUrls = ref<Record<string, string>>({})

// Avoid feedback-loop: when we emit JSON to parent, parent updates props.json,
// but we should not re-parse/reload from props in that case.
const lastSelfEmittedJson = ref('')

const revokeAllGalleryPreviews = () => {
    for (const url of Object.values(galleryPreviewUrls.value)) {
        if (url) URL.revokeObjectURL(url)
    }
    galleryPreviewUrls.value = {}
}

const revokeGalleryPreview = (id: string) => {
    const key = String(id ?? '').trim()
    if (!key) return
    const url = galleryPreviewUrls.value[key]
    if (url) URL.revokeObjectURL(url)
    const next = { ...galleryPreviewUrls.value }
    delete next[key]
    galleryPreviewUrls.value = next
}

const galleryThumbSrc = (g: any): string => {
    const id = String(g?.id ?? '').trim()
    if (id) {
        const local = galleryPreviewUrls.value[id]
        if (local) return local
    }

    // Prefer the persisted public URL when available.
    // Note: for draft (unpublished) products, /api/v1/assets/* may 404 by design.
    // In that case, we fall back to fetching via protected /api/v1/admin/assets/*
    // and cache it into galleryPreviewUrls (see onGalleryImgError).
    const url = String(g?.url ?? '').trim()
    if (url) return resolveApiUrl(url)

    const objectKey = String(g?.objectKey ?? '').replace(/^\/+/, '').trim()
    if (objectKey) return resolveApiUrl(`/api/v1/assets/${objectKey}`)

    return ''
}

// Avoid repeated failing fetch attempts for the same image.
const galleryFetchFailures = ref<Record<string, boolean>>({})

const onGalleryImgError = async (g: any) => {
    const id = String(g?.id ?? '').trim()
    if (!id) return
    // If we already have a local/object preview, no need to recover.
    if (galleryPreviewUrls.value[id]) return
    if (galleryFetchFailures.value[id]) return

    const objectKey = String(g?.objectKey ?? '').replace(/^\/+/, '').trim()
    if (!objectKey) return

    try {
        const blob = await adminGetBlob(`/api/v1/admin/assets/${objectKey}`)
        revokeGalleryPreview(id)
        galleryPreviewUrls.value = {
            ...galleryPreviewUrls.value,
            [id]: URL.createObjectURL(blob),
        }
    } catch {
        // Mark as failed so we don't keep retrying on every re-render.
        galleryFetchFailures.value = {
            ...galleryFetchFailures.value,
            [id]: true,
        }
    }
}

const syncingFromParent = ref(false)
const syncingToParent = ref(false)

const loadFromJson = (raw: string) => {
    const parsed = safeParseJsonObject(raw)
    if (!parsed.ok) {
        parseError.value = t('admin.products.detailEditor.errors.parse')
        return
    }
    parseError.value = ''
    detail.value = ensureDetailV2(parsed.value)
}

watch(
    () => props.json,
    (next) => {
        if (syncingToParent.value) return
        if (String(next ?? '') === lastSelfEmittedJson.value) return
        syncingFromParent.value = true
        try {
            // If parent replaces the JSON, drop local preview URLs to avoid mismatches/leaks.
            revokeAllGalleryPreviews()
            loadFromJson(next)
        } finally {
            syncingFromParent.value = false
        }
    },
    { immediate: true },
)

onBeforeUnmount(() => {
    revokeAllGalleryPreviews()
})

const syncToJson = () => {
    if (syncingFromParent.value) return
    syncingToParent.value = true
    try {
        const s = stringifyPretty(detail.value)
        lastSelfEmittedJson.value = s
        emit('update:json', s)
    } finally {
        syncingToParent.value = false
    }
}

watch(
    detail,
    () => {
        // Keep parent JSON in sync (but avoid doing this while typing raw JSON).
        if (activeTab.value !== 'visual') return
        syncToJson()
    },
    { deep: true },
)

const canUploadGallery = computed(() => isValidStyleNo(props.styleNo))

const ensureArray = <T>(v: unknown, fallback: T[]): T[] => (Array.isArray(v) ? (v as T[]) : fallback)

const sections = computed<DetailSection[]>({
    get: () => ensureArray<DetailSection>((detail.value as any).sections, []),
    set: (next) => {
        detail.value.sections = next
    },
})

const addSection = (type: DetailSection['type']) => {
    const id = typeof crypto !== 'undefined' && 'randomUUID' in crypto ? crypto.randomUUID() : `sec_${Date.now()}`

    if (type === 'gallery') {
        sections.value.push({
            id,
            type: 'gallery',
            area: 'media',
            title_i18n: { zh: '画廊', en: 'Gallery' },
            props: { includeCoverHover: true },
        })
        return
    }

    if (type === 'options') {
        sections.value.push({
            id,
            type: 'options',
            area: 'sticky',
            title_i18n: { zh: '可选项', en: 'Options' },
        })
        return
    }

    if (type === 'specs') {
        sections.value.push({
            id,
            type: 'specs',
            area: 'main',
            title_i18n: { zh: '规格', en: 'Specs' },
        })
        return
    }

    if (type === 'service') {
        sections.value.push({
            id,
            type: 'service',
            area: 'aside',
            title_i18n: { zh: '服务', en: 'Service' },
        })
        return
    }

    if (type === 'divider') {
        sections.value.push({
            id,
            type: 'divider',
            area: 'main',
            title_i18n: { zh: '分隔线', en: 'Divider' },
        })
        return
    }

    if (type === 'richText') {
        sections.value.push({
            id,
            type: 'richText',
            area: 'main',
            title_i18n: { zh: '文本', en: 'Text' },
            data: { text_i18n: { zh: '', en: '' } },
        })
        return
    }
}

const removeSection = (id: string) => {
    const next = sections.value.filter((s) => s.id !== id)
    // Always keep at least one gallery so PDP doesn't look empty.
    if (!next.some((s) => s.type === 'gallery')) {
        next.unshift({
            id: typeof crypto !== 'undefined' && 'randomUUID' in crypto ? crypto.randomUUID() : `sec_${Date.now()}`,
            type: 'gallery',
            area: 'media',
            title_i18n: { zh: '画廊', en: 'Gallery' },
            props: { includeCoverHover: true },
        } as any)
    }
    sections.value = next
}

const areaOptions = [
    { label: t('admin.products.detailEditor.area.media'), value: 'media' },
    { label: t('admin.products.detailEditor.area.sticky'), value: 'sticky' },
    { label: t('admin.products.detailEditor.area.main'), value: 'main' },
    { label: t('admin.products.detailEditor.area.aside'), value: 'aside' },
]

const blockTypeOptions = [
    { label: t('admin.products.detailEditor.blocks.gallery'), value: 'gallery' },
    { label: t('admin.products.detailEditor.blocks.options'), value: 'options' },
    { label: t('admin.products.detailEditor.blocks.richText'), value: 'richText' },
    { label: t('admin.products.detailEditor.blocks.specs'), value: 'specs' },
    { label: t('admin.products.detailEditor.blocks.service'), value: 'service' },
    { label: t('admin.products.detailEditor.blocks.divider'), value: 'divider' },
]

const newBlockType = ref<DetailSection['type']>('richText')

const setI18n = (obj: any, key: string, locale: 'zh' | 'en', value: string) => {
    if (!obj[key] || typeof obj[key] !== 'object') obj[key] = {}
    obj[key][locale] = value
}

const gallery = computed<any[]>({
    get: () => {
        const g = (detail.value as any).gallery
        if (!Array.isArray(g)) return []
        // Ensure stable ids for draggable.
        for (const it of g) {
            if (!it || typeof it !== 'object') continue
            if (typeof (it as any).id === 'string' && String((it as any).id).trim()) continue
                ; (it as any).id = typeof crypto !== 'undefined' && 'randomUUID' in crypto ? crypto.randomUUID() : `img_${Date.now()}_${Math.random().toString(16).slice(2)}`
        }
        return g
    },
    set: (next) => {
        ; (detail.value as any).gallery = next
    },
})

const addGalleryItem = () => {
    gallery.value = [
        ...gallery.value,
        {
            id: typeof crypto !== 'undefined' && 'randomUUID' in crypto ? crypto.randomUUID() : `img_${Date.now()}`,
            url: '',
            objectKey: '',
        },
    ]
}

const removeGalleryItem = (idx: number) => {
    const next = gallery.value.slice()
    const removed = next[idx]
    const removedId = removed && typeof removed === 'object' ? String((removed as any).id ?? '') : ''
    if (removedId) revokeGalleryPreview(removedId)
    next.splice(idx, 1)
    gallery.value = next
}

const onUploadGallery = async (idx: number, e: Event) => {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    input.value = ''
    if (!file) return

    if (!canUploadGallery.value) return

    const styleNo = normalizeStyleNo(props.styleNo)

    try {
        const webp = await compressImageToWebpUnderLimit(file)

        const cur = gallery.value[idx]
        const curId = cur && typeof cur === 'object' ? String((cur as any).id ?? '') : ''
        if (curId) {
            revokeGalleryPreview(curId)
            galleryPreviewUrls.value = {
                ...galleryPreviewUrls.value,
                [curId]: URL.createObjectURL(webp),
            }
        }

        const res = await uploadAdminImage('gallery', styleNo, webp)
        const next = gallery.value.slice()
        next[idx] = { ...next[idx], url: res.url, objectKey: res.objectKey }
        gallery.value = next
    } catch (err) {
        if (err instanceof HttpError) {
            parseError.value = t('admin.products.detailEditor.errors.uploadHttp', { status: err.status })
        } else if (err instanceof Error) {
            parseError.value = err.message
        } else {
            parseError.value = t('admin.products.detailEditor.errors.upload')
        }
    }
}

const specs = computed<any[]>({
    get: () => {
        const raw = (detail.value as any).specs
        return Array.isArray(raw) ? raw : []
    },
    set: (next) => {
        ; (detail.value as any).specs = next
    },
})

const addSpecRow = () => {
    specs.value = [
        ...specs.value,
        {
            key: `spec_${specs.value.length + 1}`,
            label_i18n: { zh: '', en: '' },
            value_i18n: { zh: '', en: '' },
        },
    ]
}

const removeSpecRow = (idx: number) => {
    const next = specs.value.slice()
    next.splice(idx, 1)
    specs.value = next
}

const optionGroups = computed<any[]>({
    get: () => {
        const raw = (detail.value as any).option_groups
        return Array.isArray(raw) ? raw : []
    },
    set: (next) => {
        ; (detail.value as any).option_groups = next
    },
})

const addOptionGroup = () => {
    optionGroups.value = [
        ...optionGroups.value,
        {
            key: `group_${optionGroups.value.length + 1}`,
            name_i18n: { zh: '', en: '' },
            options: [],
        },
    ]
}

const removeOptionGroup = (idx: number) => {
    const next = optionGroups.value.slice()
    next.splice(idx, 1)
    optionGroups.value = next
}

const addOptionItem = (groupIdx: number) => {
    const next = optionGroups.value.slice()
    const g = { ...(next[groupIdx] ?? {}) }
    const opts = Array.isArray(g.options) ? g.options.slice() : []
    opts.push({
        key: `opt_${opts.length + 1}`,
        label_i18n: { zh: '', en: '' },
    })
    g.options = opts
    next[groupIdx] = g
    optionGroups.value = next
}

const removeOptionItem = (groupIdx: number, optIdx: number) => {
    const next = optionGroups.value.slice()
    const g = { ...(next[groupIdx] ?? {}) }
    const opts = Array.isArray(g.options) ? g.options.slice() : []
    opts.splice(optIdx, 1)
    g.options = opts
    next[groupIdx] = g
    optionGroups.value = next
}

const onJsonInput = (next: string) => {
    emit('update:json', next)
    // parse on-the-fly so user can flip back to visual mode safely.
    loadFromJson(next)
}

const resetToVisualNormalized = () => {
    detail.value = ensureDetailV2(detail.value)
    const s = stringifyPretty(detail.value)
    lastSelfEmittedJson.value = s
    emit('update:json', s)
    parseError.value = ''
}
</script>

<template>
    <NCard size="small" class="mt-3">
        <div class="flex items-center justify-between gap-3">
            <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">
                {{ t('admin.products.detailEditor.title') }}
            </div>
            <div class="flex items-center gap-2">
                <NSelect v-model:value="newBlockType" :options="blockTypeOptions" size="small" style="width: 220px"
                    :disabled="disabled" />
                <NButton size="small" secondary :disabled="disabled" @click="addSection(newBlockType)">
                    {{ t('admin.products.detailEditor.addBlock') }}
                </NButton>
            </div>
        </div>

        <NTabs v-model:value="activeTab" type="line" class="mt-3">
            <NTabPane name="visual" :tab="t('admin.products.detailEditor.tabs.visual')">
                <NAlert v-if="parseError" type="error" class="mb-3">{{ parseError }}</NAlert>

                <div class="text-xs text-black/60 leading-relaxed">
                    {{ t('admin.products.detailEditor.hint') }}
                </div>

                <NDivider class="my-4" />

                <div class="text-xs font-mono uppercase tracking-[0.25em] text-black/50">
                    {{ t('admin.products.detailEditor.sectionsTitle') }}
                </div>

                <div class="mt-2">
                    <draggable v-model="sections" item-key="id" handle=".drag-handle" :disabled="disabled"
                        class="space-y-3">
                        <template #item="{ element }">
                            <div class="border border-border bg-white p-3">
                                <div class="flex items-start justify-between gap-3">
                                    <div class="flex items-center gap-2">
                                        <span class="drag-handle cursor-grab select-none text-black/40">⋮⋮</span>
                                        <div class="font-mono text-xs uppercase tracking-[0.25em]">
                                            {{ element.type }}
                                        </div>
                                    </div>
                                    <div class="flex items-center gap-2">
                                        <NSelect v-model:value="(element as any).area" :options="areaOptions"
                                            size="tiny" style="width: 140px"
                                            :disabled="disabled || element.type === 'gallery' || element.type === 'service'" />
                                        <NButton size="tiny" secondary type="error" :disabled="disabled"
                                            @click="removeSection(element.id)">
                                            {{ t('admin.actions.delete') }}
                                        </NButton>
                                    </div>
                                </div>

                                <div class="mt-3 grid md:grid-cols-2 gap-3">
                                    <div>
                                        <div class="font-mono text-[10px] uppercase tracking-[0.25em] text-black/40">ZH
                                        </div>
                                        <NInput :value="(element as any).title_i18n?.zh || ''" :disabled="disabled"
                                            @update:value="(v) => { (element as any).title_i18n = (element as any).title_i18n || {}; (element as any).title_i18n.zh = v }"
                                            :placeholder="t('admin.products.detailEditor.placeholders.titleZh')" />
                                    </div>
                                    <div>
                                        <div class="font-mono text-[10px] uppercase tracking-[0.25em] text-black/40">EN
                                        </div>
                                        <NInput :value="(element as any).title_i18n?.en || ''" :disabled="disabled"
                                            @update:value="(v) => { (element as any).title_i18n = (element as any).title_i18n || {}; (element as any).title_i18n.en = v }"
                                            :placeholder="t('admin.products.detailEditor.placeholders.titleEn')" />
                                    </div>
                                </div>

                                <div v-if="element.type === 'gallery'" class="mt-4">
                                    <div class="flex items-center justify-between">
                                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">
                                            {{ t('admin.products.detailEditor.gallery.title') }}
                                        </div>
                                        <NButton size="small" secondary :disabled="disabled" @click="addGalleryItem">
                                            {{ t('admin.products.detailEditor.gallery.add') }}
                                        </NButton>
                                    </div>

                                    <div class="mt-2">
                                        <draggable v-model="gallery" item-key="id" handle=".drag-handle"
                                            :disabled="disabled" class="space-y-2">
                                            <template #item="{ element: g, index }">
                                                <div class="border border-border p-2">
                                                    <div class="flex items-start gap-3">
                                                        <div
                                                            class="w-16 h-16 border border-border bg-white overflow-hidden">
                                                            <img v-if="galleryThumbSrc(g)" :src="galleryThumbSrc(g)"
                                                                class="w-full h-full object-cover"
                                                                @error="() => onGalleryImgError(g)" />
                                                        </div>
                                                        <div class="flex-1 min-w-0">
                                                            <div class="flex items-center gap-2">
                                                                <span
                                                                    class="drag-handle cursor-grab select-none text-black/40">⋮⋮</span>
                                                                <div
                                                                    class="font-mono text-[10px] uppercase tracking-[0.25em] text-black/40">
                                                                    {{ t('admin.products.detailEditor.gallery.item') }}
                                                                    #{{ index + 1 }}
                                                                </div>
                                                            </div>
                                                            <div class="mt-2">
                                                                <NInput v-model:value="g.url" :disabled="disabled"
                                                                    :placeholder="t('admin.products.detailEditor.gallery.urlPlaceholder')" />
                                                            </div>
                                                            <div class="mt-2 flex items-center gap-3">
                                                                <input type="file" accept="image/*"
                                                                    :disabled="disabled || !canUploadGallery"
                                                                    @change="(e) => onUploadGallery(index, e)"
                                                                    class="block w-full text-xs" />
                                                                <span v-if="!canUploadGallery"
                                                                    class="text-xs text-black/50">
                                                                    {{
                                                                    t('admin.products.detailEditor.gallery.needStyleNo')
                                                                    }}
                                                                </span>
                                                            </div>
                                                        </div>
                                                        <div>
                                                            <NButton size="tiny" secondary type="error"
                                                                :disabled="disabled" @click="removeGalleryItem(index)">
                                                                {{ t('admin.actions.delete') }}
                                                            </NButton>
                                                        </div>
                                                    </div>
                                                </div>
                                            </template>
                                        </draggable>
                                    </div>
                                </div>

                                <div v-else-if="element.type === 'richText'" class="mt-4">
                                    <div class="grid md:grid-cols-2 gap-3">
                                        <div>
                                            <div
                                                class="font-mono text-[10px] uppercase tracking-[0.25em] text-black/40">
                                                ZH
                                            </div>
                                            <NInput :value="(element as any).data?.text_i18n?.zh || ''" type="textarea"
                                                :autosize="{ minRows: 4, maxRows: 10 }" :disabled="disabled"
                                                @update:value="(v) => { (element as any).data = (element as any).data || {}; (element as any).data.text_i18n = (element as any).data.text_i18n || {}; (element as any).data.text_i18n.zh = v }"
                                                :placeholder="t('admin.products.detailEditor.placeholders.textZh')" />
                                        </div>
                                        <div>
                                            <div
                                                class="font-mono text-[10px] uppercase tracking-[0.25em] text-black/40">
                                                EN
                                            </div>
                                            <NInput :value="(element as any).data?.text_i18n?.en || ''" type="textarea"
                                                :autosize="{ minRows: 4, maxRows: 10 }" :disabled="disabled"
                                                @update:value="(v) => { (element as any).data = (element as any).data || {}; (element as any).data.text_i18n = (element as any).data.text_i18n || {}; (element as any).data.text_i18n.en = v }"
                                                :placeholder="t('admin.products.detailEditor.placeholders.textEn')" />
                                        </div>
                                    </div>
                                </div>

                                <div v-else-if="element.type === 'specs'" class="mt-4">
                                    <div class="flex items-center justify-between">
                                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">
                                            {{ t('admin.products.detailEditor.specs.title') }}
                                        </div>
                                        <NButton size="small" secondary :disabled="disabled" @click="addSpecRow">
                                            {{ t('admin.products.detailEditor.specs.add') }}
                                        </NButton>
                                    </div>

                                    <div class="mt-2">
                                        <draggable v-model="specs" item-key="key" handle=".drag-handle"
                                            :disabled="disabled" class="space-y-2">
                                            <template #item="{ element: row, index }">
                                                <div class="border border-border p-2">
                                                    <div class="flex items-center justify-between gap-2">
                                                        <div class="flex items-center gap-2">
                                                            <span
                                                                class="drag-handle cursor-grab select-none text-black/40">⋮⋮</span>
                                                            <div
                                                                class="font-mono text-[10px] uppercase tracking-[0.25em] text-black/40">
                                                                {{ t('admin.products.detailEditor.specs.row') }} #{{
                                                                index + 1 }}
                                                            </div>
                                                        </div>
                                                        <NButton size="tiny" secondary type="error" :disabled="disabled"
                                                            @click="removeSpecRow(index)">
                                                            {{ t('admin.actions.delete') }}
                                                        </NButton>
                                                    </div>

                                                    <div class="mt-2 grid md:grid-cols-2 gap-3">
                                                        <NInput v-model:value="row.key" :disabled="disabled"
                                                            :placeholder="t('admin.products.detailEditor.specs.keyPlaceholder')" />
                                                        <div class="grid grid-cols-2 gap-2">
                                                            <NInput :value="row.label_i18n?.zh || ''"
                                                                :disabled="disabled"
                                                                @update:value="(v) => setI18n(row, 'label_i18n', 'zh', v)"
                                                                :placeholder="t('admin.products.detailEditor.specs.labelZh')" />
                                                            <NInput :value="row.label_i18n?.en || ''"
                                                                :disabled="disabled"
                                                                @update:value="(v) => setI18n(row, 'label_i18n', 'en', v)"
                                                                :placeholder="t('admin.products.detailEditor.specs.labelEn')" />
                                                        </div>
                                                    </div>

                                                    <div class="mt-2 grid grid-cols-2 gap-2">
                                                        <NInput :value="row.value_i18n?.zh || ''" :disabled="disabled"
                                                            @update:value="(v) => setI18n(row, 'value_i18n', 'zh', v)"
                                                            :placeholder="t('admin.products.detailEditor.specs.valueZh')" />
                                                        <NInput :value="row.value_i18n?.en || ''" :disabled="disabled"
                                                            @update:value="(v) => setI18n(row, 'value_i18n', 'en', v)"
                                                            :placeholder="t('admin.products.detailEditor.specs.valueEn')" />
                                                    </div>
                                                </div>
                                            </template>
                                        </draggable>
                                    </div>
                                </div>

                                <div v-else-if="element.type === 'options'" class="mt-4">
                                    <div class="flex items-center justify-between">
                                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">
                                            {{ t('admin.products.detailEditor.options.title') }}
                                        </div>
                                        <NButton size="small" secondary :disabled="disabled" @click="addOptionGroup">
                                            {{ t('admin.products.detailEditor.options.addGroup') }}
                                        </NButton>
                                    </div>

                                    <div class="mt-2 space-y-3">
                                        <div v-for="(g, gIdx) in optionGroups" :key="g.key"
                                            class="border border-border p-2">
                                            <div class="flex items-center justify-between gap-2">
                                                <div
                                                    class="font-mono text-[10px] uppercase tracking-[0.25em] text-black/40">
                                                    {{ t('admin.products.detailEditor.options.group') }} #{{ gIdx + 1 }}
                                                </div>
                                                <NButton size="tiny" secondary type="error" :disabled="disabled"
                                                    @click="removeOptionGroup(Number(gIdx))">
                                                    {{ t('admin.actions.delete') }}
                                                </NButton>
                                            </div>

                                            <div class="mt-2 grid md:grid-cols-2 gap-3">
                                                <NInput v-model:value="g.key" :disabled="disabled"
                                                    :placeholder="t('admin.products.detailEditor.options.groupKeyPlaceholder')" />
                                                <div class="grid grid-cols-2 gap-2">
                                                    <NInput :value="g.name_i18n?.zh || ''" :disabled="disabled"
                                                        @update:value="(v) => setI18n(g, 'name_i18n', 'zh', v)"
                                                        :placeholder="t('admin.products.detailEditor.options.groupNameZh')" />
                                                    <NInput :value="g.name_i18n?.en || ''" :disabled="disabled"
                                                        @update:value="(v) => setI18n(g, 'name_i18n', 'en', v)"
                                                        :placeholder="t('admin.products.detailEditor.options.groupNameEn')" />
                                                </div>
                                            </div>

                                            <div class="mt-3 flex items-center justify-between">
                                                <div
                                                    class="font-mono text-[10px] uppercase tracking-[0.25em] text-black/40">
                                                    {{ t('admin.products.detailEditor.options.items') }}
                                                </div>
                                                <NButton size="tiny" secondary :disabled="disabled"
                                                    @click="addOptionItem(Number(gIdx))">
                                                    {{ t('admin.products.detailEditor.options.addItem') }}
                                                </NButton>
                                            </div>

                                            <div class="mt-2 space-y-2">
                                                <div v-for="(opt, optIdx) in (g.options || [])" :key="opt.key"
                                                    class="border border-border p-2">
                                                    <div class="flex items-center justify-between gap-2">
                                                        <div
                                                            class="font-mono text-[10px] uppercase tracking-[0.25em] text-black/40">
                                                            {{ t('admin.products.detailEditor.options.item') }} #{{
                                                            Number(optIdx) + 1 }}
                                                        </div>
                                                        <NButton size="tiny" secondary type="error" :disabled="disabled"
                                                            @click="removeOptionItem(Number(gIdx), Number(optIdx))">
                                                            {{ t('admin.actions.delete') }}
                                                        </NButton>
                                                    </div>

                                                    <div class="mt-2 grid md:grid-cols-2 gap-3">
                                                        <NInput v-model:value="opt.key" :disabled="disabled"
                                                            :placeholder="t('admin.products.detailEditor.options.itemKeyPlaceholder')" />
                                                        <div class="grid grid-cols-2 gap-2">
                                                            <NInput :value="opt.label_i18n?.zh || ''"
                                                                :disabled="disabled"
                                                                @update:value="(v) => setI18n(opt, 'label_i18n', 'zh', v)"
                                                                :placeholder="t('admin.products.detailEditor.options.itemLabelZh')" />
                                                            <NInput :value="opt.label_i18n?.en || ''"
                                                                :disabled="disabled"
                                                                @update:value="(v) => setI18n(opt, 'label_i18n', 'en', v)"
                                                                :placeholder="t('admin.products.detailEditor.options.itemLabelEn')" />
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </template>
                    </draggable>
                </div>
            </NTabPane>

            <NTabPane name="json" :tab="t('admin.products.detailEditor.tabs.json')">
                <NAlert type="warning" class="mb-3">{{ t('admin.products.detailEditor.jsonWarning') }}</NAlert>

                <NInput :value="json" type="textarea" :autosize="{ minRows: 10, maxRows: 22 }" class="font-mono text-xs"
                    :disabled="disabled" @update:value="onJsonInput" />

                <div class="mt-3 flex items-center justify-end gap-2">
                    <NButton size="small" secondary :disabled="disabled" @click="resetToVisualNormalized">
                        {{ t('admin.products.detailEditor.normalize') }}
                    </NButton>
                </div>
            </NTabPane>
        </NTabs>
    </NCard>
</template>
