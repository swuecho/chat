# ZIO-inspired chat runtime

This project does not implement a Go effect system. It adopts the operational
guarantees that make ZIO applications predictable: explicit requirements,
classified failures, structured lifetimes, resource ownership, typed streams,
policy values, and deterministic tests.

## The generation lifecycle

A normal chat request has one owner: the chat-generation workflow in
`api/handler/chat_session_helpers.go`. Its lifecycle is:

1. claim the idempotency record;
2. persist or reuse the user prompt;
3. collect the owned session history;
4. mark the request as streaming;
5. resolve and invoke the provider;
6. consume content under the HTTP request context;
7. persist the completed answer;
8. emit exactly one terminal event;
9. start bounded, non-critical title generation.

The durable boundary matters. A `completed` event means the assistant message
was successfully persisted. A provider finishing is not, by itself, completion.

## Context and structured lifetimes

Request-scoped functions accept the caller's `context.Context` and derive
timeouts from it. They must not replace it with `context.Background()`.

`context.Background()` is reserved for application startup and explicitly
detached best-effort work such as logging or bounded title generation. Detached
work must have its own timeout and concurrency bound.

Provider stream producers use `emitChunk`. It selects between delivering a
chunk and cancellation, preventing a goroutine from becoming stuck after its
consumer exits. The stream consumer also derives and cancels its own context,
checks write errors, and rejects a channel that closes without a final answer.

When adding a provider:

- create requests with `http.NewRequestWithContext`;
- close every HTTP body or SDK stream in the function that acquired it;
- send chunks with `emitChunk`, never a bare channel send;
- return promptly when the context is canceled;
- close the provider's output channel exactly once in its owner goroutine;
- never translate cancellation into a successful partial answer.

## Provider failures

All upstream failures cross the provider boundary as
`domain.ProviderFailure`. Expected kinds are:

| Kind | Normally retryable | Meaning |
| --- | --- | --- |
| `invalid_request` | no | Provider rejected model, payload, or parameters |
| `authentication` | no | Upstream credentials were rejected |
| `permission` | no | Credentials lack access |
| `rate_limited` | yes | Upstream returned HTTP 429 |
| `unavailable` | yes | Network failure or upstream 5xx |
| `timeout` | yes | Deadline or upstream timeout |
| `canceled` | no | Caller canceled the workflow |
| `invalid_response` | no | Successful response could not be decoded or was empty |
| `configuration` | no | URL, key, or client configuration is invalid |
| `internal` | no | Unclassified provider defect |

`Retryable` describes the error category, not blanket permission to replay an
operation. Never retry a streaming generation after a delta has reached the
client unless the provider offers a verified resume protocol.

The HTTP adapter maps provider failures to stable public errors. Provider code
must not choose HTTP response wording.

## Retry policies

Retry scheduling uses `github.com/cenkalti/backoff/v4`, the newest release of
that library compatible with this project's Go 1.21 baseline. It provides
bounded exponential backoff, jitter, context cancellation, and permanent-error
classification. The provider adapter marks normalized failures with
`Retryable == false` as `backoff.Permanent`.

Current automatic retry use is deliberately limited to non-streaming title
generation. It is safe to rerun because no title content has been exposed to a
client and the final database update is idempotent.

When enabling retry elsewhere, verify all three conditions:

1. no partial response has crossed an external boundary;
2. repeating the operation cannot duplicate a durable mutation;
3. the policy has a small attempt cap, exponential delay, and jitter in
   production.

## Typed answer events

`provider.AnswerEvent` is the SSE event algebra:

- `started`
- `delta`
- `reasoning_delta`
- `suggested_questions`
- `completed`
- `failed`
- `canceled`

`completed`, `failed`, and `canceled` are terminal. `AnswerEventWriter` permits
only one terminal event and requires `Persisted: true` for `completed`.

The server emits typed events for normal, replayed, regenerated, and bot
answers. The frontend reader temporarily accepts legacy OpenAI-compatible delta
frames during migration, but new server code must use `AnswerEvent`; do not
invent an untyped SSE payload in a handler.

## Dependency boundaries

Use constructor injection for capabilities and keep interfaces narrow:

- application workflows depend on session/message/provider capabilities;
- handlers own JSON, HTTP status, and SSE adaptation;
- services own application behavior and durable invariants;
- provider implementations own upstream protocol details;
- SQLC remains the direct persistence adapter for simple queries;
- transaction interfaces expose only operations required by that use case.

IDs are injected through the existing `newID` function fields in application
services. The backoff library exposes clock and timer seams for time-dependent
tests. Follow this pattern when nondeterminism affects a test; avoid a global
mock clock or a general service locator.

## Testing laws

Tests should verify lifecycle laws, not only example output:

- cancellation makes producers and consumers exit;
- a provider channel closes exactly once;
- a stream has one terminal answer or one failure;
- channel closure without a final answer is an error;
- no content follows a terminal event;
- `completed` always implies durable persistence;
- retry stops on non-retryable failures;
- retry stops at its attempt limit;
- retry waiting is cancelable;
- HTTP 429 and 5xx are retryable, while 4xx request/auth failures are not;
- resource cleanup runs on success, failure, and cancellation.

Use an injected ID function and the backoff library's clock/timer seams for
deterministic tests.
Integration tests that require PostgreSQL run through the repository's Docker
test setup; focused domain/provider tests do not require Docker.

## Review checklist

Before merging a provider or streaming change, confirm:

- the caller context reaches every database and upstream call;
- every acquired body, stream, timer, and goroutine has an obvious owner;
- every upstream error is normalized;
- retryability is explicit and partial streams are not replayed;
- every producer send observes cancellation;
- the final answer is persisted before `completed`;
- exactly one terminal event is possible;
- tests cover cancellation and the failure path, not just success.
