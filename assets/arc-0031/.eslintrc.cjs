module.exports = {
  root: true,
  extends: ['@mrcointreau/eslint-config-typescript'],
  parserOptions: {
    project: ['./tsconfig.eslint.json']
  },
  ignorePatterns: [
    '*.min.*',
    '*.d.ts',
    'dist',
    'package-lock.json'
  ]
}
