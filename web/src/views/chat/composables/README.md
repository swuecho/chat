# Chat Composables

This directory contains refactored composables that break down the large Conversation.vue component into smaller, more manageable and reusable pieces.

## Composables Overview

### Core Functionality

#### `useStreamHandling.ts`
Handles all streaming-related functionality for chat responses.

**Features:**
- Stream progress handling with proper error management
- Type-safe stream chunk processing
- Centralized error formatting with i18n support
- Robust error handling for various response statuses

**Key Functions:**
- `handleStreamProgress()` - Processes incoming stream data
- `processStreamChunk()` - Parses and validates stream chunks
- `streamChatResponse()` - Manages chat streaming lifecycle
- `streamRegenerateResponse()` - Handles message regeneration

#### `useConversationFlow.ts`
Manages the main conversation flow and user interactions.

**Features:**
- Input validation with comprehensive error handling
- Proper message state management
- Integrated error handling and user feedback
- Type-safe message structures

**Key Functions:**
- `onConversationStream()` - Main conversation handler
- `validateConversationInput()` - Input validation with user feedback
- `addUserMessage()` - Adds user messages to chat
- `initializeChatResponse()` - Sets up response placeholders

#### `useRegenerate.ts`
Handles message regeneration functionality.

**Features:**
- Smart regeneration context handling
- Proper cleanup of existing messages
- Error handling for regeneration failures
- Support for both user and AI message regeneration

**Key Functions:**
- `onRegenerate()` - Main regeneration handler
- `prepareRegenerateContext()` - Context setup for regeneration
- `handleUserMessageRegenerate()` - User message regeneration logic

#### `useSearchAndPrompts.ts`
Manages search functionality and prompt templates.

**Features:**
- Debounced search for better performance
- Memoized search results to prevent unnecessary computations
- Type-safe search options and filtering
- Support for both session and prompt searching

**Key Functions:**
- `searchOptions` - Computed search results with performance optimizations
- `renderOption()` - Renders search option labels
- `handleSelectAutoComplete()` - Handles search selection

#### `useChatActions.ts`
Contains various chat-related actions and utilities.

**Features:**
- Snapshot and bot creation functionality
- File upload handling
- Gallery and modal management
- VFS (Virtual File System) integration

**Key Functions:**
- `handleSnapshot()` - Creates chat snapshots
- `handleCreateBot()` - Bot creation functionality
- `handleVFSFileUploaded()` - File upload handling
- `toggleArtifactGallery()` - UI state management

### Utility Composables

#### `useErrorHandling.ts`
Centralized error management system.

**Features:**
- Comprehensive error logging and tracking
- User-friendly error notifications
- API error handling with proper HTTP status mapping
- Retry mechanism with exponential backoff
- Error history management

**Key Functions:**
- `handleApiError()` - Handles API errors with proper classification
- `logError()` - Logs errors with context and timestamp
- `retryOperation()` - Retry mechanism for failed operations
- `showErrorNotification()` - User notifications

#### `useValidation.ts`
Input validation system with reusable rules.

**Features:**
- Comprehensive validation rules (required, length, email, URL, etc.)
- Form field validation with reactive state
- Custom validation rule support
- Specific validators for chat messages, UUIDs, and file uploads

**Key Functions:**
- `validateChatMessage()` - Chat message validation
- `validateSessionUuid()` - UUID format validation
- `validateFileUpload()` - File validation with size and type checks
- `useField()` - Reactive form field validation

#### `usePerformanceOptimizations.ts`
Performance optimization utilities.

**Features:**
- Debouncing for high-frequency inputs
- Memoization for expensive computations
- Virtual scrolling for large lists
- Throttling for event handlers

**Key Functions:**
- `useDebounce()` - Debounces reactive values
- `useMemoized()` - Memoizes expensive computations
- `useVirtualList()` - Virtual scrolling implementation
- `useThrottle()` - Throttles function calls

## Usage Examples

### Basic Usage in Components

```typescript
// In a Vue component
import { useConversationFlow } from './composables/useConversationFlow'
import { useErrorHandling } from './composables/useErrorHandling'

export default {
  setup() {
    const sessionUuid = 'your-session-uuid'
    const conversationFlow = useConversationFlow(sessionUuid)
    const { showErrorNotification } = useErrorHandling()

    const handleSubmit = async (message: string) => {
      try {
        await conversationFlow.onConversationStream(message, dataSources.value)
      } catch (error) {
        showErrorNotification('Failed to send message')
      }
    }

    return {
      handleSubmit,
      loading: conversationFlow.loading
    }
  }
}
```

### Validation Example

```typescript
import { useValidation } from './composables/useValidation'

const { useField, rules } = useValidation()

const messageField = useField('', [
  rules.required('Message is required'),
  rules.maxLength(1000, 'Message too long')
])

// Use in template
// v-model="messageField.value.value"
// :error="messageField.showErrors.value"
```

### Performance Optimization Example

```typescript
import { useDebounce, useMemoized } from './composables/usePerformanceOptimizations'

const searchTerm = ref('')
const debouncedSearch = useDebounce(searchTerm, 300)

const expensiveComputation = useMemoized(
  (data) => computeHeavyCalculation(data),
  () => someReactiveData.value
)
```

## Benefits of This Refactoring

### 1. **Separation of Concerns**
Each composable has a single responsibility, making the code easier to understand and maintain.

### 2. **Reusability**
Composables can be reused across different components, reducing code duplication.

### 3. **Testability**
Individual composables can be tested in isolation, improving test coverage and reliability.

### 4. **Type Safety**
Comprehensive TypeScript interfaces and types provide better development experience and catch errors at compile time.

### 5. **Performance**
Optimizations like debouncing, memoization, and proper error handling improve the overall user experience.

### 6. **Error Handling**
Centralized error management provides consistent error handling across the application.

### 7. **Maintainability**
Smaller, focused files are easier to maintain and update.

## File Size Reduction

The main `Conversation.vue` file was reduced from **738 lines** to **293 lines** (60% reduction) while gaining:
- Better error handling
- Performance optimizations
- Type safety
- Improved reusability
- Enhanced maintainability

## Future Improvements

1. **Unit Tests**: Add comprehensive unit tests for each composable
2. **Documentation**: Add JSDoc comments to all public functions
3. **Logging**: Integrate with application logging system
4. **Metrics**: Add performance metrics collection
5. **Accessibility**: Enhance accessibility features
6. **Internationalization**: Improve i18n support throughout composables