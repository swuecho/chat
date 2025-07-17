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
  await page.waitForTimeout(1000);

  await page.locator('a').filter({ hasText: 'New Chat' }).click();

  // Set debug mode
  await page.getByRole('contentinfo').getByRole('button').nth(3).click();
  await page.getByTestId('debug_mode').click();
  await page.click('body', { position: { x: 0, y: 0 } });

  // Send first message
  await inputHelpers.sendMessage('test_demo_bestqa');
  
  // Wait for and verify first response
  await messageHelpers.waitForMessageCount(2); // User message + Bot response
  const firstAnswer = await messageHelpers.getMessageText(1); // Bot response is index 1
  expect(firstAnswer).toContain('test_demo_bestqa');

  // Send second message
  await inputHelpers.sendMessage('test_debug_1');
  
  // Wait for and verify second response
  await messageHelpers.waitForMessageCount(4); // 2 previous + user message + bot response
  const secondAnswer = await messageHelpers.getMessageText(3); // Bot response is index 3
  expect(secondAnswer).toContain('test_debug_1');

  // Test regenerate functionality on second response
  const isRegenerateVisible = await messageHelpers.isRegenerateButtonVisible(3);
  expect(isRegenerateVisible).toBe(true);
  
  await messageHelpers.clickRegenerate(3);
  await page.waitForTimeout(1000);

  // Verify regenerated response still contains the expected text
  const secondAnswerRegen = await messageHelpers.getMessageText(3);
  expect(secondAnswerRegen).toContain('test_debug_1');

  // Send third message
  await inputHelpers.sendMessage('test_debug_2');
  
  // Wait for and verify third response
  await messageHelpers.waitForMessageCount(6); // Previous + user message + bot response
  const thirdAnswer = await messageHelpers.getMessageText(5); // Bot response is index 5
  expect(thirdAnswer).toContain('test_debug_2');

  // Test regenerate on third response
  await messageHelpers.clickRegenerate(5);
  await page.waitForTimeout(1000);

  const thirdAnswerRegen = await messageHelpers.getMessageText(5);
  expect(thirdAnswerRegen).toContain('test_debug_2');

  // Regenerate the second answer again
  await messageHelpers.clickRegenerate(3);
  await page.waitForTimeout(1000);

  // Verify the second answer regeneration
  const secondAnswerRegen2 = await messageHelpers.getMessageText(3);
  expect(secondAnswerRegen2).toContain('test_debug_1');
  expect(secondAnswerRegen2).not.toContain('test_debug_2');

  // Final verification
  const secondAnswerRegen3 = await messageHelpers.getMessageText(3);
  expect(secondAnswerRegen3).toContain('test_debug_1');
  expect(secondAnswerRegen3).not.toContain('test_debug_2');
});