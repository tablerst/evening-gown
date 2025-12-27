<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { HttpError, httpPost } from '@/api/http'

const { t } = useI18n()

const open = ref(false)
const submitting = ref(false)
const errorMsg = ref('')
const successMsg = ref('')

const form = ref({
    name: '',
    phone: '',
    wechat: '',
    message: '',
})

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

const toggle = () => {
    open.value = !open.value
    errorMsg.value = ''
    successMsg.value = ''
}

const submit = async () => {
    errorMsg.value = ''
    successMsg.value = ''
    submitting.value = true
    try {
        const sourcePage = typeof window === 'undefined' ? '' : window.location.href
        await httpPost('/api/v1/contacts', {
            name: form.value.name,
            phone: form.value.phone,
            wechat: form.value.wechat,
            message: form.value.message,
            source_page: sourcePage,
            ...readUtm(),
        })
        successMsg.value = t('info.contactForm.success')
        form.value = { name: '', phone: '', wechat: '', message: '' }
        open.value = false
    } catch (e) {
        if (e instanceof HttpError && typeof e.payload === 'object' && e.payload && 'error' in (e.payload as any)) {
            errorMsg.value = String((e.payload as any).error)
        } else {
            errorMsg.value = t('info.contactForm.error')
        }
    } finally {
        submitting.value = false
    }
}
</script>

