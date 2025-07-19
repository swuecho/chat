# Virtual File System (VFS) User Guide

## Overview

The Virtual File System (VFS) provides file I/O capabilities for both Python and JavaScript code runners. It creates an isolated, in-memory file system that allows you to work with files and directories as if they were on a real file system, while maintaining security and isolation from the host system.

## Key Features

- **Isolated Environment**: Files exist only within the code execution session
- **Standard APIs**: Compatible with standard Python and JavaScript file operations
- **Cross-Language**: Files created in Python can be accessed from JavaScript and vice versa
- **Automatic Integration**: No special setup required - just use normal file operations
- **Session Persistence**: Files persist throughout your chat session

## Directory Structure

The VFS comes with pre-created directories:

- `/workspace` - Main working directory (default current directory)
- `/data` - For storing data files (CSV, JSON, etc.)
- `/tmp` - For temporary files

You can create additional directories as needed.

## Python Usage

### Basic File Operations

```python
# Write a text file
with open('/data/example.txt', 'w') as f:
    f.write('Hello, World!')

# Read a text file
with open('/data/example.txt', 'r') as f:
    content = f.read()
    print(content)  # Output: Hello, World!

# Check if file exists
import os
if os.path.exists('/data/example.txt'):
    print('File exists!')
```

### Working with CSV Data

```python
import csv
import os

# Write CSV data
data = [
    ['Name', 'Age', 'City'],
    ['Alice', 25, 'New York'],
    ['Bob', 30, 'San Francisco']
]

with open('/data/people.csv', 'w', newline='') as f:
    writer = csv.writer(f)
    writer.writerows(data)

# Read CSV data
with open('/data/people.csv', 'r') as f:
    reader = csv.DictReader(f)
    for row in reader:
        print(f"{row['Name']} lives in {row['City']}")
```

### Using pandas with VFS

```python
import pandas as pd

# Create a DataFrame
df = pd.DataFrame({
    'product': ['A', 'B', 'C'],
    'sales': [100, 150, 80],
    'profit': [20, 30, 15]
})

# Save to CSV
df.to_csv('/data/sales.csv', index=False)

# Read back from CSV
df_loaded = pd.read_csv('/data/sales.csv')
print(df_loaded)

# Save to JSON
df.to_json('/data/sales.json', orient='records', indent=2)
```

### Directory Operations

```python
import os

# Create directories
os.makedirs('/workspace/project/src', exist_ok=True)

# List directory contents
files = os.listdir('/data')
print('Files in /data:', files)

# Get current directory
print('Current directory:', os.getcwd())

# Change directory
os.chdir('/workspace')
print('Changed to:', os.getcwd())

# Check if path is file or directory
print('Is file:', os.path.isfile('/data/example.txt'))
print('Is directory:', os.path.isdir('/data'))
```

### Using pathlib (Modern Python)

```python
from pathlib import Path

# Create a Path object
data_dir = Path('/data')

# Create a file using pathlib
config_file = data_dir / 'config.json'
config_file.write_text('{"debug": true, "version": "1.0"}')

# Read file using pathlib
content = config_file.read_text()
print('Config:', content)

# List files with glob
txt_files = list(data_dir.glob('*.txt'))
print('Text files:', txt_files)

# Create directory
project_dir = Path('/workspace/myproject')
project_dir.mkdir(exist_ok=True)
```

## JavaScript Usage

### Node.js-style File Operations

```javascript
const fs = require('fs');

// Write a text file (synchronous)
fs.writeFileSync('/data/example.txt', 'Hello from JavaScript!');

// Read a text file (synchronous)
const content = fs.readFileSync('/data/example.txt', 'utf8');
console.log(content); // Output: Hello from JavaScript!

// Check if file exists
if (fs.existsSync('/data/example.txt')) {
    console.log('File exists!');
}

// Get file statistics
const stats = fs.statSync('/data/example.txt');
console.log('Is file:', stats.isFile);
console.log('Is directory:', stats.isDirectory);
```

### Async File Operations

```javascript
const fs = require('fs');

async function fileOperations() {
    try {
        // Write file asynchronously
        await fs.writeFile('/data/async.txt', 'Async content');
        
        // Read file asynchronously
        const content = await fs.readFile('/data/async.txt', 'utf8');
        console.log('Async content:', content);
        
        // List directory contents
        const files = await fs.readdir('/data');
        console.log('Files:', files);
        
    } catch (error) {
        console.error('Error:', error.message);
    }
}

fileOperations();
```

### Working with JSON

```javascript
const fs = require('fs');

// Create and save JSON data
const users = [
    { id: 1, name: 'John', email: 'john@example.com' },
    { id: 2, name: 'Jane', email: 'jane@example.com' }
];

fs.writeFileSync('/data/users.json', JSON.stringify(users, null, 2));

// Read and parse JSON
const loadedUsers = JSON.parse(fs.readFileSync('/data/users.json', 'utf8'));
console.log('Users:', loadedUsers);

// Filter and save subset
const johnUser = loadedUsers.filter(user => user.name === 'John');
fs.writeFileSync('/data/john.json', JSON.stringify(johnUser, null, 2));
```

### Directory Operations

