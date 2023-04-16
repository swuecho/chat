import { test, expect } from '@playwright/test';
import { Pool } from 'pg';
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

  await page.locator('a').filter({ hasText: '新对话' }).click();

  // set debug mode
  await page.getByRole('contentinfo').getByRole('button').nth(3).click();
  await page.getByText('关闭').nth(1).click();

  // click bottom of page to close the modal

  await page.click('body', { position: { x: 0, y: 0 } });

  let input_area = await page.$("#message_textarea textarea")
  await input_area?.click();
  await input_area?.fill('test_demo_bestqa');
  await input_area?.press('Enter');
  await page.waitForTimeout(1000);

  const first_answer = await page.$eval('#image-wrapper .chat-message:nth-child(2) .message-text', (el: HTMLElement) => el.innerText);
  // check the answer return by the server
  expect(first_answer).toContain('test_demo_bestqa');

  await input_area?.click();
  await input_area?.fill('test_debug_1');
  await input_area?.press('Enter');
  await page.waitForTimeout(1000);
  // check the answer return by the server
  await page.waitForSelector('#image-wrapper .chat-message:nth-child(4) .message-text');
  const sec_answer = await page.$eval('#image-wrapper .chat-message:nth-child(4) .message-text', (el: HTMLElement) => el.innerText);
  // check the sec_answer has the debug message
  expect(sec_answer).toContain('test_debug_1');

  await page.locator('#image-wrapper .chat-message:nth-child(4) .chat-message-regenerate').click();
  await page.waitForTimeout(1000);

  const sec_answer_regen = await page.$eval('#image-wrapper .chat-message:nth-child(4) .message-text', (el: HTMLElement) => el.innerText);
  // check the sec_answer has the debug message
  expect(sec_answer_regen).toContain('test_debug_1');

  // add new message "test_debug_2"
  await input_area?.click();
  await input_area?.fill('test_debug_2');
  await input_area?.press('Enter');
  await page.waitForTimeout(1000);
  // check the answer return by the server
  await page.waitForSelector('#image-wrapper .chat-message:nth-child(6) .message-text');
  const third_answer = await page.$eval('#image-wrapper .chat-message:nth-child(6) .message-text', (el: HTMLElement) => el.innerText);
  // check the third_answer has the debug message
  expect(third_answer).toContain('test_debug_2');

  await page.locator('#image-wrapper .chat-message:nth-child(6) .chat-message-regenerate').click();
  await page.waitForTimeout(1000);

  const third_answer_regen = await page.$eval('#image-wrapper .chat-message:nth-child(6) .message-text', (el: HTMLElement) => el.innerText);
  // check the third_answer has the debug message
  expect(third_answer_regen).toContain('test_debug_2');

  // regenerate the second answer
  await page.locator('#image-wrapper .chat-message:nth-child(4) .chat-message-regenerate').click();
  await page.waitForTimeout(1000);

  // check the second answer has been regenerated
  const sec_answer_regen_2 = await page.$eval('#image-wrapper .chat-message:nth-child(4) .message-text', (el: HTMLElement) => el.innerText);
  // check the sec_answer has the debug message
  expect(sec_answer_regen_2).toContain('test_debug_1');
  expect(sec_answer_regen_2).not.toContain('test_debug_2')

  // check the second answer has been regenerated
  const sec_answer_regen_3 = await page.$eval('#image-wrapper .chat-message:nth-child(4) .message-text', (el: HTMLElement) => el.innerText);
  // check the sec_answer has the debug message
  expect(sec_answer_regen_3).toContain('test_debug_1');
  expect(sec_answer_regen_2).not.toContain('test_debug_2')
});