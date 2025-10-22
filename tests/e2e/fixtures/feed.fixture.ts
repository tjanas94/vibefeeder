import { test as base } from "@playwright/test";
import { AuthPage } from "../pages/auth.page.ts";
import { DashboardPage } from "../pages/dashboard.page.ts";
import { FeedPage } from "../pages/feed.page.ts";

/**
 * Test fixtures for feed-related tests
 * Extends auth fixture and provides FeedPage with authenticated context
 */
type FeedFixtures = {
  authPage: AuthPage;
  dashboardPage: DashboardPage;
  feedPage: FeedPage;
};

/**
 * Extended test with feed fixtures
 * Provides authenticated context before each test
 */
export const test = base.extend<FeedFixtures>({
  authPage: async ({ page }, use) => {
    const authPage = new AuthPage(page);
    await use(authPage);
  },

  dashboardPage: async ({ page }, use) => {
    const dashboardPage = new DashboardPage(page);
    await use(dashboardPage);
  },

  feedPage: async ({ page, authPage }, use) => {
    // Authenticate before providing FeedPage
    const testEmail = process.env.E2E_USERNAME;
    const testPassword = process.env.E2E_PASSWORD;

    if (!testEmail || !testPassword) {
      throw new Error(
        "E2E test credentials not configured. Please set E2E_USERNAME and E2E_PASSWORD in .env.test file",
      );
    }

    // Login
    await authPage.gotoLogin();
    await authPage.login(testEmail, testPassword);

    // Wait for redirect to dashboard
    await authPage.isRedirectedToDashboard();

    // Provide FeedPage with authenticated context
    const feedPage = new FeedPage(page);
    await use(feedPage);
  },
});

export { expect } from "@playwright/test";
