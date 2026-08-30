# Typed Answer SSE Processing

The chat stream is an application protocol, not a pass-through of any LLM
provider's wire format. Providers normalize their output into `AnswerEvent` on
the backend, and the frontend accepts only those typed events.

## Event protocol

Every SSE frame has a named `event` and a JSON `data` object. The name must
match the object's `type` field.

```text
event: delta
data: {"type":"delta","answerId":"answer-1","delta":"Hello"}

event: completed
data: {"type":"completed","answerId":"answer-1","persisted":true}

```

The event types are:

- `started`
- `delta`
- `reasoning_delta`
- `suggested_questions`
- `completed`
- `failed`
- `canceled`

`completed`, `failed`, and `canceled` are terminal. A successful completion is
valid only when `persisted` is true.

## Client flow

`consumeAnswerEventStream()` in `utils/sse.ts` reads bytes from the response
body, normalizes CRLF to LF, and splits complete SSE frames on a blank line. It
retains the final incomplete frame until more bytes arrive. New answers and
regenerated answers share this transport path.

Each complete frame is passed to `readAnswerStreamEvent()` in `utils/sse.ts`.
The parser requires:

1. a recognized named event;
2. valid JSON data;
3. a matching `type` in the JSON object.

An untyped frame, including a provider-native OpenAI `choices` payload, is a
protocol error. There is no compatibility fallback.

`processAnswerEvent()` applies accepted events to the chat store:

- `delta` and `reasoning_delta` append text and refresh extracted artifacts;
- `suggested_questions` stores the suggestions;
- `started` and `completed` update answer identity and loading state;
- `failed` and `canceled` are surfaced through the stream error path.

The client rejects a stream that closes without a persisted `completed` event.
For a new answer it may retry once with the same request UUID, allowing the
server to replay a durably completed answer or reclaim a failed request.

## Backend contract

Handlers and application services emit `provider.AnswerEvent` through
`AnswerEventWriter`. Provider-specific chunks must be converted inside the
provider boundary. Do not expose raw provider payloads, unnamed `data:` frames,
or sentinel values such as `[DONE]` to the web client.

`AnswerEventWriter` enforces one terminal event and rejects a `completed` event
unless persistence has succeeded. This keeps transport completion aligned with
the durable answer state.
