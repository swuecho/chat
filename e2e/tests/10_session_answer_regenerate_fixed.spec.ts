import { test, expect } from '@playwright/test';
import { randomEmail } from '../lib/sample';
import { setupDebugChatSession, sendMessageAndWaitAssistantCount } from '../lib/chat-test-setup';

const test_email = randomEmail();

test('session answer regenerate - robust version', async ({ page }) => {
  const { inputHelpers, messageHelpers } = await setupDebugChatSession(page, test_email);
  await sendMessageAndWaitAssistantCount(inputHelpers, messageHelpers, 'test_demo_bestqa', 1);
  await sendMessageAndWaitAssistantCount(inputHelpers, messageHelpers, 'test_debug_1', 2);

  // Test regenerate functionality on second response
  const isRegenerateVisible = await messageHelpers.isAssistantRegenerateButtonVisible(1);
  expect(isRegenerateVisible).toBe(true);
  
  await messageHelpers.clickAssistantRegenerate(1);
  await page.waitForTimeout(300);
  await messageHelpers.waitForAssistantMessageCount(2);
  expect(await messageHelpers.isAssistantRegenerateButtonVisible(1)).toBe(true);

  await sendMessageAndWaitAssistantCount(inputHelpers, messageHelpers, 'test_debug_2', 3);

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
