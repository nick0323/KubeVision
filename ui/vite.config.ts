import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    extensions: ['.ts', '.tsx', '.js', '.jsx', '.json'],
    alias: {
      '@components': path.resolve(__dirname, 'src/components'),
      '@pages': path.resolve(__dirname, 'src/pages'),
      '@hooks': path.resolve(__dirname, 'src/hooks'),
      '@utils': path.resolve(__dirname, 'src/utils'),
      '@types': path.resolve(__dirname, 'src/types'),
      '@constants': path.resolve(__dirname, 'src/constants'),
      '@styles': path.resolve(__dirname, 'src/styles'),
    },
  },
  server: {
    port: 3000,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true,  // 启用 WebSocket 代理
        secure: false,  // 不验证 SSL
        configure: (proxy, _options) => {
          proxy.on('error', (err, _req, _res) => {
            console.log('[proxy]', 'error:', err);
          });
          proxy.on('proxyReq', (proxyReq, req, _res) => {
            console.log('[proxy]', 'proxyReq:', req.method, req.url);
          });
          proxy.on('proxyRes', (proxyRes, req, _res) => {
            console.log('[proxy]', 'proxyRes:', proxyRes.statusCode, req.url);
          });
        }
      }
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
});
