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

  // locate by id #message_texarea and click
  // Find the element by Id and click on it
  //const messageTextArea = await page.$('#message_texarea');
  //await messageTextArea?.click();
  //await messageTextArea?.fill('test_demo_bestqa');
  await page.getByTestId("message_textarea").click()
  const input_area = await page.$("#message_textarea textarea")
  await input_area?.fill('test_demo_bestqa');
  // await page.fill("#message_textarea", 'test_demo_bestqa');
  //await page.getByPlaceholder('来说点什么吧...（Shift + Enter = 换行）').press('Enter');
  await input_area?.press('Enter');
  // sleep 500ms
  await page.waitForTimeout(1000);
  // get by id

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
  expect(messages.length).toBe(1);


});