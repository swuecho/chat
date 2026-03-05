import { test, expect } from '@playwright/test';
import { randomEmail } from '../lib/sample';
import { MessageHelpers, InputHelpers } from '../lib/message-helpers';

const test_email = randomEmail();

test('session answer regenerate - robust version', async ({ page }) => {
  // Initialize helpers
  const messageHelpers = new MessageHelpers(page);
  const inputHelpers = new InputHelpers(page);

  // Setup
  await page.goto('/');
  await page.getByTitle('signuptab').click();
  await page.getByTestId('signup_email').click();
  await page.getByTestId('signup_email').locator('input').fill(test_email);
  await page.getByTestId('signup_password').locator('input').click();
  await page.getByTestId('signup_password').locator('input').fill('@ThisIsATestPass5');
  await page.getByTestId('repwd').locator('input').click();
  await page.getByTestId('repwd').locator('input').fill('@ThisIsATestPass5');
  await page.getByTestId('signup').click();
  
  // Wait for authentication to complete
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(3000);
  
  // Wait for the permission modal to disappear
  try {
    await page.waitForSelector('.n-modal-mask', { state: 'detached', timeout: 10000 });
  } catch (error) {
    // Modal might already be gone
  }
  
  await page.waitForTimeout(1000);

  await page.locator('a').filter({ hasText: 'New Chat' }).click();

  // Set debug mode
  await page.getByTestId('chat-settings-button').click();
  // expand the Advanced Settings section (accordion)
  await page.getByTestId('collapse-advanced').click();
  // wait for the section to expand
  await page.waitForTimeout(300);
  await page.getByTestId('debug_mode').click();
  await page.click('body', { position: { x: 0, y: 0 } });

  // Send first message
  await inputHelpers.sendMessage('test_demo_bestqa');
  
  // Wait for and verify first response
  await messageHelpers.waitForAssistantMessageCount(1);

  // Send second message
  await inputHelpers.sendMessage('test_debug_1');
  
  // Wait for and verify second response
  await messageHelpers.waitForAssistantMessageCount(2);

  // Test regenerate functionality on second response
  const isRegenerateVisible = await messageHelpers.isAssistantRegenerateButtonVisible(1);
  expect(isRegenerateVisible).toBe(true);
  
  await messageHelpers.clickAssistantRegenerate(1);
  await page.waitForTimeout(300);
  await messageHelpers.waitForAssistantMessageCount(2);
  expect(await messageHelpers.isAssistantRegenerateButtonVisible(1)).toBe(true);

  // Send third message
  await inputHelpers.sendMessage('test_debug_2');
  
  // Wait for and verify third response
  await messageHelpers.waitForAssistantMessageCount(3);

  // Test regenerate on third response
  await messageHelpers.clickAssistantRegenerate(2);
  await page.waitForTimeout(300);
  await messageHelpers.waitForAssistantMessageCount(3);
  expect(await messageHelpers.isAssistantRegenerateButtonVisible(2)).toBe(true);

  // Regenerate the second answer again
  await messageHelpers.clickAssistantRegenerate(1);
  await page.waitForTimeout(300);
  await messageHelpers.waitForAssistantMessageCount(3);
  expect(await messageHelpers.isAssistantRegenerateButtonVisible(1)).toBe(true);
  expect(await messageHelpers.isAssistantRegenerateButtonVisible(2)).toBe(true);
});
