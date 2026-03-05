import { test, expect } from '@playwright/test';
import { Pool } from 'pg';
import { randomEmail } from '../lib/sample';
import { db_config } from '../lib/db/config';
import { MessageHelpers } from '../lib/message-helpers';

const test_email = randomEmail();

const pool = new Pool(db_config);


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
  await messageHelpers.waitForAssistantMessageCount(1);
  await messageHelpers.waitForAssistantMessageTextContains(0, 'test_demo_bestqa');
  const first_answer = await messageHelpers.getAssistantMessageText(0);
  // check the answer return by the server
  expect(first_answer).toContain('test_demo_bestqa');

  await input_area?.click();
  await input_area?.fill('test_debug_1');
  await input_area?.press('Enter');
  await page.waitForTimeout(300);
  // Wait for second assistant response
  await messageHelpers.waitForAssistantMessageCount(2);
  await messageHelpers.waitForAssistantMessageTextContains(1, 'test_debug_1');
  const sec_answer = await messageHelpers.getAssistantMessageText(1);
  // check the sec_answer has the debug message
  expect(sec_answer).toContain('test_debug_1');

  // Click regenerate button with better selector and error handling
  await messageHelpers.clickAssistantRegenerate(1);
  await page.waitForTimeout(300);
  await messageHelpers.waitForAssistantMessageTextContains(1, 'test_debug_1');
  const sec_answer_regen = await messageHelpers.getAssistantMessageText(1);
  // check the sec_answer has the debug message
  expect(sec_answer_regen).toContain('test_debug_1');

  // add new message "test_debug_2"
  await input_area?.click();
  await input_area?.fill('test_debug_2');
  await input_area?.press('Enter');
  await page.waitForTimeout(300);
  // Wait for third assistant response
  await messageHelpers.waitForAssistantMessageCount(3);
  await messageHelpers.waitForAssistantMessageTextContains(2, 'test_debug_2');
  const third_answer = await messageHelpers.getAssistantMessageText(2);
  // check the third_answer has the debug message
  expect(third_answer).toContain('test_debug_2');

  await messageHelpers.clickAssistantRegenerate(2);
  await page.waitForTimeout(300);
  await messageHelpers.waitForAssistantMessageTextContains(2, 'test_debug_2');
  const third_answer_regen = await messageHelpers.getAssistantMessageText(2);
  // check the third_answer has the debug message
  expect(third_answer_regen).toContain('test_debug_2');

  // regenerate the second answer
  await messageHelpers.clickAssistantRegenerate(1);
  await page.waitForTimeout(300);
  await messageHelpers.waitForAssistantMessageTextContains(1, 'test_debug_1');

  // check the second answer has been regenerated
  const sec_answer_regen_2 = await messageHelpers.getAssistantMessageText(1);
  // check the sec_answer has the debug message
  expect(sec_answer_regen_2).toContain('test_debug_1');
  expect(sec_answer_regen_2).not.toContain('test_debug_2')

  // check the second answer has been regenerated
  const sec_answer_regen_3 = await messageHelpers.getAssistantMessageText(1);
  // check the sec_answer has the debug message
  expect(sec_answer_regen_3).toContain('test_debug_1');
  expect(sec_answer_regen_2).not.toContain('test_debug_2')
});
