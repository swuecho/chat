import { ref, computed, h } from 'vue'
import { useMessage } from 'naive-ui'
import EnhancedNotification from '@/components/common/EnhancedNotification.vue'

interface NotificationOptions {
  title?: string
  message: string
  type?: 'success' | 'error' | 'warning' | 'info'
  duration?: number
  action?: {
    text: string
    onClick: () => void
  }
  persistent?: boolean
  closable?: boolean
  enhanced?: boolean // New option to use enhanced notifications
}

interface QueuedNotification {
  id: string
  options: NotificationOptions
  timestamp: Date
}

class NotificationManager {
  private queue = ref<QueuedNotification[]>([])
  private activeNotifications = ref<Set<string>>(new Set())
  private messageInstance: any = null
  private maxConcurrent = 3
  private queueEnabled = true

  setMessageInstance(instance: any) {
    this.messageInstance = instance
  }

  private generateId(): string {
    return `notification_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  private canShowNotification(): boolean {
    return this.activeNotifications.value.size < this.maxConcurrent
  }

  private showNotification(notification: QueuedNotification) {
    if (!this.messageInstance) return

    const { id, options } = notification
    this.activeNotifications.value.add(id)

    const showFn = this.messageInstance[options.type || 'info']
    const notificationOptions: any = {
      duration: options.persistent ? 0 : (options.duration || 3000),
      closable: options.closable !== false,
      keepAliveOnHover: true,
      onLeave: () => {
        this.activeNotifications.value.delete(id)
        this.processQueue()
      }
    }

    if (options.action) {
      notificationOptions.action = options.action
    }

    try {
      // Use enhanced notification if requested
      if (options.enhanced) {
        const content = h(EnhancedNotification, {
          type: options.type || 'info',
          title: options.title,
          content: options.message,
          closable: options.closable !== false,
          action: options.action,
          onClose: () => {
            this.activeNotifications.value.delete(id)
            this.processQueue()
          }
        })
        
        showFn(content, {
          ...notificationOptions,
          closable: false // Let the component handle closing
        })
      } else {
        showFn(options.message, notificationOptions)
      }
    } catch (error) {
      console.error('Failed to show notification:', error)
      this.activeNotifications.value.delete(id)
      this.processQueue()
    }
  }

  private processQueue() {
    if (!this.queueEnabled || !this.canShowNotification()) return

    const nextNotification = this.queue.value.shift()
    if (nextNotification) {
      this.showNotification(nextNotification)
    }
  }

  show(options: NotificationOptions): string {
    const id = this.generateId()
    const notification: QueuedNotification = {
      id,
      options,
      timestamp: new Date()
    }

    if (this.canShowNotification()) {
      this.showNotification(notification)
    } else {
      this.queue.value.push(notification)
    }

    return id
  }

  success(message: string, options: Omit<NotificationOptions, 'message' | 'type'> = {}): string {
    return this.show({ message, type: 'success', ...options })
  }

  error(message: string, options: Omit<NotificationOptions, 'message' | 'type'> = {}): string {
    return this.show({ message, type: 'error', ...options })
  }

  warning(message: string, options: Omit<NotificationOptions, 'message' | 'type'> = {}): string {
    return this.show({ message, type: 'warning', ...options })
  }

  info(message: string, options: Omit<NotificationOptions, 'message' | 'type'> = {}): string {
    return this.show({ message, type: 'info', ...options })
  }

  // Enhanced notification methods with better visual hierarchy
  enhancedSuccess(title: string, message: string, options: Omit<NotificationOptions, 'message' | 'type' | 'title' | 'enhanced'> = {}): string {
    return this.show({ title, message, type: 'success', enhanced: true, ...options })
  }

  enhancedError(title: string, message: string, options: Omit<NotificationOptions, 'message' | 'type' | 'title' | 'enhanced'> = {}): string {
    return this.show({ title, message, type: 'error', enhanced: true, ...options })
  }

  enhancedWarning(title: string, message: string, options: Omit<NotificationOptions, 'message' | 'type' | 'title' | 'enhanced'> = {}): string {
    return this.show({ title, message, type: 'warning', enhanced: true, ...options })
  }

  enhancedInfo(title: string, message: string, options: Omit<NotificationOptions, 'message' | 'type' | 'title' | 'enhanced'> = {}): string {
    return this.show({ title, message, type: 'info', enhanced: true, ...options })
  }

  persistent(message: string, type: 'error' | 'warning' | 'info' = 'error', action?: { text: string; onClick: () => void }): string {
    return this.show({
      message,
      type,
      persistent: true,
      action
    })
  }

  remove(id: string): void {
    this.queue.value = this.queue.value.filter(n => n.id !== id)
    this.activeNotifications.value.delete(id)
  }

  clear(): void {
    this.queue.value = []
    this.activeNotifications.value.clear()
    if (this.messageInstance) {
      try {
        this.messageInstance.destroyAll()
      } catch (error) {
        console.warn('Failed to clear notifications:', error)
      }
    }
  }

  getStats() {
    return {
      queued: this.queue.value.length,
      active: this.activeNotifications.value.size,
      maxConcurrent: this.maxConcurrent
    }
  }

  setMaxConcurrent(max: number): void {
    this.maxConcurrent = max
    this.processQueue()
  }

  enableQueue(): void {
    this.queueEnabled = true
    this.processQueue()
  }

  disableQueue(): void {
    this.queueEnabled = false
  }
}

// Export singleton instance
export const notificationManager = new NotificationManager()

// Vue composable for easy usage in components
export function useNotification() {
  const message = useMessage()
  
  // Initialize message instance if not already set
  if (!notificationManager['messageInstance']) {
    notificationManager.setMessageInstance(message)
  }

  return {
    show: notificationManager.show.bind(notificationManager),
    success: notificationManager.success.bind(notificationManager),
    error: notificationManager.error.bind(notificationManager),
    warning: notificationManager.warning.bind(notificationManager),
    info: notificationManager.info.bind(notificationManager),
    enhancedSuccess: notificationManager.enhancedSuccess.bind(notificationManager),
    enhancedError: notificationManager.enhancedError.bind(notificationManager),
    enhancedWarning: notificationManager.enhancedWarning.bind(notificationManager),
    enhancedInfo: notificationManager.enhancedInfo.bind(notificationManager),
    persistent: notificationManager.persistent.bind(notificationManager),
    clear: notificationManager.clear.bind(notificationManager),
    stats: computed(() => notificationManager.getStats())
  }
}

// Global notification functions for non-Vue contexts
export function showNotification(options: NotificationOptions): string {
  return notificationManager.show(options)
}

export function showSuccessNotification(message: string, options?: Omit<NotificationOptions, 'message' | 'type'>): string {
  return notificationManager.success(message, options)
}

export function showErrorNotification(message: string, options?: Omit<NotificationOptions, 'message' | 'type'>): string {
  return notificationManager.error(message, options)
}

export function showWarningNotification(message: string, options?: Omit<NotificationOptions, 'message' | 'type'>): string {
  return notificationManager.warning(message, options)
}

export function showInfoNotification(message: string, options?: Omit<NotificationOptions, 'message' | 'type'>): string {
  return notificationManager.info(message, options)
}

export function showPersistentNotification(message: string, type: 'error' | 'warning' | 'info' = 'error', action?: { text: string; onClick: () => void }): string {
  return notificationManager.persistent(message, type, action)
}

// Enhanced notification functions with better visual hierarchy
export function showEnhancedSuccessNotification(title: string, message: string, options?: Omit<NotificationOptions, 'message' | 'type' | 'title' | 'enhanced'>): string {
  return notificationManager.enhancedSuccess(title, message, options)
}

export function showEnhancedErrorNotification(title: string, message: string, options?: Omit<NotificationOptions, 'message' | 'type' | 'title' | 'enhanced'>): string {
  return notificationManager.enhancedError(title, message, options)
}

export function showEnhancedWarningNotification(title: string, message: string, options?: Omit<NotificationOptions, 'message' | 'type' | 'title' | 'enhanced'>): string {
  return notificationManager.enhancedWarning(title, message, options)
}

export function showEnhancedInfoNotification(title: string, message: string, options?: Omit<NotificationOptions, 'message' | 'type' | 'title' | 'enhanced'>): string {
  return notificationManager.enhancedInfo(title, message, options)
}

export function clearAllNotifications(): void {
  notificationManager.clear()
}