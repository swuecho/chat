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

  await page.getByTitle('signuptab').click();
  await page.getByTestId('signup_email').click();
  await page.getByTestId('signup_email').locator('input').fill(test_email);
  await page.getByTestId('signup_password').locator('input').click();
  await page.getByTestId('signup_password').locator('input').fill('@ThisIsATestPass5');
  await page.getByTestId('repwd').locator('input').click();
  await page.getByTestId('repwd').locator('input').fill('@ThisIsATestPass5');
  await page.getByTestId('signup').click();

  // Wait for signup to complete - either successful or with error
  try {
    await page.waitForLoadState('networkidle', { timeout: 15000 });
  } catch (error) {
    // Continue if networkidle times out - the page might still be functional
    console.log('Network idle timeout, continuing...');
  }

  await page.waitForTimeout(3000);

  // Wait for the permission modal to disappear OR wait for message textarea to be clickable
  try {
    await page.waitForSelector('.n-modal-mask', { state: 'detached', timeout: 5000 });
  } catch (error) {
    // Modal might already be gone or not exist
    console.log('Modal mask not found, continuing...');
  }

  // Alternative approach: wait for the message textarea to be available and force click if needed
  await page.waitForSelector('#message_textarea textarea', { timeout: 10000 });

  // Try to click, and if blocked by modal, dismiss it first
  try {
    await page.getByTestId("message_textarea").click({ timeout: 5000 });
  } catch (error) {
    // If click is blocked, try to dismiss any modal and retry
    console.log('Click blocked, trying to dismiss modal...');
    try {
      // Try to click outside modal to dismiss it
      await page.click('body', { position: { x: 10, y: 10 }, timeout: 2000 });
      await page.waitForTimeout(1000);
    } catch (dismissError) {
      // Continue anyway
    }
    // Retry the click
    await page.getByTestId("message_textarea").click();
  }
  await page.waitForTimeout(1000);
  const input_area = await page.$("#message_textarea textarea")
  await input_area?.fill('test_demo_bestqa');
  // await page.fill("#message_textarea", 'test_demo_bestqa');
  //await page.getByPlaceholder('来说点什么吧...（Shift + Enter = 换行）').press('Enter');
  await input_area?.press('Enter');
  // sleep 500ms
  await page.waitForTimeout(5000); // Increased from 1000ms to 5000ms
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
  await page.waitForTimeout(1000);;
  const messages = await selectChatMessagesBySessionUUID(pool, session.uuid)
  expect(messages.length).toBe(1);
  expect(messages[0].role).toBe('assistant');


});