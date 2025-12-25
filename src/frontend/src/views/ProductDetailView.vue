<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

import { HttpError, httpGet, httpPost, resolveApiUrl } from '@/api/http'

type ProductDetail = {
    id: number
    slug: string
    styleNo: number
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
type OptionGroup = { name: string; options: string[] }

const { t, te, locale } = useI18n()
const route = useRoute()
const router = useRouter()

const productId = computed(() => Number(route.params.id))

const loading = ref(false)
const errorMsg = ref('')
const product = ref<ProductDetail | null>(null)

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
    const label = String(obj.label ?? obj.k ?? obj.key ?? obj.name ?? '').trim()
    const value = String(obj.value ?? obj.v ?? obj.val ?? '').trim()
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
    const name = String(obj.name ?? obj.title ?? obj.label ?? '').trim()
    const optionsRaw = obj.options
    const options: string[] = []
    if (Array.isArray(optionsRaw)) {
        for (const opt of optionsRaw) {
            if (typeof opt === 'string') {
                const s = opt.trim()
                if (s) options.push(s)
                continue
            }
            const optObj = asRecord(opt)
            if (!optObj) continue
            const s = String(optObj.label ?? optObj.name ?? optObj.value ?? '').trim()
            if (s) options.push(s)
        }
    }
    if (!name || options.length === 0) return null
    return { name, options }
}

const optionGroups = computed<OptionGroup[]>(() => {
    const raw = detailObj.value.option_groups ?? detailObj.value.optionGroups
    if (!Array.isArray(raw)) return []
    return raw.map(normalizeOptionGroup).filter(Boolean) as OptionGroup[]
})

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
    const brand = 'White Phantom'
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

const buildPoster = async (p: ProductDetail) => {
    // 3:5 竖版
    const w = 1080
    const h = 1800

    const canvas = document.createElement('canvas')
    canvas.width = w
    canvas.height = h
    const ctx = canvas.getContext('2d')
    if (!ctx) throw new Error('no canvas ctx')

    // base background
    ctx.fillStyle = '#ffffff'
    ctx.fillRect(0, 0, w, h)

    // top bar
    ctx.fillStyle = '#000226'
    ctx.fillRect(0, 0, w, 96)
    ctx.fillStyle = '#ffffff'
    ctx.font = 'bold 38px ui-sans-serif, system-ui, -apple-system'
    ctx.fillText('WHITE PHANTOM', 48, 62)

    let drewImage = false
    if (p.coverImage) {
        try {
            const img = await loadImage(p.coverImage)
            // cover image area
            const imgX = 0
            const imgY = 96
            const imgW = w
            const imgH = 1320
            ctx.drawImage(img, imgX, imgY, imgW, imgH)
            // subtle overlay
            ctx.fillStyle = 'rgba(0,2,38,0.10)'
            ctx.fillRect(imgX, imgY, imgW, imgH)
            drewImage = true
        } catch {
            // ignore; fallback to text-only poster
        }
    }

    // bottom info card
    const cardY = 1460
    ctx.fillStyle = '#ffffff'
    ctx.fillRect(0, cardY, w, h - cardY)
    ctx.strokeStyle = '#e5e7eb'
    ctx.lineWidth = 2
    ctx.strokeRect(0, cardY, w, h - cardY)

    ctx.fillStyle = '#111827'
    ctx.font = '700 48px ui-sans-serif, system-ui, -apple-system'
    ctx.fillText(`STYLE #${p.styleNo}`, 48, cardY + 90)

    ctx.fillStyle = '#6b7280'
    ctx.font = '500 28px ui-sans-serif, system-ui, -apple-system'
    ctx.fillText(`${p.season.toUpperCase()} · ${p.category.toUpperCase()} · ${p.availability.toUpperCase()}`, 48, cardY + 140)

    ctx.fillStyle = '#d4af37'
    ctx.font = '800 36px ui-sans-serif, system-ui, -apple-system'
    ctx.fillText(p.priceText || t('product.login'), 48, cardY + 200)

    // link
    ctx.fillStyle = '#111827'
    ctx.font = '500 22px ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace'
    const url = typeof window === 'undefined' ? '' : window.location.href
    if (url) ctx.fillText(url, 48, h - 40)

    // to data url
    const dataUrl = canvas.toDataURL('image/png')
    return { dataUrl, meta: { w, h, drewImage } }
}

