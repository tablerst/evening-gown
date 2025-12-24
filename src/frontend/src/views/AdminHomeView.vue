<script setup lang="ts">
import { useRouter } from 'vue-router'

import { NCard, NGrid, NGridItem } from 'naive-ui'

const router = useRouter()

const go = async (name: string) => {
    await router.push({ name })
}
</script>

<template>
    <div class="max-w-6xl mx-auto">
        <NCard size="large" :bordered="true">
            <div class="font-display text-xl uppercase tracking-wider">Quick Actions</div>
            <div class="mt-2 font-mono text-xs uppercase tracking-[0.25em] text-black/50">
                Manage catalog, content, leads & analytics
            </div>
        </NCard>

        <div class="mt-4">
            <NGrid :cols="2" :x-gap="12" :y-gap="12">
                <NGridItem>
                    <NCard hoverable @click="go('admin-products')">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">Catalog</div>
                        <div class="mt-2 font-display text-lg uppercase tracking-wider">Products</div>
                    </NCard>
                </NGridItem>
                <NGridItem>
                    <NCard hoverable @click="go('admin-updates')">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">Content</div>
                        <div class="mt-2 font-display text-lg uppercase tracking-wider">Updates</div>
                    </NCard>
                </NGridItem>
                <NGridItem>
                    <NCard hoverable @click="go('admin-contacts')">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">Leads</div>
                        <div class="mt-2 font-display text-lg uppercase tracking-wider">Contacts</div>
                    </NCard>
                </NGridItem>
                <NGridItem>
                    <NCard hoverable @click="go('admin-events')">
                        <div class="font-mono text-xs uppercase tracking-[0.25em] text-black/50">Analytics</div>
                        <div class="mt-2 font-display text-lg uppercase tracking-wider">Events</div>
                    </NCard>
                </NGridItem>
            </NGrid>
        </div>
    </div>
</template>
