import { defineConfig } from '@hey-api/openapi-ts'

export default defineConfig({
  input: '../api/openapi/openapi.json',
  output: {
    path: 'src/api/generated',
    format: 'prettier',
    lint: 'eslint',
  },
  plugins: [
    '@hey-api/typescript',
    {
      name: '@hey-api/sdk',
      responseStyle: 'data',
    },
    {
      name: '@hey-api/client-fetch',
      throwOnError: true,
    },
  ],
})
