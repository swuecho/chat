# Code Runner User Manual

## Overview

The Code Runner is an interactive JavaScript execution environment built into the chat application. It allows you to write, run, and experiment with JavaScript code directly in chat messages, with real-time output, graphics support, and access to popular libraries.

## Getting Started

### Creating Executable Code Artifacts

There are two ways to create executable code artifacts:

#### Method 1: Explicit Executable Syntax
```javascript
// @import lodash
console.log('Hello, World!')
const numbers = [1, 2, 3, 4, 5]
const sum = _.sum(numbers)
console.log('Sum:', sum)
return sum
```

#### Method 2: Using the Executable Marker
When asking the AI to create executable code, use this format:
```
Can you create an executable JavaScript function that calculates fibonacci numbers?
```

The AI will automatically create artifacts with the `<!-- executable: Title -->` marker.

## Basic Features

### 1. Console Output
```javascript
console.log('This is a log message')
console.error('This is an error message') 
console.warn('This is a warning')
console.info('This is info')
```

### 2. Return Values
```javascript
// The return value is automatically displayed
const result = Math.PI * 2
return result
```

### 3. Error Handling
```javascript
try {
  throw new Error('Something went wrong!')
} catch (error) {
  console.error('Caught error:', error.message)
}
```

## Advanced Features

### 1. Library Loading

The Code Runner supports 9 popular JavaScript libraries that can be loaded automatically:

- **lodash**: Utility functions
- **d3**: Data visualization
- **chart.js**: Chart creation
- **moment**: Date/time manipulation
- **axios**: HTTP requests (limited to safe operations)
- **rxjs**: Reactive programming
- **p5**: Creative coding
- **three**: 3D graphics
- **fabric**: Canvas manipulation

#### Usage:
```javascript
// @import lodash
// @import d3

// Use lodash
const numbers = [1, 2, 3, 4, 5]
const doubled = _.map(numbers, n => n * 2)
console.log('Doubled:', doubled)

// Use d3
const scale = d3.scaleLinear().domain([0, 10]).range([0, 100])
console.log('Scaled value:', scale(5))

return { doubled, scaled: scale(5) }
```

### 2. Canvas Graphics

Create visual outputs using the built-in canvas support:

```javascript
// Create a canvas
const canvas = createCanvas(400, 300)
const ctx = canvas.getContext('2d')

// Draw a colorful rectangle
ctx.fillStyle = '#FF6B6B'
ctx.fillRect(50, 50, 100, 80)

// Draw a circle
ctx.fillStyle = '#4ECDC4'
ctx.beginPath()
ctx.arc(200, 150, 40, 0, 2 * Math.PI)
ctx.fill()

// Add text
ctx.fillStyle = '#45B7D1'
ctx.fillText('Hello Canvas!', 100, 200)

// Return the canvas to display it
return canvas
```

### 3. Data Visualization with D3

```javascript
// @import d3

// Create a simple bar chart
const canvas = createCanvas(400, 300)
const ctx = canvas.getContext('2d')

const data = [10, 20, 30, 40, 50]
const scale = d3.scaleLinear().domain([0, 50]).range([0, 200])

data.forEach((value, index) => {
  const barHeight = scale(value)
  const x = index * 60 + 50
  const y = 250 - barHeight
  
  ctx.fillStyle = `hsl(${index * 60}, 70%, 50%)`
  ctx.fillRect(x, y, 40, barHeight)
  
  ctx.fillStyle = '#000'
  ctx.fillText(value, x + 15, y - 5)
})

return canvas
```

### 4. Mathematical Computations

```javascript
// @import lodash

// Generate random data
const dataset = _.range(100).map(() => Math.random() * 100)

// Calculate statistics
const stats = {
  mean: _.mean(dataset),
  median: _.sortBy(dataset)[Math.floor(dataset.length / 2)],
  min: _.min(dataset),
  max: _.max(dataset),
  sum: _.sum(dataset)
}

console.log('Dataset Statistics:', stats)

// Visualize distribution
const canvas = createCanvas(400, 200)
const ctx = canvas.getContext('2d')

const buckets = _.range(0, 101, 10)
const histogram = buckets.map(bucket => 
  dataset.filter(value => value >= bucket && value < bucket + 10).length
)

histogram.forEach((count, index) => {
  const barHeight = count * 5
  const x = index * 35 + 20
  const y = 180 - barHeight
  
  ctx.fillStyle = '#3498db'
  ctx.fillRect(x, y, 30, barHeight)
  
  ctx.fillStyle = '#000'
  ctx.fillText(count, x + 10, y - 5)
})

return { stats, histogram }
```

## User Interface Guide

### 1. Code Execution Controls

