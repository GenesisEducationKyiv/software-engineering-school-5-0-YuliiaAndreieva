import { test, expect } from '@playwright/test';
import { 
  clearMailHogMessages, 
  waitForEmail, 
  getMailHogMessages,
  extractConfirmationLink,
  debugEmailContent,
} from './utils/mailhog';

test.describe('Email Confirmation Flow', () => {
  test.beforeEach(async () => {
    await clearMailHogMessages();
  });

  test('should send confirmation email and allow confirmation via link', async ({ page }) => {
    await page.goto('/');
    
    const testEmail = `confirm-${Date.now()}@example.com`;
    await page.fill('#email', testEmail);
    await page.fill('#city', 'Kharkiv');
    await page.selectOption('#frequency', 'daily');

    const subscribeResponse = page.waitForResponse(response =>
        response.url().includes('/api/subscribe') && response.status() === 200
    );
    await page.click('#subscribeBtn');
    await subscribeResponse;

    await page.waitForTimeout(3000);

    const emailMessage = await waitForEmail(testEmail, 20000);
    expect(emailMessage.Content.Body).toContain('Please click the link below to confirm your subscription');
    
    debugEmailContent(emailMessage);
    
    const confirmationLink = extractConfirmationLink(emailMessage.Content.Body);
    console.log('Extracted confirmation link:', confirmationLink);
    expect(confirmationLink).not.toBeNull();
    
    await page.goto(confirmationLink!);
    
    await expect(page.locator('body')).toContainText('Subscription confirmed');
  });

  test('should handle invalid confirmation token', async ({ page }) => {
    await page.goto('/api/confirm/invalid-token-123');
    
    await expect(page.locator('body')).toContainText('error');
  });

  test('should verify email template content', async ({ page }) => {
    await page.goto('/');
    
    const testEmail = `template-${Date.now()}@example.com`;
    await page.fill('#email', testEmail);
    await page.fill('#city', 'Odesa');
    await page.selectOption('#frequency', 'hourly');

    const response = page.waitForResponse(response =>
        response.url().includes('/api/subscribe') && response.status() === 200
    );
    await page.click('#subscribeBtn');
    await response;

    await page.waitForTimeout(2000);

    const emailMessage = await waitForEmail(testEmail, 15000);
    
    debugEmailContent(emailMessage);
    
    expect(emailMessage.From.Mailbox + '@' + emailMessage.From.Domain).toBe('test@example.com');
    expect(emailMessage.To[0].Mailbox + '@' + emailMessage.To[0].Domain).toBe(testEmail);
    
    expect(emailMessage.Content.Headers['Subject']).toBeDefined();
    expect(emailMessage.Content.Headers['Content-Type'][0]).toContain('text/html');
    
    const emailBody = emailMessage.Content.Body;
    expect(emailBody).toContain('Odesa');
    expect(emailBody).toContain('confirm your subscription');
    expect(emailBody).toContain('href=');
  });

  test('should handle multiple subscriptions and confirmations', async ({ page }) => {
    const emails = [
      `multi1-${Date.now()}@example.com`,
      `multi2-${Date.now()}@example.com`,
      `multi3-${Date.now()}@example.com`
    ];

    for (const email of emails) {
      await page.goto('/');
      await page.fill('#email', email);
      await page.fill('#city', 'Dnipro');
      await page.selectOption('#frequency', 'daily');

      const response = page.waitForResponse(response =>
          response.url().includes('/api/subscribe') && response.status() === 200
      );
      await page.click('#subscribeBtn');
      await response;
      
      await page.waitForTimeout(1000);
    }

    await page.waitForTimeout(3000);

    const messages = await getMailHogMessages();
    expect(messages.length).toBe(3);

    for (const email of emails) {
      const emailMessage = messages.find((msg) => 
        `${msg.To[0]?.Mailbox}@${msg.To[0]?.Domain}` === email
      );
      expect(emailMessage).toBeDefined();
      
      const confirmationLink = extractConfirmationLink(emailMessage.Content.Body);
      console.log(`Confirmation link for ${email}:`, confirmationLink);
      expect(confirmationLink).not.toBeNull();
      
      await page.goto(confirmationLink!);
      await expect(page.locator('body')).toContainText('Subscription confirmed');
    }
  });
}); 