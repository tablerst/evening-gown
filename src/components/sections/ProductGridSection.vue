<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t, te } = useI18n()

type Availability = 'in_stock' | 'preorder' | 'archived'
type Season = 'ss25' | 'fw25'
type Category = 'gown' | 'couture' | 'bridal'

const CATEGORY_OPTIONS = ['all', 'gown', 'couture', 'bridal'] as const
type CategoryOption = (typeof CATEGORY_OPTIONS)[number]

const SEASON_OPTIONS = ['all', 'ss25', 'fw25'] as const
type SeasonOption = (typeof SEASON_OPTIONS)[number]

const AVAILABILITY_OPTIONS = ['all', 'in_stock', 'preorder', 'archived'] as const
type AvailabilityOption = (typeof AVAILABILITY_OPTIONS)[number]

const SORT_OPTIONS = ['newest', 'style_asc', 'style_desc'] as const
type SortKey = (typeof SORT_OPTIONS)[number]

type Product = {
    id: number
    styleNo: number
    season: Season
    category: Category
    availability: Availability
    coverImage: string
    hoverImage: string
    isNew: boolean
}

// 占位数据（硬编码）：后续可替换为真实商品/库存接口
const products = ref<Product[]>(
    Array.from({ length: 42 }, (_, idx) => {
        const id = idx + 1
        const styleNo = 2000 + id
        const season: Product['season'] = id % 2 === 0 ? 'fw25' : 'ss25'
        const category: Product['category'] = id % 3 === 0 ? 'couture' : id % 3 === 1 ? 'gown' : 'bridal'
        const availability: Availability = id % 9 === 0 ? 'archived' : id % 4 === 0 ? 'preorder' : 'in_stock'
        const isNew = id % 7 === 0

        return {
            id,
            styleNo,
            season,
            category,
            availability,
            isNew,
            coverImage:
                'https://images.unsplash.com/photo-1595777457583-95e059d581b8?q=80&w=1200&auto=format&fit=crop',
            hoverImage:
                'https://images.unsplash.com/photo-1515372039744-b8f02a3ae446?q=80&w=1200&auto=format&fit=crop',
        }
    })
)

// 端适配：参考 Seasonal 的密度（PC 5 列 / 手机 3 列）
const isDesktop = ref(false)
let mql: MediaQueryList | null = null

const syncMedia = () => {
    isDesktop.value = mql?.matches ?? false
}

const onMediaChange = (e: MediaQueryListEvent) => {
    isDesktop.value = e.matches
}

onMounted(() => {
    mql = window.matchMedia('(min-width: 768px)')
    syncMedia()

    if (!mql) return
    if (typeof mql.addEventListener === 'function') mql.addEventListener('change', onMediaChange)
    else (mql as unknown as { addListener?: (cb: (e: MediaQueryListEvent) => void) => void }).addListener?.(onMediaChange)
})

onBeforeUnmount(() => {
    if (!mql) return
    if (typeof mql.removeEventListener === 'function') mql.removeEventListener('change', onMediaChange)
    else (mql as unknown as { removeListener?: (cb: (e: MediaQueryListEvent) => void) => void }).removeListener?.(onMediaChange)
})

const cols = computed(() => (isDesktop.value ? 5 : 3))
const gapPx = computed(() => (isDesktop.value ? 16 : 10))

// 容器宽度：用像素计算卡片宽度，保持“单屏可见 5/3 张”的紧凑节奏（不做横向滑动）
const gridEl = ref<HTMLElement | null>(null)
const containerWidth = ref(0)
let ro: ResizeObserver | null = null

const syncContainerWidth = () => {
    containerWidth.value = gridEl.value?.clientWidth ?? 0
}

const itemWidthPx = computed(() => {
    const w = containerWidth.value
    const c = cols.value
    const gap = gapPx.value

    if (!w || !c) return isDesktop.value ? 220 : 140
    const raw = (w - (c - 1) * gap) / c
    const safe = Number.isFinite(raw) ? raw : isDesktop.value ? 220 : 140
    // 不强制最小宽度（避免小屏横向溢出）；但保持整数像素利于“硬切”与对齐
    return Math.max(80, Math.floor(safe))
})

const gridStyle = computed(() => ({
    '--product-cols': String(cols.value),
    '--product-gap': `${gapPx.value}px`,
    '--product-item-width': `${itemWidthPx.value}px`,
}))

onMounted(() => {
    syncContainerWidth()
    if (typeof ResizeObserver !== 'undefined') {
        ro = new ResizeObserver(() => syncContainerWidth())
        if (gridEl.value) ro.observe(gridEl.value)
    } else {
        window.addEventListener('resize', syncContainerWidth)
    }
})

onBeforeUnmount(() => {
    ro?.disconnect()
    ro = null
    window.removeEventListener('resize', syncContainerWidth)
})

// 筛选 / 排序 / 分页（先硬编码）
const categoryOptions = CATEGORY_OPTIONS
const seasonOptions = SEASON_OPTIONS
const availabilityOptions = AVAILABILITY_OPTIONS

