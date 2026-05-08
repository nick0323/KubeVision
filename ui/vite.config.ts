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
    // 生产环境生成 sourcemap，开发环境不生成
    sourcemap: process.env.NODE_ENV === 'production',
    // 代码压缩配置
    minify: 'esbuild',
    // 资源内联限制（4KB 以下内联为 base64）
    assetsInlineLimit: 4096,
    // 代码分割配置
    rollupOptions: {
      output: {
        manualChunks: {
          // React 核心库
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          // 终端模拟器
          'xterm': ['xterm', 'xterm-addon-fit'],
          // YAML 处理
          'yaml': ['js-yaml'],
          // 图标库
          'react-icons': ['react-icons/fa', 'react-icons/fi', 'react-icons/hi', 'react-icons/md', 'react-icons/ri'],
        },
      },
    },
    // 构建报告
    reportCompressedSize: true,
    // 警告阈值（KB）
    chunkSizeWarningLimit: 1000,
  },
});
