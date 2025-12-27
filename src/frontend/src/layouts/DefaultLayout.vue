<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

import { setLocale } from '@/i18n'
import logo from '@/assets/logo.svg'

const isNavCompacted = ref(false)
const isMobileMenuOpen = ref(false)

const { t, locale } = useI18n()
const route = useRoute()

const isZhLocale = computed(() => locale.value === 'zh')

const localeToggleLabel = computed(() => (locale.value === 'zh' ? t('nav.language.en') : t('nav.language.zh')))

let scrollHandler: (() => void) | null = null
let previousScrollOverflow: string | null = null
let previousScrollPaddingRight: string | null = null

let previousBodyOverflow: string | null = null
let previousHtmlOverflow: string | null = null

const closeButtonRef = ref<HTMLButtonElement | null>(null)
const navRef = ref<HTMLElement | null>(null)
const scrollRef = ref<HTMLElement | null>(null)
const rootRef = ref<HTMLElement | null>(null)

const syncNavHeight = () => {
    const navH = navRef.value?.offsetHeight ?? 0
    rootRef.value?.style.setProperty('--site-nav-h', `${navH}px`)
}

const mobileMenuTitle = computed(() => (locale.value === 'zh' ? '导航' : 'Menu'))

const onWindowKeydown = (event: KeyboardEvent) => {
    if (event.key === 'Escape') closeMobileMenu()
}

const lockScroll = () => {
    if (typeof window === 'undefined') return
    if (previousScrollOverflow !== null) return

    const scrollEl = scrollRef.value
    if (!scrollEl) return

    const scrollbarWidth = scrollEl.offsetWidth - scrollEl.clientWidth
    previousScrollOverflow = scrollEl.style.overflow
    previousScrollPaddingRight = scrollEl.style.paddingRight

    scrollEl.style.overflow = 'hidden'
    if (scrollbarWidth > 0) {
        scrollEl.style.paddingRight = `${scrollbarWidth}px`
    }
}

const unlockScroll = () => {
    if (typeof window === 'undefined') return
    if (previousScrollOverflow === null) return

    const scrollEl = scrollRef.value
    if (!scrollEl) return

    scrollEl.style.overflow = previousScrollOverflow
    scrollEl.style.paddingRight = previousScrollPaddingRight ?? ''
    previousScrollOverflow = null
    previousScrollPaddingRight = null
}

const openMobileMenu = async () => {
    isMobileMenuOpen.value = true
    await nextTick()
    closeButtonRef.value?.focus()
}

const closeMobileMenu = () => {
    isMobileMenuOpen.value = false
}

const toggleMobileMenu = () => {
    if (isMobileMenuOpen.value) closeMobileMenu()
    else void openMobileMenu()
}

onMounted(() => {
    if (typeof window === 'undefined') return

    // App-shell scrolling: prevent the document from scrolling, so browser UI is less likely
    // to collapse/expand on swipe; the actual scroll happens in `scrollRef`.
    previousHtmlOverflow = document.documentElement.style.overflow
    previousBodyOverflow = document.body.style.overflow
    document.documentElement.style.overflow = 'hidden'
    document.body.style.overflow = 'hidden'

    syncNavHeight()
    window.addEventListener('resize', syncNavHeight, { passive: true })

    scrollHandler = () => {
        const scrollTop = scrollRef.value?.scrollTop ?? 0
        isNavCompacted.value = scrollTop > 30
    }

    scrollHandler()
    scrollRef.value?.addEventListener('scroll', scrollHandler, { passive: true })
})

onBeforeUnmount(() => {
    if (typeof window !== 'undefined' && scrollHandler) {
        scrollRef.value?.removeEventListener('scroll', scrollHandler)
    }
    scrollHandler = null
    unlockScroll()

    if (typeof window !== 'undefined') {
        window.removeEventListener('resize', syncNavHeight)
    }

    if (typeof window !== 'undefined') {
        document.documentElement.style.overflow = previousHtmlOverflow ?? ''
        document.body.style.overflow = previousBodyOverflow ?? ''
        previousHtmlOverflow = null
        previousBodyOverflow = null
    }
})

const toggleLocale = () => {
    setLocale(locale.value === 'zh' ? 'en' : 'zh')
}

watch(
    () => route.fullPath,
    () => {
        if (isMobileMenuOpen.value) closeMobileMenu()
    }
)

// When navigating between pages (path changes), keep scroll container consistent.
watch(
    () => route.path,
    () => {
        scrollRef.value?.scrollTo({ top: 0 })
    }
)

