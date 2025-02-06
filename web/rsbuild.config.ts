import { defineConfig } from '@rsbuild/core';
import {pluginVue } from '@rsbuild/plugin-vue';
import { pluginLess} from "@rsbuild/plugin-less";

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
        },
      },
    },
});
