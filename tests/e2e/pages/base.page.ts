import { type Page, type Locator, expect } from "@playwright/test";

/**
 * BasePage - Base class for all page objects
 * Provides common functionality for page interactions
 */
export class BasePage {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * Navigate to a specific path
   */
  async goto(path: string = "/"): Promise<void> {
    await this.page.goto(path);
  }

  /**
   * Wait for element to be visible
   */
  async waitForElement(locator: Locator): Promise<void> {
    await locator.waitFor({ state: "visible" });
  }

  /**
   * Get page title
   */
  async getTitle(): Promise<string> {
    return await this.page.title();
  }

  /**
   * Take a screenshot
   */
  async screenshot(name: string): Promise<void> {
    await this.page.screenshot({ path: `screenshots/${name}.png` });
  }

  /**
   * Fill input field by locator
   */
  async fillInput(locator: Locator, value: string): Promise<void> {
    await locator.pressSequentially(value);
    await this.page.waitForTimeout(100); // Small delay to ensure input processing
    await expect(locator).toHaveValue(value);
  }
}
