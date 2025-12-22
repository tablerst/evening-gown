<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import TickerSection from '@/components/sections/TickerSection.vue'
import { httpGet } from '@/api/http'

type UpdateItem = {
    id: number
    date: string
    tag: string
    title: string
    body: string
    ref?: string
}

const { t } = useI18n()

const updates = ref<UpdateItem[]>([])
const loading = ref(false)
const errorMsg = ref('')

onMounted(async () => {
    loading.value = true
    errorMsg.value = ''
    try {
        const res = await httpGet<{ items: UpdateItem[] }>('/api/v1/updates?limit=3')
        updates.value = res.items ?? []
    } catch {
        errorMsg.value = '加载失败'
    } finally {
        loading.value = false
    }
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
                    <article v-for="item in updates" :key="item.id" class="bg-white p-6 rounded-none">
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

            <p v-if="errorMsg" class="mt-6 font-mono text-xs tracking-[0.25em] text-red-600">
                {{ errorMsg }}
            </p>
            <p v-else class="mt-6 font-mono text-xs tracking-[0.25em] text-black/50">
                {{ t('updates.note') }}
            </p>
        </div>
    </section>
</template>
