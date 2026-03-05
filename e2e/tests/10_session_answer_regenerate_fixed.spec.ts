import { test, expect } from '@playwright/test';
import { Pool } from 'pg';
import { randomEmail } from '../lib/sample';
import { db_config } from '../lib/db/config';
import { MessageHelpers, InputHelpers } from '../lib/message-helpers';

const test_email = randomEmail();
const pool = new Pool(db_config);

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
  await messageHelpers.waitForAssistantMessageTextContains(0, 'test_demo_bestqa');
  const firstAnswer = await messageHelpers.getAssistantMessageText(0);
  expect(firstAnswer).toContain('test_demo_bestqa');

  // Send second message
  await inputHelpers.sendMessage('test_debug_1');
  
  // Wait for and verify second response
  await messageHelpers.waitForAssistantMessageCount(2);
  await messageHelpers.waitForAssistantMessageTextContains(1, 'test_debug_1');
  const secondAnswer = await messageHelpers.getAssistantMessageText(1);
  expect(secondAnswer).toContain('test_debug_1');

  // Test regenerate functionality on second response
  const isRegenerateVisible = await messageHelpers.isAssistantRegenerateButtonVisible(1);
  expect(isRegenerateVisible).toBe(true);
  
  await messageHelpers.clickAssistantRegenerate(1);
  await page.waitForTimeout(300);
  await messageHelpers.waitForAssistantMessageTextContains(1, 'test_debug_1');

  // Verify regenerated response still contains the expected text
  const secondAnswerRegen = await messageHelpers.getAssistantMessageText(1);
  expect(secondAnswerRegen).toContain('test_debug_1');

  // Send third message
  await inputHelpers.sendMessage('test_debug_2');
  
  // Wait for and verify third response
  await messageHelpers.waitForAssistantMessageCount(3);
  await messageHelpers.waitForAssistantMessageTextContains(2, 'test_debug_2');
  const thirdAnswer = await messageHelpers.getAssistantMessageText(2);
  expect(thirdAnswer).toContain('test_debug_2');

  // Test regenerate on third response
  await messageHelpers.clickAssistantRegenerate(2);
  await page.waitForTimeout(300);
  await messageHelpers.waitForAssistantMessageTextContains(2, 'test_debug_2');

  const thirdAnswerRegen = await messageHelpers.getAssistantMessageText(2);
  expect(thirdAnswerRegen).toContain('test_debug_2');

  // Regenerate the second answer again
  await messageHelpers.clickAssistantRegenerate(1);
  await page.waitForTimeout(300);
  await messageHelpers.waitForAssistantMessageTextContains(1, 'test_debug_1');

  // Verify the second answer regeneration
  const secondAnswerRegen2 = await messageHelpers.getAssistantMessageText(1);
  expect(secondAnswerRegen2).toContain('test_debug_1');
  expect(secondAnswerRegen2).not.toContain('test_debug_2');

  // Final verification
  const secondAnswerRegen3 = await messageHelpers.getAssistantMessageText(1);
  expect(secondAnswerRegen3).toContain('test_debug_1');
  expect(secondAnswerRegen3).not.toContain('test_debug_2');
});
