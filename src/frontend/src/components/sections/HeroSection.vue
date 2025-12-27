<script setup lang="ts">
import heroUrl from '@/assets/hero.webp'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
    <section id="process" class="relative h-[100svh] w-full overflow-hidden bg-black text-white">
        <!-- Video Background (Simulated with Image for now) -->
        <div class="absolute inset-0 z-0">
            <img :src="heroUrl" class="w-full h-full object-cover filter-cold-tone opacity-90"
                alt="Factory Production" />
            <!-- Overlay for the 'Midnight Abyss' tint -->
            <div class="absolute inset-0 bg-[#000226] opacity-20 mix-blend-multiply"></div>
        </div>

        <!-- Content -->
        <div class="absolute bottom-12 left-8 md:left-12 z-10 max-w-lg">
            <div class="font-mono text-xs mb-4 tracking-widest opacity-70 text-gray-300">
                {{ t('hero.label') }}
            </div>
            <h1 class="font-mono text-sm md:text-lg leading-loose whitespace-pre-line text-white">
                {{ t('hero.body') }}
            </h1>
        </div>
    </section>
</template>

<style scoped>
.filter-cold-tone {
    filter: contrast(95%) brightness(105%) saturate(85%);
    /* 移除 hue-rotate(180deg) 以恢复正常肤色 */
}
</style>
