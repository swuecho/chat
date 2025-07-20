import { ref, computed, watch } from 'vue'
import { t } from '@/locales'

interface ValidationRule {
  validator: (value: any) => boolean
  message: string
}

interface ValidationResult {
  isValid: boolean
  errors: string[]
}

export function useValidation() {
  
  const createValidator = (rules: ValidationRule[]) => {
    return (value: any): ValidationResult => {
      const errors: string[] = []
      
      for (const rule of rules) {
        if (!rule.validator(value)) {
          errors.push(rule.message)
        }
      }
      
      return {
        isValid: errors.length === 0,
        errors
      }
    }
  }

  // Common validation rules
  const rules = {
    required: (message?: string): ValidationRule => ({
      validator: (value: any) => {
        if (typeof value === 'string') return value.trim().length > 0
        return value != null && value !== ''
      },
      message: message || t('validation.required') || 'This field is required'
    }),

    minLength: (min: number, message?: string): ValidationRule => ({
      validator: (value: string) => !value || value.length >= min,
      message: message || t('validation.minLength', { min }) || `Minimum length is ${min} characters`
    }),

    maxLength: (max: number, message?: string): ValidationRule => ({
      validator: (value: string) => !value || value.length <= max,
      message: message || t('validation.maxLength', { max }) || `Maximum length is ${max} characters`
    }),

    email: (message?: string): ValidationRule => ({
      validator: (value: string) => {
        if (!value) return true
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
        return emailRegex.test(value)
      },
      message: message || t('validation.email') || 'Please enter a valid email address'
    }),

    url: (message?: string): ValidationRule => ({
      validator: (value: string) => {
        if (!value) return true
        try {
          new URL(value)
          return true
        } catch {
          return false
        }
      },
      message: message || t('validation.url') || 'Please enter a valid URL'
    }),

    pattern: (regex: RegExp, message: string): ValidationRule => ({
      validator: (value: string) => !value || regex.test(value),
      message
    }),

    custom: (validator: (value: any) => boolean, message: string): ValidationRule => ({
      validator,
      message
    })
  }

  // Form field validation
  function useField<T>(
    initialValue: T,
    validationRules: ValidationRule[] = []
  ) {
    const value = ref<T>(initialValue)
    const touched = ref(false)
    const errors = ref<string[]>([])
    
    const validator = createValidator(validationRules)
    
    const isValid = computed(() => errors.value.length === 0)
    const hasErrors = computed(() => errors.value.length > 0)
    const showErrors = computed(() => touched.value && hasErrors.value)

    const validate = () => {
      const result = validator(value.value)
      errors.value = result.errors
      return result.isValid
    }

    const touch = () => {
      touched.value = true
    }

    const reset = () => {
      value.value = initialValue
      touched.value = false
      errors.value = []
    }

    // Validate on value change
    watch(value, () => {
      if (touched.value) {
        validate()
      }
    })

    return {
      value,
      errors: computed(() => errors.value),
      isValid,
      hasErrors,
      showErrors,
      validate,
      touch,
      reset
    }
  }

  // Chat message validation
  function validateChatMessage(message: string): ValidationResult {
    const messageRules = [
      rules.required('Message cannot be empty'),
      rules.maxLength(10000, 'Message is too long (max 10,000 characters)')
    ]
    
    return createValidator(messageRules)(message)
  }

  // Session UUID validation
  function validateSessionUuid(uuid: string): ValidationResult {
    const uuidRules = [
      rules.required('Session UUID is required'),
      rules.pattern(
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
        'Invalid UUID format'
      )
    ]
    
    return createValidator(uuidRules)(uuid)
  }

  // File upload validation
  function validateFileUpload(
    file: File,
    maxSize: number = 10 * 1024 * 1024, // 10MB
    allowedTypes: string[] = []
  ): ValidationResult {
    const fileRules = [
      rules.custom(
        () => file.size <= maxSize,
        `File size must be less than ${Math.round(maxSize / 1024 / 1024)}MB`
      )
    ]

    if (allowedTypes.length > 0) {
      fileRules.push(
        rules.custom(
          () => allowedTypes.some(type => file.type.includes(type)),
          `File type not allowed. Allowed types: ${allowedTypes.join(', ')}`
        )
      )
    }

    return createValidator(fileRules)(file)
  }

  return {
    rules,
    createValidator,
    useField,
    validateChatMessage,
    validateSessionUuid,
    validateFileUpload
  }
}