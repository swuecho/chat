# Adding a New Chat Model

This guide explains how to add a new chat model to the system.

## Prerequisites
- Admin access to the system
- API credentials for the model you want to add
- Model's API endpoint URL

## Steps to Add a Model

### 1. Access the Admin Interface
1. Log in as an admin user
2. Navigate to the Admin section
3. Go to the "Models" tab

### 2. Fill in the Model Details
Fill in the following fields in the Add Model form:

- **Name**: Internal name for the model (e.g. "gpt-3.5-turbo")
- **Label**: Display name for the model (e.g. "GPT-3.5 Turbo")
- **URL**: API endpoint URL for the model
- **API Auth Header**: Header name for authentication (e.g. "Authorization")
- **API Auth Key**: Environment variable containing the API key
- **Is Default**: Whether this should be the default model
- **Enable Per-Mode Rate Limit**: Enable rate limiting for this specific model
- **Order Number**: Position in the model list (lower numbers appear first)
- **Default Tokens**: Default token limit for requests
- **Max Tokens**: Maximum token limit for requests

### 3. Add the Model
Click "Confirm" to add the model. The system will:
1. Validate the input
2. Create the model record in the database
3. Make the model available for use

### 4. (Optional) Set Rate Limits
If you enabled per-mode rate limiting:
1. Go to the "Rate Limits" tab
2. Set rate limits for specific users 

## Example Configuration

Here's an example JSON configuration you can paste into the form:

```json
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
```

## Troubleshooting

**Model not appearing?**
- Check if the model was added successfully in the database
- Verify the API credentials are correct
- Ensure the API endpoint is accessible

**Rate limiting issues?**
- Verify rate limits are properly configured
- Check user permissions
- Review system logs for errors

