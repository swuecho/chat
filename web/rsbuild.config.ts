import { defineConfig } from '@rsbuild/core';
import { pluginVue } from '@rsbuild/plugin-vue';
import { pluginLess } from "@rsbuild/plugin-less";

export default defineConfig({
  html: {
    template: './index.html',
  },
  source: {
    entry: {
      index: './src/main.ts',
    },
  },
  plugins: [
    pluginVue(),
    pluginLess(),
  ],
  server: {
    host: '0.0.0.0',
    port: 9002,
    open: false,
    proxy: {
      '/api': {
        target: 'http://localhost:8080/',
        changeOrigin: true,
        // Disable buffering for SSE streaming
        xfwd: true,
        // Handle streaming responses properly
        onProxyReq: (proxyReq, req, res) => {
          // Set headers to prevent buffering
          proxyReq.setHeader('Connection', 'keep-alive');
          proxyReq.setHeader('Cache-Control', 'no-cache');
        },
        onProxyRes: (proxyRes, req, res) => {
          // For SSE responses, ensure no buffering
          if (proxyRes.headers['content-type']?.includes('text/event-stream')) {
            proxyRes.headers['cache-control'] = 'no-cache';
            proxyRes.headers['connection'] = 'keep-alive';
            // Remove content-length to enable streaming
            delete proxyRes.headers['content-length'];
            // Set X-Accel-Buffering header to disable nginx buffering
            proxyRes.headers['x-accel-buffering'] = 'no';
          }
        },
        // Increase timeout for long-running SSE connections
        timeout: 0,
        // Disable buffering at webpack-dev-server level
        secure: false,
        ws: true,
      },
    },
  },
});