- **▶️ Run Code**: Execute the current code
- **🗑️ Clear Output**: Remove all output
- **✏️ Edit Mode**: Switch between view and edit modes
- **📦 Libraries**: View available libraries

### 2. Library Management

Click the "Available" button next to "Libraries" to see:
- List of all available libraries
- Usage instructions
- Import syntax examples

### 3. Output Display

The output area shows:
- **Console logs**: Blue background
- **Errors**: Red background with error details
- **Return values**: Purple background
- **Canvas graphics**: Rendered inline
- **Execution stats**: Time, memory usage, operations

### 4. Keyboard Shortcuts

- **Ctrl/Cmd + Enter**: Run code while in edit mode
- **Escape**: Exit edit mode

## Performance and Limits

### Resource Limits

The Code Runner has built-in safety limits:
- **Execution Time**: 10 seconds maximum
- **Memory Usage**: ~50MB limit
- **Operations**: 100,000 operations max (prevents infinite loops)
- **Library Loading**: 30 seconds timeout

### Performance Monitoring

Each execution shows:
- **Execution time**: How long the code took to run
- **Memory usage**: Approximate memory consumption
- **Operation count**: Number of operations performed

Example output:
```
Execution completed in 45ms | ~2.3MB | 1,247 ops
```

## Error Handling and Debugging

### Common Errors

1. **Library Not Found**:
   ```
   Error: Library 'unknown' is not available
   ```
   Solution: Check available libraries and use correct names

2. **Memory Limit Exceeded**:
   ```
   Error: Memory limit exceeded: ~52MB
   ```
   Solution: Optimize code to use less memory

3. **Operation Limit Exceeded**:
   ```
   Error: Operation limit exceeded: 100000 operations
   ```
   Solution: Check for infinite loops or reduce computational complexity

4. **Canvas Errors**:
   ```
   Error: Canvas error: Cannot read property 'getContext' of null
   ```
   Solution: Ensure canvas is created before using

### Debugging Tips

1. **Use console.log liberally**:
   ```javascript
   const data = [1, 2, 3]
   console.log('Data:', data)
   
   const result = data.map(x => x * 2)
   console.log('Result:', result)
   ```

2. **Break complex code into steps**:
   ```javascript
   // Step 1: Generate data
   const data = _.range(10).map(() => Math.random())
   console.log('Generated data:', data.length, 'points')
   
   // Step 2: Process data
   const processed = data.map(x => x * 100)
   console.log('Processed data range:', _.min(processed), 'to', _.max(processed))
   
   // Step 3: Visualize
   console.log('Creating visualization...')
   // ... canvas code
   ```

3. **Check execution stats**:
   - Monitor memory usage for large datasets
   - Watch operation count for loops
   - Optimize based on execution time

## Advanced Examples

### 1. Interactive Algorithm Visualization

```javascript
// @import lodash

// Bubble sort with visualization
const canvas = createCanvas(400, 300)
const ctx = canvas.getContext('2d')

const data = _.shuffle(_.range(1, 21)) // Random array 1-20
const steps = []

// Bubble sort algorithm
for (let i = 0; i < data.length; i++) {
  for (let j = 0; j < data.length - i - 1; j++) {
    if (data[j] > data[j + 1]) {
      // Swap
      [data[j], data[j + 1]] = [data[j + 1], data[j]]
      steps.push([...data]) // Record step
    }
  }
}

// Visualize final sorted array
data.forEach((value, index) => {
  const barHeight = value * 10
  const x = index * 18 + 10
  const y = 280 - barHeight
  
  ctx.fillStyle = `hsl(${value * 18}, 70%, 50%)`
  ctx.fillRect(x, y, 15, barHeight)
  
  ctx.fillStyle = '#000'
  ctx.fillText(value, x + 2, y - 2)
})

console.log(`Sorting completed in ${steps.length} steps`)
return canvas
```

### 2. Statistical Analysis with Charts

```javascript
// @import lodash
// @import d3

// Generate sample data
const sampleSize = 1000
const data = _.range(sampleSize).map(() => 
  d3.randomNormal(50, 15)() // Normal distribution, mean=50, std=15
)

// Calculate statistics
const stats = {
  count: data.length,
  mean: d3.mean(data),
  median: d3.median(data),
  deviation: d3.deviation(data),
  min: d3.min(data),
  max: d3.max(data)
}

console.log('Statistics:', stats)

// Create histogram
const canvas = createCanvas(500, 400)
const ctx = canvas.getContext('2d')

const bins = d3.histogram()
  .domain(d3.extent(data))
  .thresholds(20)(data)

const xScale = d3.scaleLinear()
  .domain(d3.extent(data))
  .range([50, 450])

const yScale = d3.scaleLinear()
  .domain([0, d3.max(bins, d => d.length)])
  .range([350, 50])

bins.forEach(bin => {
  const x = xScale(bin.x0)
  const y = yScale(bin.length)
  const width = xScale(bin.x1) - xScale(bin.x0) - 1
  const height = 350 - y
  
  ctx.fillStyle = '#3498db'
  ctx.fillRect(x, y, width, height)
  
  ctx.fillStyle = '#000'
  ctx.fillText(bin.length, x + width/2 - 5, y - 5)
})

// Add title
ctx.fillStyle = '#000'
ctx.font = '16px Arial'
ctx.fillText('Normal Distribution Histogram', 150, 30)

return { stats, canvas }
```

