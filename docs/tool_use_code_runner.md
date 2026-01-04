# Tool Use with Code Runner (Lightweight Mode)

This guide explains how tool use works when **Code Runner** is enabled per session. It is a lightweight, Claude‑Code‑style workflow where the model can request code execution and then continue based on tool results.

## 1) Enable Code Runner for a Session

1. Open a chat session.
2. Click **Session Config**.
3. Turn on **Code Runner**.
4. Leave it off by default for other sessions.

When enabled, the system adds tool instructions to the model prompt.

## 2) How Tool Calls Look

When the model needs execution, it will emit a tool call block:

````text
```tool_call
{"name":"run_code","arguments":{"language":"python","code":"print('hello')"}}
```
````

Only `run_code` is supported:
- `language`: `python`, `javascript`, or `typescript`
- `code`: the code to execute

The UI runs the tool automatically and sends results back to the model.

## 3) Tool Result Format

Tool results are sent back to the model using `tool_result` blocks:

````text
```tool_result
{"name":"run_code","success":true,"results":[{"type":"stdout","content":"hello"}]}
```
````

The model then continues with a normal answer (and can still emit artifacts).

## 4) Example Workflow

**User prompt**
```
Load /data/iris.csv and summarize it.
```

**Model tool call**
````text
```tool_call
{"name":"run_code","arguments":{"language":"python","code":"import pandas as pd\\nprint(pd.read_csv('/data/iris.csv').describe())"}}
```
````

**Tool result** (automatic)
````text
```tool_result
{"name":"run_code","success":true,"results":[{"type":"stdout","content":"...summary table..."}]}
```
````

**Model final response**
- Summary in plain text
- Optional executable artifact for future runs

## Notes

- Tool results are hidden from the chat UI to keep the conversation clean.
- If you want the output visible, ask the model to include it in a final response or in an artifact.
