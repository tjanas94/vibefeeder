import type { Page, Locator } from "@playwright/test";
import { BasePage } from "./base.page.ts";
import { FEED_FIXTURES } from "../fixtures/feeds.fixture.ts";

/**
 * FeedPage - Page Object for feed management
 * Handles interactions with feed list, add/edit/delete operations
 */
export class FeedPage extends BasePage {
  // Search and filter locators
  readonly feedFilterForm: Locator;
  readonly searchInput: Locator;
  readonly statusFilterGroup: Locator;
  readonly statusFilterAllBtn: Locator;
  readonly statusFilterWorkingBtn: Locator;
  readonly statusFilterPendingBtn: Locator;
  readonly statusFilterErrorBtn: Locator;

  // Modal and form locators
  readonly feedFormModal: Locator;
  readonly feedForm: Locator;
  readonly feedFormNameInput: Locator;
  readonly feedFormUrlInput: Locator;
  readonly feedFormSubmitBtn: Locator;
  readonly feedFormCancelBtn: Locator;
  readonly feedFormError: Locator;

  // Delete confirmation locators
  readonly deleteConfirmation: Locator;
  readonly deleteConfirmationCancelBtn: Locator;

  constructor(page: Page) {
    super(page);

    // Search and filter
    this.feedFilterForm = page.getByTestId("feed-filter-form");
    this.searchInput = page.getByTestId("feed-search-input");
    this.statusFilterGroup = page.getByTestId("status-filter-group");
    this.statusFilterAllBtn = page.getByTestId("status-filter-all");
    this.statusFilterWorkingBtn = page.getByTestId("status-filter-working");
    this.statusFilterPendingBtn = page.getByTestId("status-filter-pending");
    this.statusFilterErrorBtn = page.getByTestId("status-filter-error");

    // Modal and form
    this.feedFormModal = page.locator("#feed-form-modal-content");
    this.feedForm = page.getByTestId("feed-form");
    this.feedFormNameInput = page.getByTestId("feed-form-name-input");
    this.feedFormUrlInput = page.getByTestId("feed-form-url-input");
    this.feedFormSubmitBtn = page.getByTestId("feed-form-submit-btn");
    this.feedFormCancelBtn = page.getByTestId("feed-form-cancel-btn");
    this.feedFormError = page.getByTestId("feed-form-error");

    // Delete confirmation
    this.deleteConfirmation = page.getByTestId("delete-confirmation");
    this.deleteConfirmationCancelBtn = page.getByTestId("delete-confirmation-cancel-btn");
  }

  /**
   * Generate a feed URL from feed name using slug-based approach
   * Creates predictable URLs based on feed name instead of timestamps
   */
  generateFeedUrlFromName(feedName: string, baseUrl = FEED_FIXTURES.basic): string {
    const slug = feedName.toLowerCase().replace(/[^a-z0-9]+/g, "-");
    return `${baseUrl}?test=${slug}`;
  }

  /**
   * Navigate to dashboard feeds page
   */
  async gotoFeeds(): Promise<void> {
    await this.goto("/dashboard");
  }

  /**
   * Get feed list item locator by feed ID (handles both table and card views)
   */
  getFeedListItem(feedId: string): Locator {
    return this.page.locator(
      `[data-testid="feed-list-item-${feedId}"], [data-testid="feed-card-${feedId}"]`,
    );
  }

  /**
   * Get feed name element by feed ID (handles both table and card views)
   */
  getFeedNameElement(feedId: string): Locator {
    // Both views use same testid for name
    return this.page.getByTestId(`feed-name-${feedId}`).first();
  }

  /**
   * Get feed URL element by feed ID (handles both table and card views)
   */
  getFeedUrlElement(feedId: string): Locator {
    // Both views use same testid for URL
    return this.page.getByTestId(`feed-url-${feedId}`).first();
  }

  /**
   * Get feed status element by feed ID (handles both table and card views)
   */
  getFeedStatusElement(feedId: string): Locator {
    // Both views use same testid for status
    return this.page.getByTestId(`feed-status-${feedId}`).first();
  }

  /**
   * Get feed edit button by feed ID (handles both table and card views)
   */
  getFeedEditBtn(feedId: string): Locator {
    // Both views use same testid for edit button
    return this.page.getByTestId(`feed-edit-btn-${feedId}`).first();
  }

  /**
   * Get feed delete button by feed ID (handles both table and card views)
   */
  getFeedDeleteBtn(feedId: string): Locator {
    // Both views use same testid for delete button
    return this.page.getByTestId(`feed-delete-btn-${feedId}`).first();
  }

