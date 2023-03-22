import { test, expect } from '@playwright/test';

//generate a random email address
function randomEmail() {
  const random = Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
  return `${random}@test.com`;
}
const test_email = randomEmail();

test('test', async ({ page }) => {
  await page.goto('/');
  await page.getByTestId('email').click();
  await page.getByTestId('email').locator('input').fill(test_email);
  await page.getByTestId('password').locator('input').click();
  await page.getByTestId('password').locator('input').fill('‘@ThisIsATestPass5'');
  await page.getByTestId('signup').click();
});

