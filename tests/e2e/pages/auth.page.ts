import { Page, Locator } from "@playwright/test";
import { BasePage } from "./base.page";

/**
 * AuthPage - Page Object for authentication pages
 * Handles login, registration, and authentication-related interactions
 */
export class AuthPage extends BasePage {
  // Locators for login page
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly loginButton: Locator;
  readonly registerLink: Locator;
  readonly errorMessage: Locator;
  readonly successMessage: Locator;

  constructor(page: Page) {
    super(page);

    // Initialize locators
    this.emailInput = page.locator('input[name="email"]');
    this.passwordInput = page.locator('input[name="password"]');
    this.loginButton = page.locator('button[type="submit"]');
    this.registerLink = page.locator('a[href*="register"]');
    this.errorMessage = page.locator(".alert-error");
    this.successMessage = page.locator(".alert-success");
  }

  /**
   * Navigate to login page
   */
  async gotoLogin(): Promise<void> {
    await this.goto("/login");
  }

  /**
   * Navigate to register page
   */
  async gotoRegister(): Promise<void> {
    await this.goto("/register");
  }

  /**
   * Perform login action
   */
  async login(email: string, password: string): Promise<void> {
    await this.emailInput.fill(email);
    await this.passwordInput.fill(password);
    await this.loginButton.click();
  }

  /**
   * Check if error message is displayed
   */
  async hasErrorMessage(): Promise<boolean> {
    return await this.errorMessage.isVisible();
  }

  /**
   * Get error message text
   */
  async getErrorMessage(): Promise<string> {
    await this.waitForElement(this.errorMessage);
    return (await this.errorMessage.textContent()) || "";
  }

  /**
   * Check if success message is displayed
   */
  async hasSuccessMessage(): Promise<boolean> {
    return await this.successMessage.isVisible();
  }

  /**
   * Check if user is redirected after successful login
   */
  async isRedirectedToDashboard(): Promise<boolean> {
    await this.page.waitForURL("**/dashboard", { timeout: 5000 });
    return this.page.url().includes("/dashboard");
  }
}
