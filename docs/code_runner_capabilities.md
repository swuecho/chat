# Code Runner Capabilities

This document summarizes what the current chat + code runner can do today.

## Execution

- JavaScript / TypeScript execution in a sandboxed worker
- Python execution in Pyodide
- Console output, return values, and error reporting
- Timeouts and basic resource limits

## Libraries and Packages

### JavaScript libraries (via `// @import`)
- lodash
- d3
- chart.js
- moment
- axios
- rxjs
- p5
- three
- fabric

### Python packages (auto‑loaded in Pyodide)
- numpy, pandas, matplotlib, scipy, scikit‑learn
- requests, beautifulsoup4, pillow, sympy, networkx
- seaborn, plotly, bokeh, altair

## Artifacts

- Code artifacts and executable code artifacts
- HTML artifacts (rendered in a sandboxed iframe)
- SVG, Mermaid, JSON, and Markdown artifacts
- Artifact Gallery for browsing, editing, and running artifacts

## Tool Use (Code Runner enabled per session)

When **Code Runner** is enabled in session settings, the model can use tools:

- `run_code` — Execute JS/TS or Python
- `read_vfs` — Read a file from the VFS
- `write_vfs` — Write a file to the VFS
- `list_vfs` — List directory entries
- `stat_vfs` — File/directory metadata

Tool results are returned automatically, and the model continues with normal responses or artifacts.

## Virtual File System (VFS)

- In‑memory filesystem with `/data`, `/workspace`, `/tmp`
- File upload via the UI
- Read/write from code (both JS and Python)
- VFS state syncs to runners before execution

## Typical Workflows

- Load CSVs and analyze with pandas
- Generate plots with matplotlib
- Prototype JS utilities or data transforms
- Build small HTML/SVG demos in artifacts

## Known Constraints

- No direct DOM access for JS
- No direct network requests from user code
- VFS only (no access to host filesystem)
