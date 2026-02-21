import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const isProd = mode === 'production'

  return {
    plugins: [vue()],
    esbuild: isProd
      ? {
          drop: ['console', 'debugger'],
          legalComments: 'none'
        }
      : undefined,
    build: {
      target: 'es2018',
      minify: 'esbuild',
      cssMinify: true,
      sourcemap: false,
      reportCompressedSize: true,
      chunkSizeWarningLimit: 1200,
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (!id.includes('node_modules')) return
            if (id.includes('vue')) return 'vue-vendor'
            return 'vendor'
          }
        }
      }
    },
    server: {
      port: 5173,
      proxy: {
        '/api': {
          target: 'http://localhost:5208',
          changeOrigin: true
        }
      }
    }
  }
})
