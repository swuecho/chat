## 使用本地Ollama 模型

1. 安装ollama 并下载模型
   
```bash
curl -fsSL https://ollama.com/install.sh | sh
ollama pull mistral
```

linux 下，默认的systemd 的配置限制了本机访问， 需要改HOST 能远程访问，如果ollama 和chat 在同一个host， 则不存在这个问题

![image](https://github.com/swuecho/chat/assets/666683/3695c088-4dcd-4ff4-9a75-6b9d44186a4b)

2. 在 Chat Admin 页面配置模型
![image](https://github.com/swuecho/chat/assets/666683/bc1d111f-7bd4-458d-bfed-0a0a5611809f)

关键配置字段：
```
id: ollama-{modelName}  # modelName 必须与pull的ollama模型一致，如mistral, ollama3, ollama2
name: 可任意命名
baseUrl: http://hostname:11434/api/chat
```

只需正确配置id和baseUrl字段即可，其他字段可保持默认。

享受本地模型的乐趣！
