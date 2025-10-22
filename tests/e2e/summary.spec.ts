import { test, expect } from "./fixtures/summary.fixture.ts";
import { FEED_FIXTURES } from "./fixtures/feeds.fixture.ts";

/**
 * Summary E2E Tests - Minimal, cost-effective coverage
 * Only 1 summary generation API call per test run
 * Focuses on critical user journeys and blocking issues
 *
 * Pattern: Arrange → Act → Assert
 */

test.describe("Summary", () => {
  test.describe("Empty State", () => {
    test("should show disabled button when no feeds exist", async ({ summaryPage }) => {
      // Arrange
      await summaryPage.goto("/dashboard");

      // Act & Assert - Button is disabled
      const isDisabled = await summaryPage.isSummaryButtonDisabled();
      expect(isDisabled).toBe(true);

      // Assert - Button should have disabled attribute
      const disabledBtn = summaryPage.summaryButtonDisabled;
      const isActuallyDisabled = await disabledBtn.evaluate((el: HTMLButtonElement) => {
        return el.disabled === true;
      });
      expect(isActuallyDisabled).toBe(true);
    });
  });

  test.describe("Generate Summary", () => {
    test("should generate summary and display content with timestamp", async ({
      feedPage,
      summaryPage,
    }) => {
      // Arrange - Add a feed and wait for it to be ready
      test.setTimeout(60000); // Extend timeout for feed addition and summary generation
      await feedPage.gotoFeeds();
      const feedName = "E2E Test - Generate Summary";
      const feedUrl = feedPage.generateFeedUrlFromName(feedName, FEED_FIXTURES.basic);

      await feedPage.searchFeeds(feedName);
      await feedPage.addFeed(feedName, feedUrl);

      const feed = await feedPage.getFeedByName(feedName);
      expect(feed).not.toBeNull();
      await feedPage.waitForFeedStatus(feed!.id, "OK");

      // Act - Open summary modal
      await summaryPage.openSummaryModal();

      // Assert - Empty state is displayed with generate button
      const isEmpty = await summaryPage.isEmptyStateDisplayed();
      expect(isEmpty).toBe(true);

      const hasGenerateBtn = await summaryPage.isGenerateButtonVisible();
      expect(hasGenerateBtn).toBe(true);

      // Act - Generate summary
      await summaryPage.generateSummary();

      // Assert - Summary content is displayed
      const content = await summaryPage.getSummaryContent();
      expect(content).not.toBeNull();
      expect(content?.length || 0).toBeGreaterThan(0);

      // Assert - Timestamp is present
      const timestamp = await summaryPage.getSummaryTimestamp();
      expect(timestamp).not.toBeNull();
      expect(timestamp?.length || 0).toBeGreaterThan(0);

      // Assert - Modal can be closed after generation
      await summaryPage.closeModal();
      const isStillOpen = await summaryPage.isModalOpen();
      expect(isStillOpen).toBe(false);
    });
  });
});
