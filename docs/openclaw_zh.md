# 在聊天应用中使用 OpenClaw

本文档说明如何在聊天应用中配置 OpenClaw 作为模型提供商。

## 什么是 OpenClaw？

OpenClaw 是一个 AI 助手平台，提供**与 OpenAI 兼容的 API**。这意味着您可以通过与 OpenAI 模型相同的接口使用 OpenClaw 模型——无需任何代码修改！

## 获取 OpenClaw API 密钥

### 方法一：从 OpenClaw Gateway 获取

1. OpenClaw 运行 Gateway 服务（默认：`http://localhost:7488`）
2. API 密钥通常存储在 OpenClaw 配置中
3. 检查 `~/.openclaw/config.yaml` 或环境变量

### 方法二：环境变量

OpenClaw 通常使用环境变量存储 API 密钥。检查以下变量：
- `OPENAI_API_KEY` - 如果 OpenClaw 配置为使用 OpenAI
- `ZHIPU_API_KEY` - 用于 GLM 模型（zai/glm-5）
- 其他提供商特定的密钥

### 方法三：直接配置

如果您自己托管 OpenClaw，可以通过以下方式生成 API 密钥：

```bash
# 检查 OpenClaw 状态
openclaw status

# 查看配置
cat ~/.openclaw/config.yaml
```

## 配置方法

由于 OpenClaw 使用与 OpenAI 兼容的 API，只需将其添加为 OpenAI 模型：

### 通过管理后台

导航到模型管理页面，添加新模型：

| 字段 | 值 |
|-------|-------|
| **Name** | 模型名称（如 `zai/glm-5`, `zai/gpt-4o`） |
| **Label** | 显示名称（如 "OpenClaw GLM-5"） |
| **URL** | OpenClaw API 端点 |
| **API Type** | `openai`（使用 OpenAI 兼容接口） |
| **API Auth Key** | API 密钥或环境变量名 |

### 配置示例

```json
{
  "name": "zai/glm-5",
  "label": "OpenClaw GLM-5",
  "url": "http://localhost:7488/v1/chat/completions",
  "apiType": "openai",
  "apiAuthKey": "ZHIPU_API_KEY"
}
```

### 常用 OpenClaw 端点

| 端点 | 说明 |
|------|------|
| `http://localhost:7488/v1/chat/completions` | 本地 OpenClaw Gateway |
| `http://your-server:7488/v1/chat/completions` | 远程 OpenClaw 实例 |

### 使用环境变量

通过环境变量设置 API 密钥：

```bash
# 用于 GLM 模型
export ZHIPU_API_KEY=your_zhipu_key

# 或使用通用名称
export OPENCLAW_API_KEY=your_key
```

然后在模型配置中，将 **API Auth Key** 设置为变量名（如 `ZHIPU_API_KEY`）。

### 直接使用 API 密钥

您也可以直接在 **API Auth Key** 字段中输入 API 密钥，无需使用环境变量。

## 可用模型

根据您的配置，OpenClaw 支持多种模型：

| 模型名称 | 说明 |
|----------|------|
| `zai/glm-5` | GLM-5（智谱 AI） |
| `zai/gpt-4o` | GPT-4o（通过 OpenClaw 代理） |
| `zai/claude-3-5-sonnet` | Claude 3.5 Sonnet（通过 OpenClaw 代理） |

请查看您的 OpenClaw 配置以了解可用模型。

## 功能特性

- **流式支持**：完整的流式响应支持，实现实时聊天
- **推理内容**：支持输出推理/思考内容的模型
- **文件上传**：支持文本和多媒体文件上传（如果模型支持）
- **速率限制**：可按模型配置

## 故障排除

### 连接被拒绝

```
Error: dial tcp 127.0.0.1:7488: connection refused
```

**解决方案：**
1. 确保 OpenClaw Gateway 正在运行：`openclaw gateway status`
2. 如需要则启动：`openclaw gateway start`
3. 检查模型配置中的 URL

### 认证错误

```
Error: 401 Unauthorized
```

**解决方案：**
1. 验证 API 密钥是否正确
2. 检查环境变量是否设置：`echo $ZHIPU_API_KEY`
3. 确保密钥具有适当的权限

### 模型未找到

```
Error: model not found
```

**解决方案：**
1. 检查 OpenClaw 配置中的可用模型
2. 验证模型名称完全匹配（包括前缀如 `zai/`）
3. 检查 OpenClaw 日志：`openclaw gateway logs`

### 超时错误

**解决方案：**
1. 在模型配置中增加 `HttpTimeOut`（默认：120 秒）
2. 检查网络连接
3. 验证 OpenClaw Gateway 是否响应

## 架构说明

```
聊天应用 → OpenAI 兼容 API → OpenClaw Gateway → LLM 提供商
                              ↓
                        API 密钥管理
                        请求路由
                        响应流式传输
```

聊天应用将 OpenClaw 完全视为 OpenAI，因为它们共享相同的 API 格式。OpenClaw Gateway 负责：
- 多个提供商的 API 密钥管理
- 将请求路由到适当的 LLM 后端
- 将响应流式传输回客户端

## 相关文档

- [添加新模型指南](./add_model_zh.md)
- [本地开发指南](./dev_locally_zh.md)
- [部署指南](./deployment_zh.md)
