<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

import { HttpError, httpGet, httpPost, resolveApiUrl } from '@/api/http'
import { normalizeStyleNo } from '@/utils/styleNo'

type ProductDetail = {
    id: number
    slug: string
    styleNo: string
    season: string
    category: string
    availability: string
    coverImage: string
    hoverImage: string
    isNew: boolean
    priceMode: string
    priceText: string
    detail: unknown
}

type SpecItem = { label: string; value: string }
type OptionItem = { key: string; label: string }
type OptionGroup = { key: string; name: string; options: OptionItem[] }

type DetailSection = {
    id?: string
    type?: string
    area?: string
    title_i18n?: unknown
    data?: unknown
}

const { t, te, locale } = useI18n()
const route = useRoute()
const router = useRouter()

const productId = computed(() => Number(route.params.id))

const loading = ref(false)
const errorMsg = ref('')
const product = ref<ProductDetail | null>(null)

// Interactive option selections (single-select per option group).
// Keyed by option group key (stable) or fallback to the displayed name.
const selectedOptions = ref<Record<string, string>>({})

const posterOpen = ref(false)
const posterDataUrl = ref('')
const posterError = ref('')

const linkHint = ref('')

const getOrCreateAnonId = () => {
    if (typeof window === 'undefined') return ''
    const key = 'anon_id'
    const existing = window.localStorage.getItem(key) ?? ''
    if (existing) return existing
    const next = typeof crypto !== 'undefined' && 'randomUUID' in crypto ? crypto.randomUUID() : String(Date.now())
    window.localStorage.setItem(key, next)
    return next
}

const readUtm = () => {
    if (typeof window === 'undefined') return {}
    const u = new URL(window.location.href)
    const p = u.searchParams
    return {
        utm_source: p.get('utm_source') ?? '',
        utm_medium: p.get('utm_medium') ?? '',
        utm_campaign: p.get('utm_campaign') ?? '',
        utm_content: p.get('utm_content') ?? '',
        utm_term: p.get('utm_term') ?? '',
    }
}

const load = async () => {
    errorMsg.value = ''
    loading.value = true
    try {
        if (!productId.value || Number.isNaN(productId.value)) {
            errorMsg.value = t('productDetail.error')
            return
        }
        const raw = await httpGet<ProductDetail>(`/api/v1/products/${productId.value}`)
        product.value = {
            ...raw,
            styleNo: normalizeStyleNo((raw as any)?.styleNo ?? (raw as any)?.style_no ?? ''),
            coverImage: resolveApiUrl(raw.coverImage),
            hoverImage: resolveApiUrl(raw.hoverImage),
        }
    } catch (e) {
        if (e instanceof HttpError && e.status === 404) errorMsg.value = 'Not Found'
        else errorMsg.value = t('productDetail.error')
    } finally {
        loading.value = false
    }
}

const asRecord = (v: unknown): Record<string, unknown> | null => {
    if (!v || typeof v !== 'object') return null
    if (Array.isArray(v)) return null
    return v as Record<string, unknown>
}

const pickLocalizedText = (v: unknown) => {
    if (typeof v === 'string') return v
    const obj = asRecord(v)
    if (!obj) return ''
    const direct = obj[locale.value]
    if (typeof direct === 'string') return direct
    const zh = obj.zh
    const en = obj.en
    if (typeof zh === 'string') return zh
    if (typeof en === 'string') return en
    return ''
}

const detailObj = computed<Record<string, unknown>>(() => asRecord(product.value?.detail) ?? {})

const descriptionText = computed(() => {
    const d = detailObj.value
    // support both plain string and *_i18n object
    const candidates = [d.description_i18n, d.description, d.desc_i18n, d.desc]
    for (const c of candidates) {
        const s = pickLocalizedText(c)
        if (s) return s
    }
    return ''
})

