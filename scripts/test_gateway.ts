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
      max_tokens: 128,
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
  let reasoning = ""
  let sawDone = false
  let eventCount = 0
  const observedDeltaFields = new Set<string>()

  function processLine(rawLine: string) {
    const line = rawLine.trimEnd()
    if (!line.startsWith("data:"))
      return
    const data = line.slice(5).trimStart()
    if (data === "[DONE]") {
      sawDone = true
      return
    }
    if (!data)
      return
    const chunk = JSON.parse(data) as {
      choices?: Array<{ delta?: { content?: string; reasoning_content?: string; reasoning?: string } }>
    }
    eventCount++
    const delta = chunk.choices?.[0]?.delta
    if (!delta)
      return
    Object.keys(delta).forEach(field => observedDeltaFields.add(field))
    content += delta.content ?? ""
    reasoning += delta.reasoning_content ?? delta.reasoning ?? ""
  }

  while (true) {
    const { value, done } = await reader.read()
    buffer += decoder.decode(value, { stream: !done })

    const lines = buffer.split("\n")
    buffer = lines.pop() ?? ""
    lines.forEach(processLine)
    if (done)
      break
  }
  if (buffer)
    processLine(buffer)

  if (!sawDone)
    throw new Error("Streaming response did not contain data: [DONE]")
  if (!content && !reasoning) {
    const fields = [...observedDeltaFields].join(", ") || "none"
    throw new Error(`Streaming response contained no text or reasoning delta (${eventCount} events; delta fields: ${fields})`)
  }
  if (content)
    console.log(`✓ Streaming content: ${JSON.stringify(content)}`)
  if (reasoning)
    console.log(`✓ Streaming reasoning: ${JSON.stringify(reasoning)}`)
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
