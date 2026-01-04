# Tool-Use Showcase: Chat + Code Runner (Lightweight Mode)

This doc demonstrates the current tool‑use capabilities in the chat interface. It uses the built‑in code runner plus VFS tooling to inspect files, run code, and return artifacts.

## Prerequisites

1. Open a chat session.
2. Open **Session Config** and enable **Code Runner**.
3. Upload one or more files to the VFS (for example `iris.csv`).

## 1) Inspect the VFS

**User**
```
List the files in /data and show file stats for iris.csv.
```

**Assistant (tool calls)**
````text
```tool_call
{"name":"list_vfs","arguments":{"path":"/data"}}
```

```tool_call
{"name":"stat_vfs","arguments":{"path":"/data/iris.csv"}}
```
````

**Tool results** (automatic)
````text
```tool_result
{"name":"list_vfs","success":true,"results":[{"type":"vfs","content":"[\"iris.csv\",\"notes.txt\"]","path":"/data"}]}
```

```tool_result
{"name":"stat_vfs","success":true,"results":[{"type":"vfs","content":"{\"isDirectory\":false,\"isFile\":true}","path":"/data/iris.csv"}]}
```
````

**Assistant (final)**
```
Found iris.csv in /data. It looks like a file (not a directory). I can load it next.
```

## 2) Read the CSV (Tool) + Analyze (Code Runner)

**User**
```
Load /data/iris.csv, summarize the numeric columns, and show the first 5 rows.
```

**Assistant (tool call)**
````text
```tool_call
{"name":"run_code","arguments":{"language":"python","code":"import pandas as pd\n\n# Load from VFS\niris = pd.read_csv('/data/iris.csv')\n\nprint('Head:')\nprint(iris.head())\n\nprint('\\nSummary:')\nprint(iris.describe())"}}
```
````

**Tool result** (automatic)
````text
```tool_result
{"name":"run_code","success":true,"results":[{"type":"stdout","content":"Head:\n..."}]}
```
````

**Assistant (final)**
```
Here’s the summary and first rows of the dataset. Let me know if you want charts or a filtered view.
```

## 3) Write Results Back to the VFS

**User**
```
Save the summary to /data/iris_summary.txt.
```

**Assistant (tool call)**
````text
```tool_call
{"name":"run_code","arguments":{"language":"python","code":"import pandas as pd\n\niris = pd.read_csv('/data/iris.csv')\nsummary = iris.describe().to_string()\n\nwith open('/data/iris_summary.txt', 'w') as f:\n    f.write(summary)\n\nprint('Saved summary to /data/iris_summary.txt')"}}
```
````

**Tool result** (automatic)
````text
```tool_result
{"name":"run_code","success":true,"results":[{"type":"stdout","content":"Saved summary to /data/iris_summary.txt"}]}
```
````

**Assistant (final)**
```
Saved. You can download the file from the VFS file manager or inspect it directly.
```

## 4) Optional: Use Artifacts for Reproducibility

When you want the result to be re‑runnable later, ask for a Python executable artifact:

````text
```python <!-- executable: Iris Summary -->
import pandas as pd

iris = pd.read_csv('/data/iris.csv')
print(iris.describe())
```
````

Artifacts show up in the message view and the Artifact Gallery.

## Notes

- Tool results are hidden by default to keep the chat clean.
- Use **Tool Debug** to view tool calls/results inline.
