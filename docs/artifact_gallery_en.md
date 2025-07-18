# Artifact Gallery

## Overview

The Artifact Gallery is a comprehensive management interface for viewing, organizing, and interacting with code artifacts generated during chat conversations. It provides a centralized location to browse, execute, and manage all artifacts created across different chat sessions.

## Features

### 1. **Artifact Management**
- **Browse artifacts** from all chat sessions in one place
- **Filter and search** by type, language, session, or content
- **Sort artifacts** by creation date, type, execution count, or rating
- **View statistics** about artifact usage and performance
- **Export artifacts** to JSON format for backup or sharing

### 2. **Code Execution**
- **Run JavaScript/Python code** directly in the gallery
- **Real-time output** with syntax highlighting
- **Error handling** with detailed error messages
- **Performance metrics** including execution time
- **Library support** for popular packages (lodash, d3, numpy, pandas, etc.)

### 3. **Artifact Viewing**
- **HTML rendering** in secure iframe sandbox
- **SVG visualization** with proper scaling
- **JSON formatting** with validation
- **Mermaid diagram** rendering
- **Syntax highlighting** for code artifacts

### 4. **Organization Tools**
- **Grid and list views** for different browsing preferences
- **Pagination** for large artifact collections
- **Tag-based categorization** with automatic tag generation
- **Session context** showing which chat created each artifact
- **Duplicate detection** and management

## Accessing the Gallery

### From Chat Interface
1. Click the **Gallery** button (gallery icon) in the chat footer
2. The gallery will open in place of the chat interface
3. Use the same gallery button to return to chat view

### Navigation
- **Toggle between chat and gallery** using the gallery button
- **Switch between grid and list views** using the view toggle
- **Use filters** to narrow down artifacts by specific criteria

## Artifact Types

### Executable Artifacts
These artifacts can be run directly in the gallery:

#### JavaScript/TypeScript
- **Supported features**: ES6+, async/await, classes, modules
- **Available libraries**: lodash, d3, chart.js, moment, axios, rxjs, p5, three, fabric
- **Canvas support**: Create interactive graphics and visualizations
- **Execution environment**: Secure Web Worker sandbox

#### Python
- **Supported features**: Python 3.x syntax, scientific computing
- **Available packages**: numpy, pandas, matplotlib, scipy, scikit-learn, requests, beautifulsoup4, pillow, sympy, networkx, seaborn, plotly, bokeh, altair
- **Plot support**: Matplotlib plots rendered as images
- **Execution environment**: Pyodide-based Python interpreter

### Viewable Artifacts
These artifacts are rendered for visual inspection:

#### HTML
- **Secure rendering**: Iframe sandbox with restricted permissions
- **Full HTML support**: CSS, basic JavaScript, forms, modals
- **Responsive design**: Adapts to different screen sizes
- **External links**: Open in new window capability

#### SVG
- **Vector graphics**: Scalable and crisp rendering
- **Interactive elements**: Hover effects and basic interactions
- **Theme adaptation**: Adjusts colors for dark/light themes
- **Export capability**: Download as SVG files

#### JSON
- **Formatted display**: Pretty-printed with syntax highlighting
- **Validation**: Automatic JSON syntax validation
- **Copy functionality**: Easy copying of formatted JSON
- **Large file support**: Handles large JSON structures efficiently

#### Mermaid
- **Diagram types**: Flowcharts, sequence diagrams, class diagrams, etc.
- **Auto-rendering**: Automatic conversion to visual diagrams
- **Responsive**: Adapts to container size
- **Theme support**: Follows application theme

## Interface Elements

### Gallery Header
- **Title and count**: Shows total number of artifacts
- **Action buttons**: Access to filters, statistics, and export
- **Search bar**: Quick text search across all artifacts

### Filter Panel
- **Search**: Text search across titles, content, and tags
- **Type filter**: Filter by artifact type (code, HTML, SVG, etc.)
- **Language filter**: Filter by programming language
- **Session filter**: Filter by chat session
- **Date range**: Filter by creation date
- **Sort options**: Multiple sorting criteria

### Statistics Panel
- **Total artifacts**: Overall count of artifacts
- **Execution stats**: Total runs, average execution time, success rate
- **Type breakdown**: Distribution of artifact types
- **Language distribution**: Popular programming languages
- **Performance charts**: Visual representation of usage patterns

### Artifact Cards (Grid View)
- **Artifact preview**: Truncated code or content preview
- **Metadata**: Creation date, language, session info
- **Action buttons**: Preview, Run/View, Edit, Duplicate, Delete
- **Tags**: Automatically generated and custom tags
- **Execution count**: Number of times the artifact has been run

### Artifact Rows (List View)
- **Compact layout**: More artifacts visible at once
- **Essential info**: Title, type, language, creation date
- **Quick actions**: All management functions accessible
- **Session context**: Clear indication of source chat session

