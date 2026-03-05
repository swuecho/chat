import { test, expect } from '@playwright/test';
import { randomEmail } from '../lib/sample';
import { setupDebugChatSession, sendMessageAndWaitAssistantCount } from '../lib/chat-test-setup';

const test_email = randomEmail();

test('test', async ({ page }) => {
  const { inputHelpers, messageHelpers } = await setupDebugChatSession(page, test_email);

  await sendMessageAndWaitAssistantCount(inputHelpers, messageHelpers, 'test_demo_bestqa', 1);
  await sendMessageAndWaitAssistantCount(inputHelpers, messageHelpers, 'test_debug_1', 2);
  await sendMessageAndWaitAssistantCount(inputHelpers, messageHelpers, 'test_debug_2', 3);

  const assistantMessages = await messageHelpers.getAssistantMessages();
  expect(assistantMessages.length).toBeGreaterThanOrEqual(3);

});
