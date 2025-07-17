import { Page, Locator } from '@playwright/test';

/**
 * Helper functions for interacting with chat messages in E2E tests
 * These functions are more robust to layout changes than direct nth-child selectors
 */

export class MessageHelpers {
  private page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * Get all chat messages in the current session
   */
  async getAllMessages(): Promise<Locator[]> {
    await this.page.waitForSelector('.chat-message', { timeout: 5000 });
    return await this.page.locator('.chat-message').all();
  }

  /**
   * Get a specific message by index (0-based)
   */
  async getMessageByIndex(index: number): Promise<Locator> {
    const messages = await this.getAllMessages();
    if (index >= messages.length) {
      throw new Error(`Message index ${index} not found. Only ${messages.length} messages exist.`);
    }
    return messages[index];
  }

  /**
   * Get the text content of a message by index
   */
  async getMessageText(index: number): Promise<string> {
    const message = await this.getMessageByIndex(index);
    const textElement = message.locator('.message-text');
    await textElement.waitFor({ timeout: 5000 });
    return await textElement.innerText();
  }

  /**
   * Get the regenerate button for a message by index
   */
  async getRegenerateButton(index: number): Promise<Locator> {
    const message = await this.getMessageByIndex(index);
    return message.locator('.chat-message-regenerate');
  }

  /**
   * Click the regenerate button for a message
   */
  async clickRegenerate(index: number): Promise<void> {
    const button = await this.getRegenerateButton(index);
    await button.waitFor({ state: 'visible', timeout: 5000 });
    await button.click();
  }

  /**
   * Wait for a message to appear and contain specific text
   */
  async waitForMessageWithText(text: string, timeout: number = 10000): Promise<void> {
    await this.page.waitForFunction(
      (searchText) => {
        const messages = document.querySelectorAll('.message-text');
        return Array.from(messages).some(msg => msg.textContent?.includes(searchText));
      },
      text,
      { timeout }
    );
  }

  /**
   * Get the last message text
   */
  async getLastMessageText(): Promise<string> {
    const messages = await this.getAllMessages();
    const lastMessage = messages[messages.length - 1];
    const textElement = lastMessage.locator('.message-text');
    return await textElement.innerText();
  }

  /**
   * Wait for a specific number of messages to be present
   */
  async waitForMessageCount(count: number, timeout: number = 10000): Promise<void> {
    await this.page.waitForFunction(
      (expectedCount) => document.querySelectorAll('.chat-message').length >= expectedCount,
      count,
      { timeout }
    );
  }

  /**
   * Check if a regenerate button is visible for a message
   */
  async isRegenerateButtonVisible(index: number): Promise<boolean> {
    try {
      const button = await this.getRegenerateButton(index);
      return await button.isVisible();
    } catch (error) {
      return false;
    }
  }

  /**
   * Get message by content (useful for finding specific responses)
   */
  async getMessageByContent(partialText: string): Promise<Locator | null> {
    const messages = await this.getAllMessages();
    
    for (const message of messages) {
      const textElement = message.locator('.message-text');
      try {
        const text = await textElement.innerText();
        if (text.includes(partialText)) {
          return message;
        }
      } catch (error) {
        // Continue if text element not found in this message
        continue;
      }
    }
    
    return null;
  }

  /**
   * Get the index of a message by its content
   */
  async getMessageIndexByContent(partialText: string): Promise<number> {
    const messages = await this.getAllMessages();
    
    for (let i = 0; i < messages.length; i++) {
      const textElement = messages[i].locator('.message-text');
      try {
        const text = await textElement.innerText();
        if (text.includes(partialText)) {
          return i;
        }
      } catch (error) {
        // Continue if text element not found in this message
        continue;
      }
    }
    
    throw new Error(`Message containing "${partialText}" not found`);
  }
}

/**
 * Authentication helpers for signup/login flows
 */
export class AuthHelpers {
  private page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * Complete signup process and wait for authentication to be ready
   */
  async signupAndWaitForAuth(email: string, password: string): Promise<void> {
    await this.page.getByTitle('signuptab').click();
    await this.page.getByTestId('signup_email').click();
    await this.page.getByTestId('signup_email').locator('input').fill(email);
    await this.page.getByTestId('signup_password').locator('input').click();
    await this.page.getByTestId('signup_password').locator('input').fill(password);
    await this.page.getByTestId('repwd').locator('input').click();
    await this.page.getByTestId('repwd').locator('input').fill(password);
    await this.page.getByTestId('signup').click();
    
    // Wait for the page reload after successful signup
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(3000);
    
    // Wait for the permission modal to disappear before proceeding
    try {
      await this.page.waitForSelector('.n-modal-mask', { state: 'detached', timeout: 10000 });
    } catch (error) {
      // If modal is not found, it might already be gone, which is fine
      console.log('Modal mask not found or already disappeared');
    }
    
    // Additional wait to ensure the interface is ready
    await this.page.waitForTimeout(1000);
  }

  /**
   * Wait for the interface to be ready for interaction after authentication
   */
  async waitForInterfaceReady(): Promise<void> {
    // Wait for the message textarea to be available and clickable
    await this.page.waitForSelector('#message_textarea textarea', { timeout: 10000 });
    await this.page.waitForTimeout(500);
  }
}

/**
 * Input helpers for sending messages
 */
export class InputHelpers {
  private page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * Get the message input textarea
   */
  async getInputArea(): Promise<Locator> {
    return this.page.locator('#message_textarea textarea');
  }

  /**
   * Send a message and wait for response
   */
  async sendMessage(text: string, waitForResponse: boolean = true): Promise<void> {
    const input = await this.getInputArea();
    await input.click();
    await input.fill(text);
    await input.press('Enter');
    
    if (waitForResponse) {
      // Wait for the message to appear in the chat
      const messageHelpers = new MessageHelpers(this.page);
      await messageHelpers.waitForMessageWithText(text);
      // Wait a bit more for the response to be generated
      await this.page.waitForTimeout(1000);
    }
  }
}