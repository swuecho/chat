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



test('test', async ({ page }) => {
  await page.goto('/');

  await page.getByTestId('email').click();
  await page.getByTestId('email').locator('input').fill(test_email);
  await page.getByTestId('password').locator('input').click();
  await page.getByTestId('password').locator('input').fill('‘@ThisIsATestPass5'');
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
  // sleep 5 seconds
  await page.waitForTimeout(500);
  const messages = await selectChatMessagesBySessionUUID(pool, session.uuid)
  expect(messages.length).toBe(3);

  // test edit session topic
  await page.getByTestId('edit_session_topic').click();
  await page.getByTestId('edit_session_topic_input').locator('input').fill('test_session_topic');
  await page.getByTestId('save_session_topic').click();
  await input_area?.click();
  await input_area?.fill('test_demo_bestqa');
  await input_area?.press('Enter');

  const sessions_1 = await selectChatSessionsByUserId(pool, user.id);
  const session_1 = sessions_1[0];
  expect(session_1.topic).toBe('test_session_topic');
  const prompts_1 = await selectChatPromptsBySessionUUID(pool, session_1.uuid)
  expect(prompts_1.length).toBe(1);
  expect(prompts_1[0].updated_by).toBe(user.id);
  // sleep 5 seconds
  await page.waitForTimeout(500);
  const messages_1 = await selectChatMessagesBySessionUUID(pool, session_1.uuid)
  expect(messages_1.length).toBe(5);

});