const normalizeSpecItem = (raw: unknown): SpecItem | null => {
    const obj = asRecord(raw)
    if (!obj) return null
    const label =
        pickLocalizedText(obj.label_i18n) ||
        String(obj.label ?? obj.k ?? obj.key ?? obj.name ?? '').trim()
    const value =
        pickLocalizedText(obj.value_i18n) ||
        String(obj.value ?? obj.v ?? obj.val ?? '').trim()
    if (!label || !value) return null
    return { label, value }
}

const specs = computed<SpecItem[]>(() => {
    const raw = detailObj.value.specs
    if (!Array.isArray(raw)) return []
    return raw.map(normalizeSpecItem).filter(Boolean) as SpecItem[]
})

const normalizeOptionGroup = (raw: unknown): OptionGroup | null => {
    const obj = asRecord(raw)
    if (!obj) return null

    const key = String(obj.key ?? obj.id ?? obj.name ?? obj.title ?? obj.label ?? '').trim()
    const name = pickLocalizedText(obj.name_i18n) || String(obj.name ?? obj.title ?? obj.label ?? '').trim()

    const optionsRaw = obj.options
    const options: OptionItem[] = []
    if (Array.isArray(optionsRaw)) {
        for (const opt of optionsRaw) {
            if (typeof opt === 'string') {
                const s = opt.trim()
                if (s) options.push({ key: s, label: s })
                continue
            }
            const optObj = asRecord(opt)
            if (!optObj) continue

            const optKey = String(optObj.key ?? optObj.id ?? optObj.value ?? optObj.name ?? optObj.label ?? '').trim()
            const optLabel =
                pickLocalizedText(optObj.label_i18n) ||
                String(optObj.label ?? optObj.name ?? optObj.value ?? '').trim()

            if (optKey && optLabel) options.push({ key: optKey, label: optLabel })
        }
    }
    if (!name || options.length === 0) return null
    return { key: key || name, name, options }
}

const optionGroups = computed<OptionGroup[]>(() => {
    const raw = detailObj.value.option_groups ?? detailObj.value.optionGroups
    if (!Array.isArray(raw)) return []
    return raw.map(normalizeOptionGroup).filter(Boolean) as OptionGroup[]
})

const isOptionSelected = (groupKey: string, optKey: string) => (selectedOptions.value[groupKey] ?? '') === optKey

const toggleOption = (groupKey: string, optKey: string) => {
    const cur = selectedOptions.value[groupKey] ?? ''
    // Click again to clear.
    selectedOptions.value[groupKey] = cur === optKey ? '' : optKey
}

watch(
    () => product.value?.id,
    () => {
        // Reset selections when switching products.
        selectedOptions.value = {}
    },
)

const effectiveSections = computed<DetailSection[]>(() => {
    const raw = (detailObj.value as any).sections
    if (Array.isArray(raw) && raw.length) return raw as DetailSection[]

    // Legacy fallback: keep the old layout order.
    return [
        { id: 'gallery', type: 'gallery', area: 'media' },
        { id: 'options', type: 'options', area: 'sticky' },
        { id: 'overview', type: 'richText', area: 'main' },
        { id: 'specs', type: 'specs', area: 'main' },
        { id: 'service', type: 'service', area: 'aside' },
    ]
})

const sectionsByArea = computed(() => {
    const by = {
        media: [] as DetailSection[],
        sticky: [] as DetailSection[],
        main: [] as DetailSection[],
        aside: [] as DetailSection[],
    }

    for (const s of effectiveSections.value) {
        const area = String(s?.area ?? '').trim()
        if (area === 'media' || area === 'sticky' || area === 'main' || area === 'aside') {
            by[area].push(s)
        } else {
            by.main.push(s)
        }
    }
    return by
})

const sectionTitle = (s: DetailSection, fallbackKey: string) => {
    const title = pickLocalizedText((s as any)?.title_i18n)
    return title || t(fallbackKey)
}

