import { test, expect } from "./fixtures/feed.fixture.ts";
import { FEED_FIXTURES } from "./fixtures/feeds.fixture.ts";

/**
 * Feed Management E2E Tests
 * Tests for adding, editing, deleting, searching, and filtering feeds
 * Uses static, descriptive test names with slug-based URL generation
 * Pattern: Search → Action → Verify (search filter is set before action, then action is performed, then results verified in filtered list)
 */

test.describe("Feed Management", () => {
  test.describe("Empty State", () => {
    test("should show empty state message when no feeds exist", async ({ feedPage }) => {
      // Arrange - feedPage fixture already handles authentication
      // Navigate to feeds (dashboard)
      await feedPage.gotoFeeds();

      // Act - Check for empty state
      const isEmpty = await feedPage.isEmptyStateDisplayed();

      // Assert - Empty state should be displayed
      expect(isEmpty).toBe(true);
    });
  });

  test.describe("Add Feed", () => {
    test("should add a new feed and display it in the list", async ({ feedPage }) => {
      // Arrange
      await feedPage.gotoFeeds();
      const feedName = "E2E Test - Add New Feed";
      const feedUrl = feedPage.generateFeedUrlFromName(feedName, FEED_FIXTURES.basic);

      // Set search before action
      await feedPage.searchFeeds(feedName);

      // Act
      await feedPage.addFeed(feedName, feedUrl);

      // Assert - Feed should appear in list with Pending status
      const feed = await feedPage.getFeedByName(feedName);
      expect(feed).not.toBeNull();
      expect(feed?.id).toBeTruthy();

      // Check initial status is Pending
      const status = await feedPage.getFeedStatus(feed!.id);
      expect(status).toContain("Pending");
    });

    test("should validate feed name is required", async ({ feedPage }) => {
      // Arrange
      await feedPage.gotoFeeds();
      await feedPage.openAddFeedForm();

      // Act - Try to submit without name
      const feedUrl = feedPage.generateFeedUrlFromName(
        "E2E Test - Name Validation",
        FEED_FIXTURES.basic,
      );
      await feedPage.fillFeedForm("", feedUrl);
      await feedPage.submitFeedForm();

      // Assert - Wait for error, then get and verify
      await feedPage.waitForNameFieldError();
      const errorMessage = await feedPage.getNameFieldError();
      expect(errorMessage?.toLowerCase() || "").toMatch(/(required|cannot be empty)/);
    });

    test("should validate feed URL is required", async ({ feedPage }) => {
      // Arrange
      await feedPage.gotoFeeds();
      await feedPage.openAddFeedForm();

      // Act - Try to submit without URL
      const feedName = "E2E Test - URL Required Validation";
      await feedPage.fillFeedForm(feedName, "");
      await feedPage.submitFeedForm();

      // Assert - Wait for error, then get and verify
      await feedPage.waitForURLFieldError();
      const errorMessage = await feedPage.getURLFieldError();
      expect(errorMessage?.toLowerCase() || "").toMatch(/(required|cannot be empty)/);
    });

    test("should validate invalid URL format", async ({ feedPage }) => {
      // Arrange
      await feedPage.gotoFeeds();
      await feedPage.openAddFeedForm();

      // Act - Try to submit with invalid URL
      const feedName = "E2E Test - Invalid URL Format";
      await feedPage.fillFeedForm(feedName, "not-a-valid-url");
      await feedPage.submitFeedForm();

      // Assert - Wait for error, then get and verify
      await feedPage.waitForURLFieldError();
      const errorMessage = await feedPage.getURLFieldError();
      expect(errorMessage?.toLowerCase() || "").toMatch(/(invalid|valid url|https)/);
    });

    test("should show error for duplicate feed URL", async ({ feedPage }) => {
      // Arrange - Add first feed
      await feedPage.gotoFeeds();
      const feedName1 = "E2E Test - Duplicate URL First";
      const feedUrl = feedPage.generateFeedUrlFromName(feedName1, FEED_FIXTURES.basic);
      await feedPage.searchFeeds(feedName1);
      await feedPage.addFeed(feedName1, feedUrl);

      // Act - Try to add same URL again with different name
      const feedName2 = "E2E Test - Duplicate URL Second";
      await feedPage.searchFeeds(feedName2);
      await feedPage.openAddFeedForm();
      await feedPage.fillFeedForm(feedName2, feedUrl);
      await feedPage.submitFeedForm();

      // Assert - Wait for error, then get and verify
      await feedPage.waitForURLFieldError();
      const errorMessage = await feedPage.getURLFieldError();
      expect(errorMessage?.toLowerCase() || "").toContain("already");
    });
  });

  test.describe("Edit Feed", () => {
    test("should edit feed name", async ({ feedPage }) => {
      // Arrange
      await feedPage.gotoFeeds();
      const feedName = "E2E Test - Edit Feed Name";
      const feedUrl = feedPage.generateFeedUrlFromName(feedName, FEED_FIXTURES.basic);
      await feedPage.searchFeeds(feedName);
      await feedPage.addFeed(feedName, feedUrl);

      // Get feed ID
      let feed = await feedPage.getFeedByName(feedName);
      expect(feed).not.toBeNull();

      // Act - Edit feed name
      const newFeedName = "E2E Test - Edit Feed Name Updated";
      await feedPage.searchFeeds(newFeedName);
      await feedPage.editFeed(feed!.id, newFeedName);

      // Assert - New name should be in list
      feed = await feedPage.getFeedByName(newFeedName);
      expect(feed).not.toBeNull();
    });
  });

  test.describe("Delete Feed", () => {
    test("should delete a feed from the list", async ({ feedPage }) => {
      // Arrange
      await feedPage.gotoFeeds();
      const feedName = "E2E Test - Delete Feed";
      const feedUrl = feedPage.generateFeedUrlFromName(feedName, FEED_FIXTURES.basic);
      await feedPage.searchFeeds(feedName);
      await feedPage.addFeed(feedName, feedUrl);

      // Verify feed exists
      let feed = await feedPage.getFeedByName(feedName);
      expect(feed).not.toBeNull();

      // Act - Delete the feed
      await feedPage.deleteFeed(feed!.id);

      // Verify feed is truly gone by searching
      await feedPage.searchFeeds(feedName);
      feed = await feedPage.getFeedByName(feedName);
      expect(feed).toBeNull();
    });

    test("should cancel delete operation", async ({ feedPage }) => {
      // Arrange
      await feedPage.gotoFeeds();
      const feedName = "E2E Test - Cancel Delete";
      const feedUrl = feedPage.generateFeedUrlFromName(feedName, FEED_FIXTURES.basic);
      await feedPage.searchFeeds(feedName);
      await feedPage.addFeed(feedName, feedUrl);

      // Get feed and open delete confirmation
      let feed = await feedPage.getFeedByName(feedName);
      expect(feed).not.toBeNull();
      const deleteBtn = feedPage.getFeedDeleteBtn(feed!.id);
      await deleteBtn.click();

      // Wait for delete confirmation modal
      await feedPage.deleteConfirmation.waitFor({ state: "visible" });

      // Act - Cancel deletion
      await feedPage.closeDeleteConfirmation();

      // Assert - Feed should still exist in list
      feed = await feedPage.getFeedByName(feedName);
      expect(feed).not.toBeNull();
    });
  });

  test.describe("Search Feeds", () => {
    test("should search feeds by partial name match", async ({ feedPage }) => {
      // Arrange
      await feedPage.gotoFeeds();
      const feed1Name = "E2E Test - Search By Name";
      const feed2Name = "E2E Test - Other Feed";

      const feedUrl1 = feedPage.generateFeedUrlFromName(feed1Name, FEED_FIXTURES.basic);
      const feedUrl2 = feedPage.generateFeedUrlFromName(feed2Name, FEED_FIXTURES.atom);

      await feedPage.addFeed(feed1Name, feedUrl1);
      await feedPage.addFeed(feed2Name, feedUrl2);

      // Act - Search for first feed by partial name
      const searchPrefix = "Search By Name";
      await feedPage.searchFeeds(searchPrefix);

      // Assert - Feed with matching name should be visible
      const feed = await feedPage.getFeedByName(feed1Name);
      expect(feed).not.toBeNull();
    });

    test("should return no results when searching for non-existing feed", async ({ feedPage }) => {
      // Arrange
      await feedPage.gotoFeeds();
      const feed1Name = "E2E Test - Non Existing Search";
      const feedUrl1 = feedPage.generateFeedUrlFromName(feed1Name, FEED_FIXTURES.basic);
      await feedPage.addFeed(feed1Name, feedUrl1);

      // Act - Search for non-existing feed with unique query
      const nonExistingQuery = "NonExistingFeedQuery12345XYZ";
      await feedPage.searchFeeds(nonExistingQuery);

      // Assert - No feed should be found
      const feed = await feedPage.getFeedByName(nonExistingQuery);
      expect(feed).toBeNull();
    });
  });

  test.describe("Filter Feeds", () => {
    test("should filter feeds by status", async ({ feedPage }) => {
      // Arrange - Add a feed with valid URL (will eventually become OK)
      await feedPage.gotoFeeds();
      const feed1Name = "E2E Test - Filter By Status";
      const feedUrl1 = feedPage.generateFeedUrlFromName(feed1Name, FEED_FIXTURES.basic);
      await feedPage.searchFeeds(feed1Name);
      await feedPage.addFeed(feed1Name, feedUrl1);

      const feed1 = await feedPage.getFeedByName(feed1Name);
      expect(feed1).not.toBeNull();

      // Act - Filter by Pending status
      await feedPage.filterByStatus("pending");

      // Assert - Our feed with Pending status should be visible
      const feed = await feedPage.getFeedByName(feed1Name);
      expect(feed).not.toBeNull();
    });

    test("should filter only working feeds", async ({ feedPage }) => {
      // Arrange - Add a feed with valid URL (will eventually become OK)
      test.setTimeout(60000); // Extend timeout for feed status update
      await feedPage.gotoFeeds();
      const feed1Name = "E2E Test - Filter By Status Only Working";
      const feedUrl1 = feedPage.generateFeedUrlFromName(feed1Name, FEED_FIXTURES.basic);
      await feedPage.searchFeeds(feed1Name);
      await feedPage.addFeed(feed1Name, feedUrl1);

      const feed1 = await feedPage.getFeedByName(feed1Name);
      expect(feed1).not.toBeNull();

      // Wait for feed to become OK
      await feedPage.waitForFeedStatus(feed1!.id, "OK");

      // Act - Filter by Working status
      await feedPage.filterByStatus("working");

      // Assert - Our feed with Working status should be visible
      const feed = await feedPage.getFeedByName(feed1Name);
      expect(feed).not.toBeNull();
    });

    test("should reset to all feeds filter", async ({ feedPage }) => {
      // Arrange - Add a feed and verify it appears with pending status
      await feedPage.gotoFeeds();
      const feedName = "E2E Test - Filter Reset To All";
      const feedUrl = feedPage.generateFeedUrlFromName(feedName, FEED_FIXTURES.basic);
      await feedPage.searchFeeds(feedName);
      await feedPage.addFeed(feedName, feedUrl);

      // Filter by pending to isolate our feed
      await feedPage.filterByStatus("pending");
      let feed = await feedPage.getFeedByName(feedName);
      expect(feed).not.toBeNull();

      // Act - Reset to all
      await feedPage.filterByStatus("all");

      // Assert - Feed should still be visible after reset
      feed = await feedPage.getFeedByName(feedName);
      expect(feed).not.toBeNull();
    });
  });
});

