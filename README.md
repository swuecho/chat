## Demo

[video.webm](https://user-images.githubusercontent.com/666683/230305516-60154c5b-7170-4d2b-9670-a5ff4c851d25.webm)

<img width="1116" alt="image" src="https://github.com/swuecho/chat/assets/666683/8064b861-3089-47a7-8331-fc714ab4fe98">

<img width="850" alt="image" src="https://user-images.githubusercontent.com/666683/234461014-d717a0ff-92ff-45a1-a981-ac480145b25c.png">

<img width="850" alt="image" src="https://github-production-user-asset-6210df.s3.amazonaws.com/666683/245065230-5a65a45e-2630-4d9e-b545-5a62bab72bde.png">

<img width="850" alt="image" src="https://user-images.githubusercontent.com/666683/233824829-f6069e25-05b3-48ef-a165-bb854a009edd.png">


## 规则

- 第一个消息是系统消息（prompt）
- 上下文默认附带最新创建的4条消息
- 第一个注册的用户是管理员
- 默认限流 100 chatGPT call /10分钟 (OPENAI_RATELIMIT=100)
- 根据对话生成可以分享的静态页面(like ShareGPT), 也可以继续会话. 
- 对话快照目录(对话集), 支持全文查找(Enlgish), 方便整理, 搜索会话记录.
- 支持OPEN AI, Claude 模型 [免费申请链接](https://www.anthropic.com/earlyaccess)
- 支持Ollama host模型, 配置参考: https://github.com/swuecho/chat/discussions/396
- 提示词管理, 提示词快捷键 '/'

## 参与开发

1. git clone
2. golang dev

```bash
cd chat; cd api
go install github.com/cosmtrek/air@latest
go mod tidy
# export env var, change base on your env
export PG_HOST=192.168.0.135
export PG_DB=hwu
export PG_USER=hwu
export PG_PASS=pass
export PG_PORT=5432
# export DATABASE_URL= postgres://user:pass@192.168.0.1:5432/db?sslmode=disable

# export OPENAI_API_KEY=sk-xxx, not required if you use `debug` model
# export OPENAI_RATELIMIT=100
#
make serve
```

3. node env

```bash
cd ..; cd web
npm install
npm run dev
```

4. e2e test

```bash
cd ..; cd e2e
# export env var, change base on your env
export PG_HOST=192.168.0.135
export PG_DB=hwu
export PG_USER=hwu
export PG_PASS=pass
export PG_PORT=5432
npm install
npx playwright test # --ui 
```

The instruction might not be accurate, ask in issue or discussion if unclear.

## 如何部署

参考 `docker-compose.yaml`

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/tk7jWU?referralCode=5DMfQv)

然后配置环境变量就可以了.

```
PORT=8080
OPENAI_RATELIMIT=0
```

别的两个 api key 有就填.

<img width="750" alt="image" src="https://user-images.githubusercontent.com/666683/232234418-941c9336-783c-4430-857c-9e7b703bb1c1.png">

部署之后,  注册用户, 第一个用户是管理员, 然后到  <https://$hostname/static/#/admin/user>,
设置 ratelimit, 公网部署, 只对信任的email 增加 ratelimit, 这样即使有人注册, 也是不能用的.

<img width="750" alt="image" src="https://user-images.githubusercontent.com/666683/232227529-284289a8-1336-49dd-b5c6-8e8226b9e862.png">


## 使用本地Ollama 模型

1. 安装ollama 并下载模型
   
```
curl -fsSL https://ollama.com/install.sh | sh
ollama pull mistral
```

linux 下，默认的systemd 的配置限制了本机访问， 需要改HOST 能远程访问，如果ollama 和chat 在同一个host， 则不存在这个问题

![image](https://github.com/swuecho/chat/assets/666683/3695c088-4dcd-4ff4-9a75-6b9d44186a4b)

2. 在 Chat Admin 页面配置模型

<img width="1293" alt="image" src="https://github.com/swuecho/chat/assets/666683/65cfca8a-edbf-4fa2-844c-c3b536b32cd0">

```
id: ollama-{modelName}  # modelName 与 pull的 ollama 模型 一致， 比如 mistral, ollama3, ollama2
name: does not matter, naming as you like, 
baseUrl: http://hostname:11434/api/generate
other fields is irrelevant.
```
id 和 baseUrl 这两个地方配置对即可。

enjoy!



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
- Support OPEN AI, Claude model [free application link](https://www.anthropic.com/earlyaccess)

## How to Deploy

Refer to `docker-compose.yaml`

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/tk7jWU?referralCode=5DMfQv)

Then configure the environment variables.

```
PORT=8080
OPENAI_RATELIMIT=0
```

Fill in the other two keys if you have them.

<img width="750" alt="image" src="https://user-images.githubusercontent.com/666683/232234418-941c9336-783c-4430-857c-9e7b703bb1c1.png">

After deployment, registering users, the first user is an administrator, then go to
<https://$hostname/static/#/admin/user> to set rate limiting. Public deployment,
only adds rate limiting to trusted emails, so even if someone registers, it will not be available.

<img width="750" alt="image" src="https://user-images.githubusercontent.com/666683/232227529-284289a8-1336-49dd-b5c6-8e8226b9e862.png">

This helps ensure only authorized users can access the deployed system by limiting registration to trusted emails and enabling rate limiting controls.

1. download ollama and pull mistral model.
   
```
curl -fsSL https://ollama.com/install.sh | sh
ollama pull mistral
```

![image](https://github.com/swuecho/chat/assets/666683/3695c088-4dcd-4ff4-9a75-6b9d44186a4b)

2. config ollama model in chat admin

<img width="1293" alt="image" src="https://github.com/swuecho/chat/assets/666683/65cfca8a-edbf-4fa2-844c-c3b536b32cd0">

```
id: ollama-{modelName}
name: does not matter, naming as you like
baseUrl: http://hostname:11434/api/generate
other fields is irrelevant.
```

enjoy!

## Acknowledgments

- web: copied from chatgpt-web <https://github.com/Chanzhaoyu/chatgpt-web>
- api: based on the node version of [Kerwin1202](https://github.com/Kerwin1202)'s [Chanzhaoyu/chatgpt-web#589](https://github.com/Chanzhaoyu/chatgpt-web/pull/589)
and written with the help of chatgpt.
