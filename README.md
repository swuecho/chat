## Demo

[video.webm](https://user-images.githubusercontent.com/666683/230305516-60154c5b-7170-4d2b-9670-a5ff4c851d25.webm)


<img width="850" alt="Screenshot 2023-04-12 at 12 43 31" src="https://user-images.githubusercontent.com/666683/231352196-5a6101db-9f5b-4eae-9198-59afed5822a6.png">


## 规则

- 第一个消息是系统消息（prompt）
- 上下文默认附带最新创建的10条消息
- 第一个注册的用户是管理员
- 默认限流 100 chatGPT call /10分钟 (OPENAI_RATELIMIT=100)
- 支持OPEN AI, Claude 模型 [免费申请链接](https://www.anthropic.com/earlyaccess)

## 如何部署

参考 `docker-compose.yaml`

## 致谢

- web: [ChatGPT-Web](https://github.com/Chanzhaoyu/chatgpt-web) 复制过来的 。
- api : 参考 [Kerwin1202](https://github.com/Kerwin1202)'s [Chanzhaoyu/chatgpt-web#589](https://github.com/Chanzhaoyu/chatgpt-web/pull/589) 的node版本在chatgpt帮助下写的

## LICENCE: MIT 

## How to Use

- The first message is a system message (prompt)
- by default, the latest 10 messages are context
- First user is superuser.
- 100 chatgpt api call / 10 mins (OPENAI_RATELIMIT=100)
-  Support OPEN AI, Claude model [free application link](https://www.anthropic.com/earlyaccess)

## How to Deploy

Refer to `docker-compose.yaml`

## Acknowledgments

- web: copied from chatgpt-web https://github.com/Chanzhaoyu/chatgpt-web
- api: based on the node version of [Kerwin1202](https://github.com/Kerwin1202)'s [Chanzhaoyu/chatgpt-web#589](https://github.com/Chanzhaoyu/chatgpt-web/pull/589)
and written with the help of chatgpt.