<template>
    <section id="contact" class="grid md:grid-cols-2 lg:grid-cols-4 border-t border-border bg-border gap-[1px]">
        <!-- Designer -->
        <div
            class="bg-white p-8 md:p-12 flex flex-col justify-between group hover:bg-gray-50 transition-colors duration-500">
            <div>
                <h3 class="font-display text-2xl mb-8 uppercase tracking-[0.15em] font-bold text-black">{{
                    t('info.designerTitle') }}</h3>
                <p class="font-sans text-sm leading-8 mb-10 text-black/80 text-justify">
                    {{ t('info.designerBody') }}
                </p>
            </div>
            <div class="mt-auto pt-6 border-t border-black/5">
                <span class="font-mono text-[10px] uppercase tracking-[0.25em] text-black block">{{
                    t('info.designerSign') }}</span>
            </div>
        </div>

        <!-- Appointment -->
        <div
            class="bg-white p-8 md:p-12 flex flex-col justify-between group hover:bg-gray-50 transition-colors duration-500">
            <div>
                <h3 class="font-display text-2xl mb-8 uppercase tracking-[0.15em] font-bold text-black">{{
                    t('info.appointmentTitle') }}</h3>
                <p class="font-sans text-sm leading-8 mb-10 text-black/80">
                    {{ t('info.appointmentBody') }}
                </p>
            </div>

            <div class="mt-auto">
                <button @click="toggle"
                    class="w-full h-12 bg-brand text-white font-mono text-xs hover:bg-black transition-colors duration-300 uppercase tracking-[0.25em] flex items-center justify-center">
                    {{ open ? t('info.contactForm.close') : t('info.appointmentCta') }}
                </button>

                <div v-if="open" class="mt-4 space-y-3">
                    <label class="block">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                            t('info.contactForm.name') }}</div>
                        <input v-model.trim="form.name" class="mt-1 w-full h-10 px-3 border border-border" />
                    </label>
                    <label class="block">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                            t('info.contactForm.phone') }}</div>
                        <input v-model.trim="form.phone" class="mt-1 w-full h-10 px-3 border border-border" />
                    </label>
                    <label class="block">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                            t('info.contactForm.wechat') }}</div>
                        <input v-model.trim="form.wechat" class="mt-1 w-full h-10 px-3 border border-border" />
                    </label>
                    <label class="block">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/60">{{
                            t('info.contactForm.message') }}</div>
                        <textarea v-model.trim="form.message" rows="3"
                            class="mt-1 w-full px-3 py-2 border border-border"></textarea>
                    </label>

                    <button :disabled="submitting" @click="submit"
                        class="w-full py-3 border border-black bg-white font-mono text-xs uppercase tracking-[0.3em] hover:bg-brand hover:text-white transition-none disabled:opacity-60">
                        {{ submitting ? t('info.contactForm.submitting') : t('info.contactForm.submit') }}
                    </button>

                    <p v-if="errorMsg" class="font-mono text-xs text-red-600">{{ errorMsg }}</p>
                </div>

                <p v-if="successMsg" class="mt-3 font-mono text-xs text-brand">{{ successMsg }}</p>
            </div>
        </div>

        <!-- Brand -->
        <div
            class="bg-white p-8 md:p-12 flex flex-col justify-between group hover:bg-gray-50 transition-colors duration-500">
            <div>
                <h3 class="font-display text-2xl mb-8 uppercase tracking-[0.15em] font-bold text-black">{{
                    t('info.brandTitle') }}</h3>
                <ul class="font-mono text-sm space-y-0">
                    <li class="flex justify-between items-center border-b border-black/10 py-4 first:border-t">
                        <span class="text-[10px] uppercase tracking-[0.15em] text-black/50">{{ t('info.brandStats.est')
                            }}</span>
                        <span class="font-bold text-black">{{ t('info.brandStats.estValue') }}</span>
                    </li>
                    <li class="flex justify-between items-center border-b border-black/10 py-4">
                        <span class="text-[10px] uppercase tracking-[0.15em] text-black/50">{{ t('info.brandStats.hq')
                            }}</span>
                        <span class="font-bold text-black">{{ t('info.brandStats.hqValue') }}</span>
                    </li>
                    <li class="flex justify-between items-center border-b border-black/10 py-4">
                        <span class="text-[10px] uppercase tracking-[0.15em] text-black/50">{{
                            t('info.brandStats.clients') }}</span>
                        <span class="font-bold text-black">{{ t('info.brandStats.clientsValue') }}</span>
                    </li>
                    <li class="flex justify-between items-center border-b border-black/10 py-4">
                        <span class="text-[10px] uppercase tracking-[0.15em] text-black/50">{{
                            t('info.brandStats.capacity') }}</span>
                        <span class="font-bold text-black">{{ t('info.brandStats.capacityValue') }}</span>
                    </li>
                </ul>
            </div>
        </div>

        <!-- GIS / Map -->
        <div class="bg-white relative overflow-hidden group min-h-[300px]">
            <!-- Grid Background -->
            <div class="absolute inset-0 opacity-10"
                style="background-image: linear-gradient(#000 1px, transparent 1px), linear-gradient(90deg, #000 1px, transparent 1px); background-size: 40px 40px;">
            </div>

            <!-- Radar/Sonar Effect -->
            <div class="absolute inset-0 flex items-center justify-center">
                <div class="w-64 h-64 border border-black/10 rounded-full animate-pulse"></div>
                <div class="absolute w-48 h-48 border border-black/20 rounded-full"></div>
                <div class="absolute w-2 h-2 bg-black rounded-full"></div>
                <!-- Crosshair lines -->
                <div class="absolute w-full h-[1px] bg-black/10"></div>
                <div class="absolute h-full w-[1px] bg-black/10"></div>
            </div>

            <!-- Content Overlay -->
            <div class="relative z-10 h-full p-8 md:p-12 flex flex-col justify-between pointer-events-none">
                <div>
                    <h3 class="font-display text-2xl mb-3 uppercase tracking-[0.15em] text-black font-bold">{{
                        t('info.gisTitle') }}
                    </h3>
                    <div class="flex items-center gap-2">
                        <div class="w-1.5 h-1.5 bg-brand rounded-none animate-pulse"></div>
                        <p class="font-mono text-[10px] uppercase tracking-[0.2em] text-brand">{{ t('info.gisStatus') }}
                        </p>
                    </div>
                </div>

                <div class="font-mono text-xs">
                    <div class="mb-2 text-black/50 uppercase tracking-widest">{{ t('info.gisLocation') }}</div>
                    <div class="text-black tracking-[0.15em] font-bold text-sm">{{ t('info.gisCoordinates') }}</div>
                </div>
            </div>
        </div>

        <!-- <div class="bg-white relative overflow-hidden group min-h-[300px]">
            <iframe width="800" height="460" frameborder='0' scrolling='no' marginheight='0' marginwidth='0'
                src="https://surl.amap.com/4yX3MlhjTdBi"></iframe>
        </div> -->
    </section>
</template>
