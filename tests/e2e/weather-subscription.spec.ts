import { test, expect } from '@playwright/test';

test.describe('Weather Subscription Page', () => {
  test('should display subscription form', async ({ page }) => {
    await page.goto('/');
    
    await expect(page.locator('h1')).toHaveText('Weather Subscription');
    await expect(page.locator('#email')).toBeVisible();
    await expect(page.locator('#city')).toBeVisible();
    await expect(page.locator('#frequency')).toBeVisible();
    await expect(page.locator('#subscribeBtn')).toBeVisible();
  });

  test('should show validation errors for empty form', async ({ page }) => {
    await page.goto('/');
    
    await page.click('#subscribeBtn');
    
    await expect(page.locator('#emailError')).toBeVisible();
    await expect(page.locator('#cityError')).toBeVisible();
  });

  test('should submit form successfully', async ({ page }) => {
    await page.goto('/');
    
    await page.fill('#email', 'test@example.com');
    await page.fill('#city', 'Kyiv');
    await page.selectOption('#frequency', 'daily');

    const firstResponsePromise = page.waitForResponse(response =>
        response.url().includes('/api/subscribe') && response.status() === 200
    );
    await page.click('#subscribeBtn');

    const firstResponse = await firstResponsePromise;
    const firstData = await firstResponse.json();
    expect(firstData.message).toContain('Subscription successful. Confirmation email sent.');
  });

  test('should handle duplicate subscription', async ({ page }) => {
    await page.goto('/');
    
    const testEmail = `test-${Date.now()}@example.com`;
    
    await page.fill('#email', testEmail);
    await page.fill('#city', 'Kyiv');
    await page.selectOption('#frequency', 'daily');
    
    const firstResponsePromise = page.waitForResponse(response =>
      response.url().includes('/api/subscribe') && response.status() === 200
    );
    await page.click('#subscribeBtn');
    
    const firstResponse = await firstResponsePromise;
    const firstData = await firstResponse.json();
    expect(firstData.message).toContain('Subscription successful. Confirmation email sent.');

    try {
      const alertPromise = page.waitForEvent('dialog', { timeout: 5000 });
      const dialog = await alertPromise;
      await dialog.accept();
    } catch (error) {
      console.log('First subscription completed');
    }

    await page.goto('/');
    await page.fill('#email', testEmail);
    await page.fill('#city', 'Kyiv');
    await page.selectOption('#frequency', 'daily');
    
    const secondResponsePromise = page.waitForResponse(response =>
      response.url().includes('/api/subscribe')
    );
    await page.click('#subscribeBtn');
    
    const secondResponse = await secondResponsePromise;
    
    const secondData = await secondResponse.json();
    expect(secondData).toHaveProperty('error');
    expect(secondData.error).toContain('email already subscribed');
  });
}); 