const openPoster = async () => {
    posterError.value = ''
    posterDataUrl.value = ''
    const p = product.value
    if (!p || typeof window === 'undefined') return

    try {
        const { dataUrl, meta } = await buildPoster(p)
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

                        <div v-if="optionGroups.length" class="mt-10">
                            <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                                t('productDetail.sections.options') }}</h2>
                            <div class="mt-4 space-y-4">
                                <div v-for="g in optionGroups" :key="g.name">
                                    <div class="font-mono text-xs uppercase tracking-[0.22em] text-black/50">
                                        {{ g.name }}</div>
                                    <div class="mt-2 flex flex-wrap gap-2">
                                        <span v-for="opt in g.options" :key="opt"
                                            class="px-2.5 py-1 border border-border bg-white text-xs">
                                            {{ opt }}
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </aside>
                </div>

                <!-- Lower modules -->
                <div class="mt-14 grid lg:grid-cols-12 gap-10">
                    <section class="lg:col-span-7 space-y-10">
                        <div v-if="descriptionText" class="border-t border-border pt-8">
                            <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                                t('productDetail.sections.overview') }}</h2>
                            <p class="mt-4 text-sm leading-relaxed text-black/80 whitespace-pre-line">
                                {{ descriptionText }}
                            </p>
                        </div>

                        <div v-if="specs.length" class="border-t border-border pt-8">
                            <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                                t('productDetail.sections.specs') }}</h2>
                            <div class="mt-4 border border-border bg-white">
                                <div v-for="row in specs" :key="row.label"
                                    class="grid grid-cols-12 gap-4 px-4 py-3 border-b border-border last:border-b-0">
                                    <div class="col-span-5 font-mono text-xs uppercase tracking-[0.18em] text-black/60">
                                        {{ row.label }}</div>
                                    <div class="col-span-7 text-sm text-black/80">{{ row.value }}</div>
                                </div>
                            </div>
                        </div>
                    </section>

                    <aside class="lg:col-span-5">
                        <div class="border-t border-border pt-8">
                            <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                                t('productDetail.sections.service') }}</h2>
                            <ul class="mt-4 space-y-2 text-sm text-black/80">
                                <li v-for="(it, idx) in serviceItems" :key="idx" class="flex gap-3">
                                    <span class="text-black/40">—</span>
                                    <span>{{ it }}</span>
                                </li>
                            </ul>
                        </div>
                    </aside>
                </div>
            </div>

            <div v-if="posterOpen" class="fixed inset-0 bg-black/60 flex items-center justify-center p-4"
                @click.self="closePoster">
                <div class="bg-white max-w-lg w-full p-4 border border-border">
                    <div class="flex items-center justify-between">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                            t('productDetail.posterTitle') }}</div>
                        <button @click="closePoster"
                            class="h-9 px-3 border border-border font-mono text-xs uppercase tracking-[0.25em]">Close</button>
                    </div>
                    <div class="mt-4 border border-border overflow-hidden">
                        <img v-if="posterDataUrl" :src="posterDataUrl" class="w-full h-auto block" />
                    </div>
                    <a v-if="posterDataUrl" :href="posterDataUrl" download="poster.png"
                        class="mt-4 inline-flex h-10 items-center px-4 bg-brand text-white font-mono text-xs uppercase tracking-[0.25em]">
                        {{ t('productDetail.downloadPoster') }}
                    </a>
                </div>
            </div>
        </div>
    </main>
</template>
