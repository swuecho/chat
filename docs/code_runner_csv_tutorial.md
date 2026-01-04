# Tutorial: Chat + File Upload + Python Artifact (Iris CSV)

This walkthrough shows how to create a chat with a system prompt, upload `iris.csv`, and ask the model to generate a Python executable artifact that reads the CSV from the VFS and prints results.

## 1) Create a New Chat Session

1. Open the app and click **New Chat**.
2. Set a clear title, for example: `Iris CSV Analysis`.
3. Open the session settings and add a **System Message** like this:

```text
You are a data assistant. Always answer with a Python executable artifact when asked to analyze data files. Use VFS paths like /workspace or /data. Print a small table and summary stats.
```

Why this works:
- The artifact extractor looks for fenced code blocks with `<!-- executable: Title -->`.
- The Python runner reads files from the in-memory VFS at `/workspace`, `/data`, or `/tmp`.

## 2) Upload the CSV File to the VFS

1. In the chat view, find the **VFS upload** area (often labeled “Upload files for code runners”).
2. Upload your `iris.csv` file.
3. Note the upload location:
   - Most uploads land under `/data/` or `/workspace/` in the VFS.
4. For this tutorial, we will assume the file is available at:

```text
/data/iris.csv
```

If your UI shows a different path, use that path in the next step.

## 3) Prompt the LLM to Generate a Python Artifact

Send a user message like this:

```text
Use Python to load the CSV at /data/iris.csv, show the first 5 rows, and print summary stats. 
Return your answer as an executable artifact.
```

Expected artifact structure (example):
````python
```python <!-- executable: Iris CSV Summary -->
import pandas as pd

df = pd.read_csv('/data/iris.csv')
print("Head:")
print(df.head())
print("\nSummary:")
print(df.describe())
```
````

## 4) Run the Artifact

1. The message will show an artifact panel.
2. Click **Run**.
3. Review output in the execution panel:
   - `stdout` for printed tables
   - `return` if a value is returned
   - `error` if the run fails

## Troubleshooting

- **File not found**: Confirm the path from the upload UI and update the code.
- **Missing package**: Use supported packages (pandas, numpy, matplotlib, etc.). The runner auto-loads common imports.
- **No output**: Ensure the code uses `print()` or returns a value.
