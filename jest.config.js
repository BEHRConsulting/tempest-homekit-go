module.exports = {
  testEnvironment: 'node',
  setupFiles: ['<rootDir>/jest.setup.js'],
  testMatch: [
    '<rootDir>/pkg/web/static/__tests__/**/*.test.js'
  ]
};
