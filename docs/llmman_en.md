## Using Local llmman Models

[llmman](https://github.com/llmmanorg/llmman) is a local model runner that serves the Ollama API (alongside OpenAI- and Anthropic-compatible ones) on port 17434. Chat talks to it through the existing `ollama` API type; only the port differs from Ollama.

1. Install llmman, start the server and pull a model

```bash
curl -fsSL https://raw.githubusercontent.com/llmmanorg/llmman/main/install.sh | sh
llmman serve
llmman pull gemma4
```

Models can also be pulled straight from Hugging Face, e.g. `llmman pull hf.co/unsloth/Qwen3.5-0.8B-GGUF`.

llmman binds to `127.0.0.1:17434` by default. If llmman and Chat run on different hosts, set `LLMMAN_HOST=0.0.0.0:17434` before `llmman serve`.

2. Configure the model in the Chat Admin page

The key fields to configure are:
```
name: gemma4                                 # must match the model you pulled
label: Can be any name you prefer
url: http://hostname:17434/api/chat
apiType: ollama
```

No API key is required; leave the auth header and auth key fields empty.

Enjoy your local models!
