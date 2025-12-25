import { watch } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

import { appEnv } from '@/config/env'
import { getAdminToken } from '@/admin/auth'
import { i18n } from '@/i18n'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        name: 'home',
        component: () => import('../views/HomeView.vue'),
        meta: {
            layout: 'default',
            title: 'White Phantom',
        },
    },
    {
        path: '/preview',
        name: 'preview',
        component: () => import('../views/PreviewView.vue'),
        meta: {
            layout: 'blank',
            title: 'Preview · White Phantom',
        },
    },
    {
        path: '/products/:id',
        name: 'product-detail',
        component: () => import('../views/ProductDetailView.vue'),
        meta: {
            layout: 'default',
            titleKey: 'productDetail.routeTitle',
        },
    },
    {
        path: '/updates',
        name: 'updates',
        component: () => import('../views/UpdatesView.vue'),
        meta: {
            layout: 'default',
            title: 'Updates · White Phantom',
        },
    },
    {
        path: '/updates/:id',
        name: 'update-detail',
        component: () => import('../views/UpdateDetailView.vue'),
        meta: {
            layout: 'default',
            title: 'Update · White Phantom',
        },
    },

    // Admin backoffice
    {
        path: '/admin/login',
        name: 'admin-login',
        component: () => import('../views/AdminLoginView.vue'),
        meta: {
            layout: 'blank',
            titleKey: 'admin.titles.login',
        },
    },
    {
        path: '/admin',
        name: 'admin-home',
        component: () => import('../views/AdminHomeView.vue'),
        meta: {
            layout: 'admin',
            titleKey: 'admin.titles.home',
        },
    },
    {
        path: '/admin/products',
        name: 'admin-products',
        component: () => import('../views/AdminProductsView.vue'),
        meta: {
            layout: 'admin',
            titleKey: 'admin.titles.products',
        },
    },
    {
        path: '/admin/updates',
        name: 'admin-updates',
        component: () => import('../views/AdminUpdatesView.vue'),
        meta: {
            layout: 'admin',
            titleKey: 'admin.titles.updates',
        },
    },
    {
        path: '/admin/contacts',
        name: 'admin-contacts',
        component: () => import('../views/AdminContactsView.vue'),
        meta: {
            layout: 'admin',
            titleKey: 'admin.titles.contacts',
        },
    },
    {
        path: '/admin/events',
        name: 'admin-events',
        component: () => import('../views/AdminEventsView.vue'),
        meta: {
            layout: 'admin',
            titleKey: 'admin.titles.events',
        },
    },
    {
        path: '/:pathMatch(.*)*',
        redirect: '/',
    },
]

const router = createRouter({
    history: createWebHistory(),
    routes,
    scrollBehavior(to) {
        if (to.hash) {
            return {
                el: to.hash,
                behavior: 'smooth',
            }
        }
        return { top: 0 }
    },
})

router.beforeEach((to) => {
    // 预览模式：站点全局只允许访问 /preview
    if (appEnv.previewMode && to.name !== 'preview') {
        return { name: 'preview', replace: true }
    }

    // 非预览模式：避免用户误入 /preview
    if (!appEnv.previewMode && to.name === 'preview') {
        return { name: 'home', replace: true }
    }

    // Admin routes guard (simple token check).
    if (typeof to.path === 'string' && to.path.startsWith('/admin')) {
        if (to.name === 'admin-login') return
        const token = getAdminToken()
        if (!token) {
            return { name: 'admin-login', replace: true }
        }
    }
})

const setDocumentTitle = (to: any) => {
    if (typeof document === 'undefined') return

    const titleKey = typeof to?.meta?.titleKey === 'string' ? (to.meta.titleKey as string) : null
    if (titleKey) {
        document.title = i18n.global.t(titleKey)
        return
    }

    const title = typeof to?.meta?.title === 'string' ? (to.meta.title as string) : null
    if (title) {
        document.title = title
    }
}

router.afterEach((to) => {
    setDocumentTitle(to)
})

// When locale changes, keep title in sync without requiring navigation.
watch(i18n.global.locale, () => {
    setDocumentTitle(router.currentRoute.value)
})

export default router
