import { Page } from '@playwright/test';

/**
 * Helper functions for interacting with footer buttons in the chat interface
 */

/**
 * Gets the clear conversation button from the footer
 * @param page Playwright page object
 * @returns Locator for the clear conversation button
 */
export async function getClearConversationButton(page: Page) {
  // Use test ID for reliable button selection instead of fragile position-based selection
  return page.getByTestId('clear-conversation-button');
}

/**
 * Gets the snapshot button from the footer (desktop only)
 * @param page Playwright page object
 * @returns Locator for the snapshot button
 */
export async function getSnapshotButton(page: Page) {
  // Snapshot button is hidden on mobile, so position may vary
  return page.getByTestId('snpashot-button');
}

/**
 * Gets the VFS upload button from the footer (desktop only)
 * @param page Playwright page object
 * @returns Locator for the VFS upload button
 */
export async function getVFSUploadButton(page: Page) {
  // VFS upload button is now hidden on mobile
  return page.getByRole('button').filter({ hasText: 'Upload files to VFS' }).first();
}

/**
 * Gets the artifact gallery toggle button from the footer (desktop only)
 * @param page Playwright page object
 * @returns Locator for the artifact gallery button
 */
export async function getArtifactGalleryButton(page: Page) {
  // Artifact gallery button is hidden on mobile
  return page.getByRole('button').filter({ hasText: /Hide Gallery|Show Gallery/ }).first();
}

/**
 * Footer button positions (0-indexed) for different screen sizes
 * Note: These positions may change based on which buttons are visible
 */
export const FOOTER_BUTTON_POSITIONS = {
  DESKTOP: {
    CLEAR_CONVERSATION: 0,
    SNAPSHOT: 1, // May not be visible on mobile
    VFS_UPLOAD: 2, // Hidden on mobile
    ARTIFACT_GALLERY: 3, // Hidden on mobile
  },
  MOBILE: {
    CLEAR_CONVERSATION: 0,
    // Other buttons are hidden on mobile
  }
} as const;