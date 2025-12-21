<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import bgUrl from '@/assets/bg.webp'
import logoUrl from '@/assets/logo.svg'

const { t } = useI18n()
</script>

<template>
    <div class="relative min-h-screen w-full overflow-hidden flex items-center justify-center bg-brand">
        <!-- 1. Background Layer: Cold Silk (Liquid Metal) -->
        <div class="absolute inset-0 z-0 hero-background" :style="{ backgroundImage: `url(${bgUrl})` }"></div>

        <!-- 2. Geometric Overlay: Vector Grid -->
        <div class="absolute inset-0 z-10 overlay-grid pointer-events-none"></div>

        <!-- 3. Main Card: The White Archive Document -->
        <div
            class="relative z-20 bg-white border border-brand shadow-[4px_4px_0px_#000226] md:shadow-[8px_8px_0px_#000226] flex flex-col items-center p-6 md:p-12 w-[90%] max-w-[480px] select-none">
            <!-- Header Specs -->
            <div class="w-full flex justify-between items-center mb-6 md:mb-8 border-b border-slate pb-4">
                <span class="font-mono text-[10px] text-gray-400 tracking-wider">{{ t('preview.ref') }}</span>
                <span class="font-mono text-[10px] text-brand font-bold tracking-wider">{{ t('preview.security')
                }}</span>
            </div>

            <!-- Logo (Inverted for White Background) -->
            <div class="mb-8 md:mb-10 mt-4">
                <img :src="logoUrl" alt="FLEURLIS" class="w-56 md:w-72 invert opacity-90" />
            </div>

            <!-- Slogan / Identity -->
            <div class="text-center mb-8 md:mb-12 space-y-3">
                <h2 class="font-display text-base md:text-xl tracking-widest text-brand uppercase whitespace-nowrap">
                    {{ t('preview.heading') }}</h2>
                <p class="font-sans text-[10px] tracking-[0.3em] text-gray-500 uppercase font-medium">
                    {{ t('preview.subheading') }}
                </p>
            </div>

            <!-- Action Button -->
            <a href="/"
                class="w-full group relative flex items-center justify-center bg-brand text-white h-14 border border-brand hover:bg-white hover:text-brand transition-colors duration-0">
                <span class="font-mono text-xs tracking-[0.2em] font-bold uppercase">
                    {{ t('preview.enter') }}
                </span>
            </a>

            <!-- Footer Specs -->
            <div class="mt-6 flex flex-col items-center gap-1">
                <div class="flex items-center gap-2">
                    <div class="w-1.5 h-1.5 bg-orange-500 rounded-none animate-pulse"></div>
                    <span class="font-mono text-[10px] text-gray-400 uppercase">{{ t('preview.ready') }}</span>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* 1. Background Processing */
.hero-background {
    background-size: cover;
    background-position: center;
    /* Japanese Cold Drama Filter + Darkening for Liquid Metal look */
    filter: contrast(110%) saturate(0) brightness(60%) sepia(20%) hue-rotate(180deg);
    opacity: 0.9;
}

/* 2. Geometric Grid */
.overlay-grid {
    background-image:
        linear-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px),
        linear-gradient(90deg, rgba(255, 255, 255, 0.05) 1px, transparent 1px);
    background-size: 40px 40px;
    opacity: 0.3;
}
</style>
