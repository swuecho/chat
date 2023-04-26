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

let snapshot_url = ""
test('test', async ({ page }) => {
  await page.goto('/');

  await page.getByTestId('email').click();
  await page.getByTestId('email').locator('input').fill(test_email);
  await page.getByTestId('password').locator('input').click();
  await page.getByTestId('password').locator('input').fill('@ThisIsATestPass5');
  await page.getByTestId('signup').click();

  await page.waitForTimeout(1000);
  let input_area = await page.$("#message_textarea textarea")
  await input_area?.click();
  await input_area?.fill('test_demo_bestqa');
  // await page.fill("#message_textarea", 'test_demo_bestqa');
  //await page.getByPlaceholder('来说点什么吧...（Shift + Enter = 换行）').press('Enter');
  await input_area?.press('Enter');
  // sleep 500ms
  await page.waitForTimeout(1000);
  await input_area?.click();
  await input_area?.fill('test_demo_bestqa');
  await input_area?.press('Enter');

  await page.waitForTimeout(1000);

  const user = await selectUserByEmail(pool, test_email);
  expect(user.email).toBe(test_email);
  // expect(user.id).toBe(37);
  const sessions = await selectChatSessionsByUserId(pool, user.id);
  const session = sessions[0];
  const prompts = await selectChatPromptsBySessionUUID(pool, session.uuid)
  expect(prompts.length).toBe(1);
  expect(prompts[0].updated_by).toBe(user.id);
  // sleep 500ms
  await page.waitForTimeout(500);
  const messages = await selectChatMessagesBySessionUUID(pool, session.uuid)
  expect(messages.length).toBe(3);
  const page1Promise = page.waitForEvent('popup');
  await page.getByTestId('snpashot-button').getByRole('button').click();
  await page.waitForTimeout(500)
  const page_snapshot = await page1Promise;
  await page_snapshot.waitForTimeout(500)
  snapshot_url = page_snapshot.url()
  expect(snapshot_url).toMatch(/snapshot/)
  // continue to chat
  await page_snapshot.locator('.floating-button > div > .flex').click();
  const page_back_to_chat_promise = page_snapshot.waitForEvent('popup');
  const page_back = await page_back_to_chat_promise
  expect(page_back.url()).toMatch(/chat/)
  await page_back.waitForTimeout(500)
  const sessions_new = await selectChatSessionsByUserId(pool, user.id);
  // two conversion now
  expect(sessions_new.length).toBe(2)
});
