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

  await page.getByTestId('email').click();
  await page.getByTestId('email').locator('input').fill(test_email);
  await page.getByTestId('password').locator('input').click();
  await page.getByTestId('password').locator('input').fill('‘@ThisIsATestPass5'');
  await page.getByTestId('signup').click();


  await page.getByRole('complementary').getByRole('button').nth(1).click();
  await page.getByPlaceholder('请输入').dblclick();
  await page.getByPlaceholder('请输入').fill('This is a test topic');
  await page.getByRole('main').filter({ hasText: 'New chat' }).locator('a').getByRole('button').click();

  // sleep 500ms
  await page.waitForTimeout(500);
  const user = await selectUserByEmail(pool, test_email);
  expect(user.email).toBe(test_email);
  // expect(user.id).toBe(37);
  const sessions = await selectChatSessionsByUserId(pool, user.id);
  const session = sessions[0];
  expect(session.topic).toBe('This is a test topic');
});