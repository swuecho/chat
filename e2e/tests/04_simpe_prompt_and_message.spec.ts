import { test, expect } from '@playwright/test';
import { Pool } from 'pg';
import { selectUserByEmail } from '../lib/db/user';
import { selectChatSessionByUserId as selectChatSessionsByUserId } from '../lib/db/chat_session';
import { selectChatPromptsBySessionUUID } from '../lib/db/chat_prompt';
import { selectChatMessagesBySessionUUID } from '../lib/db/chat_message';
import { randomEmail } from '../lib/sample';
import { db_config } from '../lib/db/config';


const test_email = randomEmail();


const pool = new Pool(db_config);


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

test('test', async ({ page }) => {
  await page.goto('/');

  // Wait for the page reload after successful signup and permission modal to disappear
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
  await input_area?.click();
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

  const user = await selectUserByEmail(pool, test_email);
  expect(user.email).toBe(test_email);
  const sessions = await selectChatSessionsByUserId(pool, user.id);
  const session = sessions[0];
  const prompts = await selectChatPromptsBySessionUUID(pool, session.uuid)
  expect(prompts.length).toBe(1);
  expect(prompts[0].updated_by).toBe(user.id);

  // Poll database until all 4 messages are saved
  await waitForMessageCount(pool, session.uuid, 4);

  // test edit session topic
  await page.getByTestId('edit_session_topic').click();
  await page.getByTestId('edit_session_topic_input').locator('input').fill('test_session_topic');
  await page.getByTestId('save_session_topic').click();
  await input_area?.click();
  await input_area?.fill('test_demo_bestqa');
  await input_area?.press('Enter');

  // Wait for third assistant response to appear (total 3)
  await page.waitForFunction(
    () => {
      const messages = Array.from(document.querySelectorAll('.chat-message'));
      const assistantCount = messages.filter((message) => {
        const row = message.querySelector('.flex.w-full');
        return row && !row.classList.contains('flex-row-reverse');
      }).length;
      return assistantCount >= 3;
    },
    { timeout: 15000 }
  );

  const sessions_1 = await selectChatSessionsByUserId(pool, user.id);
  const session_1 = sessions_1[0];
  expect(session_1.topic).toBe('test_session_topic');
  const prompts_1 = await selectChatPromptsBySessionUUID(pool, session_1.uuid)
  expect(prompts_1.length).toBe(1);
  expect(prompts_1[0].updated_by).toBe(user.id);

  // Poll database until all 6 messages are saved
  await waitForMessageCount(pool, session_1.uuid, 6);

});
