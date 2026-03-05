import { test, expect } from '@playwright/test';
import { randomEmail } from '../lib/sample';
import { MessageHelpers } from '../lib/message-helpers';

const test_email = randomEmail();

test('test', async ({ page }) => {
  const messageHelpers = new MessageHelpers(page);
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
  await page.getByTestId('chat-settings-button').click();
  // expand the Advanced Settings section (accordion)
  await page.getByTestId('collapse-advanced').click();
  // wait for the section to expand
  await page.waitForTimeout(300);
  await page.getByTestId('debug_mode').click();


  // click bottom of page to close the modal

  await page.click('body', { position: { x: 0, y: 0 } });

  let input_area = await page.$("#message_textarea textarea")
  await input_area?.click();
  await input_area?.fill('test_demo_bestqa');
  await input_area?.press('Enter');
  await page.waitForTimeout(300);

  // Wait for first assistant response
  await messageHelpers.waitForAssistantMessageWithText('test_demo_bestqa');
  const firstMessage = await messageHelpers.getAssistantMessageByContent('test_demo_bestqa');
  const first_answer = firstMessage ? await firstMessage.locator('.message-text').innerText() : '';
  // check the answer return by the server
  expect(first_answer).toContain('test_demo_bestqa');

  await input_area?.click();
  await input_area?.fill('test_debug_1');
  await input_area?.press('Enter');
  await page.waitForTimeout(300);
  // Wait for second assistant response
  await messageHelpers.waitForAssistantMessageWithText('test_debug_1');
  const secondMessage = await messageHelpers.getAssistantMessageByContent('test_debug_1');
  const sec_answer = secondMessage ? await secondMessage.locator('.message-text').innerText() : '';
  // check the sec_answer has the debug message
  expect(sec_answer).toContain('test_debug_1');

  // add new message "test_debug_2"
  await input_area?.click();
  await input_area?.fill('test_debug_2');
  await input_area?.press('Enter');
  await page.waitForTimeout(300);
  // Wait for third assistant response
  await messageHelpers.waitForAssistantMessageWithText('test_debug_2');
  const thirdMessage = await messageHelpers.getAssistantMessageByContent('test_debug_2');
  const third_answer = thirdMessage ? await thirdMessage.locator('.message-text').innerText() : '';
  // check the third_answer has the debug message
  expect(third_answer).toContain('test_debug_2');

});
