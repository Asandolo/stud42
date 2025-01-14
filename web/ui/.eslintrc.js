module.exports = {
  root: true,
  parser: '@typescript-eslint/parser',

  parserOptions: {
    project: './tsconfig.json',
    ecmaVersion: 2020,
    ecmaFeatures: {
      jsx: true,
    },
  },

  extends: ['airbnb-typescript', 'next', 'next/core-web-vitals', 'prettier'],
  rules: {
    'react/no-danger': 'off', // it's self explainatory that no-danger should be used sparingly
    'react/react-in-jsx-scope': 'off', // next.js does not require react in most components
    'react/prop-types': 'off', // as long as TS strict mode is off this is not required
    'no-console': 'error', // no console statements allowed
    'prettier/prettier': 'off', // don't show prettier errors as it will be fixed when saved anyway
    'import/extensions': 'off', // don't require files extensions
    'import/no-extraneous-dependencies': [
      'error',
      {
        devDependencies: ['**/*.test.*', '**/*.spec.*'],
        peerDependencies: true,
      },
    ],
  },
  settings: {
    react: {
      version: 'detect',
    },
  },
  env: {
    node: true,
    browser: true,
    amd: true,
  },
};
