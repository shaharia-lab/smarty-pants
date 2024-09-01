// cypress.config.ts
import { defineConfig } from "cypress";

export default defineConfig({
  e2e: {
    baseUrl: 'http://localhost:3000',
    setupNodeEvents(on, config) {
      // You can set up any custom Node event listeners here
    },
  },
  video: false,
  defaultCommandTimeout: 10000,
  viewportWidth: 1920,
});