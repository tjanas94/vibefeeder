import { test, expect } from "./fixtures/auth.fixture.ts";

/**
 * Authentication E2E Tests
 * Tests login, registration, and authentication flows
 */

test.describe("Authentication", () => {
  test.beforeEach(async ({ authPage }) => {
    // Navigate to login page before each test
    await authPage.gotoLogin();
  });

  test("should display login page correctly", async ({ authPage, page }) => {
    // Verify page title
    await expect(page).toHaveTitle(/login/i);

    // Verify form elements are visible
    await expect(authPage.emailInput).toBeVisible();
    await expect(authPage.passwordInput).toBeVisible();
    await expect(authPage.loginButton).toBeVisible();
  });

  test("should show error for invalid login credentials", async ({ authPage }) => {
    // Attempt login with invalid credentials
    await authPage.login("invalid@example.com", "wrongpassword");

    // Verify error message is displayed
    await expect(authPage.errorMessage).toBeVisible();
    const errorText = await authPage.getErrorMessage();
    expect(errorText.toLowerCase()).toContain("invalid");
  });

  test("should show validation error for empty email", async ({ authPage }) => {
    // Try to submit with empty email
    await authPage.passwordInput.fill("somepassword");
    await authPage.loginButton.click();

    // Browser validation or server-side error should appear
    // Check if email input has validation state or error message appears
    const emailValidity = await authPage.emailInput.evaluate((el: HTMLInputElement) => {
      return el.validity.valid;
    });
    expect(emailValidity).toBe(false);
  });

  test("should show validation error for empty password", async ({ authPage }) => {
    // Try to submit with empty password
    await authPage.emailInput.fill("test@example.com");
    await authPage.loginButton.click();

    // Browser validation or server-side error should appear
    const passwordValidity = await authPage.passwordInput.evaluate((el: HTMLInputElement) => {
      return el.validity.valid;
    });
    expect(passwordValidity).toBe(false);
  });

  test("should navigate to register page", async ({ authPage, page }) => {
    // Check if register link exists
    const registerLinkExists = await authPage.registerLink.count();

    if (registerLinkExists > 0) {
      // Click register link
      await authPage.registerLink.click();

      // Verify navigation to register page
      await expect(page).toHaveURL(/register/);
    } else {
      test.skip();
    }
  });

  // This test uses credentials configured in .env.test file
  test("should successfully login with valid credentials", async ({ authPage, dashboardPage }) => {
    // Get E2E test user credentials from environment
    const testEmail = process.env.E2E_USERNAME;
    const testPassword = process.env.E2E_PASSWORD;

    // Validate credentials are configured
    if (!testEmail || !testPassword) {
      throw new Error(
        "E2E test credentials not configured. Please set E2E_USERNAME and E2E_PASSWORD in .env.test file",
      );
    }

    // Perform login
    await authPage.login(testEmail, testPassword);

    // Verify redirect to dashboard
    const isRedirected = await authPage.isRedirectedToDashboard();
    expect(isRedirected).toBe(true);

    // Verify dashboard is accessible
    const isLoggedIn = await dashboardPage.isLoggedIn();
    expect(isLoggedIn).toBe(true);
  });
});

test.describe("Dashboard Access", () => {
  test("should redirect to login when accessing dashboard without authentication", async ({
    page,
  }) => {
    // Try to access dashboard directly
    await page.goto("/dashboard");

    // Should redirect to login page
    await page.waitForURL("**/login", { timeout: 5000 });
    expect(page.url()).toContain("/login");
  });
});
