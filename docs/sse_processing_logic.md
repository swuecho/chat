# Server-Sent Events (SSE) Processing Logic

This document explains how the chat application handles Server-Sent Events (SSE) streaming responses from the backend.

## Overview

The SSE processing logic is implemented in `useStreamHandling.ts` and handles real-time streaming of chat responses. It manages buffering, message parsing, and error handling for continuous data streams.

## SSE Protocol Basics

Server-Sent Events follow this format:
```
data: {"id": "chatcmpl-123", "choices": [{"delta": {"content": "Hello"}}]}

data: {"id": "chatcmpl-123", "choices": [{"delta": {"content": " world"}}]}

data: [DONE]

```

- Each message starts with `data: `
- Messages are separated by double newlines (`\n\n`)
- The stream ends with `data: [DONE]`

## Processing Flow

### 1. Stream Reading Setup

```javascript
const response = await fetch(streamingUrl, requestConfig)
const reader = response.body.getReader()
const decoder = new TextDecoder()
let buffer = ''
```

### 2. Chunk Processing Loop

```javascript
while (true) {
  const { done, value } = await reader.read()
  if (done) break
  
  const chunk = decoder.decode(value, { stream: true })
  buffer += chunk
  
  // Process complete messages
  processBuffer(buffer)
}
```

### 3. Buffer Management

The key challenge is handling partial messages that arrive across multiple chunks:

```javascript
// Input buffer: "data: {partial}\n\ndata: {complete}\n\nda"
const lines = buffer.split('\n\n')
buffer = lines.pop() || ''  // Keep incomplete part: "da"

// Process complete messages: ["data: {partial}", "data: {complete}"]
for (const line of lines) {
  if (line.trim()) {
    processMessage(line)
  }
}
```

### 4. Data Extraction

Each SSE message is processed by `extractStreamingData()`:

```javascript
function extractStreamingData(streamResponse: string): string {
  // Handle standard SSE format: "data: {...}"
  if (streamResponse.startsWith('data:')) {
    return streamResponse.slice(5).trim()
  }
  
  // Handle multiple segments (fallback)
  const lastDataPosition = streamResponse.lastIndexOf('\n\ndata:')
  if (lastDataPosition === -1) {
    return streamResponse.trim()
  }
  
  return streamResponse.slice(lastDataPosition + 8).trim()
}
```

### 5. Message Processing

Once data is extracted, it's parsed and processed:

```javascript
function processStreamChunk(chunk: string, responseIndex: number, sessionUuid: string) {
  const data = extractStreamingData(chunk)
  if (!data) return
  
  try {
    const parsedData = JSON.parse(data)
    
    // Validate structure
    if (!parsedData.choices?.[0]?.delta?.content || !parsedData.id) {
      console.warn('Invalid stream chunk structure')
      return
    }
    
    const content = parsedData.choices[0].delta.content
    const messageId = parsedData.id.replace('chatcmpl-', '')
    
    // Update chat with new content
    updateChat(sessionUuid, responseIndex, {
      uuid: messageId,
      text: content,
      // ... other properties
    })
  } catch (error) {
    console.error('Failed to parse stream chunk:', error)
  }
}
```

## Error Handling

### Stream Errors
- HTTP errors (non-2xx responses)
- Network connectivity issues
- Reader errors

### Parsing Errors
- Malformed JSON in SSE data
- Invalid message structure
- Missing required fields

### Recovery Strategies
- Graceful degradation on parse errors
- User notification for critical errors
- Automatic cleanup of incomplete messages

## Key Functions

### `streamChatResponse()`
Main streaming function for new chat messages:
- Sets up fetch request with streaming enabled
- Manages ReadableStream reader
- Handles progressive response processing
- Calls progress callbacks for real-time updates

### `streamRegenerateResponse()`
Specialized streaming for message regeneration:
- Similar to `streamChatResponse()` but with `regenerate: true`
- Updates existing message instead of creating new one

### `processStreamChunk()`
Core message processing logic:
- Extracts JSON data from SSE format
- Validates message structure
- Updates chat store with new content
- Handles artifacts extraction

### `extractStreamingData()`
Utility for parsing SSE data:
- Removes `data: ` prefix
- Handles both single messages and multi-segment responses
- Trims whitespace and normalizes output

## Buffer Management Strategy

The buffering strategy handles the asynchronous nature of streaming:

1. **Accumulate**: Add incoming chunks to buffer
2. **Split**: Divide buffer by SSE delimiters (`\n\n`)
3. **Process**: Handle complete messages
4. **Retain**: Keep incomplete message for next iteration
5. **Cleanup**: Process remaining buffer when stream ends

This ensures no message data is lost and all messages are processed in order.

## Integration Points

### Progress Callbacks
```javascript
onProgress?: (chunk: string, responseIndex: number) => void
```
- Called for each processed SSE message
- Allows custom handling in different contexts
- Used for real-time UI updates

### Chat Store Integration
- Updates stored chat messages
- Manages message state (loading, error, complete)
- Handles message artifacts and metadata

### Error Notification
- User-facing error messages
- Developer console logging
- Graceful fallback behaviors

## Performance Considerations

- **Memory**: Buffer size is automatically managed
- **Processing**: Minimal parsing overhead per chunk
- **Updates**: Batched UI updates prevent excessive re-renders
- **Cleanup**: Proper reader resource management

## Example SSE Flow

```
1. User sends message: "Explain SSE"

2. Server responds with stream:
   data: {"id":"msg-1","choices":[{"delta":{"content":"Server-Sent"}}]}
   
   data: {"id":"msg-1","choices":[{"delta":{"content":" Events"}}]}
   
   data: {"id":"msg-1","choices":[{"delta":{"content":" allow..."}}]}

3. Each chunk updates the UI:
   "Server-Sent" → "Server-Sent Events" → "Server-Sent Events allow..."

4. Stream ends, message is complete
```

This architecture enables real-time chat experiences while maintaining reliability and error resilience.