const galleryUrls = computed(() => {
    const urls: string[] = []

    const push = (u: unknown) => {
        const s = String(u ?? '').trim()
        if (!s) return
        const resolved = resolveApiUrl(s)
        if (!resolved) return
        if (!urls.includes(resolved)) urls.push(resolved)
    }

    const d = detailObj.value

    // Always include cover/hover first.
    push(product.value?.coverImage)
    push(product.value?.hoverImage)

    const rawGallery = d.gallery ?? d.images
    if (Array.isArray(rawGallery)) {
        for (const item of rawGallery) {
            if (typeof item === 'string') {
                push(item)
                continue
            }
            const obj = asRecord(item)
            if (!obj) continue
            push(obj.url ?? obj.src)
        }
    }

    return urls
})

const displayPriceText = computed(() => {
    const mode = String(product.value?.priceMode ?? '').toLowerCase()
    if (mode === 'negotiable' || mode === 'inquiry' || mode === 'on_request') return t('product.login')
    const s = String(product.value?.priceText ?? '').trim()
    return s || t('product.login')
})

const safeUpper = (v: unknown) => String(v ?? '').trim().toUpperCase()

const labelFromKeyOrFallback = (key: string, fallback: string) => (te(key) ? t(key) : fallback)

const seasonLabel = (season: unknown) => {
    const s = String(season ?? '').trim().toLowerCase()
    return labelFromKeyOrFallback(`product.filters.season.${s}`, safeUpper(season))
}

const categoryLabel = (category: unknown) => {
    const c = String(category ?? '').trim().toLowerCase()
    return labelFromKeyOrFallback(`product.filters.category.${c}`, safeUpper(category))
}

const availabilityLabel = (availability: unknown) => {
    const a = String(availability ?? '').trim().toLowerCase()
    return labelFromKeyOrFallback(`product.filters.availability.${a}`, safeUpper(availability))
}

const SERVICE_ITEM_KEYS = ['appointment', 'leadTime', 'showroom', 'shipping'] as const

const serviceItems = computed<string[]>(() => {
    const items: string[] = []
    for (const k of SERVICE_ITEM_KEYS) {
        const key = `productDetail.service.items.${k}`
        if (!te(key)) continue
        const s = String(t(key) ?? '').trim()
        if (s) items.push(s)
    }
    return items
})

const upsertMetaTag = (selector: string, attrs: Record<string, string>, content: string) => {
    if (typeof document === 'undefined') return
    let el = document.querySelector(selector) as HTMLMetaElement | null
    if (!el) {
        el = document.createElement('meta')
        for (const [k, v] of Object.entries(attrs)) el.setAttribute(k, v)
        document.head.appendChild(el)
    }
    el.setAttribute('content', content)
}

const buildDynamicTitle = () => {
    const styleNo = product.value?.styleNo
    if (!styleNo) return ''
    const prefix = t('productDetail.titlePrefix')
    const brand = 'FLEURLIS'
    return `${prefix} #${styleNo} · ${brand}`
}

const buildMetaDescription = () => {
    const base = descriptionText.value || t('productDetail.metaDescriptionFallback')
    const normalized = String(base ?? '').replace(/\s+/g, ' ').trim()
    if (normalized.length <= 160) return normalized
    return normalized.slice(0, 157) + '…'
}

const syncHead = () => {
    if (typeof document === 'undefined') return

    const title = buildDynamicTitle()
    if (title) document.title = title

    const desc = buildMetaDescription()
    if (desc) {
        upsertMetaTag('meta[name="description"]', { name: 'description' }, desc)
        upsertMetaTag('meta[property="og:title"]', { property: 'og:title' }, title || document.title)
        upsertMetaTag('meta[property="og:description"]', { property: 'og:description' }, desc)
    }
}

watch([() => product.value?.styleNo, locale], async () => {
    // Ensure this runs after router's locale watcher so our PDP title wins.
    await nextTick()
    syncHead()
}, { flush: 'post' })

const copyLink = async () => {
    linkHint.value = ''
    if (typeof window === 'undefined') return
    const url = window.location.href

    try {
        if (navigator?.clipboard?.writeText) {
            await navigator.clipboard.writeText(url)
        } else {
            const input = document.createElement('input')
            input.value = url
            input.style.position = 'fixed'
            input.style.left = '-9999px'
            document.body.appendChild(input)
            input.select()
            document.execCommand('copy')
            document.body.removeChild(input)
        }
        linkHint.value = t('productDetail.copySuccess')
    } catch {
        linkHint.value = t('productDetail.copyFail')
    }
    window.setTimeout(() => {
        linkHint.value = ''
    }, 2000)
}

