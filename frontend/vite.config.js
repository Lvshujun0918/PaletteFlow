import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import viteCompression from 'vite-plugin-compression'

export default defineConfig(({ mode }) => {
  const isProd = mode === 'production'

  return {
    plugins: [
      vue(),
      ...(isProd
        ? [
            viteCompression({
              verbose: false,
              threshold: 1024,
              algorithm: 'gzip',
              ext: '.gz',
              deleteOriginFile: false
            }),
            viteCompression({
              verbose: false,
              threshold: 1024,
              algorithm: 'brotliCompress',
              ext: '.br',
              deleteOriginFile: false
            })
          ]
        : [])
    ],
    esbuild: isProd
      ? {
          drop: ['console', 'debugger'],
          legalComments: 'none'
        }
      : undefined,
    build: {
      target: 'es2018',
      minify: isProd ? 'terser' : false,
      terserOptions: isProd
        ? {
            compress: {
              drop_console: true,
              drop_debugger: true,
              passes: 2
            },
            format: {
              comments: false
            }
          }
        : undefined,
      cssMinify: true,
      sourcemap: false,
      reportCompressedSize: true,
      chunkSizeWarningLimit: 1200,
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (!id.includes('node_modules')) return
            if (id.includes('driver.js')) return 'driver-vendor'
            if (id.includes('lucide-vue-next')) return 'icons-vendor'
            if (id.includes('axios')) return 'http-vendor'
            if (id.includes('vue-router')) return 'router-vendor'
            if (id.includes('/vue/')) return 'vue-vendor'
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