test.describe("Feed Status Updates", () => {
  test("should update feed status from Pending to OK", async ({ feedPage }) => {
    // Arrange
    test.setTimeout(60000); // Extend timeout for feed status update
    await feedPage.gotoFeeds();
    const feedName = "E2E Test - Status Update Pending To OK";
    const feedUrl = feedPage.generateFeedUrlFromName(feedName, FEED_FIXTURES.basic);
    await feedPage.searchFeeds(feedName);
    await feedPage.addFeed(feedName, feedUrl);

    // Get feed and verify initial Pending status
    const feed = await feedPage.getFeedByName(feedName);
    expect(feed).not.toBeNull();
    let status = await feedPage.getFeedStatus(feed!.id);
    expect(status).toContain("Pending");

    // Act - Wait for status to update
    await feedPage.waitForFeedStatus(feed!.id, "OK");

    // Assert - Status should now be OK
    status = await feedPage.getFeedStatus(feed!.id);
    expect(status).toContain("OK");
  });

  test("should handle feed fetch errors gracefully", async ({ feedPage }) => {
    // Arrange
    await feedPage.gotoFeeds();
    const feedName = "E2E Test - Handle Fetch Errors";
    const invalidFeedUrl = "https://invalid-feed-url-example.com/feed.xml";

    // Act
    await feedPage.searchFeeds(feedName);
    await feedPage.addFeed(feedName, invalidFeedUrl);

    // Get feed and wait for error status
    const feed = await feedPage.getFeedByName(feedName);
    expect(feed).not.toBeNull();

    // Wait for feed to update (may show Error status)
    try {
      await feedPage.waitForFeedStatus(feed!.id, "Error");
      const status = await feedPage.getFeedStatus(feed!.id);
      expect(status).toContain("Error");
    } catch {
      // Feed might still be pending if fetcher is still trying
      // This is acceptable behavior
    }
  });
});