  /**
   * Get delete confirmation button by feed ID
   */
  getDeleteConfirmationBtn(feedId: string): Locator {
    return this.page.getByTestId(`delete-confirmation-confirm-btn-${feedId}`);
  }

  /**
   * Search for feeds by name
   */
  async searchFeeds(query: string): Promise<void> {
    await this.searchInput.clear();
    await this.fillInput(this.searchInput, query);
  }

  /**
   * Filter feeds by status
   */
  async filterByStatus(status: "all" | "working" | "pending" | "error"): Promise<void> {
    const filterBtn = {
      all: this.statusFilterAllBtn,
      working: this.statusFilterWorkingBtn,
      pending: this.statusFilterPendingBtn,
      error: this.statusFilterErrorBtn,
    }[status];

    await filterBtn.click();
  }

  /**
   * Open add feed form
   */
  async openAddFeedForm(): Promise<void> {
    await this.page.getByTestId("add-feed-button").click();
    await this.feedForm.waitFor({ state: "visible" });
  }

  /**
   * Open edit feed form by feed ID
   */
  async openEditFeedForm(feedId: string): Promise<void> {
    const editBtn = this.getFeedEditBtn(feedId);
    await editBtn.click();
    await this.feedForm.waitFor({ state: "visible" });
  }

  /**
   * Fill feed form with name and URL
   */
  async fillFeedForm(name: string, url: string): Promise<void> {
    await this.fillInput(this.feedFormNameInput, name);
    await this.fillInput(this.feedFormUrlInput, url);
  }

  /**
   * Submit feed form and wait for network idle
   */
  async submitFeedForm(): Promise<void> {
    await this.feedFormSubmitBtn.click();
  }

  /**
   * Add a new feed (open form, fill, submit, and wait for completion)
   */
  async addFeed(name: string, url: string): Promise<void> {
    // Click add feed button to open form
    await this.page.getByTestId("add-feed-button").click();

    // Wait for form to appear
    await this.feedForm.waitFor({ state: "visible" });

    // Fill form
    await this.fillFeedForm(name, url);

    // Submit form
    await this.feedFormSubmitBtn.click();

    // Wait for success (form should close and list should refresh)
    await this.feedForm.waitFor({ state: "hidden" });
  }

  /**
   * Get feed by name from the currently visible list (handles both table and card views)
   * NOTE: Caller must explicitly call searchFeeds() before this method to filter results
   */
  async getFeedByName(feedName: string): Promise<{ id: string; element: Locator } | null> {
    // Find feed items from either table or cards (searches only in currently visible items)
    const tableRows = await this.page.locator('[data-testid^="feed-list-item-"]').all();
    const cardRows = await this.page.locator('[data-testid^="feed-card-"]').all();
    const allRows = [...tableRows, ...cardRows];

    for (const row of allRows) {
      const nameText = await row.locator(`[data-testid^="feed-name-"]`).textContent();
      if (nameText?.includes(feedName)) {
        const testId = await row.getAttribute("data-testid");
        // Extract feed ID from either "feed-list-item-{id}" or "feed-card-{id}"
        const feedId =
          testId?.replace("feed-list-item-", "") || testId?.replace("feed-card-", "") || "";
        return { id: feedId, element: row };
      }
    }

    return null;
  }

  /**
   * Get feed status text
   */
  async getFeedStatus(feedId: string): Promise<string | null> {
    const statusElement = this.getFeedStatusElement(feedId);
    await statusElement.waitFor({ state: "visible" });
    // Extract text from badge (e.g., "✅ OK", "⏳ Pending", "❌ Error")
    const text = await statusElement.textContent();
    return text?.trim() || null;
  }

  /**
   * Wait for feed status to change (without reloading page)
   */
  async waitForFeedStatus(
    feedId: string,
    expectedStatus: "OK" | "Pending" | "Error",
    timeoutMs = 30000,
  ): Promise<void> {
    const startTime = Date.now();
    const pollInterval = 2000; // Poll every 2 seconds

    while (Date.now() - startTime < timeoutMs) {
      await this.page.waitForTimeout(pollInterval);
      await this.page.reload();

      try {
        const status = await this.getFeedStatus(feedId);
        if (status?.includes(expectedStatus)) {
          return;
        }
      } catch {
        // Element not found, continue polling
        continue;
      }
    }

    throw new Error(`Feed status did not change to "${expectedStatus}" within ${timeoutMs}ms`);
  }

