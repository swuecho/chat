<template>
  <div class="notification-demo">
    <n-card title="Notification System Demo">
      <n-space vertical>
        <div class="demo-section">
          <h3>Basic Notifications</h3>
          <n-space>
            <n-button @click="showSuccess" type="success">Success</n-button>
            <n-button @click="showError" type="error">Error</n-button>
            <n-button @click="showWarning" type="warning">Warning</n-button>
            <n-button @click="showInfo" type="info">Info</n-button>
          </n-space>
        </div>

        <div class="demo-section">
          <h3>Persistent Notifications</h3>
          <n-space>
            <n-button @click="showPersistentError" type="error">Persistent Error</n-button>
            <n-button @click="showPersistentWarning" type="warning">Persistent Warning</n-button>
          </n-space>
        </div>

        <div class="demo-section">
          <h3>Action Notifications</h3>
          <n-space>
            <n-button @click="showActionError" type="error">Error with Action</n-button>
            <n-button @click="showActionSuccess" type="success">Success with Action</n-button>
          </n-space>
        </div>

        <div class="demo-section">
          <h3>Batch Notifications</h3>
          <n-space>
            <n-button @click="showBatch" type="primary">Show Batch (5)</n-button>
            <n-button @click="clearAll" type="default">Clear All</n-button>
          </n-space>
        </div>

        <div class="demo-section">
          <h3>Network Status</h3>
          <n-space>
            <n-button @click="simulateOffline" type="warning">Simulate Offline</n-button>
            <n-button @click="simulateOnline" type="success">Simulate Online</n-button>
          </n-space>
        </div>

        <div class="demo-section">
          <h3>Error Simulation</h3>
          <n-space>
            <n-button @click="simulateNetworkError" type="error">Network Error</n-button>
            <n-button @click="simulateTimeout" type="warning">Timeout Error</n-button>
            <n-button @click="simulateAuthError" type="error">Auth Error</n-button>
            <n-button @click="simulateServerError" type="error">Server Error</n-button>
          </n-space>
        </div>

        <div class="demo-section">
          <h3>Notification Stats</h3>
          <n-space vertical>
            <div>Queued: {{ stats.queued }}</div>
            <div>Active: {{ stats.active }}</div>
            <div>Max Concurrent: {{ stats.maxConcurrent }}</div>
          </n-space>
        </div>
      </n-space>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { useNotification } from '@/utils/notificationManager'
import { useErrorHandling } from '@/views/chat/composables/useErrorHandling'
import { ref } from 'vue'

const notification = useNotification()
const { handleApiError } = useErrorHandling()

const stats = ref(notification.stats)

// Basic notifications
function showSuccess() {
  notification.success('Operation completed successfully!')
}

function showError() {
  notification.error('Something went wrong!')
}

function showWarning() {
  notification.warning('Please be careful with this action.')
}

function showInfo() {
  notification.info('Here is some information for you.')
}

// Persistent notifications
function showPersistentError() {
  notification.persistent('This is a persistent error message. It will stay until you close it.', 'error', {
    text: 'Retry',
    onClick: () => notification.success('Retried successfully!')
  })
}

function showPersistentWarning() {
  notification.persistent('This is a persistent warning message.', 'warning', {
    text: 'Dismiss',
    onClick: () => notification.info('Warning dismissed')
  })
}

// Action notifications
function showActionError() {
  notification.error('Failed to save changes', {
    duration: 8000,
    action: {
      text: 'Retry',
      onClick: () => notification.success('Changes saved successfully!')
    }
  })
}

function showActionSuccess() {
  notification.success('File uploaded successfully!', {
    action: {
      text: 'View File',
      onClick: () => notification.info('Opening file...')
    }
  })
}

// Batch notifications
function showBatch() {
  const messages = [
    { type: 'success', text: 'Item 1 created' },
    { type: 'error', text: 'Item 2 failed' },
    { type: 'warning', text: 'Item 3 has warnings' },
    { type: 'info', text: 'Item 4 processed' },
    { type: 'success', text: 'Item 5 completed' }
  ]

  messages.forEach((msg, index) => {
    setTimeout(() => {
      notification[msg.type](msg.text)
    }, index * 500)
  })
}

// Clear all notifications
function clearAll() {
  notification.clear()
}

// Network status simulation
function simulateOffline() {
  notification.persistent('You are offline. Please check your internet connection.', 'error', {
    text: 'Retry',
    onClick: () => notification.success('Connection restored!')
  })
}

function simulateOnline() {
  notification.success('You are back online!')
}

// Error simulation
function simulateNetworkError() {
  const error = new Error('Network Error')
  error.name = 'Network Error'
  error.message = 'Failed to connect to server'
  handleApiError(error, 'demo')
}

function simulateTimeout() {
  const error = new Error('Timeout Error')
  error.name = 'Timeout Error'
  error.message = 'Request timed out after 30 seconds'
  handleApiError(error, 'demo')
}

function simulateAuthError() {
  const error = new Error('Auth Error')
  error.name = 'Auth Error'
  error.message = 'Authentication failed'
  error.response = {
    status: 401,
    data: { message: 'Invalid credentials' }
  }
  handleApiError(error, 'demo')
}

function simulateServerError() {
  const error = new Error('Server Error')
  error.name = 'Server Error'
  error.message = 'Internal server error'
  error.response = {
    status: 500,
    data: { message: 'Something went wrong on our end' }
  }
  handleApiError(error, 'demo')
}
</script>

<style scoped>
.notification-demo {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}

.demo-section {
  margin-bottom: 24px;
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #f9fafb;
}

.demo-section h3 {
  margin-top: 0;
  margin-bottom: 12px;
  color: #374151;
  font-size: 16px;
  font-weight: 600;
}
</style>