import { createClient } from "@supabase/supabase-js";

/**
 * Global teardown for Playwright E2E tests
 * Cleans up test data from Supabase after all tests complete
 *
 * Deletes data for the E2E test user:
 * - feeds (articles cascade delete automatically via FK constraint)
 * - summaries
 * - events
 */
async function globalTeardown() {
  const supabaseUrl = process.env.SUPABASE_URL;
  const supabaseKey = process.env.SUPABASE_KEY;
  const e2eUserId = process.env.E2E_USERNAME_ID;

  // Validate required environment variables
  if (!supabaseUrl || !supabaseKey || !e2eUserId) {
    console.warn("⚠️  Skipping database cleanup: Missing required environment variables");
    console.warn("   Required: SUPABASE_URL, SUPABASE_KEY, E2E_USERNAME_ID");
    return;
  }

  try {
    // Initialize Supabase client with admin key
    const supabase = createClient(supabaseUrl, supabaseKey);

    console.log("🧹 Cleaning up test data for E2E user...");

    // Delete feeds (articles cascade delete automatically via FK constraint)
    const { error: feedsError } = await supabase.from("feeds").delete().eq("user_id", e2eUserId);

    if (feedsError) {
      console.error("❌ Error deleting feeds:", feedsError);
      throw feedsError;
    }
    console.log("✓ Deleted feeds and cascaded articles");

    // Delete summaries
    const { error: summariesError } = await supabase
      .from("summaries")
      .delete()
      .eq("user_id", e2eUserId);

    if (summariesError) {
      console.error("❌ Error deleting summaries:", summariesError);
      throw summariesError;
    }
    console.log("✓ Deleted summaries");

    // Delete events
    const { error: eventsError } = await supabase.from("events").delete().eq("user_id", e2eUserId);

    if (eventsError) {
      console.error("❌ Error deleting events:", eventsError);
      throw eventsError;
    }
    console.log("✓ Deleted events");

    console.log("✅ Test data cleanup completed successfully");
  } catch (error) {
    console.error("❌ Database cleanup failed:", error);
    // Don't throw - allow tests to complete even if cleanup fails
    process.exit(1);
  }
}

export default globalTeardown;