const loadImage = (src: string) =>
    new Promise<HTMLImageElement>((resolve, reject) => {
        const img = new Image()
        img.crossOrigin = 'anonymous'
        img.onload = () => resolve(img)
        img.onerror = () => reject(new Error('image load failed'))
        img.src = src
    })

const buildPoster = async (
    p: ProductDetail,
    extra: {
        selectedOptionLines: string[]
        specsLines: string[]
    },
) => {
    const w = 1080
    const h = 1920
    const padding = 60

    const canvas = document.createElement('canvas')
    canvas.width = w
    canvas.height = h
    const ctx = canvas.getContext('2d')
    if (!ctx) throw new Error('no canvas ctx')

    // 1. Background
    ctx.fillStyle = '#FFFFFF'
    ctx.fillRect(0, 0, w, h)

    // 2. Outer Frame
    ctx.strokeStyle = '#000000'
    ctx.lineWidth = 2
    ctx.strokeRect(padding, padding, w - padding * 2, h - padding * 2)

    // 3. Header
    ctx.fillStyle = '#000000'
    ctx.font = 'bold 90px "Times New Roman", serif'
    ctx.textAlign = 'left'
    ctx.fillText('FLEURLIS', padding + 40, padding + 120)

    const dateStr = new Date().toLocaleDateString('en-GB').toUpperCase()
    ctx.font = '400 30px "Courier New", monospace'
    ctx.textAlign = 'right'
    ctx.fillText(`ARCHIVE / ${dateStr}`, w - padding - 40, padding + 110)

    ctx.beginPath()
    ctx.moveTo(padding + 40, padding + 160)
    ctx.lineTo(w - padding - 40, padding + 160)
    ctx.lineWidth = 4
    ctx.stroke()

    // 4. Image
    let imgY = padding + 220
    const maxImgH = 900
    if (p.coverImage) {
        try {
            const img = await loadImage(p.coverImage)
            const availW = w - (padding * 2) - 80
            const scale = Math.min(availW / img.width, maxImgH / img.height)
            const drawW = img.width * scale
            const drawH = img.height * scale
            const drawX = (w - drawW) / 2

            ctx.drawImage(img, drawX, imgY, drawW, drawH)
            ctx.strokeStyle = '#000000'
            ctx.lineWidth = 1
            ctx.strokeRect(drawX, imgY, drawW, drawH)

            imgY += drawH + 60
        } catch {
            imgY += 400
        }
    } else {
        imgY += 400
    }

    // 5. Product Info
    const infoX = padding + 40

    ctx.fillStyle = '#000000'
    ctx.textAlign = 'left'
    ctx.font = 'bold 100px "Times New Roman", serif'
    ctx.fillText(`STYLE #${p.styleNo}`, infoX, imgY + 80)

    ctx.font = '500 36px "Courier New", monospace'
    ctx.fillStyle = '#444444'
    const metaText = `${p.season} · ${p.category} · ${p.availability}`.toUpperCase()
    ctx.fillText(metaText, infoX, imgY + 150)

    ctx.fillStyle = '#000226'
    ctx.font = 'bold 60px "Times New Roman", serif'
    const price = p.priceText || t('product.login')
    ctx.fillText(price, infoX, imgY + 250)

    // 6. Specs & Options
    let cursorY = imgY + 350
    const lineHeight = 50

    ctx.font = '400 32px "Courier New", monospace'
    ctx.fillStyle = '#000000'

    const drawRow = (label: string, value: string) => {
        if (cursorY > h - padding - 150) return
        ctx.fillStyle = '#666666'
        ctx.fillText(label.toUpperCase(), infoX, cursorY)
        const valueX = infoX + 300
        ctx.fillStyle = '#000000'
        ctx.fillText(value, valueX, cursorY)
        cursorY += lineHeight
    }

    if (extra.selectedOptionLines.length) {
        ctx.fillStyle = '#000000'
        ctx.font = 'bold 36px "Courier New", monospace'
        ctx.fillText('OPTIONS', infoX, cursorY)
        cursorY += 60
        ctx.font = '400 32px "Courier New", monospace'
        extra.selectedOptionLines.forEach(line => {
            const parts = line.split(':')
            if (parts.length >= 2) {
                drawRow(parts[0] + ':', parts.slice(1).join(':').trim())
            } else {
                drawRow('', line)
            }
        })
        cursorY += 40
    }

    if (extra.specsLines.length) {
        ctx.fillStyle = '#000000'
        ctx.font = 'bold 36px "Courier New", monospace'
        ctx.fillText('SPECIFICATIONS', infoX, cursorY)
        cursorY += 60
        ctx.font = '400 32px "Courier New", monospace'
        extra.specsLines.forEach(line => {
            const parts = line.split(':')
            if (parts.length >= 2) {
                drawRow(parts[0] + ':', parts.slice(1).join(':').trim())
            } else {
                drawRow('', line)
            }
        })
    }

    // 7. Footer
    const url = typeof window === 'undefined' ? '' : window.location.href
    if (url) {
        ctx.fillStyle = '#000000'
        ctx.font = '400 28px "Courier New", monospace'
        ctx.textAlign = 'center'
        ctx.fillText(url, w / 2, h - padding - 40)
    }

    return { dataUrl: canvas.toDataURL('image/png'), meta: { w, h, drewImage: !!p.coverImage } }
}

