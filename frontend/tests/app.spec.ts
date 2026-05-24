import { test, expect } from '@playwright/test';

test.describe('Login Page', () => {
  test('should display login form', async ({ page }) => {
    await page.goto('/login');

    await expect(page.locator('text=go-diamond')).toBeVisible();
    await expect(page.locator('text=Configuration Center')).toBeVisible();
    await expect(page.locator('input[type="email"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('should show demo credentials', async ({ page }) => {
    await page.goto('/login');

    await expect(page.locator('text=admin@example.com')).toBeVisible();
    await expect(page.locator('text=user@example.com')).toBeVisible();
  });

  test('should login successfully with valid credentials', async ({ page }) => {
    await page.goto('/login');

    await page.fill('input[type="email"]', 'admin@example.com');
    await page.fill('input[type="password"]', 'admin123');
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('text=go-diamond')).toBeVisible();
  });

  test('should show error with invalid credentials', async ({ page }) => {
    await page.goto('/login');

    await page.fill('input[type="email"]', 'admin@example.com');
    await page.fill('input[type="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');

    await expect(page.locator('text=Invalid email or password')).toBeVisible();
  });

  test('should have persistent login in localStorage', async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('input[type="email"]', 'admin@example.com');
    await page.fill('input[type="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/dashboard');

    // Verify token exists in localStorage
    const token = await page.evaluate(() => localStorage.getItem('token'));
    expect(token).toBeTruthy();

    // Verify user exists in localStorage
    const userStr = await page.evaluate(() => localStorage.getItem('user'));
    expect(userStr).toBeTruthy();
  });
});

test.describe('Dashboard Page', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.fill('input[type="email"]', 'admin@example.com');
    await page.fill('input[type="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/dashboard');
  });

  test('should display dashboard with header', async ({ page }) => {
    await expect(page.locator('h1:has-text("go-diamond")')).toBeVisible();
    await expect(page.locator('text=admin@example.com')).toBeVisible();
    await expect(page.locator('button:has-text("Logout")')).toBeVisible();
  });

  test('should display search filters', async ({ page }) => {
    await expect(page.locator('label:has-text("Namespace")')).toBeVisible();
    await expect(page.locator('label:has-text("Group")')).toBeVisible();
    await expect(page.locator('button:has-text("Search")')).toBeVisible();
    await expect(page.locator('button:has-text("New Config")')).toBeVisible();
  });

  test('should open create modal when clicking New Config', async ({ page }) => {
    await page.click('button:has-text("New Config")');

    await expect(page.locator('h2:has-text("Create Config")')).toBeVisible();
    await expect(page.locator('input[placeholder="app.json"]')).toBeVisible();
  });

  test('should close create modal on cancel', async ({ page }) => {
    await page.click('button:has-text("New Config")');
    await expect(page.locator('h2:has-text("Create Config")')).toBeVisible();

    await page.click('button:has-text("Cancel")');
    await expect(page.locator('h2:has-text("Create Config")')).not.toBeVisible();
  });

  test('should logout and redirect to login', async ({ page }) => {
    await page.click('button:has-text("Logout")');

    await expect(page).toHaveURL('/login');
    await expect(page.locator('input[type="email"]')).toBeVisible();
  });
});

test.describe('Navigation', () => {
  test('should redirect to login when not authenticated', async ({ page }) => {
    await page.goto('/dashboard');

    await expect(page).toHaveURL('/login');
  });

  test('should redirect root to dashboard when authenticated', async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('input[type="email"]', 'admin@example.com');
    await page.fill('input[type="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/dashboard');

    // Navigate to root
    await page.goto('/');
    await expect(page).toHaveURL('/dashboard');
  });
});