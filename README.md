## Demo


<img width="850" alt="Snipaste_2024-08-03_18-32-02" src="https://github.com/user-attachments/assets/d9ccb0ef-1409-43e4-94a1-80aed2fd33d6">

<img width="850" alt="image" src="https://github.com/user-attachments/assets/65b7286e-9df6-429c-98a4-64bd8ad1518b">

![image](https://github.com/user-attachments/assets/ad38194e-dd13-4eb0-b946-81c29a37955d)

<img width="850" alt="image" src="https://github.com/swuecho/chat/assets/666683/45dd865e-7f9f-4209-8587-4781e37dd928">

<img width="850" alt="image" src="https://github.com/swuecho/chat/assets/666683/0c4f546a-e884-4dc1-91c0-d4b07e63a1a9.png">


![image](https://github.com/user-attachments/assets/5b3751e4-eaa1-4a79-b47a-9b073c63eb04)


<img width="850" alt="image" src="https://github.com/user-attachments/assets/e5145dc9-ca4e-4fc3-a40c-ef28693d811a" />


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

> （可选）对话标题用 `gemini-2.0-flash` 生成， 所以需要配置该模型， 不配置默认用提示词前100个字符

## 文档

- [添加新模型指南](https://github.com/swuecho/chat/blob/master/docs/add_model_zh.md)
- [快照 vs 聊天机器人](https://github.com/swuecho/chat/blob/master/docs/snapshots_vs_chatbots_zh.md)
- [使用本地Ollama](https://github.com/swuecho/chat/blob/master/docs/ollama_zh.md)
- [Community Discussions](https://github.com/swuecho/chat/discussions)

## 开发指南

- [本地开发指南](https://github.com/swuecho/chat/blob/master/docs/dev_locally_zh.md)

## 部署指南

- [部署指南](https://github.com/swuecho/chat/blob/master/docs/deployment_zh.md)

## 致谢

- web: [ChatGPT-Web](https://github.com/Chanzhaoyu/chatgpt-web) 复制过来的 。
- api : 参考 [Kerwin1202](https://github.com/Kerwin1202)'s [Chanzhaoyu/chatgpt-web#589](https://github.com/Chanzhaoyu/chatgpt-web/pull/589) 的node版本在chatgpt帮助下写的

## LICENCE: MIT

## How to Use

- The first message is a system message (prompt)
- by default, the latest 4 messages are context
- First user is superuser.
- 100 chatgpt api call / 10 mins (OPENAI_RATELIMIT=100)
- Snapshot conversation and Share (like ShareGPT)
- Support OPEN AI, Claude model 
- Support Upload File
- Support MultiMedia File (rely on Model support)

## User Manual

For instructions on how to add new models, please refer to:
- [Adding New Models Guide](https://github.com/swuecho/chat/blob/master/docs/add_model_en.md)
- [Snapshots vs ChatBots](https://github.com/swuecho/chat/blob/master/docs/snapshots_vs_chatbots_en.md)
- [Using Local Ollama](https://github.com/swuecho/chat/blob/master/docs/ollama_en.md)
- [Community Discussions](https://github.com/swuecho/chat/discussions)

## Deployment

- [Deployment Guide](https://github.com/swuecho/chat/blob/master/docs/deployment_en.md)

## Development

- [Local Development Guide](https://github.com/swuecho/chat/blob/master/docs/dev_locally_en.md)

## Acknowledgments

- web: copied from chatgpt-web <https://github.com/Chanzhaoyu/chatgpt-web>
- api: based on the node version of [Kerwin1202](https://github.com/Kerwin1202)'s [Chanzhaoyu/chatgpt-web#589](https://github.com/Chanzhaoyu/chatgpt-web/pull/589)
and written with the help of chatgpt.
