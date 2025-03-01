# Snapshots vs ChatBots

## Snapshots (Chat Records)

Snapshots are static records of chat conversations. They are useful for:

- Archiving important conversations
- Sharing chat histories with others
- Reference and review of past discussions
- Exporting conversations in various formats (Markdown, PNG)

Key characteristics:
- Read-only after creation
- Contains the full conversation history
- Can be exported/shared
- Useful for documentation and record-keeping
- Accessed via `/snapshot/{uuid}` route

## ChatBots

ChatBots are interactive AI assistants created from chat histories. They are useful for:

- Creating specialized AI assistants
- Continuing conversations with context
- Building custom AI tools
- Sharing interactive AI experiences

Key characteristics:
- Created from existing chat histories
- Can continue conversations with new inputs
- Maintains context from original chat
- Interactive and dynamic
- Accessed via `/bot/{uuid}` route

## Key Differences

| Feature              | Snapshot                     | ChatBot                     |
|----------------------|------------------------------|-----------------------------|
| Interactivity        | Read-only                    | Interactive                 |
| Purpose              | Record keeping               | AI assistant                |
| Data                 | Static conversation history  | Dynamic conversation        |
| Export Options       | Markdown, PNG                | API access                  |
| Route                | /snapshot/{uuid}             | /bot/{uuid}                 |
| Use Case             | Documentation, Sharing       | Custom AI, Continued Chat   |

## Creating from Chat History

Both Snapshots and ChatBots can be created from existing chat histories:

1. **Snapshot** - Creates a permanent record of the conversation
2. **ChatBot** - Creates an interactive AI assistant based on the conversation

## API Access

ChatBots provide additional API access for integration:

```javascript
// Example API usage
const response = await fetch('/api/bot/{uuid}', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    message: 'Your question here'
  })
});
```

This allows programmatic interaction with the ChatBot while Snapshots remain static records.
