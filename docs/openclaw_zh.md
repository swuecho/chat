# OpenClaw 集成指南

本指南介绍如何将 [OpenClaw](https://github.com/openclaw/openclaw) 集成到聊天应用程序中。

## 概述

[OpenClaw](https://github.com/openclaw/openclaw) 是一个 AI 网关，通过 OpenAI 兼容的 API 提供对多个 LLM 提供商的统一访问。此集成允许您使用 OpenClaw 作为聊天应用程序的后端。

## 前提条件

- OpenClaw 网关正在运行并可访问
- 在 OpenClaw 中配置了 API 密钥（如果启用了身份验证）
- 对聊天应用程序的管理员访问权限

## 配置

### 1. 启动 OpenClaw 网关

确保您的 OpenClaw 网关正在运行。默认情况下，它在 8080 端口运行：

```bash
openclaw gateway start
```

OpenAI 兼容的 API 端点将在以下地址可用：
```
http://localhost:8080/v1/chat/completions
```

### 2. 添加 OpenClaw 作为模型

以管理员身份登录并导航到 **Admin > Models > Add Model**。

填写以下配置：

```json
{
  "name": "openclaw-default",
  "label": "OpenClaw (默认)",
  "url": "http://localhost:8080/v1/chat/completions",
  "apiAuthHeader": "Authorization",
  "apiAuthKey": "OPENCLAW_API_KEY",
  "isDefault": false,
  "enablePerModeRatelimit": false,
  "isEnable": true,
  "orderNumber": 10,
  "defaultToken": 4096,
  "maxToken": 8192
}
```

**字段说明：**

| 字段 | 值 | 描述 |
|------|-----|------|
| `name` | `openclaw-default` | 内部标识符（必须以 `openclaw-` 开头以自动检测 api_type） |
| `label` | `OpenClaw (默认)` | UI 中显示的名称 |
| `url` | `http://localhost:8080/v1/chat/completions` | OpenClaw 的 OpenAI 兼容端点 |
| `apiAuthHeader` | `Authorization` | 身份验证的标头名称 |
| `apiAuthKey` | `OPENCLAW_API_KEY` | 包含 API 密钥的环境变量 |
| `api_type` | `openclaw` | 如果不使用 `openclaw-` 前缀，则显式设置 |

### 3. 设置环境变量

在环境中设置 API 密钥：

```bash
export OPENCLAW_API_KEY="your-openclaw-api-key"
```

或添加到 `.env` 文件：
```
OPENCLAW_API_KEY=your-openclaw-api-key
```

### 4. 重启 API 服务器

添加模型配置后，重启 API 服务器以应用更改：

```bash
# 如果使用 docker-compose
docker-compose restart api

# 如果本地运行
go run ./api
```

## 高级配置

### 使用远程 OpenClaw 实例

如果您的 OpenClaw 网关在另一台主机上运行：

```json
{
  "name": "openclaw-remote",
  "label": "OpenClaw (远程)",
  "url": "https://openclaw.your-domain.com/v1/chat/completions",
  "apiAuthHeader": "Authorization",
  "apiAuthKey": "OPENCLAW_REMOTE_API_KEY",
  "isDefault": false,
  "enablePerModeRatelimit": false,
  "isEnable": true,
  "orderNumber": 11,
  "defaultToken": 4096,
  "maxToken": 8192
}
```

### 多个 OpenClaw 配置

您可以添加具有不同设置的多个 OpenClaw 配置：

```json
{
  "name": "openclaw-high-capacity",
  "label": "OpenClaw (高容量)",
  "url": "http://localhost:8080/v1/chat/completions",
  "apiAuthHeader": "Authorization",
  "apiAuthKey": "OPENCLAW_API_KEY",
  "isDefault": false,
  "enablePerModeRatelimit": false,
  "isEnable": true,
  "orderNumber": 12,
  "defaultToken": 16384,
  "maxToken": 32768
}
```

## 故障排除

### 连接被拒绝

如果您看到连接错误：
1. 验证 OpenClaw 是否正在运行：`openclaw gateway status`
2. 检查 URL 是否正确（默认：`http://localhost:8080`）
3. 确保没有防火墙阻止连接

### 身份验证错误

如果您收到 401/403 错误：
1. 验证 `OPENCLAW_API_KEY` 是否正确设置
2. 检查 OpenClaw 中的 API 密钥是否匹配
3. 确保 `apiAuthHeader` 设置为 `Authorization`

### 模型未显示

如果模型没有出现在下拉列表中：
1. 检查 `isEnable` 是否设置为 `true`
2. 验证名称是否以 `openclaw-` 开头或 `api_type` 显式设置为 `openclaw`
3. 检查 API 服务器日志以获取错误信息

### 流式传输问题

如果流式响应不起作用：
1. 确保 OpenClaw 支持 SSE（服务器发送事件）
2. 检查 URL 是否使用 `/v1/chat/completions` 端点
3. 验证 `stream: true` 是否受您的 OpenClaw 配置支持

## API 类型路由

聊天应用程序使用 `api_type` 字段来确定使用哪个处理程序：

- 名称前缀为 `openclaw-` 的模型自动获得 `api_type = 'openclaw'`
- 您也可以在数据库中手动设置 `api_type = 'openclaw'`
- OpenClaw 处理程序使用 OpenAI 兼容的流式传输格式

## 环境变量参考

| 变量 | 必需 | 描述 |
|------|------|------|
| `OPENCLAW_API_KEY` | 否* | OpenClaw 身份验证的 API 密钥（如果 OpenClaw 启用了身份验证则必需） |

## 另请参阅

- [OpenClaw 文档](https://docs.openclaw.ai)
- [OpenClaw GitHub](https://github.com/openclaw/openclaw)
- [添加新模型指南](./add_model_zh.md)
