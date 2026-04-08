import { test, expect } from '@playwright/test';
import { Pool } from 'pg';
import { selectUserByEmail } from '../lib/db/user';
import { selectChatSessionByUserId as selectChatSessionsByUserId } from '../lib/db/chat_session';
import { selectChatPromptsBySessionUUID } from '../lib/db/chat_prompt';
import { selectChatMessagesBySessionUUID } from '../lib/db/chat_message';
import { randomEmail } from '../lib/sample';
import { db_config } from '../lib/db/config';
import { getClearConversationButton } from '../lib/button-helpers';

const pool = new Pool(db_config);

const test_email = randomEmail();

async function waitForMessageCount(pool: Pool, sessionUuid: string, expectedCount: number, timeout = 10000): Promise<void> {
  const startTime = Date.now();
  while (Date.now() - startTime < timeout) {
    const messages = await selectChatMessagesBySessionUUID(pool, sessionUuid);
    if (messages.length >= expectedCount) {
      return;
    }
    await new Promise(resolve => setTimeout(resolve, 500));
  }
  const messages = await selectChatMessagesBySessionUUID(pool, sessionUuid);
  expect(messages.length).toBe(expectedCount);
}

test('after clear conversation, only system message remains', async ({ page }) => {
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

  await page.waitForSelector('[data-testid="chat-settings-button"]', { timeout: 10000 });
  await page.waitForSelector('#message_textarea textarea', { timeout: 10000 });
  await page.waitForTimeout(500);
  let input_area = await page.$("#message_textarea textarea")
  await input_area?.click();
  await input_area?.fill('test_demo_bestqa');
  await input_area?.press('Enter');

  // Wait for first assistant response to appear
  await page.waitForFunction(
    () => {
      const messages = Array.from(document.querySelectorAll('.chat-message'));
      const assistantCount = messages.filter((message) => {
        const row = message.querySelector('.flex.w-full');
        return row && !row.classList.contains('flex-row-reverse');
      }).length;
      return assistantCount >= 1;
    },
    { timeout: 15000 }
  );

  // Send second message
  await input_area?.fill('test_demo_bestqa');
  await input_area?.press('Enter');

  // Wait for second assistant response to appear
  await page.waitForFunction(
    () => {
      const messages = Array.from(document.querySelectorAll('.chat-message'));
      const assistantCount = messages.filter((message) => {
        const row = message.querySelector('.flex.w-full');
        return row && !row.classList.contains('flex-row-reverse');
      }).length;
      return assistantCount >= 2;
    },
    { timeout: 15000 }
  );

  const message_counts = await page.$$eval('.message-text', (messages) => messages.length);
  expect(message_counts).toBe(4);

  const user = await selectUserByEmail(pool, test_email);
  expect(user.email).toBe(test_email);
  const sessions = await selectChatSessionsByUserId(pool, user.id);
  const session = sessions[0];

  // clear
  const clearButton = await getClearConversationButton(page);
  await clearButton.click();
  await page.getByRole('button', { name: 'Yes' }).click();

  await page.waitForTimeout(1000);
  const message_count_after_clear = await page.$$eval('.message-text', (messages) => messages.length);
  expect(message_count_after_clear).toBe(1);

  const prompts = await selectChatPromptsBySessionUUID(pool, session.uuid)
  expect(prompts.length).toBe(1);
  expect(prompts[0].updated_by).toBe(user.id);

  // Poll database until messages are cleared (0 messages expected)
  await waitForMessageCount(pool, session.uuid, 0);

});
