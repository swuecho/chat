/**
 * VFS Test Examples
 * Demonstrates usage of the Virtual File System in both Python and JavaScript runners
 */

// Python VFS Examples
const pythonVFSExamples = [
  {
    title: "Basic File Operations",
    code: `
# Write a simple text file
with open('/data/hello.txt', 'w') as f:
    f.write('Hello, Virtual File System!')

# Read the file back
with open('/data/hello.txt', 'r') as f:
    content = f.read()
    print(f"File content: {content}")

# Check if file exists
import os
print(f"File exists: {os.path.exists('/data/hello.txt')}")
`
  },
  {
    title: "CSV Data Processing",
    code: `
import csv
import os

# Create sample data
data = [
    ['Name', 'Age', 'City'],
    ['Alice', 25, 'New York'],
    ['Bob', 30, 'San Francisco'],
    ['Charlie', 35, 'Chicago']
]

# Write CSV file
with open('/data/people.csv', 'w', newline='') as f:
    writer = csv.writer(f)
    writer.writerows(data)

# Read and process CSV
with open('/data/people.csv', 'r') as f:
    reader = csv.DictReader(f)
    for row in reader:
        print(f"{row['Name']} is {row['Age']} years old and lives in {row['City']}")

# List directory contents
print("Files in /data:")
for file in os.listdir('/data'):
    print(f"  {file}")
`
  },
  {
    title: "JSON Configuration",
    code: `
import json
import os

# Create configuration
config = {
    "app_name": "VFS Demo",
    "version": "1.0.0",
    "debug": True,
    "database": {
        "host": "localhost",
        "port": 5432,
        "name": "demo_db"
    }
}

# Create config directory
os.makedirs('/workspace/config', exist_ok=True)

# Save configuration
with open('/workspace/config/app.json', 'w') as f:
    json.dump(config, f, indent=2)

# Load and modify configuration
with open('/workspace/config/app.json', 'r') as f:
    loaded_config = json.load(f)

loaded_config['debug'] = False
loaded_config['version'] = '1.0.1'

# Save updated configuration
with open('/workspace/config/app.json', 'w') as f:
    json.dump(loaded_config, f, indent=2)

print("Configuration updated successfully")
print(f"New version: {loaded_config['version']}")
`
  },
  {
    title: "Working with pandas and VFS",
    code: `
import pandas as pd
import json

# Create sample data
data = {
    'product': ['A', 'B', 'C', 'D', 'E'],
    'sales': [100, 150, 80, 200, 120],
    'profit': [20, 30, 15, 40, 25]
}

df = pd.DataFrame(data)

# Save to CSV
df.to_csv('/data/sales.csv', index=False)

# Save to JSON
df.to_json('/data/sales.json', orient='records', indent=2)

# Read back from CSV
df_csv = pd.read_csv('/data/sales.csv')
print("Data from CSV:")
print(df_csv)

# Calculate summary statistics
summary = df_csv.describe()
print("\\nSummary statistics:")
print(summary)

# Save summary
summary.to_csv('/data/sales_summary.csv')
print("\\nSummary saved to /data/sales_summary.csv")
`
  },
  {
    title: "File System Navigation",
    code: `
import os
from pathlib import Path

# Show current directory
print(f"Current directory: {os.getcwd()}")

# Create directory structure
os.makedirs('/workspace/project/src', exist_ok=True)
os.makedirs('/workspace/project/tests', exist_ok=True)
os.makedirs('/workspace/project/docs', exist_ok=True)

# Create some files
files_to_create = [
    '/workspace/project/README.md',
    '/workspace/project/src/main.py',
    '/workspace/project/src/utils.py',
    '/workspace/project/tests/test_main.py',
    '/workspace/project/docs/api.md'
]

for file_path in files_to_create:
    with open(file_path, 'w') as f:
        f.write(f"# {Path(file_path).name}\\n\\nContent for {file_path}")

# Navigate and explore
os.chdir('/workspace/project')
print(f"\\nChanged to: {os.getcwd()}")

print("\\nProject structure:")
for root, dirs, files in os.walk('.'):
    level = root.replace('.', '').count(os.sep)
    indent = ' ' * 2 * level
    print(f"{indent}{os.path.basename(root)}/")
    subindent = ' ' * 2 * (level + 1)
    for file in files:
        print(f"{subindent}{file}")
`
  }
]

