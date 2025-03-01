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

<img width="924" alt="Image" src="https://github.com/user-attachments/assets/51b221ab-603e-41aa-8b68-b3f32dce5f5c" />

这允许通过编程方式与聊天机器人交互，而快照保持静态记录。
