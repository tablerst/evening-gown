<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

type MetricsDay = {
    date: string
    total: number
    byType: Record<string, number>
}

const props = withDefaults(
    defineProps<{
        title?: string
        tz?: string
        series: MetricsDay[]
        loading?: boolean
        height?: string
    }>(),
    {
        title: '',
        tz: 'UTC',
        loading: false,
        height: '260px',
    }
)

const elRef = ref<HTMLDivElement | null>(null)
let chart: any | null = null
let removeResize: (() => void) | null = null

const types = computed(() => {
    const set = new Set<string>()
    for (const d of props.series) {
        for (const k of Object.keys(d.byType || {})) set.add(k)
    }
    return Array.from(set).sort()
})

const categories = computed(() => props.series.map((d) => d.date))

const buildOption = () => {
    const seriesByType = types.value.map((t) => ({
        name: t,
        type: 'bar',
        stack: 'events',
        emphasis: { focus: 'series' },
        data: props.series.map((d) => (d.byType && typeof d.byType[t] === 'number' ? d.byType[t] : 0)),
    }))

    const totalLine = {
        name: 'total',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        yAxisIndex: 0,
        data: props.series.map((d) => d.total ?? 0),
    }

    return {
        backgroundColor: 'transparent',
        tooltip: { trigger: 'axis' },
        legend: {
            top: 0,
            left: 0,
            textStyle: { fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace', fontSize: 11 },
        },
        grid: { left: 36, right: 16, top: 40, bottom: 28 },
        xAxis: {
            type: 'category',
            data: categories.value,
            axisLabel: { fontSize: 10 },
        },
        yAxis: {
            type: 'value',
            axisLabel: { fontSize: 10 },
            splitLine: { lineStyle: { color: '#00000010' } },
        },
        series: [...seriesByType, totalLine],
    }
}

const render = async () => {
    if (!elRef.value) return
    const echarts = await import('echarts')

    if (!chart) {
        chart = echarts.init(elRef.value)
    }

    chart.setOption(buildOption(), { notMerge: true })
    chart.resize()
}

const dispose = () => {
    if (removeResize) removeResize()
    removeResize = null

    if (chart) {
        chart.dispose()
        chart = null
    }
}

onMounted(() => {
    void render()

    const onResize = () => {
        chart?.resize?.()
    }
    window.addEventListener('resize', onResize, { passive: true })
    removeResize = () => window.removeEventListener('resize', onResize)
})

onBeforeUnmount(() => {
    dispose()
})

watch(
    () => props.series,
    () => {
        void render()
    },
    { deep: true }
)
</script>

<template>
    <div class="border border-border p-4">
        <div class="flex items-center justify-between gap-4">
            <div>
                <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">{{ title }}</div>
                <div class="mt-1 font-mono text-[11px] text-black/50">TZ: {{ tz }}</div>
            </div>
            <div v-if="loading" class="font-mono text-xs text-black/50">Loading…</div>
        </div>

        <div class="mt-3" :style="{ height }" ref="elRef" />

        <div v-if="!loading && series.length === 0" class="mt-3 font-mono text-xs text-black/50">
            No data
        </div>
    </div>
</template>
