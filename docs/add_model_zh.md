# 添加新聊天模型

本指南介绍如何向系统添加新的聊天模型。

## 先决条件
- 系统管理员权限
- 要添加模型的API凭证
- 模型的API端点URL

## 添加模型的步骤

### 1. 访问管理员界面
1. 以管理员用户登录
2. 导航到管理员部分
3. 进入"模型"标签页

<img width="1880" alt="image" src="https://github.com/user-attachments/assets/a9ca268e-9a8c-4ab1-bcc8-d847b905dc6a" />

### 2. 填写模型详情
在添加模型表单中填写以下字段：

- **名称**: 模型的内部名称 (如 "gpt-3.5-turbo")
- **标签**: 模型的显示名称 (如 "GPT-3.5 Turbo")
- **URL**: 模型的API端点URL
- **API认证头**: 认证头名称 (如 "Authorization", "x-api-key")
- **API认证密钥**: 包含API密钥的环境变量
- **是否默认**: 是否设为默认模型
- **启用模式限速**: 为此特定模型启用速率限制
- **排序号**: 在模型列表中的位置（数字越小越靠前）
- **默认token数**: 请求的默认token限制
- **最大token数**: 请求的最大token限制

<img width="665" alt="image" src="https://github.com/user-attachments/assets/d6646e82-487f-4c47-bf4a-075b9437b340" />

### 3. 添加模型
点击"确认"添加模型。系统将：
1. 验证输入
2. 在数据库中创建模型记录
3. 使模型可供使用

### 4. （可选）设置速率限制
如果启用了模式限速：
1. 进入"速率限制"标签页
2. 为特定用户设置速率限制

## 示例配置

以下是可粘贴到表单中的示例JSON配置：

```json
# openai
{
  "name": "gpt-4",
  "label": "GPT-4",
  "url": "https://api.openai.com/v1/chat/completions",
  "apiAuthHeader": "Authorization",
  "apiAuthKey": "OPENAI_API_KEY",
  "isDefault": false,
  "enablePerModeRatelimit": true,
  "orderNumber": 5,
  "defaultToken": 4096,
  "maxToken": 8192
}

# claude
{
  "name": "claude-3-7-sonnet-20250219",
  "label": "claude-3-7-sonnet-20250219",
  "url": "https://api.anthropic.com/v1/messages",
  "apiAuthHeader": "x-api-key",
  "apiAuthKey": "CLAUDE_API_KEY",
  "isDefault": false,
  "enablePerModeRatelimit": false,
  "isEnable": true,
  "orderNumber": 0,
  "defaultToken": 4096,
  "maxToken": 4096
}

# gemini
{
  "name": "gemini-2.0-flash",
  "label": "gemini-2.0-flash",
  "url": "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash",
  "apiAuthHeader": "GEMINI_API_KEY",
  "apiAuthKey": "GEMINI_API_KEY",
  "isDefault": true,
  "enablePerModeRatelimit": false,
  "isEnable": true,
  "orderNumber": 0,
  "defaultToken": 4096,
  "maxToken": 4096
}

# deepseek
{
  "name": "deepseek-chat",
  "label": "deepseek-chat",
  "url": "https://api.deepseek.com/v1/chat/completions",
  "apiAuthHeader": "Authorization",
  "apiAuthKey": "DEEPSEEK_API_KEY",
  "isDefault": false,
  "enablePerModeRatelimit": false,
  "isEnable": true,
  "orderNumber": 0,
  "defaultToken": 8192,
  "maxToken": 8192
}

# open router
{
  "name": "deepseek/deepseek-r1:free",
  "label": "deepseek/deepseek-r1(OR)",
  "url": "https://openrouter.ai/api/v1/chat/completions",
  "apiAuthHeader": "Authorization",
  "apiAuthKey": "OPENROUTER_API_KEY",
  "isDefault": false,
  "enablePerModeRatelimit": false,
  "isEnable": true,
  "orderNumber": 1,
  "defaultToken": 8192,
  "maxToken": 8192
}
```

## 故障排除

**模型未出现？**
- 检查模型是否成功添加到数据库
- 验证API凭证是否正确
- 确保API端点可访问

**速率限制问题？**
- 验证速率限制是否正确配置
- 检查用户权限
- 查看系统日志中的错误