// Hash navigation should scroll the internal container, not the document.
watch(
    () => route.hash,
    async (hash) => {
        if (!hash) return
        await nextTick()
        const container = scrollRef.value
        if (!container) return
        const target = container.querySelector(hash) as HTMLElement | null
        target?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    },
    { flush: 'post', immediate: true }
)

watch([isNavCompacted, locale], async () => {
    await nextTick()
    syncNavHeight()
})

watch(isMobileMenuOpen, (open) => {
    if (typeof window === 'undefined') return

    if (open) {
        lockScroll()
        window.addEventListener('keydown', onWindowKeydown)
    } else {
        unlockScroll()
        window.removeEventListener('keydown', onWindowKeydown)
    }
})

const mobileNavItems = computed(() => [
    { hash: '#process', label: t('nav.links.process') },
    { hash: '#brief', label: t('nav.links.brief') },
    { hash: '#seasonal', label: t('nav.links.seasonal') },
    { hash: '#catalog', label: t('nav.links.catalog') },
    { hash: '#contact', label: t('nav.links.appointment') },
])
</script>

<template>
    <div ref="rootRef" class="site-root min-h-screen bg-atelier text-charcoal">
        <nav ref="navRef" :class="[
            'lens-nav w-full z-50 px-6 md:px-10 flex justify-between items-center',
            isNavCompacted ? 'lens-nav--compact' : ''
        ]">
            <RouterLink :to="{ name: 'home' }" class="lens-nav__brandTag">
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
            <div class="hidden lg:flex items-center gap-3">
                <button class="nav-ghost nav-link" type="button" @click="toggleLocale">
                    {{ localeToggleLabel }}
                </button>
                <RouterLink class="nav-ghost nav-link" :to="{ name: 'home', hash: '#contact' }">
                    {{ t('nav.cta') }}
                </RouterLink>
            </div>

            <div class="flex lg:hidden items-center gap-2">
                <button class="nav-icon mobile-nav-actions" type="button" @click="toggleLocale"
                    :aria-label="isZhLocale ? 'Switch to English' : '切换到中文'">
                    <span class="text-[0.7rem] tracking-[0.35em]">
                        {{ isZhLocale ? 'EN' : '中' }}
                    </span>
                </button>
                <button class="nav-icon" type="button" @click="toggleMobileMenu" :aria-expanded="isMobileMenuOpen"
                    aria-controls="mobile-nav-drawer" aria-label="Menu">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"
                        aria-hidden="true">
                        <path d="M4 7H20" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
                        <path d="M4 12H20" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
                        <path d="M4 17H20" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
                    </svg>
                </button>
            </div>
        </nav>

        <Teleport to="body">
            <Transition name="mobile-drawer-fade">
                <div v-if="isMobileMenuOpen" class="mobile-drawer lg:hidden" aria-hidden="false">
                    <button class="mobile-drawer__backdrop" type="button" @click="closeMobileMenu"
                        aria-label="Close menu"></button>
                    <Transition name="mobile-drawer-slide">
                        <aside v-if="isMobileMenuOpen" id="mobile-nav-drawer" class="mobile-drawer__panel" role="dialog"
                            aria-modal="true" :aria-label="mobileMenuTitle" @click.stop>
                            <header class="mobile-drawer__header">
                                <span class="mobile-drawer__title">{{ mobileMenuTitle }}</span>
                                <button ref="closeButtonRef" class="nav-icon" type="button" @click="closeMobileMenu"
                                    aria-label="Close">
                                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none"
                                        xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
                                        <path d="M6 6L18 18" stroke="currentColor" stroke-width="1.6"
                                            stroke-linecap="round" />
                                        <path d="M18 6L6 18" stroke="currentColor" stroke-width="1.6"
                                            stroke-linecap="round" />
                                    </svg>
                                </button>
                            </header>

                            <nav class="mobile-drawer__nav" aria-label="Primary">
                                <RouterLink v-for="item in mobileNavItems" :key="item.hash"
                                    :to="{ name: 'home', hash: item.hash }" class="mobile-drawer__link nav-link"
                                    @click="closeMobileMenu">
                                    {{ item.label }}
                                </RouterLink>
                            </nav>

                            <div class="mobile-drawer__actions">
                                <button class="nav-ghost w-full" type="button" @click="toggleLocale">
                                    {{ localeToggleLabel }}
                                </button>
                                <RouterLink class="nav-ghost nav-link w-full text-center"
                                    :to="{ name: 'home', hash: '#contact' }" @click="closeMobileMenu">
                                    {{ t('nav.cta') }}
                                </RouterLink>
                            </div>
                        </aside>
                    </Transition>
                </div>
            </Transition>
        </Teleport>

        <main ref="scrollRef" class="site-scroll">
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
        </main>
    </div>
</template>