  /**
   * Edit feed
   */
  async editFeed(feedId: string, newName?: string, newUrl?: string): Promise<void> {
    // Click edit button
    const editBtn = this.getFeedEditBtn(feedId);
    await editBtn.click();

    // Wait for form to appear
    await this.feedForm.waitFor({ state: "visible" });

    // Update fields if provided
    if (newName) {
      await this.feedFormNameInput.clear();
      await this.fillInput(this.feedFormNameInput, newName);
    }
    if (newUrl) {
      await this.feedFormUrlInput.clear();
      await this.fillInput(this.feedFormUrlInput, newUrl);
    }

    // Submit form
    await this.feedFormSubmitBtn.click();

    // Wait for success (form should close and list should refresh)
    await this.feedForm.waitFor({ state: "hidden" });
  }

  /**
   * Delete feed
   */
  async deleteFeed(feedId: string): Promise<void> {
    // Click delete button
    const deleteBtn = this.getFeedDeleteBtn(feedId);
    await deleteBtn.click();

    // Wait for delete confirmation modal
    await this.deleteConfirmation.waitFor({ state: "visible" });

    // Click confirm delete button
    const confirmBtn = this.getDeleteConfirmationBtn(feedId);
    await confirmBtn.click();

    // Wait for modal to close and list to refresh
    await this.deleteConfirmation.waitFor({ state: "hidden" });
  }

  /**
   * Check if feed exists in list (without searching)
   */
  async feedExists(feedName: string): Promise<boolean> {
    try {
      const feed = await this.getFeedByName(feedName);
      return feed !== null;
    } catch {
      return false;
    }
  }

  /**
   * Get all visible feeds count
   */
  async getFeedsCount(): Promise<number> {
    const feedItems = await this.page.locator('[data-testid^="feed-list-item-"]').count();
    return feedItems;
  }

  /**
   * Check if empty state is displayed
   */
  async isEmptyStateDisplayed(): Promise<boolean> {
    try {
      const emptyState = this.page.locator('text="You don\'t have any feeds yet"');
      await emptyState.waitFor({ state: "visible", timeout: 2000 });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get form-level error message
   */
  async getFormError(): Promise<string | null> {
    try {
      await this.feedFormError.waitFor({ state: "visible", timeout: 2000 });
      return await this.feedFormError.textContent();
    } catch {
      return null;
    }
  }

  /**
   * Get name field error message
   */
  async getNameFieldError(): Promise<string | null> {
    try {
      const error = await this.feedFormNameInput.evaluate((el: HTMLInputElement) => {
        return el.nextElementSibling?.textContent?.trim() || null;
      });
      return error;
    } catch {
      return null;
    }
  }

  /**
   * Get URL field error message
   */
  async getURLFieldError(): Promise<string | null> {
    try {
      const error = await this.feedFormUrlInput.evaluate((el: HTMLInputElement) => {
        return el.nextElementSibling?.textContent?.trim() || null;
      });
      return error;
    } catch {
      return null;
    }
  }

  /**
   * Check if name field has error
   */
  async hasNameFieldError(): Promise<boolean> {
    const error = await this.getNameFieldError();
    return error !== null && error.length > 0;
  }

  /**
   * Check if URL field has error
   */
  async hasURLFieldError(): Promise<boolean> {
    const error = await this.getURLFieldError();
    return error !== null && error.length > 0;
  }

  /**
   * Wait for name field error to appear
   */
  async waitForNameFieldError(timeoutMs = 5000): Promise<void> {
    const startTime = Date.now();
    while (Date.now() - startTime < timeoutMs) {
      const error = await this.getNameFieldError();
      if (error) {
        return;
      }
      await this.page.waitForTimeout(100);
    }
    throw new Error(`Name field error did not appear within ${timeoutMs}ms`);
  }

  /**
   * Wait for URL field error to appear
   */
  async waitForURLFieldError(timeoutMs = 5000): Promise<void> {
    const startTime = Date.now();
    while (Date.now() - startTime < timeoutMs) {
      const error = await this.getURLFieldError();
      if (error) {
        return;
      }
      await this.page.waitForTimeout(100);
    }
    throw new Error(`URL field error did not appear within ${timeoutMs}ms`);
  }

  /**
   * Close feed form modal
   */
  async closeFeedForm(): Promise<void> {
    await this.feedFormCancelBtn.click();
    await this.feedForm.waitFor({ state: "hidden" });
  }

  /**
   * Close delete confirmation modal
   */
  async closeDeleteConfirmation(): Promise<void> {
    await this.deleteConfirmationCancelBtn.click();
    await this.deleteConfirmation.waitFor({ state: "hidden" });
  }
}
