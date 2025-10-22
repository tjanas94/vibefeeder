import type { Page, Locator } from "@playwright/test";
import { BasePage } from "./base.page.ts";

/**
 * SummaryPage - Page Object for summary generation and display
 * Handles interactions with the summary modal, generation, and content display
 */
export class SummaryPage extends BasePage {
  // Summary button locators
  readonly summaryButton: Locator;
  readonly summaryButtonDisabled: Locator;

  // Summary modal locators
  readonly summaryModal: Locator;
  readonly summaryModalContent: Locator;
  readonly summaryModalTitle: Locator;

  // Summary content locators
  readonly summaryContent: Locator;
  readonly summaryTimestamp: Locator;
  readonly summaryEmptyState: Locator;
  readonly summaryErrorState: Locator;

  // Summary action buttons
  readonly summaryGenerateButton: Locator;

  constructor(page: Page) {
    super(page);

    // Summary button
    this.summaryButton = page.getByTestId("summary-button");
    this.summaryButtonDisabled = page.getByTestId("summary-button-disabled");

    // Modal
    this.summaryModal = page.locator("#summary-modal");
    this.summaryModalContent = page.getByTestId("summary-modal-content");
    this.summaryModalTitle = page.locator("#summary-modal-title");

    // Content
    this.summaryContent = page.getByTestId("summary-content");
    this.summaryTimestamp = page.getByTestId("summary-timestamp");
    this.summaryEmptyState = page.getByTestId("summary-empty-state");
    this.summaryErrorState = page.getByTestId("summary-error-state");

    // Actions
    this.summaryGenerateButton = page.getByTestId("summary-generate-button");
  }

  /**
   * Open summary modal by clicking the summary button
   */
  async openSummaryModal(): Promise<void> {
    await this.summaryButton.click();
    await this.summaryModalContent.waitFor({ state: "visible", timeout: 10000 });
  }

  /**
   * Check if summary button is disabled
   */
  async isSummaryButtonDisabled(): Promise<boolean> {
    try {
      // Check if disabled button element exists (it's in a template that shows when no feeds)
      const count = await this.summaryButtonDisabled.count();
      return count > 0;
    } catch {
      return false;
    }
  }

  /**
   * Check if summary button is enabled
   */
  async isSummaryButtonEnabled(): Promise<boolean> {
    try {
      await this.summaryButton.waitFor({ state: "visible", timeout: 2000 });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Generate summary by clicking generate button
   */
  async generateSummary(): Promise<void> {
    await this.summaryGenerateButton.click();
  }

  /**
   * Get summary content text
   */
  async getSummaryContent(): Promise<string | null> {
    try {
      await this.summaryContent.waitFor({ state: "visible", timeout: 30000 });
      const text = await this.summaryContent.textContent();
      // Extract just the summary text (skip timestamp)
      const lines = text?.split("\n") || [];
      // Skip the first line which contains timestamp
      return lines.slice(1).join("\n").trim() || null;
    } catch {
      return null;
    }
  }

  /**
   * Get summary timestamp
   */
  async getSummaryTimestamp(): Promise<string | null> {
    try {
      await this.summaryTimestamp.waitFor({ state: "visible", timeout: 5000 });
      return await this.summaryTimestamp.textContent();
    } catch {
      return null;
    }
  }

  /**
   * Check if empty state is displayed
   */
  async isEmptyStateDisplayed(): Promise<boolean> {
    try {
      await this.summaryEmptyState.waitFor({ state: "visible", timeout: 5000 });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Check if modal is open (checks for modal-open class via Alpine binding)
   */
  async isModalOpen(): Promise<boolean> {
    try {
      // Modal uses Alpine x-bind:class, so we need to check computed classes
      const isVisible = await this.summaryModal.evaluate((el) => {
        return !!(el as HTMLDialogElement).offsetParent;
      });
      return isVisible;
    } catch {
      return false;
    }
  }

  /**
   * Close summary modal by clicking close button
   * Waits for the modal to no longer be visible (modal-open class removed by Alpine)
   */
  async closeModal(): Promise<void> {
    const closeButton = this.page.locator("#summary-modal .btn-circle.btn-ghost");
    await closeButton.click();
    // Wait for modal to be closed (Alpine removes modal-open class)
    await this.page.waitForFunction(
      () => {
        const modal = document.querySelector("#summary-modal") as HTMLDialogElement;
        return modal && !modal.offsetParent; // offsetParent is null when display:none
      },
      { timeout: 5000 },
    );
  }

  /**
   * Wait for generate button to be visible and clickable
   */
  async waitForGenerateButton(timeoutMs = 10000): Promise<void> {
    await this.summaryGenerateButton.waitFor({ state: "visible", timeout: timeoutMs });
  }

  /**
   * Check if generate button is visible
   */
  async isGenerateButtonVisible(): Promise<boolean> {
    try {
      await this.summaryGenerateButton.waitFor({ state: "visible", timeout: 2000 });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get modal title text
   */
  async getModalTitle(): Promise<string | null> {
    try {
      await this.summaryModalTitle.waitFor({ state: "visible", timeout: 5000 });
      return await this.summaryModalTitle.textContent();
    } catch {
      return null;
    }
  }
}