const selectedCategory = ref<CategoryOption>('all')
const selectedSeason = ref<SeasonOption>('all')
const selectedAvailability = ref<AvailabilityOption>('all')

const sortKey = ref<SortKey>('newest')

const labelCategory = (value: CategoryOption) => {
    if (value === 'all') return te('product.filters.category.all') ? t('product.filters.category.all') : 'all'
    const key = `product.filters.category.${value}` as const
    return te(key) ? t(key) : value
}

const labelSeason = (value: SeasonOption) => {
    if (value === 'all') return te('product.filters.season.all') ? t('product.filters.season.all') : 'all'
    const key = `product.filters.season.${value}` as const
    return te(key) ? t(key) : value
}

const labelAvailability = (value: AvailabilityOption) => {
    if (value === 'all') return te('product.filters.availability.all') ? t('product.filters.availability.all') : 'all'
    const key = `product.filters.availability.${value}` as const
    return te(key) ? t(key) : value
}

const labelSort = (value: SortKey) => {
    const key = `product.sortOptions.${value}` as const
    return te(key) ? t(key) : value
}

const filteredProducts = computed(() => {
    return products.value.filter((p) => {
        if (selectedCategory.value !== 'all' && p.category !== selectedCategory.value) return false
        if (selectedSeason.value !== 'all' && p.season !== selectedSeason.value) return false
        if (selectedAvailability.value !== 'all' && p.availability !== selectedAvailability.value) return false
        return true
    })
})

const sortedProducts = computed(() => {
    const list = [...filteredProducts.value]
    if (sortKey.value === 'style_asc') list.sort((a, b) => a.styleNo - b.styleNo)
    if (sortKey.value === 'style_desc') list.sort((a, b) => b.styleNo - a.styleNo)
    if (sortKey.value === 'newest') list.sort((a, b) => Number(b.isNew) - Number(a.isNew) || b.id - a.id)
    return list
})

const pageSizeOptions = [10, 20, 50] as const
const selectedPageSize = ref<(typeof pageSizeOptions)[number]>(10)
const pageSize = computed(() => selectedPageSize.value)
const currentPage = ref(1)

const totalCount = computed(() => sortedProducts.value.length)
const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize.value)))

watch([selectedCategory, selectedSeason, selectedAvailability, sortKey, selectedPageSize], () => {
    currentPage.value = 1
})

watch(totalPages, (n) => {
    if (currentPage.value > n) currentPage.value = n
})

const pagedProducts = computed(() => {
    const start = (currentPage.value - 1) * pageSize.value
    return sortedProducts.value.slice(start, start + pageSize.value)
})

const canPrev = computed(() => currentPage.value > 1)
const canNext = computed(() => currentPage.value < totalPages.value)

const prevLabel = computed(() => (te('product.prev') ? t('product.prev') : '上一页'))
const nextLabel = computed(() => (te('product.next') ? t('product.next') : '下一页'))
const pageLabel = computed(() => (te('product.page') ? t('product.page') : '页'))
const pageSizeLabel = computed(() => (te('product.pageSize') ? t('product.pageSize') : '每页'))

const pageSizeOptionText = (n: number) => {
    if (te('product.pageSizeOption')) return t('product.pageSizeOption', { count: n })
    return `${pageSizeLabel.value} ${n}`
}

const newBadgeText = computed(() => (te('product.badge.new') ? t('product.badge.new') : 'NEW'))
</script>

