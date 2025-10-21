/**
 * Test fixtures for feed fetching tests
 * Provides URLs to real test RSS/Atom feeds hosted on GitHub
 * Supports optional GitHub authentication to avoid rate limiting
 */

const GITHUB_TOKEN = process.env.GITHUB_ACCESS_TOKEN;
const BASE_GITHUB_URL = GITHUB_TOKEN
  ? `https://${GITHUB_TOKEN}@raw.githubusercontent.com`
  : "https://raw.githubusercontent.com";

export const FEED_FIXTURES = {
  /**
   * Basic RSS 2.0 feed with 3 simple articles
   * Tests standard feed parsing without publication dates
   */
  basic: `${BASE_GITHUB_URL}/tjanas94/vibefeeder/master/tests/e2e/fixtures/test-feed-basic.xml`,

  /**
   * Atom 1.0 feed with 2 entries
   * Tests format compatibility without date elements
   */
  atom: `${BASE_GITHUB_URL}/tjanas94/vibefeeder/master/tests/e2e/fixtures/test-feed-atom.xml`,

  /**
   * RSS 2.0 feed with edge cases
   * Tests robustness: minimal items, special characters, CDATA, long titles
   */
  edgeCases: `${BASE_GITHUB_URL}/tjanas94/vibefeeder/master/tests/e2e/fixtures/test-feed-edge-cases.xml`,
};
