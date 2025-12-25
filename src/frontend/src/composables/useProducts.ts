import { ref } from 'vue'

import { httpGet, resolveApiUrl } from '@/api/http'
import { normalizeStyleNo } from '@/utils/styleNo'

export type Availability = 'in_stock' | 'preorder' | 'archived'
export type Season = 'ss25' | 'fw25'
export type Category = 'gown' | 'couture' | 'bridal'

export type Product = {
    id: number
    styleNo: string
    season: Season
    category: Category
    availability: Availability
    coverImage: string
    hoverImage: string
    isNew: boolean
}

type ProductsResponse = { items: unknown[] }

type UseProductsOptions = {
    limit?: number
}

// Module-level singletons so all components share the same state.
const products = ref<Product[]>([])
const loading = ref(false)
const error = ref('')

let loaded = false
let lastLimit = 0
let inflight: Promise<void> | null = null
let inflightLimit = 0

const normalizeProduct = (raw: any): Product => {
    return {
        id: Number(raw?.id ?? 0),
        styleNo: normalizeStyleNo(raw?.styleNo ?? raw?.style_no ?? ''),
        season: raw?.season as Season,
        category: raw?.category as Category,
        availability: raw?.availability as Availability,
        coverImage: resolveApiUrl(String(raw?.coverImage ?? raw?.cover_image ?? '')),
        hoverImage: resolveApiUrl(String(raw?.hoverImage ?? raw?.hover_image ?? '')),
        // Backend public JSON currently uses `isNew`; query/db uses `is_new`.
        // Be defensive to avoid silent filter failures.
        isNew: Boolean(raw?.isNew ?? raw?.is_new ?? false),
    }
}

const loadProducts = async (limit: number) => {
    error.value = ''
    loading.value = true
    try {
        const res = await httpGet<ProductsResponse>(`/api/v1/products?limit=${limit}`)
        const items = Array.isArray(res?.items) ? res.items : []
        products.value = items.map(normalizeProduct)
        loaded = true
        lastLimit = limit
    } catch {
        // Keep existing data if any; allow later retry.
        loaded = false
        error.value = '商品加载失败'
    } finally {
        loading.value = false
    }
}

export const useProducts = (options: UseProductsOptions = {}) => {
    const limit = Math.max(1, Math.floor(options.limit ?? 200))

    const ensureLoaded = async (force = false) => {
        if (!force && loaded && lastLimit >= limit) return

        const nextLimit = Math.max(limit, lastLimit)

        // If there's already a request in flight, wait for it.
        // If the in-flight request is for a smaller limit, chain another request after it.
        if (inflight) {
            if (inflightLimit >= nextLimit) return inflight
            await inflight
            // Retry after inflight finishes (may start a larger request).
            return ensureLoaded(force)
        }

        inflightLimit = nextLimit
        inflight = loadProducts(nextLimit).finally(() => {
            inflight = null
            inflightLimit = 0
        })
        return inflight
    }

    const refresh = async () => ensureLoaded(true)

    return {
        products,
        loading,
        error,
        ensureLoaded,
        refresh,
    }
}