<template>
    <section id="catalog" class="py-12 bg-white">
        <div class="px-4 md:px-8 mb-4 sticky top-0 bg-white z-10 py-4 border-b border-border">
            <div class="flex justify-between items-end gap-6">
                <div class="min-w-0">
                    <h2 class="text-2xl font-display uppercase tracking-wider">{{ t('product.title') }}</h2>
                    <div class="mt-2 font-mono text-xs text-gray-500 flex flex-wrap gap-x-6 gap-y-1">
                        <span>{{ t('product.total', { count: totalCount }) }}</span>
                        <span class="uppercase tracking-wider">{{ t('product.filter') }}</span>
                        <span class="uppercase tracking-wider">{{ t('product.sort') }}</span>
                        <span class="uppercase tracking-wider">{{ pageLabel }} {{ currentPage }}/{{ totalPages }}</span>
                    </div>
                </div>

                <!-- Controls: 极简、硬切、三色体系 -->
                <div class="flex flex-wrap justify-end gap-2 font-mono text-xs text-black">
                    <label class="sr-only">Category</label>
                    <select v-model="selectedCategory"
                        class="h-8 px-2 bg-white border border-border focus:outline-none">
                        <option v-for="c in categoryOptions" :key="c" :value="c">{{ labelCategory(c) }}</option>
                    </select>

                    <label class="sr-only">Season</label>
                    <select v-model="selectedSeason" class="h-8 px-2 bg-white border border-border focus:outline-none">
                        <option v-for="s in seasonOptions" :key="s" :value="s">{{ labelSeason(s) }}</option>
                    </select>

                    <label class="sr-only">Availability</label>
                    <select v-model="selectedAvailability"
                        class="h-8 px-2 bg-white border border-border focus:outline-none">
                        <option v-for="a in availabilityOptions" :key="a" :value="a">{{ labelAvailability(a) }}</option>
                    </select>

                    <label class="sr-only">Sort</label>
                    <select v-model="sortKey" class="h-8 px-2 bg-white border border-border focus:outline-none">
                        <option v-for="s in SORT_OPTIONS" :key="s" :value="s">{{ labelSort(s) }}</option>
                    </select>

                    <label class="sr-only">Page size</label>
                    <select v-model.number="selectedPageSize"
                        class="h-8 px-2 bg-white border border-border focus:outline-none">
                        <option v-for="n in pageSizeOptions" :key="n" :value="n">{{ pageSizeOptionText(n) }}</option>
                    </select>

                    <button
                        class="h-8 px-3 border border-black bg-white hover:bg-brand hover:text-white transition-none"
                        :disabled="!canPrev" :aria-disabled="!canPrev"
                        @click="canPrev && (currentPage = currentPage - 1)">
                        {{ prevLabel }}
                    </button>
                    <button
                        class="h-8 px-3 border border-black bg-white hover:bg-brand hover:text-white transition-none"
                        :disabled="!canNext" :aria-disabled="!canNext"
                        @click="canNext && (currentPage = currentPage + 1)">
                        {{ nextLabel }}
                    </button>
                </div>
            </div>
        </div>

        <div ref="gridEl" class="px-4 md:px-8" :style="gridStyle">
            <div class="product-grid">
                <div v-for="p in pagedProducts" :key="p.id" class="product-card group cursor-pointer">
                    <div class="relative aspect-[3/5] overflow-hidden border border-border bg-white">
                        <!-- Default Image -->
                        <img :src="p.coverImage"
                            class="w-full h-full object-cover block group-hover:hidden filter-cold-drama" />

                        <!-- Hover Detail -->
                        <img :src="p.hoverImage"
                            class="w-full h-full object-cover hidden group-hover:block scale-110 filter-cold-drama" />

                        <!-- Hard-cut overlay (tri-color) -->
                        <div
                            class="absolute inset-0 bg-[#000226] opacity-0 group-hover:opacity-15 transition-opacity duration-0 mix-blend-multiply">
                        </div>

                        <div
                            class="absolute bottom-0 left-0 right-0 p-2 bg-white/95 font-mono text-xs flex justify-between items-center border-t border-border">
                            <div class="flex flex-col min-w-0">
                                <div class="flex items-center gap-2">
                                    <span class="font-bold text-black truncate">{{ t('product.style', { id: p.styleNo })
                                        }}</span>
                                    <span v-if="p.isNew" class="text-brand uppercase tracking-wider">{{ newBadgeText
                                        }}</span>
                                </div>
                                <div class="mt-1 flex items-center gap-3 text-gray-500">
                                    <span class="text-black">{{ labelSeason(p.season) }}</span>
                                    <span>{{ labelCategory(p.category) }}</span>
                                    <span class="uppercase">{{ labelAvailability(p.availability) }}</span>
                                </div>
                                <span class="text-brand font-bold mt-1">{{ t('product.login') }}</span>
                            </div>
                            <button
                                class="w-8 h-8 border border-black flex items-center justify-center hover:bg-brand hover:text-white transition-none text-lg"
                                :aria-label="t('product.add')">
                                +
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            <div class="mt-6 flex justify-between items-center font-mono text-xs border-t border-border pt-4">
                <div class="text-gray-500">
                    {{ t('product.total', { count: totalCount }) }} · {{ pageLabel }} {{ currentPage }}/{{ totalPages }}
                </div>
                <div class="flex gap-2">
                    <button
                        class="h-8 px-3 border border-black bg-white hover:bg-brand hover:text-white transition-none"
                        :disabled="!canPrev" :aria-disabled="!canPrev"
                        @click="canPrev && (currentPage = currentPage - 1)">
                        {{ prevLabel }}
                    </button>
                    <button
                        class="h-8 px-3 border border-black bg-white hover:bg-brand hover:text-white transition-none"
                        :disabled="!canNext" :aria-disabled="!canNext"
                        @click="canNext && (currentPage = currentPage + 1)">
                        {{ nextLabel }}
                    </button>
                </div>
            </div>
        </div>
    </section>
</template>

<style scoped>
.filter-cold-drama {
    filter: contrast(95%) brightness(105%) saturate(80%);
    /* 移除 hue-rotate(180deg) 以恢复正常肤色，保留低饱和度冷感 */
}

.product-grid {
    display: grid;
    grid-template-columns: repeat(var(--product-cols), var(--product-item-width));
    gap: var(--product-gap);
}

.product-card {
    width: var(--product-item-width);
}

@media (max-width: 360px) {
    .product-grid {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .product-card {
        width: auto;
    }
}
</style>
