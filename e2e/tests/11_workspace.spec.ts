import { test, expect } from '@playwright/test';
import { randomEmail } from '../lib/sample';
import { Pool } from 'pg';
import { selectUserByEmail } from '../lib/db/user';
import { selectWorkspacesByUserId, selectWorkspaceByUuid, countSessionsInWorkspace } from '../lib/db/chat_workspace';
import { selectChatSessionByUserId } from '../lib/db/chat_session';
import { db_config } from '../lib/db/config';

const test_email = randomEmail();
const pool = new Pool(db_config);

test('workspace management - create workspace and manage sessions', async ({ page }) => {
  // Register user
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

  // Create new workspace named 'test_workspace_1' via workspace selector dropdown
  await page.locator('.workspace-button').click();
  await page.getByText('Create workspace').click(); // Based on t('workspace.create')
  
  // Fill workspace form in modal
  await page.locator('input[placeholder*="workspace name"], input[placeholder*="Name"]').fill('test_workspace_1');
  await page.getByRole('button', { name: /create|save/i }).click();
  
  await page.waitForTimeout(1000);

  // Verify the workspace was created in the database
  const user = await selectUserByEmail(pool, test_email);
  expect(user.email).toBe(test_email);
  
  const workspaces = await selectWorkspacesByUserId(pool, user.id);
  const testWorkspace = workspaces.find(w => w.name === 'test_workspace_1');
  expect(testWorkspace).toBeDefined();
  expect(testWorkspace!.name).toBe('test_workspace_1');

  // Test 1: Verify 1 default session is automatically created in new workspace
  // Wait a bit for the default session to be created
  await page.waitForTimeout(1500);
  
  const sessionCount = await countSessionsInWorkspace(pool, testWorkspace!.id);
  expect(sessionCount).toBe(1);
  
  // Check that one session is displayed in the session list
  const sessionItems = page.locator('.relative.flex.items-center.gap-2.p-2.break-all.border.rounded-sm.cursor-pointer');
  await expect(sessionItems).toHaveCount(1);

  // Test 2: Add a new session with title 'first session in workspace test_workspace_1'
  // Click the "New Chat" button to add another session
  await page.getByRole('button', { name: /new|add/i }).first().click();
  await page.waitForTimeout(1000);

  // Edit the new session title
  await page.getByTestId('edit_session_topic').click();
  await page.getByTestId('edit_session_topic_input').locator('input').fill('first session in workspace test_workspace_1');
  await page.getByTestId('save_session_topic').click();
  
  await page.waitForTimeout(1000);

  // Verify the new session was created and is in the correct workspace
  const sessions = await selectChatSessionByUserId(pool, user.id);
  const testSession = sessions.find((s: any) => s.topic === 'first session in workspace test_workspace_1');
  expect(testSession).toBeDefined();
  expect(testSession!.topic).toBe('first session in workspace test_workspace_1');
  expect(testSession!.workspace_id).toBe(testWorkspace!.id);

  // Test 3: Verify now two sessions are displayed in the workspace (default + new one)
  const updatedSessionCount = await countSessionsInWorkspace(pool, testWorkspace!.id);
  expect(updatedSessionCount).toBe(2);
  
  // Check that exactly two sessions are visible in the UI
  const updatedSessionItems = page.locator('.relative.flex.items-center.gap-2.p-2.break-all.border.rounded-sm.cursor-pointer');
  await expect(updatedSessionItems).toHaveCount(2);
  
  // Verify the new session title is displayed correctly
  await expect(page.locator('text=first session in workspace test_workspace_1')).toBeVisible();

  // Test 4: Verify that page refresh doesn't change the route/workspace
  // Get the current URL before refresh
  const urlBeforeRefresh = page.url();
  console.log('URL before refresh:', urlBeforeRefresh);
  
  // Perform page refresh
  await page.reload();
  await page.waitForTimeout(2000); // Wait for page to fully load
  
  // Get the URL after refresh
  const urlAfterRefresh = page.url();
  console.log('URL after refresh:', urlAfterRefresh);
  
  // Verify the URL hasn't changed (should stay in the same workspace)
  expect(urlAfterRefresh).toBe(urlBeforeRefresh);
  
  // Verify we're still in the correct workspace by checking the workspace name is displayed
  await expect(page.locator('text=test_workspace_1')).toBeVisible();
  
  // Verify both sessions are still visible after refresh
  const sessionsAfterRefresh = page.locator('.relative.flex.items-center.gap-2.p-2.break-all.border.rounded-sm.cursor-pointer');
  await expect(sessionsAfterRefresh).toHaveCount(2);
  
  // Verify the custom session title is still visible
  await expect(page.locator('text=first session in workspace test_workspace_1')).toBeVisible();
});