## Actions and Operations

### Running Code Artifacts
1. **Click the Run button** (play icon) on JavaScript/Python artifacts
2. **Run modal opens** with code preview and execution controls
3. **Execute code** using the "Run Code" button
4. **View output** in real-time with syntax highlighting
5. **Clear results** to run again with fresh environment

### Viewing Visual Artifacts
1. **Click the View button** (external link icon) on HTML/SVG/JSON artifacts
2. **View modal opens** with proper rendering
3. **Interact with content** (for HTML artifacts)
4. **Copy content** using the copy button
5. **Close modal** to return to gallery

### Managing Artifacts
- **Preview**: Quick view of artifact content
- **Edit**: Modify artifact content and save changes
- **Duplicate**: Create a copy for experimentation
- **Delete**: Remove artifact permanently (with confirmation)
- **Copy**: Copy artifact content to clipboard
- **Download**: Save artifact as file with proper extension

## Advanced Features

### Automatic Tagging
The gallery automatically generates tags based on:
- **Programming language** (javascript, python, html, etc.)
- **Framework usage** (react, vue, pandas, matplotlib, etc.)
- **Code patterns** (async, functions, classes, loops, etc.)
- **Libraries** (lodash, d3, numpy, etc.)
- **Execution results** (error, success, slow, fast, etc.)

### Session Context
Each artifact maintains connection to its source:
- **Session title**: Name of the chat session
- **Message context**: Which message created the artifact
- **Timestamp**: When the artifact was created
- **Edit tracking**: Whether the artifact has been modified

### Performance Tracking
- **Execution time**: How long code takes to run
- **Success rate**: Percentage of successful executions
- **Error patterns**: Common error types and frequencies
- **Usage statistics**: Most frequently run artifacts

### Export and Backup
- **JSON export**: Complete artifact data with metadata
- **Filtered export**: Export only selected or filtered artifacts
- **Backup format**: Structured for easy import/restore
- **Sharing**: Share artifact collections with others

## Best Practices

### Organization
- **Use descriptive titles** for artifacts when creating them
- **Add custom tags** for better organization
- **Regular cleanup** of outdated or experimental artifacts
- **Session naming** to provide better context

### Code Execution
- **Test incrementally** when developing complex code
- **Use console.log** for debugging JavaScript
- **Handle errors gracefully** in your code
- **Be mindful of execution time** for complex operations

### Performance
- **Use pagination** for large artifact collections
- **Clear execution results** when not needed
- **Filter artifacts** to focus on relevant items
- **Regular exports** for backup and archival

## Security Considerations

### Code Execution
- **Sandboxed environment**: All code runs in isolated Web Workers
- **No network access**: Direct network requests are blocked
- **No file system access**: Cannot read/write local files
- **Limited DOM access**: Cannot modify the parent application
- **Timeout protection**: Long-running code is automatically terminated

### Content Rendering
- **HTML sandboxing**: Iframe sandbox prevents malicious scripts
- **SVG sanitization**: Removes potentially dangerous elements
- **Content validation**: JSON and other formats are validated
- **Safe origins**: External resources are restricted

## Troubleshooting

### Common Issues
- **Code not running**: Check language support and syntax
- **Artifacts not loading**: Refresh the page or check browser console
- **Performance issues**: Reduce artifact count or use filters
- **Export failures**: Check browser download permissions

### Browser Compatibility
- **Modern browsers**: Chrome, Firefox, Safari, Edge (latest versions)
- **JavaScript required**: Gallery requires JavaScript to function
- **LocalStorage**: Used for preferences and temporary data
- **WebWorkers**: Required for code execution features

## API Integration

The gallery integrates with the chat application's API:

```javascript
// Example: Getting artifacts from a specific session
const artifacts = await fetch(`/api/chat/sessions/${sessionId}/artifacts`);

// Example: Executing code artifact
const result = await fetch(`/api/artifacts/${artifactId}/execute`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    code: artifactContent,
    language: 'javascript'
  })
});
```

## Future Enhancements

### Planned Features
- **Artifact versioning**: Track changes and maintain history
- **Collaborative editing**: Multiple users editing artifacts
- **Advanced analytics**: Detailed usage and performance metrics
- **Template system**: Create reusable artifact templates
- **AI suggestions**: Automatic code improvements and suggestions

### Community Features
- **Artifact sharing**: Public gallery for sharing useful artifacts
- **Rating system**: Community ratings for popular artifacts
- **Comments**: Collaborative discussion on artifacts
- **Collections**: Curated sets of related artifacts

This documentation provides a comprehensive guide to using the Artifact Gallery effectively. For technical implementation details, see the source code in `/src/views/chat/components/ArtifactGallery.vue`.