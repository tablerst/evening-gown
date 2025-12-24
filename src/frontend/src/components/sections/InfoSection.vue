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
        <div class="bg-white p-8 md:p-12 flex flex-col justify-between">
            <div>
                <h3 class="font-display text-xl mb-6 uppercase tracking-wider">{{ t('info.designerTitle') }}</h3>
                <p class="font-body text-sm leading-relaxed mb-8 text-gray-800">
                    {{ t('info.designerBody') }}
                </p>
            </div>
            <span class="font-mono text-xs uppercase tracking-wider text-gray-500 block mt-auto">{{
                t('info.designerSign') }}</span>
        </div>

        <!-- Appointment -->
        <div class="bg-white p-8 md:p-12 flex flex-col justify-between">
            <div>
                <h3 class="font-display text-xl mb-6 uppercase tracking-wider">{{ t('info.appointmentTitle') }}</h3>
                <p class="font-body text-sm leading-relaxed mb-8 text-gray-800">
                    {{ t('info.appointmentBody') }}
                </p>
            </div>

            <div class="mt-auto">
                <button @click="toggle"
                    class="w-full py-4 bg-brand text-white font-mono text-sm hover:bg-black transition-none uppercase tracking-widest">
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
        <div class="bg-white p-8 md:p-12 flex flex-col justify-between">
            <div>
                <h3 class="font-display text-xl mb-6 uppercase tracking-wider">{{ t('info.brandTitle') }}</h3>
                <ul class="font-mono text-sm space-y-4 text-gray-800">
                    <li class="flex justify-between border-b border-gray-100 pb-2"><span>{{ t('info.brandStats.est')
                    }}</span><span>{{ t('info.brandStats.estValue') }}</span>
                    </li>
                    <li class="flex justify-between border-b border-gray-100 pb-2"><span>{{ t('info.brandStats.hq')
                    }}</span><span>{{ t('info.brandStats.hqValue') }}</span>
                    </li>
                    <li class="flex justify-between border-b border-gray-100 pb-2"><span>{{
                        t('info.brandStats.clients')
                            }}</span><span>{{ t('info.brandStats.clientsValue') }}</span></li>
                    <li class="flex justify-between border-b border-gray-100 pb-2">
                        <span>{{ t('info.brandStats.capacity') }}</span><span>{{ t('info.brandStats.capacityValue')
                        }}</span>
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
                    <h3 class="font-display text-xl mb-2 uppercase tracking-wider text-black">{{ t('info.gisTitle') }}
                    </h3>
                    <p class="font-mono text-xs text-gray-500">{{ t('info.gisStatus') }}</p>
                </div>

                <div class="font-mono text-sm">
                    <div class="mb-1 text-gray-600">{{ t('info.gisLocation') }}</div>
                    <div class="text-black tracking-wider">{{ t('info.gisCoordinates') }}</div>
                </div>
            </div>
        </div>

        <!-- <div class="bg-white relative overflow-hidden group min-h-[300px]">
            <iframe width="800" height="460" frameborder='0' scrolling='no' marginheight='0' marginwidth='0'
                src="https://surl.amap.com/4yX3MlhjTdBi"></iframe>
        </div> -->
    </section>
</template>
