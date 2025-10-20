import { test as base } from "@playwright/test";
import { AuthPage } from "../pages/auth.page";
import { DashboardPage } from "../pages/dashboard.page";

/**
 * Test fixtures for authentication-related tests
 * Extends base Playwright test with page objects
 */
type AuthFixtures = {
  authPage: AuthPage;
  dashboardPage: DashboardPage;
};

/**
 * Extended test with auth fixtures
 */
export const test = base.extend<AuthFixtures>({
  authPage: async ({ page }, use) => {
    const authPage = new AuthPage(page);
    await use(authPage);
  },

  dashboardPage: async ({ page }, use) => {
    const dashboardPage = new DashboardPage(page);
    await use(dashboardPage);
  },
});

export { expect } from "@playwright/test";