// JavaScript VFS Examples
const javascriptVFSExamples = [
  {
    title: "Node.js-style File Operations",
    code: `
const fs = require('fs');

// Write a simple text file
fs.writeFileSync('/data/greeting.txt', 'Hello from JavaScript VFS!');

// Read the file back
const content = fs.readFileSync('/data/greeting.txt', 'utf8');
console.log('File content:', content);

// Check if file exists
console.log('File exists:', fs.existsSync('/data/greeting.txt'));

// Get file stats
const stats = fs.statSync('/data/greeting.txt');
console.log('Is file:', stats.isFile);
console.log('Is directory:', stats.isDirectory);
`
  },
  {
    title: "JSON Data Management",
    code: `
const fs = require('fs');

// Create user data
const users = [
  { id: 1, name: 'John Doe', email: 'john@example.com', active: true },
  { id: 2, name: 'Jane Smith', email: 'jane@example.com', active: false },
  { id: 3, name: 'Bob Johnson', email: 'bob@example.com', active: true }
];

// Save users to JSON file
fs.writeFileSync('/data/users.json', JSON.stringify(users, null, 2));

// Read and filter users
const loadedUsers = JSON.parse(fs.readFileSync('/data/users.json', 'utf8'));
const activeUsers = loadedUsers.filter(user => user.active);

console.log('Active users:', activeUsers);

// Save filtered results
fs.writeFileSync('/data/active_users.json', JSON.stringify(activeUsers, null, 2));

console.log('Filtered data saved to /data/active_users.json');
`
  },
  {
    title: "CSV Processing",
    code: `
const fs = require('fs');

// Create CSV data
const csvData = [
  'Name,Age,Department',
  'Alice Johnson,28,Engineering',
  'Bob Smith,32,Marketing',
  'Carol Brown,29,Design',
  'David Wilson,35,Engineering'
].join('\\n');

// Write CSV file
fs.writeFileSync('/data/employees.csv', csvData);

// Read and parse CSV
const csvContent = fs.readFileSync('/data/employees.csv', 'utf8');
const lines = csvContent.split('\\n');
const headers = lines[0].split(',');
const employees = lines.slice(1).map(line => {
  const values = line.split(',');
  const employee = {};
  headers.forEach((header, index) => {
    employee[header] = values[index];
  });
  return employee;
});

console.log('Employees:', employees);

// Filter by department
const engineers = employees.filter(emp => emp.Department === 'Engineering');
console.log('Engineers:', engineers);

// Calculate average age
const avgAge = employees.reduce((sum, emp) => sum + parseInt(emp.Age), 0) / employees.length;
console.log('Average age:', avgAge.toFixed(1));
`
  },
  {
    title: "Directory Operations",
    code: `
const fs = require('fs');
const path = require('path');

// Create directory structure
fs.mkdirSync('/workspace/myapp', { recursive: true });
fs.mkdirSync('/workspace/myapp/src', { recursive: true });
fs.mkdirSync('/workspace/myapp/public', { recursive: true });
fs.mkdirSync('/workspace/myapp/config', { recursive: true });

// Create package.json
const packageJson = {
  name: 'my-vfs-app',
  version: '1.0.0',
  description: 'A demo app using VFS',
  main: 'src/index.js',
  scripts: {
    start: 'node src/index.js'
  }
};

fs.writeFileSync('/workspace/myapp/package.json', JSON.stringify(packageJson, null, 2));

// Create main application file
const appCode = \`
console.log('Welcome to VFS App!');
console.log('Current directory:', process.cwd());

const fs = require('fs');
const config = JSON.parse(fs.readFileSync('./config/app.json', 'utf8'));
console.log('App config:', config);
\`;

fs.writeFileSync('/workspace/myapp/src/index.js', appCode);

// Create configuration
const appConfig = {
  port: 3000,
  environment: 'development',
  features: {
    logging: true,
    debugging: true
  }
};

fs.writeFileSync('/workspace/myapp/config/app.json', JSON.stringify(appConfig, null, 2));

// List directory contents
console.log('Project structure:');
function listDirectory(dir, indent = '') {
  const items = fs.readdirSync(dir);
  items.forEach(item => {
    const itemPath = path.join(dir, item);
    const stats = fs.statSync(itemPath);
    if (stats.isDirectory) {
      console.log(\`\${indent}\${item}/\`);
      listDirectory(itemPath, indent + '  ');
    } else {
      console.log(\`\${indent}\${item}\`);
    }
  });
}

listDirectory('/workspace/myapp');
`
  },
  {
    title: "Async File Operations",
    code: `
const fs = require('fs');

async function demonstrateAsyncOps() {
  try {
    // Write multiple files asynchronously
    const tasks = [
      fs.writeFile('/tmp/file1.txt', 'Content of file 1'),
      fs.writeFile('/tmp/file2.txt', 'Content of file 2'),
      fs.writeFile('/tmp/file3.txt', 'Content of file 3')
    ];
    
    await Promise.all(tasks);
    console.log('All files written successfully');
    
    // Read files asynchronously
    const files = await fs.readdir('/tmp');
    console.log('Files in /tmp:', files);
    
    // Read file contents
    for (const file of files) {
      if (file.startsWith('file')) {
        const content = await fs.readFile(\`/tmp/\${file}\`, 'utf8');
        console.log(\`\${file}: \${content}\`);
      }
    }
    
    // Demonstrate fs.promises API
    const content = await fs.promises.readFile('/tmp/file1.txt', 'utf8');
    console.log('Using fs.promises:', content);
    
  } catch (error) {
    console.error('Error:', error.message);
  }
}

// Run the async demonstration
demonstrateAsyncOps();
`
  }
]

// Export for use in documentation
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { pythonVFSExamples, javascriptVFSExamples }
} else {
  window.VFSExamples = { pythonVFSExamples, javascriptVFSExamples }
}