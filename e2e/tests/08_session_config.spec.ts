import { test, expect } from '@playwright/test';
import { selectUserByEmail } from '../lib/db/user';
import { Pool } from 'pg';

import { selectChatSessionByUserId as selectChatSessionsByUserId } from '../lib/db/chat_session';

import { randomEmail } from '../lib/sample';
import { db_config } from '../lib/db/config';

const test_email = randomEmail();

const pool = new Pool(db_config);


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
  await input_area?.press('Enter');
  await page.waitForTimeout(1000);

  const user = await selectUserByEmail(pool, test_email);
  expect(user.email).toBe(test_email);

  const sessions = await selectChatSessionsByUserId(pool, user.id);
  expect(sessions.length).toBe(1);
  const new_sesion = sessions[0]
  expect(new_sesion.debug).toBe(false);
  expect(new_sesion.temperature).toBe(1);
  // click the button to open the modal 
  await page.getByRole('contentinfo').getByRole('button').nth(3).click();
  await page.getByText('关闭').nth(1).click();
  // sleep 1s
  await page.waitForTimeout(1000);
  const sessions_2 = await selectChatSessionsByUserId(pool, user.id);
  const new_sesion_2 = sessions_2[0]
  expect(new_sesion_2.temperature).toBe(1);
  expect(new_sesion_2.n).toBe(1);
  expect(new_sesion_2.debug).toBe(true);
});

