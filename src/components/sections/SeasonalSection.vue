<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t, te } = useI18n()

// 占位数据：后续可替换为真实系列/Lookbook 数据源
const items = computed(() => Array.from({ length: 18 }, (_, idx) => idx + 1))

// 端适配：PC 默认可见 5 / 最大 10；手机默认可见 3 / 最大 6
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

const maxCols = computed(() => (isDesktop.value ? 10 : 6))
const visibleCols = computed(() => (isDesktop.value ? 5 : 3))

// 容器宽度：用于精确计算“可见 5/3 个”的卡片宽度（避免用 % 导致网格宽度循环计算出现大留白）
const scrollEl = ref<HTMLElement | null>(null)
const containerWidth = ref(0)
let ro: ResizeObserver | null = null

const syncContainerWidth = () => {
    containerWidth.value = scrollEl.value?.clientWidth ?? 0
}

// 行数：以“每行最多 maxCols”推导（满足：超过则行数 + 1）
const seasonalRows = computed(() => Math.max(1, Math.ceil((items.value.length || 1) / maxCols.value)))

// 列数：在既定行数下尽量均匀分布，避免最后一行出现大面积空白
const seasonalCols = computed(() => {
    const n = items.value.length || 1
    const cols = Math.ceil(n / seasonalRows.value)
    return Math.max(1, Math.min(maxCols.value, cols))
})

// 小于默认可见数时，按实际列数放大卡片，避免“少量数据仍强制缩成 5/3 等份”
const effectiveVisibleCols = computed(() => Math.min(visibleCols.value, seasonalCols.value))

const canScroll = computed(() => seasonalCols.value > visibleCols.value)

const swipeHintText = computed(() => {
    if (te('seasonal.swipeHint')) return t('seasonal.swipeHint')
    return '→ 右滑查看更多'
})

// 收紧留白：间距更紧凑（符合档案库/高密度浏览）
const gapPx = computed(() => (isDesktop.value ? 16 : 10))
const itemMinPx = computed(() => (isDesktop.value ? 180 : 120))

const itemWidthPx = computed(() => {
    const w = containerWidth.value
    const cols = effectiveVisibleCols.value
    const gap = gapPx.value

    if (!w || !cols) return itemMinPx.value
    const raw = (w - (cols - 1) * gap) / cols
    const safe = Number.isFinite(raw) ? raw : itemMinPx.value
    return Math.max(itemMinPx.value, Math.floor(safe))
})

const gridStyle = computed(() => ({
    '--seasonal-cols': String(seasonalCols.value),
    '--seasonal-rows': String(seasonalRows.value),
    '--seasonal-visible-cols': String(effectiveVisibleCols.value),
    '--seasonal-gap': `${gapPx.value}px`,
    '--seasonal-item-width': `${itemWidthPx.value}px`,
}))

onMounted(() => {
    syncContainerWidth()
    if (typeof ResizeObserver !== 'undefined') {
        ro = new ResizeObserver(() => syncContainerWidth())
        if (scrollEl.value) ro.observe(scrollEl.value)
    } else {
        window.addEventListener('resize', syncContainerWidth)
    }
})

onBeforeUnmount(() => {
    ro?.disconnect()
    ro = null
    window.removeEventListener('resize', syncContainerWidth)
})
</script>

<template>
    <section id="seasonal" class="py-12 border-b border-border bg-white">
        <div class="px-4 md:px-8 mb-6 flex justify-between items-end">
            <h2 class="text-2xl font-display uppercase tracking-wider">{{ t('seasonal.title') }}</h2>
            <span class="font-mono text-sm text-gray-500">{{ t('seasonal.season') }}</span>
        </div>
        <div class="px-4 md:px-8 pb-2">
            <div v-if="canScroll" class="seasonal-tip font-mono text-xs text-brand">
                {{ swipeHintText }}
            </div>
            <div ref="scrollEl" class="seasonal-scroll overflow-x-auto" :style="gridStyle">
                <!-- 横向滑动 + 固定列数网格：每行最多 N 个（端适配），超出自动增加行数 -->
                <div class="seasonal-grid">
                    <div v-for="i in items" :key="i" class="seasonal-card group cursor-pointer">
                        <div class="aspect-[3/4] bg-gray-100 relative overflow-hidden border border-border">
                            <img :src="`https://images.unsplash.com/photo-1566174053879-31528523f8ae?q=80&w=800&auto=format&fit=crop&ixlib=rb-4.0.3`"
                                class="w-full h-full object-cover filter-cold-drama transition-transform duration-0 group-hover:scale-105" />
                            <div
                                class="absolute inset-0 bg-[#000226] opacity-0 group-hover:opacity-20 transition-opacity duration-0 mix-blend-multiply">
                            </div>
                        </div>
                        <div class="mt-3 flex justify-between font-mono text-sm border-t border-black pt-2">
                            <span class="font-bold">{{ t('seasonal.series', { index: i }) }}</span>
                            <span class="text-brand">{{ t('seasonal.lookbook') }}</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>
</template>

<style scoped>
.filter-cold-drama {
    filter: contrast(95%) brightness(105%) saturate(80%);
    /* 移除 hue-rotate(180deg) 以恢复正常肤色 */
}

.seasonal-tip {
    display: flex;
    justify-content: flex-end;
    padding-bottom: 8px;
    user-select: none;
}

.seasonal-scroll {
    /* 默认兜底（实际运行时由内联 style 注入像素值） */
    --seasonal-gap: 10px;
    --seasonal-item-width: 140px;

    /* 让滚动条占位稳定，避免内容跳动 */
    scrollbar-gutter: stable;
    overscroll-behavior-x: contain;

    /* Firefox */
    scrollbar-width: thin;
    scrollbar-color: #000226 transparent;
}

/* WebKit (Chrome / Edge / Safari) */
.seasonal-scroll::-webkit-scrollbar {
    height: 10px;
}

.seasonal-scroll::-webkit-scrollbar-track {
    background: transparent;
    border-top: 1px solid #e2e8f0;
}

.seasonal-scroll::-webkit-scrollbar-thumb {
    background: #000226;
    border: 1px solid #000000;
    border-radius: 0;
}

.seasonal-scroll::-webkit-scrollbar-thumb:hover {
    background: #000226;
}

.seasonal-grid {
    display: grid;
    grid-template-columns: repeat(var(--seasonal-cols), var(--seasonal-item-width));
    gap: var(--seasonal-gap);
    width: max-content;
    padding-bottom: 6px;
    /* 给滚动条一点呼吸空间 */
}

@media (min-width: 768px) {
    .seasonal-scroll {
        --seasonal-gap: 16px;
        --seasonal-item-width: 220px;
    }
}

.seasonal-card {
    width: var(--seasonal-item-width);
}
</style>
