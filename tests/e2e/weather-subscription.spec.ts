import { test, expect } from '@playwright/test';
import { 
  clearMailHogMessages, 
  waitForEmail, 
} from './utils/mailhog';

test.describe('Weather Subscription Page', () => {
  test.beforeEach(async () => {
    await clearMailHogMessages();
  });

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

  test('should submit form successfully and send confirmation email', async ({ page }) => {
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

    const emailMessage = await waitForEmail(testEmail);
    expect(emailMessage.From.Mailbox + '@' + emailMessage.From.Domain).toBe('test@example.com');
    expect(emailMessage.To[0].Mailbox + '@' + emailMessage.To[0].Domain).toBe(testEmail);
    expect(emailMessage.Content.Body).toContain('Please click the link below to confirm your subscription');
    expect(emailMessage.Content.Body).toContain('Kyiv');
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

    await waitForEmail(testEmail);

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

  test('should verify email content and headers', async ({ page }) => {
    await page.goto('/');
    
    const testEmail = `test-${Date.now()}@example.com`;
    await page.fill('#email', testEmail);
    await page.fill('#city', 'Lviv');
    await page.selectOption('#frequency', 'hourly');

    const responsePromise = page.waitForResponse(response =>
        response.url().includes('/api/subscribe') && response.status() === 200
    );
    await page.click('#subscribeBtn');

    await responsePromise;

    const emailMessage = await waitForEmail(testEmail);
    
    expect(emailMessage.Content.Headers['Subject']).toBeDefined();
    expect(emailMessage.Content.Headers['From']).toBeDefined();
    expect(emailMessage.Content.Headers['To']).toBeDefined();
    expect(emailMessage.Content.Headers['Content-Type'][0]).toContain('text/html');
    
    expect(emailMessage.Content.Body).toContain('Lviv');
    expect(emailMessage.Content.Body).toContain('confirm your subscription');
  });
}); 