import path from 'path';
import { defineConfig } from '@rsbuild/core';
import {pluginVue } from '@rsbuild/plugin-vue';
import { pluginLess} from "@rsbuild/plugin-less";

export default defineConfig((env) => {
  console.log(env);
  return {
    html: {
      template: './index.html',
    },
    source: {
      entry: {
        index: './src/main.ts',
      },
    },
    output: {
      path: 'dist',
      filename: '[name].js',
      publicPath: '/static/',
    },
    plugins: [
      pluginVue(),
      pluginLess(),
    ],
    resolve: {
      alias: {
        '@': path.resolve(process.cwd(), 'src'),
      },
    },
    server: {
      host: '0.0.0.0',
      port: 1002,
      open: false,
      proxy: {
        '/api': {
          target: 'http://localhost:8080/',
          changeOrigin: true,
        },
      },
    },
  };
});
