import { test, expect } from '@playwright/test';
import { Pool } from 'pg';
import { selectUserByEmail } from '../lib/db/user';
import { selectChatSessionByUserId as selectChatSessionsByUserId } from '../lib/db/chat_session';
import { selectChatPromptsBySessionUUID } from '../lib/db/chat_prompt';
import { selectChatMessagesBySessionUUID } from '../lib/db/chat_message';
import { randomEmail } from '../lib/sample';
import { db_config } from '../lib/db/config';

const pool = new Pool(db_config);

const test_email = randomEmail();

test('after clear conversation, only system message remains', async ({ page }) => {
  await page.goto('/');
  await page.getByTestId('email').click();
  await page.getByTestId('email').locator('input').fill(test_email);
  await page.getByTestId('password').locator('input').click();
  await page.getByTestId('password').locator('input').fill('@ThisIsATestPass5');
  await page.getByTestId('signup').click();
  // sleep 1 second
  await page.waitForTimeout(1000);
  let input_area = await page.$("#message_textarea textarea")
  await input_area?.click();
  await input_area?.fill('test_demo_bestqa');
  await input_area?.press('Enter');
  await page.waitForTimeout(1000);
  await input_area?.fill('test_demo_bestqa');
  await input_area?.press('Enter');
  // get message counts in the conversation

  await page.waitForTimeout(1000);

  const message_counts = await page.$$eval('.message-text', (messages) => messages.length);
  expect(message_counts).toBe(4);

  const user = await selectUserByEmail(pool, test_email);
  expect(user.email).toBe(test_email);
  // expect(user.id).toBe(37);
  const sessions = await selectChatSessionsByUserId(pool, user.id);
  const session = sessions[0];

  // clear
  await page.getByRole('contentinfo').getByRole('button').first().click();
  await page.getByRole('button', { name: 'Yes' }).click();

  // sleep 500 ms
  await page.waitForTimeout(1000);
  // get message counts in the conversation
  const message_count_after_clear = await page.$$eval('.message-text', (messages) => messages.length);
  expect(message_count_after_clear).toBe(1);

  const prompts = await selectChatPromptsBySessionUUID(pool, session.uuid)
  expect(prompts.length).toBe(1);
  expect(prompts[0].updated_by).toBe(user.id);
  // sleep 5 seconds
  await page.waitForTimeout(500);
  const messages = await selectChatMessagesBySessionUUID(pool, session.uuid)
  expect(messages.length).toBe(0);

});