const openPoster = async () => {
    posterError.value = ''
    posterDataUrl.value = ''
    const p = product.value
    if (!p || typeof window === 'undefined') return

    try {
        const selectedOptionLines = optionGroups.value
            .map((g) => {
                const pickedKey = selectedOptions.value[g.key] ?? ''
                if (!pickedKey) return ''
                const picked = g.options.find((o) => o.key === pickedKey)
                return picked ? `${g.name}: ${picked.label}` : ''
            })
            .filter(Boolean)

        const specsLines = specs.value
            .filter((s) => String(s.value ?? '').trim() !== '')
            .map((s) => `${s.label}: ${s.value}`)

        const { dataUrl, meta } = await buildPoster(p, { selectedOptionLines, specsLines })
        posterDataUrl.value = dataUrl
        posterOpen.value = true

        // report metadata only (no image bytes)
        await httpPost('/api/v1/events', {
            event_type: 'poster_generated',
            occurred_at: new Date().toISOString(),
            anon_id: getOrCreateAnonId(),
            product_id: p.id,
            page_url: window.location.href,
            referrer: document.referrer ?? '',
            ...readUtm(),
            payload: {
                poster: meta,
                locale: locale.value,
                selections: {
                    options: selectedOptionLines,
                    specs: specsLines,
                },
            },
        })
    } catch {
        posterError.value = t('productDetail.posterError')
    }
}

const closePoster = () => {
    posterOpen.value = false
}

onMounted(load)
</script>

