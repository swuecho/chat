module.exports = {
  root: true,
  extends: ['@antfu'],
  rules: {
    // Browser diagnostics are part of the current error-reporting strategy.
    'no-console': 'off',
    // Composition API callbacks are commonly referenced before declaration.
    '@typescript-eslint/no-use-before-define': 'off',
    // Vue template event listeners use kebab-case throughout this application.
    'vue/custom-event-name-casing': 'off',
  },
}