### 3. Real-time Data Processing

```javascript
// @import lodash
// @import moment

// Simulate time-series data
const now = moment()
const timePoints = _.range(24).map(hour => ({
  time: now.clone().subtract(24 - hour, 'hours'),
  value: Math.sin(hour * Math.PI / 12) * 50 + 50 + Math.random() * 20
}))

// Process data
const processed = timePoints.map(point => ({
  hour: point.time.format('HH:mm'),
  value: Math.round(point.value * 100) / 100,
  trend: point.value > 50 ? 'up' : 'down'
}))

// Calculate rolling average
const windowSize = 3
const rollingAverage = processed.map((point, index) => {
  const start = Math.max(0, index - windowSize + 1)
  const window = processed.slice(start, index + 1)
  const avg = _.meanBy(window, 'value')
  return { ...point, rollingAvg: Math.round(avg * 100) / 100 }
})

console.log('Data points:', processed.length)
console.log('Sample:', processed.slice(0, 3))

// Visualize trends
const canvas = createCanvas(600, 300)
const ctx = canvas.getContext('2d')

rollingAverage.forEach((point, index) => {
  const x = index * 24 + 50
  const y = 250 - (point.value * 2)
  const avgY = 250 - (point.rollingAvg * 2)
  
  // Draw data point
  ctx.fillStyle = point.trend === 'up' ? '#2ecc71' : '#e74c3c'
  ctx.fillRect(x - 2, y - 2, 4, 4)
  
  // Draw rolling average
  ctx.fillStyle = '#3498db'
  ctx.fillRect(x - 1, avgY - 1, 2, 2)
  
  // Connect points
  if (index > 0) {
    const prevX = (index - 1) * 24 + 50
    const prevY = 250 - (rollingAverage[index - 1].rollingAvg * 2)
    
    ctx.strokeStyle = '#3498db'
    ctx.beginPath()
    ctx.moveTo(prevX, prevY)
    ctx.lineTo(x, avgY)
    ctx.stroke()
  }
})

return { processed: rollingAverage, canvas }
```

## Best Practices

### 1. Code Organization

- Break complex tasks into smaller functions
- Use descriptive variable names
- Add comments for complex logic
- Return meaningful results

### 2. Performance Optimization

- Avoid unnecessary loops
- Use efficient algorithms
- Monitor memory usage for large datasets
- Consider using libraries like lodash for optimized operations

### 3. Error Prevention

- Validate inputs before processing
- Use try-catch blocks for risky operations
- Check array lengths before accessing elements
- Handle edge cases explicitly

### 4. Visualization Guidelines

- Choose appropriate canvas sizes (400x300 is standard)
- Use contrasting colors for better visibility
- Add labels and legends when helpful
- Scale graphics appropriately for the data

## Security and Limitations

### What's Allowed

- All standard JavaScript features
- Mathematical computations
- Data manipulation and analysis
- Canvas graphics and visualizations
- Supported library functions
- Console output and debugging

### What's Not Allowed

- Direct DOM manipulation
- Network requests (fetch, XMLHttpRequest)
- File system access
- Local storage access
- WebSocket connections
- Worker creation
- Eval or Function constructor (except internally)

### Resource Limits

- Maximum execution time: 10 seconds
- Memory limit: ~50MB
- Operation limit: 100,000 operations
- Library loading timeout: 30 seconds

## Troubleshooting

### Common Issues

1. **Code doesn't run**: Check for syntax errors
2. **Canvas doesn't display**: Ensure you return the canvas object
3. **Library not loading**: Verify library name and internet connection
4. **Performance issues**: Check operation count and optimize loops

### Getting Help

- Use console.log to debug step by step
- Check the execution statistics for performance insights
- Simplify complex code to isolate issues
- Refer to library documentation for specific functions

This Code Runner provides a powerful environment for learning, prototyping, and demonstrating JavaScript concepts with real-time feedback and rich visualizations. Experiment with different features and libraries to discover what's possible!