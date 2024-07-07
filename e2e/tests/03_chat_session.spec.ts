import { test, expect } from '@playwright/test';
import { randomEmail } from '../lib/sample';
import { Pool } from 'pg';
import { selectUserByEmail } from '../lib/db/user';
import { selectChatSessionByUserId as selectChatSessionsByUserId } from '../lib/db/chat_session';
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

  await page.waitForTimeout(1000);;
  await page.getByTestId('edit_session_topic').click();
  await page.getByTestId('edit_session_topic_input').locator('input').fill('This is a test topic');
  await page.getByTestId('save_session_topic').click();

  // sleep 500ms
  await page.waitForTimeout(1000);;
  const user = await selectUserByEmail(pool, test_email);
  expect(user.email).toBe(test_email);
  // expect(user.id).toBe(37);
  const sessions = await selectChatSessionsByUserId(pool, user.id);
  const session = sessions[0];
  expect(session.topic).toBe('This is a test topic');
});