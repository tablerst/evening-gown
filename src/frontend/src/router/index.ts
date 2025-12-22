import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

import { appEnv } from '@/config/env'
import { getAdminToken } from '@/admin/auth'

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
            title: 'Product · White Phantom',
        },
    },

    // Admin backoffice
    {
        path: '/admin/login',
        name: 'admin-login',
        component: () => import('../views/AdminLoginView.vue'),
        meta: {
            layout: 'blank',
            title: 'Admin Login · White Phantom',
        },
    },
    {
        path: '/admin',
        name: 'admin-home',
        component: () => import('../views/AdminHomeView.vue'),
        meta: {
            layout: 'blank',
            title: 'Admin · White Phantom',
        },
    },
    {
        path: '/admin/products',
        name: 'admin-products',
        component: () => import('../views/AdminProductsView.vue'),
        meta: {
            layout: 'blank',
            title: 'Admin Products · White Phantom',
        },
    },
    {
        path: '/admin/updates',
        name: 'admin-updates',
        component: () => import('../views/AdminUpdatesView.vue'),
        meta: {
            layout: 'blank',
            title: 'Admin Updates · White Phantom',
        },
    },
    {
        path: '/admin/contacts',
        name: 'admin-contacts',
        component: () => import('../views/AdminContactsView.vue'),
        meta: {
            layout: 'blank',
            title: 'Admin Contacts · White Phantom',
        },
    },
    {
        path: '/admin/events',
        name: 'admin-events',
        component: () => import('../views/AdminEventsView.vue'),
        meta: {
            layout: 'blank',
            title: 'Admin Events · White Phantom',
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

router.afterEach((to) => {
    if (typeof document === 'undefined') return

    const title = typeof to.meta.title === 'string' ? to.meta.title : null
    if (title) {
        document.title = title
    }
})

export default router
