<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

import HeroSection from '@/components/sections/HeroSection.vue'
import GallerySection from '@/components/sections/GallerySection.vue'
import AtelierSection from '@/components/sections/AtelierSection.vue'
import CoutureSection from '@/components/sections/CoutureSection.vue'

import { useHomeAnimations } from '@/composables/useHomeAnimations'
import { useSilkEnvironment } from '@/composables/useSilkEnvironment'

import { createSilkRenderer } from '../modules/silk/silkRenderer'

const silkContainer = ref<HTMLDivElement | null>(null)
const isReducedMotion = ref(false)
const shouldUseStaticSilk = ref(false)

const onSilkContainerEl = (el: HTMLDivElement | null) => {
    silkContainer.value = el
}
const {
    heroScrollProgress,
    evaluateFallback: evaluateSilkFallback,
    syncMotionPreference,
    dispose: disposeSilk,
} = createSilkRenderer({
    containerRef: silkContainer,
    isReducedMotion,
    shouldUseStaticSilk,
})

const { initAnimations, disposeAnimations } = useHomeAnimations(heroScrollProgress)

useSilkEnvironment({
    evaluateSilkFallback,
    syncMotionPreference,
})

onMounted(() => {
    initAnimations()
})

onBeforeUnmount(() => {
    disposeSilk()
    disposeAnimations()
})
</script>

<template>
    <HeroSection :on-silk-container-el="onSilkContainerEl" :is-reduced-motion="isReducedMotion"
        :should-use-static-silk="shouldUseStaticSilk" />

    <main class="bg-atelier">
        <GallerySection />
        <AtelierSection />
        <CoutureSection />
    </main>
</template>
