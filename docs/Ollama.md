## 使用本地Ollama 模型

1. 安装ollama 并下载模型
   
```
curl -fsSL https://ollama.com/install.sh | sh
ollama pull mistral
```

linux 下，默认的systemd 的配置限制了本机访问， 需要改HOST 能远程访问，如果ollama 和chat 在同一个host， 则不存在这个问题

![image](https://github.com/swuecho/chat/assets/666683/3695c088-4dcd-4ff4-9a75-6b9d44186a4b)

2. 在 Chat Admin 页面配置模型
![image](https://github.com/swuecho/chat/assets/666683/bc1d111f-7bd4-458d-bfed-0a0a5611809f)


```
id: ollama-{modelName}  # modelName 与 pull的 ollama 模型 一致， 比如 mistral, ollama3, ollama2
name: does not matter, naming as you like, 
baseUrl: http://hostname:11434/api/chat
other fields is irrelevant.
```
id 和 baseUrl 这两个地方配置对即可。

enjoy!

## Using Local Ollama Models (English Version)

1. Install Ollama and download a model
   
```bash
curl -fsSL https://ollama.com/install.sh | sh
ollama pull mistral
```

On Linux, the default systemd configuration restricts local access. You need to modify the HOST to allow remote access. If Ollama and Chat are on the same host, this is not an issue.

![image](https://github.com/swuecho/chat/assets/666683/3695c088-4dcd-4ff4-9a75-6b9d44186a4b)


2. Configure the model in the Chat Admin page

![image](https://github.com/swuecho/chat/assets/666683/bc1d111f-7bd4-458d-bfed-0a0a5611809f)

The key fields to configure are:
```
id: ollama-{modelName}  # modelName must match the Ollama model you pulled, e.g. mistral, ollama3, ollama2
name: Can be any name you prefer
baseUrl: http://hostname:11434/api/chat
```

Only the id and baseUrl fields need to be configured correctly. Other fields can be left as default.

Enjoy!
