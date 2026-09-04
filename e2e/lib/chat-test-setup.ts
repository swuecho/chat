import type { Page } from '@playwright/test'
import { Pool } from 'pg'
import { AuthHelpers, InputHelpers, MessageHelpers } from './message-helpers'
import { selectUserByEmail } from './db/user'
import { db_config } from './db/config'

const DEFAULT_PASSWORD = '@ThisIsATestPass5'

export async function setupDebugChatSession(page: Page, email: string) {
  const authHelpers = new AuthHelpers(page)
  const inputHelpers = new InputHelpers(page)
  const messageHelpers = new MessageHelpers(page)

  await page.goto('/')
  await authHelpers.signupAndWaitForAuth(email, DEFAULT_PASSWORD)
  await page.locator('a').filter({ hasText: 'New Chat' }).click()
  const pool = new Pool(db_config)
  try {
    const user = await selectUserByEmail(pool, email)
    await pool.query('UPDATE chat_session SET debug = true WHERE user_id = $1', [user.id])
  }
  finally {
    await pool.end()
  }

  return { inputHelpers, messageHelpers }
}

export async function sendMessageAndWaitAssistantCount(
  inputHelpers: InputHelpers,
  messageHelpers: MessageHelpers,
  text: string,
  assistantCount: number
) {
  await inputHelpers.sendMessage(text)
  await messageHelpers.waitForAssistantMessageCount(assistantCount)
  await inputHelpers.waitForComposerReady()
}
