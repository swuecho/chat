import { test, expect } from '@playwright/test';
import { randomEmail } from '../lib/sample';

const test_email = randomEmail();

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
  
  // Wait for authentication to complete
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(3000);
  
  // Wait for the permission modal to disappear
  try {
    await page.waitForSelector('.n-modal-mask', { state: 'detached', timeout: 10000 });
  } catch (error) {
    // Modal might already be gone
  }
  
  await page.waitForTimeout(1000);

  await page.locator('a').filter({ hasText: 'New Chat' }).click();

  // set debug mode
  await page.getByTestId('config-button').click();
  await page.getByTestId('debug_mode').click();


  // click bottom of page to close the modal

  await page.click('body', { position: { x: 0, y: 0 } });

  let input_area = await page.$("#message_textarea textarea")
  await input_area?.click();
  await input_area?.fill('test_demo_bestqa');
  await input_area?.press('Enter');
  await page.waitForTimeout(1000);

  // Wait for response and get the text more reliably
  await page.waitForSelector('.chat-message:nth-of-type(2) .message-text', { timeout: 10000 });
  const first_answer = await page.$eval('.chat-message:nth-of-type(2) .message-text', (el: HTMLElement) => el.innerText);
  // check the answer return by the server
  expect(first_answer).toContain('test_demo_bestqa');

  await input_area?.click();
  await input_area?.fill('test_debug_1');
  await input_area?.press('Enter');
  await page.waitForTimeout(1000);
  // check the answer return by the server
  await page.waitForSelector('.chat-message:nth-of-type(4) .message-text', { timeout: 10000 });
  const sec_answer = await page.$eval('.chat-message:nth-of-type(4) .message-text', (el: HTMLElement) => el.innerText);
  // check the sec_answer has the debug message
  expect(sec_answer).toContain('test_debug_1');

  // add new message "test_debug_2"
  await input_area?.click();
  await input_area?.fill('test_debug_2');
  await input_area?.press('Enter');
  await page.waitForTimeout(1000);
  // check the answer return by the server
  await page.waitForSelector('.chat-message:nth-of-type(6) .message-text', { timeout: 10000 });
  const third_answer = await page.$eval('.chat-message:nth-of-type(6) .message-text', (el: HTMLElement) => el.innerText);
  // check the third_answer has the debug message
  expect(third_answer).toContain('test_debug_2');

});