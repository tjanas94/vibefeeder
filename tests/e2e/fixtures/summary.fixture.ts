import { test as base } from "@playwright/test";
import { AuthPage } from "../pages/auth.page.ts";
import { DashboardPage } from "../pages/dashboard.page.ts";
import { FeedPage } from "../pages/feed.page.ts";
import { SummaryPage } from "../pages/summary.page.ts";

/**
 * Test fixtures for summary-related tests
 * Extends base fixtures and provides SummaryPage with authenticated context
 */
type SummaryFixtures = {
  authPage: AuthPage;
  dashboardPage: DashboardPage;
  feedPage: FeedPage;
  summaryPage: SummaryPage;
};

/**
 * Extended test with summary fixtures
 * Provides authenticated context before each test
 */
export const test = base.extend<SummaryFixtures>({
  authPage: async ({ page }, use) => {
    const authPage = new AuthPage(page);
    await use(authPage);
  },

  dashboardPage: async ({ page }, use) => {
    const dashboardPage = new DashboardPage(page);
    await use(dashboardPage);
  },

  feedPage: async ({ page }, use) => {
    const feedPage = new FeedPage(page);
    await use(feedPage);
  },

  summaryPage: async ({ page, authPage }, use) => {
    // Authenticate before providing SummaryPage
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

    // Provide SummaryPage with authenticated context
    const summaryPage = new SummaryPage(page);
    await use(summaryPage);
  },
});

export { expect } from "@playwright/test";
