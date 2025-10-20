import { Page, Locator } from "@playwright/test";
import { BasePage } from "./base.page";

/**
 * DashboardPage - Page Object for the main dashboard
 * Handles interactions with the dashboard page
 */
export class DashboardPage extends BasePage {
  // Locators
  readonly pageTitle: Locator;
  readonly addFeedButton: Locator;
  readonly feedList: Locator;
  readonly logoutButton: Locator;

  constructor(page: Page) {
    super(page);

    // Initialize locators
    this.pageTitle = page.locator("h1");
    this.addFeedButton = page.locator('button:has-text("Add Feed"), a:has-text("Add Feed")');
    this.feedList = page.locator('[data-testid="feed-list"], .feed-list');
    this.logoutButton = page.locator('button:has-text("Logout"), a:has-text("Logout")');
  }

  /**
   * Navigate to dashboard
   */
  async gotoDashboard(): Promise<void> {
    await this.goto("/dashboard");
  }

  /**
   * Check if user is logged in (dashboard is accessible)
   */
  async isLoggedIn(): Promise<boolean> {
    try {
      await this.page.waitForURL("**/dashboard", { timeout: 5000 });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get page title text
   */
  async getPageTitle(): Promise<string> {
    return (await this.pageTitle.textContent()) || "";
  }

  /**
   * Click add feed button
   */
  async clickAddFeed(): Promise<void> {
    await this.addFeedButton.click();
  }

  /**
   * Perform logout
   */
  async logout(): Promise<void> {
    await this.logoutButton.click();
    await this.page.waitForURL("**/login", { timeout: 5000 });
  }
}
