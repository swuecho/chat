#!/usr/bin/env bun

const baseUrl = (process.env.GATEWAY_BASE_URL ?? "http://localhost:8080/v1").replace(/\/$/, "")
const apiKey = process.env.GATEWAY_API_KEY
const requestedModel = process.env.GATEWAY_MODEL

if (!apiKey) {
  console.error("Missing GATEWAY_API_KEY")
  console.error("Example: GATEWAY_API_KEY='sk-chat-...' bun scripts/test_gateway.ts")
  process.exit(1)
}

const headers = {
  Authorization: `Bearer ${apiKey}`,
  "Content-Type": "application/json",
}

async function readError(response: Response): Promise<string> {
  const text = await response.text()
  try {
    const body = JSON.parse(text)
    return body?.error?.message ?? text
  }
  catch {
    return text
  }
}

async function listModels(): Promise<string> {
  const response = await fetch(`${baseUrl}/models`, { headers })
  if (!response.ok)
    throw new Error(`GET /models returned ${response.status}: ${await readError(response)}`)

  const body = await response.json() as { data?: Array<{ id: string }> }
  const models = body.data?.map(model => model.id) ?? []
  if (models.length === 0)
    throw new Error("GET /models returned no available models")

  console.log(`✓ Models: ${models.join(", ")}`)

  if (requestedModel) {
    if (!models.includes(requestedModel))
      throw new Error(`GATEWAY_MODEL=${requestedModel} is not available`)
    return requestedModel
  }
  return models[0]
}

async function testNonStreaming(model: string): Promise<void> {
  const response = await fetch(`${baseUrl}/chat/completions`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      model,
      messages: [{ role: "user", content: "Reply with exactly: gateway non-stream ok" }],
      temperature: 0,
      max_tokens: 32,
      stream: false,
    }),
  })

  if (!response.ok)
    throw new Error(`Non-streaming request returned ${response.status}: ${await readError(response)}`)

  const body = await response.json() as {
    id?: string
    choices?: Array<{ message?: { content?: string } }>
    usage?: { prompt_tokens?: number; completion_tokens?: number; total_tokens?: number }
  }
  const content = body.choices?.[0]?.message?.content
  if (!content)
    throw new Error("Non-streaming response has no choices[0].message.content")

  console.log(`✓ Non-streaming: ${JSON.stringify(content)}`)
  if (body.usage)
    console.log(`  Usage: ${body.usage.prompt_tokens ?? 0} input + ${body.usage.completion_tokens ?? 0} output = ${body.usage.total_tokens ?? 0}`)
}

async function testStreaming(model: string): Promise<void> {
  const response = await fetch(`${baseUrl}/chat/completions`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      model,
      messages: [{ role: "user", content: "Reply with exactly: gateway stream ok" }],
      temperature: 0,
      max_tokens: 32,
      stream: true,
    }),
  })

  if (!response.ok)
    throw new Error(`Streaming request returned ${response.status}: ${await readError(response)}`)
  if (!response.body)
    throw new Error("Streaming response has no body")

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ""
  let content = ""
  let sawDone = false

  while (true) {
    const { value, done } = await reader.read()
    buffer += decoder.decode(value, { stream: !done })

    const lines = buffer.split("\n")
    buffer = lines.pop() ?? ""
    for (const rawLine of lines) {
      const line = rawLine.trimEnd()
      if (!line.startsWith("data: "))
        continue
      const data = line.slice(6)
      if (data === "[DONE]") {
        sawDone = true
        continue
      }
      const chunk = JSON.parse(data) as { choices?: Array<{ delta?: { content?: string } }> }
      content += chunk.choices?.[0]?.delta?.content ?? ""
    }
    if (done)
      break
  }

  if (!sawDone)
    throw new Error("Streaming response did not contain data: [DONE]")
  if (!content)
    throw new Error("Streaming response contained no text delta")
  console.log(`✓ Streaming: ${JSON.stringify(content)}`)
}

try {
  console.log(`Testing ${baseUrl}`)
  const model = await listModels()
  console.log(`Using model: ${model}`)
  await testNonStreaming(model)
  await testStreaming(model)
  console.log("✓ Gateway test passed")
}
catch (error) {
  console.error(`✗ Gateway test failed: ${error instanceof Error ? error.message : String(error)}`)
  process.exit(1)
}
