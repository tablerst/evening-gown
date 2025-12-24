<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import { RouterView, useRoute } from 'vue-router'

const DefaultLayout = defineAsyncComponent(() => import('@/layouts/DefaultLayout.vue'))
const BlankLayout = defineAsyncComponent(() => import('@/layouts/BlankLayout.vue'))
const AdminLayout = defineAsyncComponent(() => import('@/layouts/AdminLayout.vue'))

const route = useRoute()

const layoutComponent = computed(() => {
    const layout = route.meta.layout
    if (layout === 'blank') return BlankLayout
    if (layout === 'admin') return AdminLayout
    return DefaultLayout
})
</script>

<template>
    <component :is="layoutComponent">
        <RouterView />
    </component>
</template>
