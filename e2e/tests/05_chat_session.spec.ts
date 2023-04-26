import { test, expect } from '@playwright/test';
import { Pool } from 'pg';
import { selectUserByEmail } from '../lib/db/user';
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
        const user = await selectUserByEmail(pool, test_email);
        expect(user.email).toBe(test_email);

        const sessions = await selectChatSessionsByUserId(pool, user.id);
        expect(sessions.length).toBe(1);

        // test edit session topic
        await page.getByTestId('edit_session_topic').click();
        await page.getByTestId('edit_session_topic_input').locator('input').fill('test_session_topic');
        await page.getByTestId('save_session_topic').click();

        await page.waitForTimeout(200);
        const sessions_1 = await selectChatSessionsByUserId(pool, user.id);
        expect(sessions_1.length).toBe(1);
        const session_1 = sessions_1[0];
        expect(session_1.topic).toBe('test_session_topic');

        await page.getByRole('button', { name: '新对话' }).click();
        await page.getByTestId('edit_session_topic').click();
        await page.getByTestId('edit_session_topic_input').locator('input').click();
        await page.getByTestId('edit_session_topic_input').locator('input').fill('test_session_topic_2');
        await page.getByTestId('save_session_topic').click();

        await page.getByRole('button', { name: '新对话' }).click();
        await page.getByTestId('edit_session_topic').click();
        await page.getByTestId('edit_session_topic_input').locator('input').fill('test_session_topic_3');
        await page.getByTestId('save_session_topic').click();
        // sleep 500ms
        await page.waitForTimeout(500);
        // should have three sessions
        const sessions_3 = await selectChatSessionsByUserId(pool, user.id);
        expect(sessions_3.length).toBe(3);
        expect(sessions_3[0].topic).toBe('test_session_topic');
        expect(sessions_3[1].topic).toBe('test_session_topic_2');
        expect(sessions_3[2].topic).toBe('test_session_topic_3');

});