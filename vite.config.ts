import path from 'node:path'
import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const isDev = mode === 'development'
  const rootDir = fileURLToPath(new URL('.', import.meta.url))

  return {
    plugins: [
      vue(),
      VueI18nPlugin({
        // 仅处理本项目语言包目录，避免无关文件被扫描
        include: [path.resolve(rootDir, './src/i18n/locales/**')],
      }),
      // devtools 仅在开发环境启用，避免影响生产构建与 CSP
      isDev ? vueDevTools() : null,
    ].filter(Boolean),
    define: {
      // 让 vue-i18n 的消息编译器代码在构建期被 tree-shaking 掉，避免生成/调用 new Function
      __INTLIFY_JIT_COMPILATION__: false,
      __INTLIFY_DROP_MESSAGE_COMPILER__: true,
      __INTLIFY_PROD_DEVTOOLS__: false,

      // vue-i18n bundler flags（可选，但能进一步减小体积/减少分支）
      __VUE_I18N_FULL_INSTALL__: true,
      __VUE_I18N_LEGACY_API__: false,
      __VUE_I18N_PROD_DEVTOOLS__: false,
    },
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        // 使用 runtime 构建，避免携带消息编译器（配合 locales 预编译输出的 AST）
        'vue-i18n': 'vue-i18n/dist/vue-i18n.runtime.esm-bundler.js',
      },
    },
  }
})
