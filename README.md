## Demo

[video.webm](https://user-images.githubusercontent.com/666683/228706339-453e3a2d-5801-42f1-9906-86e2ac140aec.webm)


## 规则

- 第一个消息是系统消息（prompt）
- OPENAI 上下文默认附带最新创建的10条消息, (Claude 模型不带上下文)
- 第一个注册的用户是管理员
- 默认限流 100 chatGPT call /10分钟 (OPENAI_RATELIMIT=100)

## 如何部署

参考 `docker-compose.yaml`

## 致谢

- web: [ChatGPT-Web](https://github.com/Chanzhaoyu/chatgpt-web) 复制过来的 。
- api : 参考 [Kerwin1202](https://github.com/Kerwin1202)'s [Chanzhaoyu/chatgpt-web#589](https://github.com/Chanzhaoyu/chatgpt-web/pull/589) 的node版本在chatgpt帮助下写的


## How to Use

- The first message is a system message (prompt)
- by default, the latest 10 messages are context for opeanai model. no context for claude v1.
- First user is superuser.
- 100 chatgpt api call / 10 mins (OPENAI_RATELIMIT=100)

## How to Deploy

Refer to `docker-compose.yaml`

## Acknowledgments

- web: copied from chatgpt-web https://github.com/Chanzhaoyu/chatgpt-web
- api: based on the node version of [Kerwin1202](https://github.com/Kerwin1202)'s [Chanzhaoyu/chatgpt-web#589](https://github.com/Chanzhaoyu/chatgpt-web/pull/589)
and written with the help of chatgpt.