```javascript
const fs = require('fs');
const path = require('path');

// Create directories
fs.mkdirSync('/workspace/app', { recursive: true });
fs.mkdirSync('/workspace/app/src', { recursive: true });

// List directory contents
const files = fs.readdirSync('/workspace');
console.log('Workspace contents:', files);

// Working directory operations
console.log('Current directory:', process.cwd());
process.chdir('/workspace/app');
console.log('Changed to:', process.cwd());
```

### Path Utilities

```javascript
const path = require('path');

// Join paths
const filePath = path.join('/data', 'subfolder', 'file.txt');
console.log('Joined path:', filePath); // /data/subfolder/file.txt

// Get directory name
console.log('Directory:', path.dirname('/data/file.txt')); // /data

// Get file name
console.log('Basename:', path.basename('/data/file.txt')); // file.txt

// Get file extension
console.log('Extension:', path.extname('/data/file.txt')); // .txt

// Check if path is absolute
console.log('Is absolute:', path.isAbsolute('/data/file.txt')); // true
```

## Cross-Language File Sharing

Files created in one language can be accessed from the other:

### Python → JavaScript

```python
# In Python: Create data
import json

data = {"message": "Hello from Python!", "numbers": [1, 2, 3, 4, 5]}
with open('/data/shared.json', 'w') as f:
    json.dump(data, f)
```

```javascript
// In JavaScript: Read the data
const fs = require('fs');

const data = JSON.parse(fs.readFileSync('/data/shared.json', 'utf8'));
console.log('Message from Python:', data.message);
console.log('Sum of numbers:', data.numbers.reduce((a, b) => a + b, 0));
```

### JavaScript → Python

```javascript
// In JavaScript: Create CSV data
const fs = require('fs');

const csvData = [
    'name,score',
    'Alice,95',
    'Bob,87',
    'Charlie,92'
].join('\n');

fs.writeFileSync('/data/scores.csv', csvData);
```

```python
# In Python: Process the CSV
import pandas as pd

df = pd.read_csv('/data/scores.csv')
print('Scores:')
print(df)

average_score = df['score'].mean()
print(f'Average score: {average_score:.1f}')
```

## Best Practices

### File Organization

- Use `/data` for persistent data files
- Use `/tmp` for temporary/intermediate files
- Use `/workspace` for project files and code
- Create subdirectories to organize related files

### Error Handling

```python
# Python error handling
try:
    with open('/data/config.json', 'r') as f:
        config = json.load(f)
except FileNotFoundError:
    print('Config file not found, using defaults')
    config = {'debug': False}
```

```javascript
// JavaScript error handling
const fs = require('fs');

try {
    const config = JSON.parse(fs.readFileSync('/data/config.json', 'utf8'));
    console.log('Config loaded:', config);
} catch (error) {
    if (error.message.includes('File not found')) {
        console.log('Config file not found, using defaults');
        const defaultConfig = { debug: false };
        fs.writeFileSync('/data/config.json', JSON.stringify(defaultConfig, null, 2));
    } else {
        console.error('Error reading config:', error.message);
    }
}
```

### Performance Tips

- Use absolute paths (starting with `/`) for clarity
- Batch multiple file operations when possible
- Use appropriate file encodings for your data
- Close files promptly (automatic with `with` statements in Python)

## Limitations

- **Memory-based**: Files exist only in memory during the session
- **Size limits**: Individual files are limited to 10MB, total storage to 50-100MB
- **No persistence**: Files are lost when the session ends (future versions will add persistence)
- **No network access**: Cannot read from URLs or external sources directly
- **Limited binary support**: Basic binary file support (images, etc.)

## Common Use Cases

### Data Analysis Workflow

1. **Data Preparation**: Create or import CSV/JSON data files
2. **Processing**: Use pandas (Python) or native parsing (JavaScript) 
3. **Analysis**: Perform calculations and transformations
4. **Export**: Save results in various formats
5. **Visualization**: Create charts and save as images

### Configuration Management

1. **Settings**: Store application configuration in JSON files
2. **Environments**: Manage different configurations for development/production
3. **Validation**: Load and validate configuration on startup

### File Processing Pipeline

1. **Input**: Read data from multiple sources
2. **Transform**: Process and clean the data
3. **Aggregate**: Combine results from multiple files
4. **Output**: Export processed data in the desired format

## Troubleshooting

### Common Issues

**File not found errors**:
- Check the file path is correct and starts with `/`
- Ensure the file was created successfully
- Verify the directory exists

**Permission errors**:
- All files in VFS are readable and writable
- If you see permission errors, it's likely a path issue

**Memory limits**:
- Files are limited to 10MB each
- Total storage is limited to 50-100MB
- Use streaming for large datasets when possible

**Path issues**:
- Always use forward slashes (`/`) in paths
- Use absolute paths starting with `/`
- Avoid using `..` or `.` in paths

### Getting Help

If you encounter issues with the VFS:

1. Check that your file paths are absolute (start with `/`)
2. Verify the file exists using `os.path.exists()` (Python) or `fs.existsSync()` (JavaScript)
3. Try creating a simple test file first to verify VFS is working
4. Check the file size limits if you're working with large files

The VFS provides a powerful way to work with files in a secure, isolated environment. Use it to enhance your data processing workflows, configuration management, and file-based operations in both Python and JavaScript!