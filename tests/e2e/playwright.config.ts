import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './specs',
  outputDir: './test-results',
  timeout: 30000,
  retries: 1,
  use: {
    baseURL: 'http://localhost:3010',
    video: 'on',
    screenshot: 'on',
    trace: 'on-first-retry',
    headless: true,
  },
  reporter: [['html', { outputFolder: './reports' }], ['list']],
  projects: [{ name: 'chromium', use: { browserName: 'chromium' } }],
});
