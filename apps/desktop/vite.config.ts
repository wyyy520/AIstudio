import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
      },
    },
    server: {
      port: 5173,
      strictPort: false,
      host: '0.0.0.0',
      proxy: {
        '/api': {
          target: env.VITE_API_BASE_URL || 'http://localhost:8081',
          changeOrigin: true,
        },
        '/ws': {
          target: env.VITE_WS_URL || 'ws://localhost:8081',
          ws: true,
        },
      },
    },
    build: {
      target: 'esnext',
      outDir: 'dist',
      rollupOptions: {
        external: [
          '@tauri-apps/plugin-fs',
          '@tauri-apps/plugin-dialog',
        ],
      },
    },
  }
})