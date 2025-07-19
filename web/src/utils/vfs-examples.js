/**
 * VFS Import/Export Examples
 * Demonstrates all import/export capabilities of the Virtual File System
 */

// Example usage of VFS Import/Export system
const vfsImportExportExamples = {
  
  // Basic file upload and download
  fileOperations: {
    title: "Basic File Upload/Download",
    description: "Upload files from your computer and download files from VFS",
    javascript: `
// Initialize VFS and Import/Export
const vfs = new VirtualFileSystem();
const importExport = new VFSImportExport(vfs);

// Simulate file upload (in real usage, this would be from a file input)
const fileContent = "Hello, VFS!\\nThis is uploaded content.";
const mockFile = new Blob([fileContent], { type: 'text/plain' });
mockFile.name = 'uploaded.txt';

// Upload file to VFS
const uploadResult = await importExport.uploadFile(mockFile, '/data/uploaded.txt');
console.log('Upload result:', uploadResult);

// Verify file exists in VFS
console.log('File exists:', vfs.exists('/data/uploaded.txt'));

// Read file content from VFS
const content = await vfs.readFile('/data/uploaded.txt', 'utf8');
console.log('File content:', content);

// Download file (in real usage, this would trigger browser download)
const downloadResult = await importExport.downloadFile('/data/uploaded.txt');
console.log('Download result:', downloadResult);
`,
    python: `
# In Python, files uploaded via the UI are automatically available
import os

# Check if uploaded file exists
if os.path.exists('/data/uploaded.txt'):
    print("Uploaded file is available in Python!")
    
    # Read the uploaded content
    with open('/data/uploaded.txt', 'r') as f:
        content = f.read()
        print(f"Content: {content}")
    
    # Process and save new file
    processed_content = content.upper()
    with open('/data/processed.txt', 'w') as f:
        f.write(processed_content)
    
    print("Processed file saved to /data/processed.txt")
`
  },

  // Data format conversion
  dataConversion: {
    title: "Data Format Conversion",
    description: "Convert between CSV, JSON, XML formats",
    javascript: `
const fs = require('fs');

// Create sample CSV data
const csvData = \`name,age,city,salary
John Doe,30,New York,75000
Jane Smith,25,San Francisco,85000
Bob Johnson,35,Chicago,65000
Alice Brown,28,Boston,70000\`;

// Write CSV file
fs.writeFileSync('/data/employees.csv', csvData);

// Convert CSV to JSON using import/export system
const convertResult = await importExport.convertToFormat('/data/employees.csv', 'json', '/data/employees.json');
console.log('Conversion result:', convertResult);

// Read and display JSON data
const jsonData = fs.readFileSync('/data/employees.json', 'utf8');
console.log('JSON data:', JSON.parse(jsonData));

// Convert back to CSV with different name
await importExport.convertToFormat('/data/employees.json', 'csv', '/data/employees_converted.csv');

// Verify the round-trip conversion
const convertedCsv = fs.readFileSync('/data/employees_converted.csv', 'utf8');
console.log('Converted back to CSV:', convertedCsv);
`,
    python: `
import json
import csv
import pandas as pd

# Read the JSON file created by JavaScript
with open('/data/employees.json', 'r') as f:
    employees = json.load(f)

print("Employee data from JSON:")
for employee in employees:
    print(f"  {employee['name']}: {employee['salary']}")

# Use pandas for advanced processing
df = pd.read_csv('/data/employees.csv')

# Calculate statistics
avg_salary = df['salary'].mean()
max_salary = df['salary'].max()
min_salary = df['salary'].min()

# Create summary report
summary = {
    "total_employees": len(df),
    "average_salary": avg_salary,
    "max_salary": max_salary,
    "min_salary": min_salary,
    "cities": df['city'].unique().tolist(),
    "avg_age": df['age'].mean()
}

# Save summary as JSON
with open('/data/salary_summary.json', 'w') as f:
    json.dump(summary, f, indent=2)

print("Summary report saved to /data/salary_summary.json")

# Export high earners to separate CSV
high_earners = df[df['salary'] > 70000]
high_earners.to_csv('/data/high_earners.csv', index=False)

print(f"Found {len(high_earners)} high earners")
`
  },

  // Import from URL
  urlImport: {
    title: "Import Data from URL",
    description: "Fetch data from external URLs",
    javascript: `
// Import CSV data from a public dataset URL
const publicDataUrl = 'https://raw.githubusercontent.com/holtzy/data_to_viz/master/Example_dataset/7_OneCatOneNum_header.csv';

try {
  const importResult = await importExport.importFromURL(publicDataUrl, '/data/public_dataset.csv');
  
  if (importResult.success) {
    console.log('Import successful:', importResult);
    
    // Process the imported data
    const csvContent = fs.readFileSync('/data/public_dataset.csv', 'utf8');
    const lines = csvContent.split('\\n');
    console.log(\`Imported \${lines.length - 1} data rows\`);
    console.log('First few lines:', lines.slice(0, 5));
    
    // Convert to JSON for easier processing
    await importExport.convertToFormat('/data/public_dataset.csv', 'json', '/data/public_dataset.json');
    
  } else {
    console.error('Import failed:', importResult.message);
  }
} catch (error) {
  console.error('URL import error:', error.message);
  
  // Fallback: Create sample data
  console.log('Creating sample data instead...');
  const sampleData = \`group,value
A,23
B,45
C,56
D,78
E,32\`;
  
  fs.writeFileSync('/data/sample_dataset.csv', sampleData);
  console.log('Sample data created at /data/sample_dataset.csv');
}
`,
    python: `
import pandas as pd
import json

# Check if the imported data exists
import os
if os.path.exists('/data/public_dataset.csv'):
    print("Processing imported public dataset...")
    df = pd.read_csv('/data/public_dataset.csv')
    
    # Analyze the data
    print(f"Dataset shape: {df.shape}")
    print("\\nColumn info:")
    print(df.info())
    
    print("\\nFirst few rows:")
    print(df.head())
    
    # Generate statistics
    stats = df.describe()
    print("\\nStatistics:")
    print(stats)
    
    # Save analysis results
    analysis = {
        "shape": df.shape,
        "columns": df.columns.tolist(),
        "dtypes": df.dtypes.astype(str).to_dict(),
        "null_counts": df.isnull().sum().to_dict(),
        "summary_stats": stats.to_dict() if len(df.select_dtypes(include='number').columns) > 0 else {}
    }
    
    with open('/data/dataset_analysis.json', 'w') as f:
        json.dump(analysis, f, indent=2)
    
    print("Analysis saved to /data/dataset_analysis.json")

elif os.path.exists('/data/sample_dataset.csv'):
    print("Processing sample dataset...")
    df = pd.read_csv('/data/sample_dataset.csv')
    
    # Simple analysis of sample data
    print("Sample data overview:")
    print(df)
    
    # Calculate total and mean
    total_value = df['value'].sum()
    mean_value = df['value'].mean()
    
    print(f"\\nTotal value: {total_value}")
    print(f"Mean value: {mean_value:.2f}")
    
    # Create visualization data
    viz_data = df.to_dict('records')
    with open('/data/visualization_data.json', 'w') as f:
        json.dump(viz_data, f, indent=2)
    
    print("Visualization data saved to /data/visualization_data.json")
else:
    print("No dataset found. Please run the JavaScript import first.")
`
  },

  // Session management
  sessionManagement: {
    title: "Session Import/Export",
    description: "Save and restore entire VFS sessions",
    javascript: `
// Create a complex file structure for demo
const fs = require('fs');

// Setup project structure
fs.mkdirSync('/workspace/myproject', { recursive: true });
fs.mkdirSync('/workspace/myproject/src', { recursive: true });
fs.mkdirSync('/workspace/myproject/data', { recursive: true });
fs.mkdirSync('/workspace/myproject/docs', { recursive: true });

// Create project files
const packageJson = {
  name: "vfs-demo-project",
  version: "1.0.0",
  description: "A demo project using VFS",
  main: "src/index.js",
  scripts: {
    start: "node src/index.js",
    test: "echo \\"No tests yet\\""
  }
};

fs.writeFileSync('/workspace/myproject/package.json', JSON.stringify(packageJson, null, 2));

const mainCode = \`
console.log('Welcome to VFS Demo Project!');

const fs = require('fs');
const path = require('path');

// Read project configuration
const packageInfo = JSON.parse(fs.readFileSync('./package.json', 'utf8'));
console.log(\`Project: \${packageInfo.name} v\${packageInfo.version}\`);

// Read data files
const dataDir = './data';
if (fs.existsSync(dataDir)) {
  const dataFiles = fs.readdirSync(dataDir);
  console.log(\`Found \${dataFiles.length} data files:\`, dataFiles);
}
\`;

fs.writeFileSync('/workspace/myproject/src/index.js', mainCode);

// Create README
const readme = \`# VFS Demo Project

This is a demonstration project showing the capabilities of the Virtual File System.

## Features

- File management
- Data processing
- Session persistence

## Usage

\\\`\\\`\\\`bash
npm start
\\\`\\\`\\\`

## Files

- \`src/index.js\` - Main application
- \`data/\` - Data files
- \`docs/\` - Documentation
\`;

fs.writeFileSync('/workspace/myproject/README.md', readme);

// Create some data files
const sampleData = [
  { id: 1, name: 'Sample 1', value: 100 },
  { id: 2, name: 'Sample 2', value: 200 },
  { id: 3, name: 'Sample 3', value: 300 }
];

fs.writeFileSync('/workspace/myproject/data/samples.json', JSON.stringify(sampleData, null, 2));

// Create CSV data
const csvData = 'id,name,value\\n1,Sample 1,100\\n2,Sample 2,200\\n3,Sample 3,300';
fs.writeFileSync('/workspace/myproject/data/samples.csv', csvData);

console.log('Project structure created!');

// Export the entire VFS session
const exportResult = await importExport.exportVFSSession('demo-project');
console.log('Session export result:', exportResult);

// Show current VFS statistics
const stats = importExport.getImportStats();
console.log('VFS Statistics:', stats);
`,
    python: `
import os
import json

# Explore the project structure created by JavaScript
project_root = '/workspace/myproject'

if os.path.exists(project_root):
    print("Project structure:")
    
    # Walk through all files and directories
    for root, dirs, files in os.walk(project_root):
        level = root.replace(project_root, '').count(os.sep)
        indent = ' ' * 2 * level
        print(f"{indent}{os.path.basename(root)}/")
        
        sub_indent = ' ' * 2 * (level + 1)
        for file in files:
            print(f"{sub_indent}{file}")
    
    print("\\n" + "="*50 + "\\n")
    
    # Read and analyze project files
    package_file = os.path.join(project_root, 'package.json')
    if os.path.exists(package_file):
        with open(package_file, 'r') as f:
            package_info = json.load(f)
        print(f"Project: {package_info['name']} v{package_info['version']}")
        print(f"Description: {package_info['description']}")
    
    # Process data files
    data_dir = os.path.join(project_root, 'data')
    if os.path.exists(data_dir):
        print(f"\\nData files in {data_dir}:")
        
        for file in os.listdir(data_dir):
            file_path = os.path.join(data_dir, file)
            print(f"  {file}")
            
            if file.endswith('.json'):
                with open(file_path, 'r') as f:
                    data = json.load(f)
                print(f"    JSON records: {len(data)}")
            elif file.endswith('.csv'):
                with open(file_path, 'r') as f:
                    lines = f.readlines()
                print(f"    CSV rows: {len(lines) - 1}")  # Subtract header
    
    # Create a Python analysis report
    report = {
        "analysis_date": "2024-01-01",
        "project_name": package_info.get('name', 'unknown'),
        "total_files": sum(len(files) for _, _, files in os.walk(project_root)),
        "total_directories": sum(len(dirs) for _, dirs, _ in os.walk(project_root)),
        "file_types": {}
    }
    
    # Count file types
    for root, dirs, files in os.walk(project_root):
        for file in files:
            ext = os.path.splitext(file)[1] or 'no_extension'
            report["file_types"][ext] = report["file_types"].get(ext, 0) + 1
    
    # Save analysis report
    report_path = os.path.join(project_root, 'docs', 'analysis_report.json')
    os.makedirs(os.path.dirname(report_path), exist_ok=True)
    
    with open(report_path, 'w') as f:
        json.dump(report, f, indent=2)
    
    print(f"\\nAnalysis report saved to {report_path}")
    print("Report contents:", json.dumps(report, indent=2))

else:
    print("Project not found. Please run the JavaScript setup first.")
`
  },

  // Multiple file upload
  bulkOperations: {
    title: "Bulk Upload and Processing",
    description: "Handle multiple files and batch operations",
    javascript: `
// Simulate multiple file upload
const files = [
  { name: 'config.json', content: '{"debug": true, "version": "1.0"}', type: 'application/json' },
  { name: 'users.csv', content: 'id,name,email\\n1,John,john@example.com\\n2,Jane,jane@example.com', type: 'text/csv' },
  { name: 'readme.txt', content: 'This is a readme file\\nwith multiple lines\\nof documentation.', type: 'text/plain' },
  { name: 'data.log', content: '2024-01-01 10:00:00 INFO Application started\\n2024-01-01 10:01:00 INFO User logged in\\n2024-01-01 10:02:00 ERROR Database connection failed', type: 'text/plain' }
];

console.log('Uploading multiple files...');

const uploadResults = [];
for (const fileInfo of files) {
  // Create mock file object
  const mockFile = new Blob([fileInfo.content], { type: fileInfo.type });
  mockFile.name = fileInfo.name;
  
  // Upload to /data directory
  const result = await importExport.uploadFile(mockFile, \`/data/\${fileInfo.name}\`);
  uploadResults.push(result);
  
  console.log(\`Upload \${fileInfo.name}:\`, result.success ? 'SUCCESS' : 'FAILED');
}

console.log('\\nUpload Summary:');
const successful = uploadResults.filter(r => r.success).length;
const failed = uploadResults.filter(r => !r.success).length;
console.log(\`Successful: \${successful}, Failed: \${failed}\`);

// Process uploaded files
console.log('\\nProcessing uploaded files...');

// Read and parse JSON config
const config = JSON.parse(fs.readFileSync('/data/config.json', 'utf8'));
console.log('Config loaded:', config);

// Convert CSV to JSON
await importExport.convertToFormat('/data/users.csv', 'json', '/data/users.json');
const users = JSON.parse(fs.readFileSync('/data/users.json', 'utf8'));
console.log('Users:', users);

// Analyze log file
const logContent = fs.readFileSync('/data/data.log', 'utf8');
const logLines = logContent.split('\\n').filter(line => line.trim());
const errorCount = logLines.filter(line => line.includes('ERROR')).length;
const infoCount = logLines.filter(line => line.includes('INFO')).length;

console.log(\`Log analysis: \${infoCount} INFO, \${errorCount} ERROR messages\`);

// Create processing summary
const summary = {
  timestamp: new Date().toISOString(),
  files_processed: files.length,
  config: config,
  user_count: users.length,
  log_stats: { info: infoCount, errors: errorCount }
};

fs.writeFileSync('/data/processing_summary.json', JSON.stringify(summary, null, 2));
console.log('Processing summary saved to /data/processing_summary.json');
`,
    python: `
import os
import json
import csv
from datetime import datetime

# Process the uploaded files
data_dir = '/data'

print("Processing bulk uploaded files...")

# Check what files are available
if os.path.exists(data_dir):
    files = os.listdir(data_dir)
    print(f"Found {len(files)} files: {files}")
    
    # Read processing summary from JavaScript
    summary_file = os.path.join(data_dir, 'processing_summary.json')
    if os.path.exists(summary_file):
        with open(summary_file, 'r') as f:
            js_summary = json.load(f)
        print("\\nJavaScript processing summary:", js_summary)
    
    # Advanced processing with Python
    
    # 1. Enhanced user data processing
    users_file = os.path.join(data_dir, 'users.json')
    if os.path.exists(users_file):
        with open(users_file, 'r') as f:
            users = json.load(f)
        
        # Add domain analysis
        domains = {}
        for user in users:
            email = user.get('email', '')
            domain = email.split('@')[-1] if '@' in email else 'unknown'
            domains[domain] = domains.get(domain, 0) + 1
        
        print("\\nEmail domain analysis:", domains)
        
        # Create enhanced user report
        user_report = {
            "total_users": len(users),
            "domains": domains,
            "users_by_domain": {domain: [u for u in users if u.get('email', '').endswith(f'@{domain}')] 
                              for domain in domains if domain != 'unknown'}
        }
        
        with open(os.path.join(data_dir, 'user_analysis.json'), 'w') as f:
            json.dump(user_report, f, indent=2)
    
    # 2. Advanced log analysis
    log_file = os.path.join(data_dir, 'data.log')
    if os.path.exists(log_file):
        with open(log_file, 'r') as f:
            log_lines = f.readlines()
        
        # Parse log entries
        log_entries = []
        for line in log_lines:
            line = line.strip()
            if line:
                parts = line.split(' ', 3)
                if len(parts) >= 4:
                    log_entries.append({
                        'date': parts[0],
                        'time': parts[1],
                        'level': parts[2],
                        'message': parts[3]
                    })
        
        # Generate log statistics
        log_stats = {
            'total_entries': len(log_entries),
            'by_level': {},
            'by_hour': {},
            'errors': [entry for entry in log_entries if entry['level'] == 'ERROR']
        }
        
        for entry in log_entries:
            level = entry['level']
            hour = entry['time'][:2]  # Extract hour
            
            log_stats['by_level'][level] = log_stats['by_level'].get(level, 0) + 1
            log_stats['by_hour'][hour] = log_stats['by_hour'].get(hour, 0) + 1
        
        print("\\nAdvanced log analysis:")
        print(f"  Total entries: {log_stats['total_entries']}")
        print(f"  By level: {log_stats['by_level']}")
        print(f"  By hour: {log_stats['by_hour']}")
        
        # Save detailed log analysis
        with open(os.path.join(data_dir, 'log_analysis.json'), 'w') as f:
            json.dump(log_stats, f, indent=2)
    
    # 3. Create comprehensive processing report
    final_report = {
        "processed_at": datetime.now().isoformat(),
        "python_version": "3.x",
        "total_files": len(files),
        "file_list": files,
        "analyses_completed": [
            "user_domain_analysis",
            "log_pattern_analysis", 
            "configuration_validation"
        ]
    }
    
    with open(os.path.join(data_dir, 'python_processing_report.json'), 'w') as f:
        json.dump(final_report, f, indent=2)
    
    print("\\nPython processing complete!")
    print("Generated reports:")
    print("  - user_analysis.json")
    print("  - log_analysis.json") 
    print("  - python_processing_report.json")

else:
    print("Data directory not found. Please run the JavaScript upload first.")
`
  }
};

// Export examples
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { vfsImportExportExamples }
} else {
  window.VFSImportExportExamples = { vfsImportExportExamples }
}