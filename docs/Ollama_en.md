## Using Local Ollama Models

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

Enjoy your local models!