<template>
    <main class="bg-white min-h-screen">
        <div class="px-4 md:px-10 pt-24 pb-16 max-w-6xl mx-auto">
            <div class="flex items-center justify-between gap-4">
                <button @click="router.back()"
                    class="h-9 px-3 border border-border font-mono text-xs uppercase tracking-[0.25em]">
                    {{ t('productDetail.back') }}
                </button>
                <router-link :to="{ name: 'home' }"
                    class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">Home</router-link>
            </div>

            <p v-if="loading" class="mt-6 font-mono text-xs text-black/60">{{ t('productDetail.loading') }}</p>
            <p v-else-if="errorMsg" class="mt-6 font-mono text-xs text-red-600">{{ errorMsg }}</p>

            <div v-else-if="product" class="mt-10">
                <div class="grid lg:grid-cols-12 gap-10">
                    <!-- Media column -->
                    <section class="lg:col-span-7">
                        <div class="space-y-4">
                            <div v-for="src in galleryUrls" :key="src"
                                class="aspect-[3/5] border border-border overflow-hidden bg-white">
                                <img :src="src" class="w-full h-full object-cover" loading="lazy" />
                            </div>
                        </div>
                    </section>

                    <!-- Sticky info column -->
                    <aside class="lg:col-span-5 lg:sticky lg:top-28 self-start">
                        <h1 class="font-display text-2xl md:text-3xl uppercase tracking-wider">STYLE #{{
                            product.styleNo }}</h1>

                        <p class="mt-3 font-mono text-xs text-black/60">
                            {{ seasonLabel(product.season) }} · {{ categoryLabel(product.category) }} · {{
                                availabilityLabel(product.availability) }}
                        </p>

                        <div class="mt-6 border border-border bg-white/70 backdrop-blur px-5 py-4">
                            <p class="font-mono text-[0.7rem] uppercase tracking-[0.25em] text-black/50">
                                {{ t('productDetail.priceLabel') }}
                            </p>
                            <p class="mt-2 text-brand font-semibold tracking-wide">{{ displayPriceText }}</p>
                            <p class="mt-3 text-xs text-black/60 leading-relaxed">
                                {{ t('productDetail.priceNote') }}
                            </p>

                            <div class="mt-6 grid grid-cols-1 sm:grid-cols-2 gap-3">
                                <RouterLink :to="{ name: 'home', hash: '#contact' }"
                                    class="h-11 inline-flex items-center justify-center bg-brand text-white font-mono text-xs uppercase tracking-[0.25em]">
                                    {{ t('productDetail.ctaPrimary') }}
                                </RouterLink>
                                <button @click="copyLink"
                                    class="h-11 px-4 border border-border bg-white font-mono text-xs uppercase tracking-[0.25em]">
                                    {{ t('productDetail.copyLink') }}
                                </button>
                            </div>

                            <div class="mt-3">
                                <button @click="openPoster"
                                    class="h-11 w-full px-4 border border-border bg-white font-mono text-xs uppercase tracking-[0.25em]">
                                    {{ t('productDetail.generatePoster') }}
                                </button>
                            </div>

                            <p v-if="linkHint" class="mt-3 font-mono text-xs text-black/60">{{ linkHint }}</p>
                            <p v-if="posterError" class="mt-3 font-mono text-xs text-red-600">{{ posterError }}</p>
                        </div>

                        <div v-for="(sec, idx) in sectionsByArea.sticky" :key="sec.id || `${sec.type}-${idx}`"
                            class="mt-10">
                            <div v-if="sec.type === 'options' && optionGroups.length">
                                <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                                    sectionTitle(sec, 'productDetail.sections.options') }}</h2>
                                <div class="mt-4 space-y-4">
                                    <div v-for="g in optionGroups" :key="g.key">
                                        <div class="font-mono text-xs uppercase tracking-[0.22em] text-black/50">
                                            {{ g.name }}</div>
                                        <div class="mt-2 flex flex-wrap gap-2">
                                            <button v-for="opt in g.options" :key="opt.key" type="button"
                                                @click="toggleOption(g.key, opt.key)"
                                                :aria-pressed="isOptionSelected(g.key, opt.key)"
                                                class="h-9 min-w-[2.25rem] px-3 border text-xs font-mono transition-colors rounded-none flex items-center justify-center"
                                                :class="isOptionSelected(g.key, opt.key)
                                                    ? 'bg-brand text-white border-brand'
                                                    : 'bg-white border-border hover:border-black text-black'">
                                                {{ opt.label }}
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div v-else-if="sec.type === 'richText'">
                                <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                                    sectionTitle(sec, 'productDetail.sections.overview') }}</h2>
                                <div class="mt-5 text-sm leading-7 text-black/80 whitespace-pre-line font-sans">
                                    {{ pickLocalizedText((sec as any)?.data?.text_i18n) || descriptionText }}
                                </div>
                            </div>

                            <div v-else-if="sec.type === 'divider'" class="border-t border-border" />
                        </div>
                    </aside>
                </div>

                <!-- Lower modules -->
                <div class="mt-14 grid lg:grid-cols-12 gap-10">
                    <section class="lg:col-span-7 space-y-10">
                        <div v-for="(sec, idx) in sectionsByArea.main" :key="sec.id || `${sec.type}-${idx}`">
                            <div v-if="sec.type === 'richText'" class="border-t border-border pt-8">
                                <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                                    sectionTitle(sec, 'productDetail.sections.overview') }}</h2>
                                <div class="mt-5 text-sm leading-7 text-black/80 whitespace-pre-line font-sans">
                                    {{ pickLocalizedText((sec as any)?.data?.text_i18n) || descriptionText }}
                                </div>
                            </div>

                            <div v-else-if="sec.type === 'specs' && specs.length" class="border-t border-border pt-8">
                                <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                                    sectionTitle(sec, 'productDetail.sections.specs') }}</h2>
                                <div class="mt-5 border-t border-border">
                                    <div v-for="row in specs" :key="row.label"
                                        class="grid grid-cols-12 border-b border-border group hover:bg-gray-50 transition-colors">
                                        <div
                                            class="col-span-4 sm:col-span-3 px-4 py-3 border-r border-border font-mono text-xs uppercase tracking-wider text-black/50 flex items-center">
                                            {{ row.label }}</div>
                                        <div
                                            class="col-span-8 sm:col-span-9 px-4 py-3 text-sm text-black font-mono flex items-center">
                                            {{ row.value }}</div>
                                    </div>
                                </div>
                            </div>

                            <div v-else-if="sec.type === 'divider'" class="border-t border-border" />
                        </div>
                    </section>

                    <aside class="lg:col-span-5">
                        <div v-for="(sec, idx) in sectionsByArea.aside" :key="sec.id || `${sec.type}-${idx}`">
                            <div v-if="sec.type === 'service'" class="border-t border-border pt-8">
                                <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                                    sectionTitle(sec, 'productDetail.sections.service') }}</h2>
                                <ul class="mt-5 space-y-3">
                                    <li v-for="(it, j) in serviceItems" :key="j" class="flex gap-4 items-start group">
                                        <span class="font-mono text-brand/40 mt-0.5">—</span>
                                        <span class="text-sm text-black/80 group-hover:text-black transition-colors">{{
                                            it }}</span>
                                    </li>
                                </ul>
                            </div>

                            <div v-else-if="sec.type === 'divider'" class="border-t border-border" />
                        </div>
                    </aside>
                </div>
            </div>

            <div v-if="posterOpen"
                class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4"
                @click.self="closePoster">
                <div class="bg-white w-full max-w-[400px] flex flex-col max-h-[90vh] shadow-2xl">
                    <div class="flex items-center justify-between p-4 border-b border-border">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black">
                            {{ t('productDetail.posterTitle') }}
                        </div>
                        <button @click="closePoster"
                            class="w-8 h-8 flex items-center justify-center hover:bg-gray-100 transition-colors">
                            <span class="text-xl leading-none">&times;</span>
                        </button>
                    </div>
                    <div class="flex-1 overflow-y-auto p-6 bg-gray-50 flex items-center justify-center">
                        <div class="shadow-lg border border-gray-200">
                            <img v-if="posterDataUrl" :src="posterDataUrl" class="w-full h-auto block" />
                        </div>
                    </div>
                    <div class="p-4 border-t border-border bg-white">
                        <a v-if="posterDataUrl" :href="posterDataUrl" download="FLEURLIS_ARCHIVE.png"
                            class="flex h-12 items-center justify-center bg-brand text-white font-mono text-xs uppercase tracking-[0.25em] hover:bg-brand/90 transition-colors w-full">
                            {{ t('productDetail.downloadPoster') }}
                        </a>
                        <p class="mt-3 text-center font-mono text-[10px] text-gray-400 uppercase tracking-widest">
                            Long press to save image
                        </p>
                    </div>
                </div>
            </div>
        </div>
    </main>
</template>
