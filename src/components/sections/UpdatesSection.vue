<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import TickerSection from '@/components/sections/TickerSection.vue'

type UpdateItem = {
    date: string
    tag: string
    title: string
    body: string
    ref?: string
}

const { t, locale } = useI18n()

const updates = computed<UpdateItem[]>(() => {
    if (locale.value === 'zh') {
        return [
            {
                date: '2025-12-18',
                tag: '交付',
                title: '春节排期预告',
                body: '1 月下旬起产线进入节前高峰。建议 2 月中旬前出货的订单在 01/05 前确认面料与尺码表。',
                ref: 'OPS-2025-CNY',
            },
            {
                date: '2025-12-10',
                tag: '展会',
                title: '上海时装周展位更新',
                body: '现场样衣以 2025 S/S 系列为主，支持面料触感档案与工序说明现场查阅。',
                ref: 'EVENT-SHFW-A12',
            },
            {
                date: '2025-12-02',
                tag: '新品',
                title: '2025 春夏小批量补单开放',
                body: '核心款支持 10 件起订；颜色与尺码按标准档案执行，确保批量一致性与交付稳定。',
                ref: 'CAT-SS25-REPLENISH',
            },
        ]
    }

    return [
        {
            date: '2025-12-18',
            tag: 'Delivery',
            title: 'CNY production schedule',
            body: 'Late January enters peak capacity. For shipments before mid-Feb, please confirm fabrics and size charts by 01/05.',
            ref: 'OPS-2025-CNY',
        },
        {
            date: '2025-12-10',
            tag: 'Trade Show',
            title: 'Shanghai Fashion Week booth update',
            body: 'On-site samples focus on 2025 S/S. Spec sheets and process notes are available for quick review.',
            ref: 'EVENT-SHFW-A12',
        },
        {
            date: '2025-12-02',
            tag: 'New',
            title: 'S/S 2025 replenishment opened',
            body: 'Core styles support MOQ from 10 pcs. Colorways and sizing follow the standard archive to keep batch consistency.',
            ref: 'CAT-SS25-REPLENISH',
        },
    ]
})
</script>

<template>
    <section id="brief" class="bg-white border-b border-border">
        <TickerSection />

        <div class="px-4 md:px-8 py-10">
            <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-black pb-4">
                <div>
                    <h2
                        class="font-sans font-bold uppercase tracking-[0.28em] text-brand text-xl md:text-2xl leading-none">
                        {{ t('updates.title') }}
                    </h2>
                </div>

                <a href="#"
                    class="inline-flex items-center font-mono text-xs uppercase tracking-[0.3em] leading-none text-brand border border-brand px-4 py-3 hover:bg-brand hover:text-white transition-none self-start md:self-auto">
                    {{ t('updates.cta') }}
                </a>
            </div>

            <div class="mt-6 border border-border">
                <div class="grid md:grid-cols-3 gap-[1px] bg-border">
                    <article v-for="(item, index) in updates" :key="index" class="bg-white p-6 rounded-none">
                        <div class="flex items-start justify-between gap-4">
                            <time class="font-mono text-xs tracking-[0.25em] text-black/60">{{ item.date }}</time>
                            <span
                                class="font-mono text-[10px] tracking-[0.3em] uppercase bg-brand text-white px-2 py-1 rounded-none">
                                {{ item.tag }}
                            </span>
                        </div>

                        <h3
                            class="mt-4 font-sans font-semibold uppercase tracking-[0.22em] text-sm text-black hover:text-brand transition-none">
                            {{ item.title }}
                        </h3>

                        <p class="mt-4 text-sm leading-relaxed text-black/70">
                            {{ item.body }}
                        </p>

                        <div v-if="item.ref"
                            class="mt-6 pt-4 border-t border-border font-mono text-xs tracking-[0.25em] text-black/60">
                            REF: {{ item.ref }}
                        </div>
                    </article>
                </div>
            </div>

            <p class="mt-6 font-mono text-xs tracking-[0.25em] text-black/50">
                {{ t('updates.note') }}
            </p>
        </div>
    </section>
</template>
