<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { setLocale } from '@/i18n'
import logo from '@/assets/logo.svg'

const isNavCompacted = ref(false)

const { t, locale } = useI18n()

const localeToggleLabel = computed(() => (locale.value === 'zh' ? t('nav.language.en') : t('nav.language.zh')))

let scrollHandler: (() => void) | null = null

onMounted(() => {
    if (typeof window === 'undefined') return

    scrollHandler = () => {
        isNavCompacted.value = window.scrollY > 30
    }

    scrollHandler()
    window.addEventListener('scroll', scrollHandler, { passive: true })
})

onBeforeUnmount(() => {
    if (typeof window !== 'undefined' && scrollHandler) {
        window.removeEventListener('scroll', scrollHandler)
    }
    scrollHandler = null
})

const toggleLocale = () => {
    setLocale(locale.value === 'zh' ? 'en' : 'zh')
}
</script>

<template>
    <div class="site-root min-h-screen bg-atelier text-charcoal">
        <nav :class="[
            'lens-nav fixed top-0 w-full z-50 px-6 md:px-10 flex justify-between items-center transition-all duration-300',
            isNavCompacted ? 'py-3 lens-nav--compact' : 'py-5'
        ]">
            <RouterLink :to="{ name: 'home' }" :class="[
                'lens-nav__brandTag',
                isNavCompacted ? 'lens-nav__brandTag--compact' : 'lens-nav__brandTag--relaxed',
            ]">
                <span class="lens-nav__brandLogoCrop">
                    <img :src="logo" :alt="t('nav.brandLine1')" class="lens-nav__brandLogo" />
                </span>
            </RouterLink>
            <div class="lens-nav__links">
                <RouterLink :to="{ name: 'home', hash: '#process' }" class="nav-link">{{ t('nav.links.process') }}
                </RouterLink>
                <RouterLink :to="{ name: 'home', hash: '#brief' }" class="nav-link">{{ t('nav.links.brief') }}
                </RouterLink>
                <RouterLink :to="{ name: 'home', hash: '#seasonal' }" class="nav-link">{{ t('nav.links.seasonal') }}
                </RouterLink>
                <RouterLink :to="{ name: 'home', hash: '#catalog' }" class="nav-link">{{ t('nav.links.catalog') }}
                </RouterLink>
                <RouterLink :to="{ name: 'home', hash: '#contact' }" class="nav-link">{{ t('nav.links.appointment') }}
                </RouterLink>
            </div>
            <div class="flex items-center gap-3">
                <button class="nav-ghost nav-link" type="button" @click="toggleLocale">
                    {{ localeToggleLabel }}
                </button>
                <button class="nav-ghost nav-link" type="button">{{ t('nav.cta') }}</button>
            </div>
        </nav>

        <slot />

        <footer class="site-footer">
            <div class="site-footer__glow" aria-hidden="true"></div>
            <p class="eyebrow">{{ t('footer.tagline') }}</p>
            <h2 class="text-3xl md:text-5xl font-serif tracking-[0.3em] mt-4">{{ t('footer.brand') }}</h2>
            <div class="site-footer__links mt-6">
                <a href="#" class="nav-link">{{ t('footer.instagram') }}</a>
                <span>•</span>
                <a href="#" class="nav-link">{{ t('footer.wechat') }}</a>
                <span>•</span>
                <a href="#" class="nav-link">{{ t('footer.email') }}</a>
            </div>
            <p class="site-footer__legal">{{ t('footer.legal') }}</p>
        </footer>
    </div>
</template>
