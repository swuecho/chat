/**
 * Utility functions for generating and handling workspace-aware URLs
 */

// Get base URL for the application
function getBaseUrl(): string {
  return `${window.location.protocol}//${window.location.host}`
}

// Generate shareable URL for a session within a workspace
export function generateSessionUrl(sessionUuid: string, workspaceUuid?: string): string {
  const baseUrl = getBaseUrl()
  
  if (workspaceUuid) {
    return `${baseUrl}/#/workspace/${workspaceUuid}/chat/${sessionUuid}`
  }
  
  return `${baseUrl}/#/chat/${sessionUuid}`
}

// Generate shareable URL for a workspace
export function generateWorkspaceUrl(workspaceUuid: string): string {
  const baseUrl = getBaseUrl()
  return `${baseUrl}/#/workspace/${workspaceUuid}/chat`
}

// Extract workspace and session UUIDs from a URL
export function parseWorkspaceUrl(url: string): { workspaceUuid?: string; sessionUuid?: string } {
  try {
    const urlObj = new URL(url)
    const hash = urlObj.hash.substring(1) // Remove the # character
    
    // Match patterns: /workspace/:workspaceUuid/chat/:sessionUuid? or /chat/:sessionUuid?
    const workspaceMatch = hash.match(/^\/workspace\/([^\/]+)\/chat\/?([^\/]+)?/)
    const chatMatch = hash.match(/^\/chat\/?([^\/]+)?/)
    
    if (workspaceMatch) {
      return {
        workspaceUuid: workspaceMatch[1],
        sessionUuid: workspaceMatch[2]
      }
    }
    
    if (chatMatch) {
      return {
        sessionUuid: chatMatch[1]
      }
    }
    
    return {}
  } catch (error) {
    console.error('Error parsing workspace URL:', error)
    return {}
  }
}

// Check if a URL is a valid workspace URL
export function isValidWorkspaceUrl(url: string): boolean {
  const parsed = parseWorkspaceUrl(url)
  return parsed.workspaceUuid !== undefined || parsed.sessionUuid !== undefined
}

// Copy URL to clipboard with error handling
export async function copyUrlToClipboard(url: string): Promise<boolean> {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(url)
      return true
    } else {
      // Fallback for older browsers or non-HTTPS
      const textArea = document.createElement('textarea')
      textArea.value = url
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()
      
      const success = document.execCommand('copy')
      document.body.removeChild(textArea)
      return success
    }
  } catch (error) {
    console.error('Failed to copy URL to clipboard:', error)
    return false
  }
}

// Generate QR code data URL for sharing (requires qr-code library)
export function generateQRCodeUrl(url: string): string {
  // This would require a QR code library like 'qrcode'
  // For now, return a placeholder
  return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(url)}`
}

// Validate workspace UUID format
export function isValidWorkspaceUuid(uuid: string): boolean {
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i
  return uuidRegex.test(uuid)
}

// Validate session UUID format
export function isValidSessionUuid(uuid: string): boolean {
  return isValidWorkspaceUuid(uuid) // Same format
}

// Create a URL-safe workspace name for potential future slug-based URLs
export function createWorkspaceSlug(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^\w\s-]/g, '') // Remove special characters
    .replace(/[\s_-]+/g, '-') // Replace spaces and underscores with hyphens
    .replace(/^-+|-+$/g, '') // Remove leading/trailing hyphens
}

// Social sharing URLs
export const socialShareUrls = {
  twitter: (url: string, text: string = 'Check out this chat workspace') => 
    `https://twitter.com/intent/tweet?url=${encodeURIComponent(url)}&text=${encodeURIComponent(text)}`,
    
  facebook: (url: string) => 
    `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(url)}`,
    
  linkedin: (url: string, title: string = 'Chat Workspace') => 
    `https://www.linkedin.com/sharing/share-offsite/?url=${encodeURIComponent(url)}&title=${encodeURIComponent(title)}`,
    
  email: (url: string, subject: string = 'Chat Workspace', body: string = 'Check out this workspace:') => 
    `mailto:?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}%20${encodeURIComponent(url)}`
}