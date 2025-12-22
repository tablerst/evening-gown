<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t, tm } = useI18n()

const bullets = computed(() => tm('couture.bullets') as string[])
</script>

<template>
    <section id="couture" class="section-block couture-section">
        <div class="page-shell">
            <div class="section-heading">
                <div>
                    <p class="eyebrow">{{ t('couture.eyebrow') }}</p>
                    <h2 class="text-2xl md:text-4xl font-serif">{{ t('couture.title') }}</h2>
                </div>
            </div>
            <div class="couture-grid">
                <div class="couture-text project-text-reveal">
                    <p>
                        {{ t('couture.copy') }}
                    </p>
                    <ul>
                        <li v-for="item in bullets" :key="item">{{ item }}</li>
                    </ul>
                </div>
                <div class="couture-panel project-item">
                    <div class="couture-panel__badge">{{ t('couture.badge') }}</div>
                    <h3 class="text-xl font-serif mb-3">{{ t('couture.formTitle') }}</h3>
                    <p class="text-sm text-stone leading-relaxed">
                        {{ t('couture.formCopy') }}
                    </p>
                    <form class="couture-form" :aria-label="t('couture.formTitle')">
                        <label>
                            <span>{{ t('couture.form.city') }}</span>
                            <input type="text" name="city" :placeholder="t('couture.form.placeholder')" />
                        </label>
                        <label>
                            <span>{{ t('couture.form.date') }}</span>
                            <input type="date" name="date" />
                        </label>
                        <button class="hero-button hero-button--primary w-full nav-link" type="button">
                            {{ t('couture.form.submit') }}
                        </button>
                    </form>
                </div>
            </div>
        </div>
    </section>
</template>
