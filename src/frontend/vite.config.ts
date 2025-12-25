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
      // 允许注册 messageCompiler（用于将预编译生成的 AST 格式化为字符串）。
      // 同时保持 DROP_MESSAGE_COMPILER=true，避免在运行时对「字符串消息」做编译，确保不引入 unsafe-eval。
      __INTLIFY_JIT_COMPILATION__: true,
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
    server: {
      // Scheme A: keep frontend requests same-origin in dev, and proxy API calls to backend.
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
  }
})
