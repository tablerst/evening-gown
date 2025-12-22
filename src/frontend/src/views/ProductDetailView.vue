<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

import { HttpError, httpGet, httpPost } from '@/api/http'

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
    detail: any
}

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()

const productId = computed(() => Number(route.params.id))

const loading = ref(false)
const errorMsg = ref('')
const product = ref<ProductDetail | null>(null)

const posterOpen = ref(false)
const posterDataUrl = ref('')
const posterError = ref('')

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
        product.value = await httpGet<ProductDetail>(`/api/v1/products/${productId.value}`)
    } catch (e) {
        if (e instanceof HttpError && e.status === 404) errorMsg.value = 'Not Found'
        else errorMsg.value = t('productDetail.error')
    } finally {
        loading.value = false
    }
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
        <div class="px-4 md:px-8 py-10 max-w-5xl mx-auto">
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

            <div v-else-if="product" class="mt-8 grid md:grid-cols-2 gap-8">
                <div>
                    <div class="aspect-[3/5] border border-border overflow-hidden bg-white">
                        <img :src="product.coverImage" class="w-full h-full object-cover" />
                    </div>
                </div>

                <div>
                    <h1 class="font-display text-2xl uppercase tracking-wider">STYLE #{{ product.styleNo }}</h1>
                    <p class="mt-2 font-mono text-xs text-black/60">
                        {{ product.season.toUpperCase() }} · {{ product.category.toUpperCase() }} · {{
                            product.availability.toUpperCase() }}
                    </p>
                    <p class="mt-4 text-brand font-bold">{{ product.priceText || t('product.login') }}</p>

                    <div class="mt-8 flex flex-wrap gap-3">
                        <button @click="openPoster"
                            class="h-10 px-4 bg-brand text-white font-mono text-xs uppercase tracking-[0.25em]">
                            {{ t('productDetail.generatePoster') }}
                        </button>
                    </div>

                    <p v-if="posterError" class="mt-4 font-mono text-xs text-red-600">{{ posterError }}</p>

                    <div class="mt-10 border-t border-border pt-6">
                        <h2 class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">DETAIL</h2>
                        <pre
                            class="mt-3 p-4 border border-border bg-white overflow-auto text-xs">{{ JSON.stringify(product.detail, null, 2) }}</pre>
                    </div>
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
