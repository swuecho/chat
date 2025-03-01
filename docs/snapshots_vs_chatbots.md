# 快照(Chat Records) vs 聊天机器人(ChatBots)

## 快照 (聊天记录)

快照是聊天对话的静态记录。它们适用于：

- 存档重要对话
- 与他人分享聊天记录
- 回顾和参考过去的讨论
- 以多种格式导出对话（Markdown、PNG）

主要特点：
- 包含完整的对话历史
- 可以导出/分享
- 适用于文档和记录保存
- 通过 `/snapshot/{uuid}` 路由访问

## 聊天机器人

聊天机器人是基于聊天历史创建的交互式AI助手。它们适用于：

- 创建专门的AI助手
- 在上下文中继续对话
- 构建自定义AI工具
- 分享交互式AI体验

主要特点：
- 从现有聊天历史创建
- 可以继续对话并接受新输入
- 保持原始聊天的上下文
- 交互式和动态的
- 通过 `/bot/{uuid}` 路由访问

## 主要区别

| 特性              | 快照                     | 聊天机器人                     |
|------------------|--------------------------|------------------------------|
| 目的              | 记录保存                  | AI助手                       |
| 数据              | 静态对话历史               | 动态对话                      |
| 导出选项          | Markdown, PNG             | API访问                      |
| 路由              | /snapshot/{uuid}          | /bot/{uuid}                  |
| 使用场景          | 文档、分享                 | 自定义AI、API使用              |

## 从聊天历史创建

快照和聊天机器人都可以从现有聊天历史创建：

1. **快照** - 创建对话的永久记录
2. **聊天机器人** - 基于对话创建交互式AI助手

## API访问

聊天机器人提供额外的API访问用于集成：

```javascript
// API使用示例
const response = await fetch('/api/bot/{uuid}', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    message: '您的问题'
  })
});
```

这允许通过编程方式与聊天机器人交互，而快照保持静态记录。

---

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
| Purpose              | Record keeping               | AI assistant                |
| Data                 | Static conversation history  | Dynamic conversation        |
| Export Options       | Markdown, PNG                | API access                  |
| Route                | /snapshot/{uuid}             | /bot/{uuid}                 |
| Use Case             | Documentation, Sharing       | Custom AI, Use by API       |

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
