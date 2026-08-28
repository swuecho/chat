## Demo


<img width="850" alt="image" src="https://github.com/user-attachments/assets/98940012-a1d9-41c0-b5c7-fc060e74546a" />


<img width="850" alt="image" src="https://github.com/user-attachments/assets/65b7286e-9df6-429c-98a4-64bd8ad1518b">

<img width="850" alt="thinking" src="https://github.com/user-attachments/assets/e5145dc9-ca4e-4fc3-a40c-ef28693d811a" />


![image](https://github.com/user-attachments/assets/ad38194e-dd13-4eb0-b946-81c29a37955d)


<img width="850" alt="image" src="https://github.com/swuecho/chat/assets/666683/0c4f546a-e884-4dc1-91c0-d4b07e63a1a9.png">

<img width="850" alt="Screenshot 2025-09-11 at 8 05 03 PM" src="https://github.com/user-attachments/assets/d3ae5c15-7498-4352-95b4-bb96b7a4c2bb" />


![image](https://github.com/user-attachments/assets/5b3751e4-eaa1-4a79-b47a-9b073c63eb04)

<img width="850" alt="image" src="https://github.com/user-attachments/assets/13b0aff8-93c4-4406-acce-b48389ae0c88" />

<img width="850" alt="chat records" src="https://github.com/swuecho/chat/assets/666683/45dd865e-7f9f-4209-8587-4781e37dd928">

<img width="1601" alt="chat record comments" src="https://github.com/user-attachments/assets/9ce940b9-2023-47ba-bcbe-32f4846354b1" />





## 规则

- 第一个消息是系统消息（prompt）
- 上下文默认附带最新创建的4条消息
- 第一个注册的用户是管理员
- 默认限流 100 chatGPT call /10分钟 (OPENAI_RATELIMIT=100)
- 根据对话生成可以分享的静态页面(like ShareGPT), 也可以继续会话. 
- 对话快照目录(对话集), 支持全文查找(English), 方便整理, 搜索会话记录.
- 支持OPEN AI, Claude 模型 
- 支持Ollama host模型, 配置参考: https://github.com/swuecho/chat/discussions/396
- 支持上传文本文件
- 支持多媒体文件, 需要模型支持
- 提示词管理, 提示词快捷键 '/'
- 内置 OpenAI 兼容网关，可通过虚拟 API Key 调用已配置的模型
- 网关支持流式透传、限流、请求审计，以及限时保留的请求/响应样本

> （可选）可在管理后台的「标题生成模型」页面，从已启用的模型中选择用于自动生成对话标题的模型。若未配置标题生成模型，或标题生成失败，将使用提示词的前 100 个字符作为默认标题。

## 文档

- [添加新模型指南](https://github.com/swuecho/chat/blob/master/docs/add_model_zh.md)
- [快照 vs 聊天机器人](https://github.com/swuecho/chat/blob/master/docs/snapshots_vs_chatbots_zh.md)
- [使用本地Ollama](https://github.com/swuecho/chat/blob/master/docs/ollama_zh.md)
- [OpenAI 兼容网关](https://github.com/swuecho/chat/blob/master/docs/gateway_en.md)
- [论坛](https://github.com/swuecho/chat/discussions)

## 开发指南

- [本地开发指南](https://github.com/swuecho/chat/blob/master/docs/dev_locally_zh.md)
- 移动端聊天应用使用 Flutter， 见 [mobile/README.md](/Users/hwu/dev/chat/mobile/README.md)

## 部署指南

- [部署指南](https://github.com/swuecho/chat/blob/master/docs/deployment_zh.md)

## 致谢

- web: [ChatGPT-Web](https://github.com/Chanzhaoyu/chatgpt-web) 复制过来的 。
- api : 参考 [Kerwin1202](https://github.com/Kerwin1202)'s [Chanzhaoyu/chatgpt-web#589](https://github.com/Chanzhaoyu/chatgpt-web/pull/589) 的node版本在chatgpt帮助下写的

## LICENCE: MIT

## Rules

- The first message is a system message (prompt)
- By default, the latest 4 messages are included in context
- The first registered user becomes administrator
- Default rate limit: 100 ChatGPT calls per 10 minutes (OPENAI_RATELIMIT=100)
- Generate shareable static pages from conversations (like ShareGPT), or continue conversations
- Conversation snapshots directory supports full-text search (English), making it easy to organize and search conversation history
- Supports OpenAI and Claude models
- Supports Ollama host models, configuration reference: https://github.com/swuecho/chat/discussions/396
- Supports text file uploads
- Supports multimedia files (requires model support)
- Prompt management with '/' shortcut
- Built-in OpenAI-compatible gateway using revocable virtual API keys
- Transparent streaming proxy with per-key limits, request auditing, and time-limited request/response samples

> (Optional) In the admin panel, open **Title Generation Model** to select any enabled model for automatic conversation titles. If no title-generation model is configured, or title generation fails, the first 100 characters of the prompt are used as the default title.

## Documentation

- [Adding New Models Guide](https://github.com/swuecho/chat/blob/master/docs/add_model_en.md)
- [Snapshots vs ChatBots](https://github.com/swuecho/chat/blob/master/docs/snapshots_vs_chatbots_en.md)
- [Using Local Ollama](https://github.com/swuecho/chat/blob/master/docs/ollama_en.md)
- [OpenAI-Compatible Gateway](https://github.com/swuecho/chat/blob/master/docs/gateway_en.md)
- [Community Discussions](https://github.com/swuecho/chat/discussions)

## Development Guide

- [Local Development Guide](https://github.com/swuecho/chat/blob/master/docs/dev_locally_en.md)
- The mobile chat app is built with Flutter. See [mobile/README.md](/Users/hwu/dev/chat/mobile/README.md).

## Deployment Guide

- [Deployment Guide](https://github.com/swuecho/chat/blob/master/docs/deployment_en.md)

## Acknowledgments

- web: copied from chatgpt-web <https://github.com/Chanzhaoyu/chatgpt-web>
- api: based on the node version of [Kerwin1202](https://github.com/Kerwin1202)'s [Chanzhaoyu/chatgpt-web#589](https://github.com/Chanzhaoyu/chatgpt-web/pull/589)
and written with the help of chatgpt.
