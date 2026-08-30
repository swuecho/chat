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
    '@hey-api/sdk',
    '@hey-api/client-fetch',
  ],
})
