import { test, expect } from '@playwright/test';

//generate a random email address
function randomEmail() {
        const random = Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
        return `${random}@test.com`;
}
const test_email = randomEmail();

test.skip('test', async ({ page }) => {
        await page.goto('/');
        await page.getByTitle('signuptab').click();
        await page.getByTestId('signup_email').click();
        await page.getByTestId('signup_email').locator('input').fill(test_email);
        await page.getByTestId('signup_password').locator('input').click();
        await page.getByTestId('signup_password').locator('input').fill('@ThisIsATestPass5');
        await page.getByTestId('repwd').locator('input').click();
        await page.getByTestId('repwd').locator('input').fill('@ThisIsATestPass5');
        await page.getByTestId('signup').click();
        await page.waitForTimeout(1000);

        await page.getByRole('contentinfo').getByRole('button').nth(3).click();
        // change the value of the slider
        // Find the slider element and adjust its value
        const sliderRailFill = await page.$('.n-slider-rail__fill')
        expect(sliderRailFill).toBeTruthy()
        await sliderRailFill?.evaluate((element) => {
                element.setAttribute('style', 'width: 25%;')
        }
        )
        // sliderRailFill?.setAttribute('style', 'width: 25%;')
        await page.waitForTimeout(1000);
        await page.locator('.n-slider-handles').click();
        await page.locator('.n-slider').click();
        await page.locator('.n-slider').click();
        await page.locator('.n-slider-handles').click();
        await page.locator('.n-slider-handles').click();
        await page.locator('.n-slider').click();

});
