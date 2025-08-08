<template>
  <div class="enhanced-notification" :class="notificationClass">
    <div class="notification-banner" :class="bannerClass">
      <div class="banner-content">
        <div class="banner-icon">
          <component :is="iconComponent" :size="16" />
        </div>
        <div class="banner-title">{{ title }}</div>
        <div class="banner-actions" v-if="closable">
          <n-button 
            quaternary 
            circle 
            size="tiny" 
            @click="handleClose"
            class="close-button"
          >
            <template #icon>
              <n-icon><CloseIcon /></n-icon>
            </template>
          </n-button>
        </div>
      </div>
    </div>
    
    <div class="notification-content" v-if="content || $slots.default">
      <div class="content-text" v-if="content">{{ content }}</div>
      <slot v-else />
      
      <div class="content-actions" v-if="action">
        <n-button 
          :type="actionButtonType" 
          size="small" 
          @click="handleAction"
          class="action-button"
        >
          {{ action.text }}
        </n-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NButton, NIcon } from 'naive-ui'
import { 
  CheckmarkCircle as SuccessIcon,
  CloseCircle as ErrorIcon,
  Warning as WarningIcon,
  InformationCircle as InfoIcon,
  Close as CloseIcon
} from '@vicons/ionicons5'

interface NotificationAction {
  text: string
  onClick: () => void
}

interface Props {
  type?: 'success' | 'error' | 'warning' | 'info'
  title: string
  content?: string
  closable?: boolean
  action?: NotificationAction
}

interface Emits {
  (e: 'close'): void
}

const props = withDefaults(defineProps<Props>(), {
  type: 'info',
  closable: true
})

const emit = defineEmits<Emits>()

const iconComponent = computed(() => {
  const icons = {
    success: SuccessIcon,
    error: ErrorIcon,
    warning: WarningIcon,
    info: InfoIcon
  }
  return icons[props.type]
})

const notificationClass = computed(() => `notification-${props.type}`)

const bannerClass = computed(() => `banner-${props.type}`)

const actionButtonType = computed(() => {
  const buttonTypes = {
    success: 'success',
    error: 'error', 
    warning: 'warning',
    info: 'primary'
  }
  return buttonTypes[props.type]
})

const titles = {
  success: 'Success',
  error: 'Error',
  warning: 'Warning',
  info: 'Information'
}

const title = computed(() => props.title || titles[props.type])

function handleClose() {
  emit('close')
}

function handleAction() {
  if (props.action) {
    props.action.onClick()
  }
}
</script>

<style scoped>
.enhanced-notification {
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  background: white;
  min-width: 320px;
  max-width: 480px;
  transition: all 0.2s ease;
}

.enhanced-notification:hover {
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.15);
}

.notification-banner {
  padding: 12px 16px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.banner-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.banner-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.banner-title {
  flex: 1;
  font-weight: 600;
  font-size: 14px;
  color: white;
}

.banner-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.close-button {
  color: rgba(255, 255, 255, 0.8) !important;
}

.close-button:hover {
  color: white !important;
  background: rgba(255, 255, 255, 0.1) !important;
}

.notification-content {
  padding: 16px;
  background: white;
}

.content-text {
  color: #374151;
  font-size: 14px;
  line-height: 1.5;
  margin-bottom: 12px;
}

.content-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.action-button {
  min-width: 64px;
}

/* Success styling */
.notification-success .notification-banner {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}

.banner-success .banner-icon {
  color: white;
}

/* Error styling */
.notification-error .notification-banner {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
}

.banner-error .banner-icon {
  color: white;
}

/* Warning styling */
.notification-warning .notification-banner {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
}

.banner-warning .banner-icon {
  color: white;
}

/* Info styling */
.notification-info .notification-banner {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
}

.banner-info .banner-icon {
  color: white;
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
  .enhanced-notification {
    background: #1f2937;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
  }

  .notification-content {
    background: #1f2937;
  }

  .content-text {
    color: #d1d5db;
  }

  .notification-banner {
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }
}
</style>