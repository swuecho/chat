import type { Page } from '@playwright/test'
import { AuthHelpers, InputHelpers, MessageHelpers } from './message-helpers'

const DEFAULT_PASSWORD = '@ThisIsATestPass5'

export async function setupDebugChatSession(page: Page, email: string) {
  const authHelpers = new AuthHelpers(page)
  const inputHelpers = new InputHelpers(page)
  const messageHelpers = new MessageHelpers(page)

  await page.goto('/')
  await authHelpers.signupAndWaitForAuth(email, DEFAULT_PASSWORD)
  await page.locator('a').filter({ hasText: 'New Chat' }).click()

  await page.getByTestId('chat-settings-button').click()
  await page.getByTestId('collapse-advanced').click()
  await page.waitForTimeout(300)
  await page.getByTestId('debug_mode').click()
  await page.keyboard.press('Escape')

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
