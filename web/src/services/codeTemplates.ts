/**
 * Code Templates and Snippets Service
 * Manages pre-built code templates and user-defined snippets
 */

import { reactive, computed } from 'vue'

export interface CodeTemplate {
  id: string
  name: string
  description: string
  language: string
  code: string
  category: string
  tags: string[]
  author?: string
  createdAt: string
  updatedAt: string
  difficulty: 'beginner' | 'intermediate' | 'advanced'
  estimatedRunTime?: number
  requirements?: string[]
  isBuiltIn: boolean
  usageCount: number
  rating?: number
  examples?: {
    input?: string
    output?: string
    description?: string
  }[]
}

export interface TemplateCategory {
  id: string
  name: string
  description: string
  icon: string
  color: string
  templates: CodeTemplate[]
}

class CodeTemplatesService {
  private static instance: CodeTemplatesService
  private templates: CodeTemplate[] = reactive([])
  private categories: TemplateCategory[] = reactive([])
  private storageKey = 'code-templates'
  private userTemplatesKey = 'user-code-templates'

  private constructor() {
    this.initializeBuiltInTemplates()
    this.loadUserTemplates()
  }

  static getInstance(): CodeTemplatesService {
    if (!CodeTemplatesService.instance) {
      CodeTemplatesService.instance = new CodeTemplatesService()
    }
    return CodeTemplatesService.instance
  }

  /**
   * Initialize built-in templates
   */
  private initializeBuiltInTemplates(): void {
    const builtInTemplates: CodeTemplate[] = [
      // JavaScript Templates
      {
        id: 'js-hello-world',
        name: 'Hello World',
        description: 'Basic Hello World example',
        language: 'javascript',
        code: `console.log('Hello, World!')
console.log('Welcome to JavaScript!')`,
        category: 'basics',
        tags: ['beginner', 'console', 'output'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        difficulty: 'beginner',
        estimatedRunTime: 50,
        isBuiltIn: true,
        usageCount: 0,
        rating: 5,
        examples: [
          {
            output: 'Hello, World!\nWelcome to JavaScript!',
            description: 'Simple console output'
          }
        ]
      },
      {
        id: 'js-async-fetch',
        name: 'Async Data Fetching',
        description: 'Demonstrate async/await with simulated API calls',
        language: 'javascript',
        code: `// Simulate an API call
async function fetchData(url) {
  console.log(\`Fetching data from: \${url}\`)
  
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 1000))
  
  // Simulate API response
  return {
    data: [1, 2, 3, 4, 5],
    timestamp: new Date().toISOString(),
    status: 'success'
  }
}

// Using async/await
async function main() {
  try {
    console.log('Starting data fetch...')
    const result = await fetchData('https://api.example.com/data')
    console.log('Data received:', result)
    console.log('Processing complete!')
  } catch (error) {
    console.error('Error fetching data:', error)
  }
}

main()`,
        category: 'async',
        tags: ['async', 'await', 'promises', 'intermediate'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        difficulty: 'intermediate',
        estimatedRunTime: 1200,
        isBuiltIn: true,
        usageCount: 0,
        rating: 4.5
      },
      {
        id: 'js-canvas-animation',
        name: 'Canvas Animation',
        description: 'Create animated graphics using Canvas API',
        language: 'javascript',
        code: `// Create a canvas for animation
const canvas = createCanvas(400, 300)
const ctx = canvas.getContext('2d')

// Animation parameters
let frame = 0
const colors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7']

// Animation loop
function animate() {
  // Clear canvas
  ctx.clearRect(0, 0, canvas.width, canvas.height)
  
  // Draw animated circles
  for (let i = 0; i < 5; i++) {
    const x = 200 + Math.sin(frame * 0.02 + i) * 100
    const y = 150 + Math.cos(frame * 0.03 + i) * 50
    const radius = 20 + Math.sin(frame * 0.05 + i) * 10
    
    ctx.beginPath()
    ctx.arc(x, y, radius, 0, Math.PI * 2)
    ctx.fillStyle = colors[i]
    ctx.fill()
  }
  
  frame++
  
  // Continue animation
  if (frame < 200) {
    setTimeout(animate, 50)
  }
}

console.log('Starting canvas animation...')
animate()`,
        category: 'graphics',
        tags: ['canvas', 'animation', 'graphics', 'intermediate'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        difficulty: 'intermediate',
        estimatedRunTime: 10000,
        isBuiltIn: true,
        usageCount: 0,
        rating: 4.8
      },
      {
        id: 'js-data-structures',
        name: 'Data Structures Demo',
        description: 'Demonstrate common data structures and algorithms',
        language: 'javascript',
        code: `// Stack implementation
class Stack {
  constructor() {
    this.items = []
  }
  
  push(item) {
    this.items.push(item)
  }
  
  pop() {
    return this.items.pop()
  }
  
  peek() {
    return this.items[this.items.length - 1]
  }
  
  isEmpty() {
    return this.items.length === 0
  }
}

// Queue implementation
class Queue {
  constructor() {
    this.items = []
  }
  
  enqueue(item) {
    this.items.push(item)
  }
  
  dequeue() {
    return this.items.shift()
  }
  
  front() {
    return this.items[0]
  }
  
  isEmpty() {
    return this.items.length === 0
  }
}

// Demonstrate usage
console.log('=== Stack Demo ===')
const stack = new Stack()
stack.push(1)
stack.push(2)
stack.push(3)
console.log('Stack after pushes:', stack.items)
console.log('Popped:', stack.pop())
console.log('Stack after pop:', stack.items)

console.log('\\n=== Queue Demo ===')
const queue = new Queue()
queue.enqueue('A')
queue.enqueue('B')
queue.enqueue('C')
console.log('Queue after enqueues:', queue.items)
console.log('Dequeued:', queue.dequeue())
console.log('Queue after dequeue:', queue.items)

// Binary search
function binarySearch(arr, target) {
  let left = 0
  let right = arr.length - 1
  
  while (left <= right) {
    const mid = Math.floor((left + right) / 2)
    
    if (arr[mid] === target) {
      return mid
    } else if (arr[mid] < target) {
      left = mid + 1
    } else {
      right = mid - 1
    }
  }
  
  return -1
}

console.log('\\n=== Binary Search Demo ===')
const sortedArray = [1, 3, 5, 7, 9, 11, 13, 15, 17, 19]
console.log('Array:', sortedArray)
console.log('Search for 7:', binarySearch(sortedArray, 7))
console.log('Search for 12:', binarySearch(sortedArray, 12))`,
        category: 'algorithms',
        tags: ['data-structures', 'algorithms', 'stack', 'queue', 'search', 'advanced'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        difficulty: 'advanced',
        estimatedRunTime: 200,
        isBuiltIn: true,
        usageCount: 0,
        rating: 4.7
      },
      // Python Templates
      {
        id: 'py-hello-world',
        name: 'Hello World',
        description: 'Basic Hello World example in Python',
        language: 'python',
        code: `print("Hello, World!")
print("Welcome to Python!")

# Variables and basic operations
name = "Python"
version = 3.9
print(f"Language: {name}, Version: {version}")

# List operations
numbers = [1, 2, 3, 4, 5]
print(f"Numbers: {numbers}")
print(f"Sum: {sum(numbers)}")
print(f"Max: {max(numbers)}")`,
        category: 'basics',
        tags: ['beginner', 'print', 'variables', 'lists'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        difficulty: 'beginner',
        estimatedRunTime: 100,
        isBuiltIn: true,
        usageCount: 0,
        rating: 5,
        requirements: []
      },
      {
        id: 'py-data-analysis',
        name: 'Data Analysis with Pandas',
        description: 'Basic data analysis using pandas and numpy',
        language: 'python',
        code: `import pandas as pd
import numpy as np

# Create sample data
data = {
    'Name': ['Alice', 'Bob', 'Charlie', 'Diana', 'Eve'],
    'Age': [25, 30, 35, 28, 32],
    'City': ['New York', 'London', 'Paris', 'Tokyo', 'Sydney'],
    'Salary': [50000, 60000, 70000, 55000, 65000]
}

# Create DataFrame
df = pd.DataFrame(data)
print("Original DataFrame:")
print(df)
print()

# Basic statistics
print("Basic Statistics:")
print(df.describe())
print()

# Data filtering
print("People over 30:")
print(df[df['Age'] > 30])
print()

# Data aggregation
print("Average salary by city:")
city_avg = df.groupby('City')['Salary'].mean()
print(city_avg)
print()

# Adding a new column
df['Salary_USD'] = df['Salary']
df['Salary_EUR'] = df['Salary'] * 0.85  # Example conversion
print("DataFrame with new columns:")
print(df[['Name', 'Salary_USD', 'Salary_EUR']])`,
        category: 'data-science',
        tags: ['pandas', 'numpy', 'data-analysis', 'intermediate'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        difficulty: 'intermediate',
        estimatedRunTime: 500,
        requirements: ['pandas', 'numpy'],
        isBuiltIn: true,
        usageCount: 0,
        rating: 4.6
      },
      {
        id: 'py-matplotlib-plots',
        name: 'Data Visualization with Matplotlib',
        description: 'Create various types of plots using matplotlib',
        language: 'python',
        code: `import matplotlib.pyplot as plt
import numpy as np

# Create sample data
x = np.linspace(0, 10, 100)
y1 = np.sin(x)
y2 = np.cos(x)
y3 = np.sin(x) * np.cos(x)

# Create the plot
plt.figure(figsize=(10, 6))

# Plot multiple lines
plt.plot(x, y1, 'b-', label='sin(x)', linewidth=2)
plt.plot(x, y2, 'r--', label='cos(x)', linewidth=2)
plt.plot(x, y3, 'g:', label='sin(x)*cos(x)', linewidth=2)

# Customize the plot
plt.title('Trigonometric Functions', fontsize=16, fontweight='bold')
plt.xlabel('x', fontsize=12)
plt.ylabel('y', fontsize=12)
plt.legend(fontsize=10)
plt.grid(True, alpha=0.3)

# Add some styling
plt.tight_layout()
plt.show()

print("Matplotlib plot created successfully!")
print("The plot shows three trigonometric functions:")
print("- Blue solid line: sin(x)")
print("- Red dashed line: cos(x)")
print("- Green dotted line: sin(x)*cos(x)")`,
        category: 'visualization',
        tags: ['matplotlib', 'visualization', 'plots', 'numpy', 'intermediate'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        difficulty: 'intermediate',
        estimatedRunTime: 800,
        requirements: ['matplotlib', 'numpy'],
        isBuiltIn: true,
        usageCount: 0,
        rating: 4.8
      },
      {
        id: 'py-machine-learning',
        name: 'Simple Machine Learning',
        description: 'Basic machine learning example with scikit-learn',
        language: 'python',
        code: `import numpy as np
from sklearn.model_selection import train_test_split
from sklearn.linear_model import LinearRegression
from sklearn.metrics import mean_squared_error, r2_score
import matplotlib.pyplot as plt

# Generate sample data
np.random.seed(42)
X = np.random.rand(100, 1) * 10  # Features
y = 2 * X.ravel() + 1 + np.random.randn(100) * 2  # Target with noise

print("Generated dataset:")
print(f"Features shape: {X.shape}")
print(f"Target shape: {y.shape}")
print(f"Feature range: [{X.min():.2f}, {X.max():.2f}]")
print(f"Target range: [{y.min():.2f}, {y.max():.2f}]")
print()

# Split the data
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

# Create and train the model
model = LinearRegression()
model.fit(X_train, y_train)

# Make predictions
y_pred = model.predict(X_test)

# Calculate metrics
mse = mean_squared_error(y_test, y_pred)
r2 = r2_score(y_test, y_pred)

print("Model Performance:")
print(f"Mean Squared Error: {mse:.2f}")
print(f"R² Score: {r2:.2f}")
print(f"Model coefficients: {model.coef_[0]:.2f}")
print(f"Model intercept: {model.intercept_:.2f}")
print()

# Visualize results
plt.figure(figsize=(10, 6))
plt.scatter(X_test, y_test, color='blue', label='Actual', alpha=0.7)
plt.scatter(X_test, y_pred, color='red', label='Predicted', alpha=0.7)
plt.plot(X_test, y_pred, color='red', linewidth=2)
plt.xlabel('Feature')
plt.ylabel('Target')
plt.title('Linear Regression: Actual vs Predicted')
plt.legend()
plt.grid(True, alpha=0.3)
plt.show()

print("Machine learning model trained and visualized successfully!")`,
        category: 'machine-learning',
        tags: ['scikit-learn', 'machine-learning', 'regression', 'matplotlib', 'advanced'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        difficulty: 'advanced',
        estimatedRunTime: 1500,
        requirements: ['scikit-learn', 'matplotlib', 'numpy'],
        isBuiltIn: true,
        usageCount: 0,
        rating: 4.9
      }
    ]

    this.templates.push(...builtInTemplates)
    this.initializeCategories()
  }

  /**
   * Initialize template categories
   */
  private initializeCategories(): void {
    this.categories.push(
      {
        id: 'basics',
        name: 'Basics',
        description: 'Fundamental programming concepts',
        icon: 'ri:book-line',
        color: '#4ECDC4',
        templates: this.templates.filter(t => t.category === 'basics')
      },
      {
        id: 'async',
        name: 'Async Programming',
        description: 'Asynchronous programming patterns',
        icon: 'ri:time-line',
        color: '#45B7D1',
        templates: this.templates.filter(t => t.category === 'async')
      },
      {
        id: 'graphics',
        name: 'Graphics & Animation',
        description: 'Canvas graphics and animations',
        icon: 'ri:palette-line',
        color: '#FF6B6B',
        templates: this.templates.filter(t => t.category === 'graphics')
      },
      {
        id: 'algorithms',
        name: 'Algorithms',
        description: 'Data structures and algorithms',
        icon: 'ri:mind-map',
        color: '#96CEB4',
        templates: this.templates.filter(t => t.category === 'algorithms')
      },
      {
        id: 'data-science',
        name: 'Data Science',
        description: 'Data analysis and manipulation',
        icon: 'ri:bar-chart-line',
        color: '#FFEAA7',
        templates: this.templates.filter(t => t.category === 'data-science')
      },
      {
        id: 'visualization',
        name: 'Data Visualization',
        description: 'Charts and plots',
        icon: 'ri:line-chart-line',
        color: '#DDA0DD',
        templates: this.templates.filter(t => t.category === 'visualization')
      },
      {
        id: 'machine-learning',
        name: 'Machine Learning',
        description: 'ML algorithms and models',
        icon: 'ri:brain-line',
        color: '#FFA07A',
        templates: this.templates.filter(t => t.category === 'machine-learning')
      }
    )
  }

  /**
   * Get all templates
   */
  getTemplates(): CodeTemplate[] {
    return [...this.templates]
  }

  /**
   * Get templates by category
   */
  getTemplatesByCategory(categoryId: string): CodeTemplate[] {
    return this.templates.filter(t => t.category === categoryId)
  }

  /**
   * Get templates by language
   */
  getTemplatesByLanguage(language: string): CodeTemplate[] {
    return this.templates.filter(t => t.language.toLowerCase() === language.toLowerCase())
  }

  /**
   * Get template by ID
   */
  getTemplate(id: string): CodeTemplate | undefined {
    return this.templates.find(t => t.id === id)
  }

  /**
   * Search templates
   */
  searchTemplates(query: string, filters?: {
    language?: string
    category?: string
    difficulty?: string
    tags?: string[]
  }): CodeTemplate[] {
    let results = this.templates

    // Text search
    if (query.trim()) {
      const searchTerm = query.toLowerCase()
      results = results.filter(template => 
        template.name.toLowerCase().includes(searchTerm) ||
        template.description.toLowerCase().includes(searchTerm) ||
        template.tags.some(tag => tag.toLowerCase().includes(searchTerm)) ||
        template.code.toLowerCase().includes(searchTerm)
      )
    }

    // Apply filters
    if (filters) {
      if (filters.language) {
        results = results.filter(t => t.language.toLowerCase() === filters.language!.toLowerCase())
      }
      
      if (filters.category) {
        results = results.filter(t => t.category === filters.category)
      }
      
      if (filters.difficulty) {
        results = results.filter(t => t.difficulty === filters.difficulty)
      }
      
      if (filters.tags && filters.tags.length > 0) {
        results = results.filter(t => 
          filters.tags!.some(tag => t.tags.includes(tag))
        )
      }
    }

    return results
  }

  /**
   * Get all categories
   */
  getCategories(): TemplateCategory[] {
    // Update template counts
    this.categories.forEach(category => {
      category.templates = this.templates.filter(t => t.category === category.id)
    })
    return [...this.categories]
  }

  /**
   * Get popular templates
   */
  getPopularTemplates(limit = 10): CodeTemplate[] {
    return this.templates
      .sort((a, b) => b.usageCount - a.usageCount)
      .slice(0, limit)
  }

  /**
   * Get recent templates
   */
  getRecentTemplates(limit = 10): CodeTemplate[] {
    return this.templates
      .sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
      .slice(0, limit)
  }

  /**
   * Add a new template
   */
  addTemplate(template: Omit<CodeTemplate, 'id' | 'createdAt' | 'updatedAt' | 'isBuiltIn' | 'usageCount'>): string {
    const id = this.generateId()
    const now = new Date().toISOString()
    
    const newTemplate: CodeTemplate = {
      id,
      ...template,
      createdAt: now,
      updatedAt: now,
      isBuiltIn: false,
      usageCount: 0
    }
    
    this.templates.push(newTemplate)
    this.saveUserTemplates()
    
    return id
  }

  /**
   * Update a template
   */
  updateTemplate(id: string, updates: Partial<CodeTemplate>): boolean {
    const template = this.templates.find(t => t.id === id)
    if (!template || template.isBuiltIn) {
      return false
    }
    
    Object.assign(template, updates, { updatedAt: new Date().toISOString() })
    this.saveUserTemplates()
    
    return true
  }

  /**
   * Delete a template
   */
  deleteTemplate(id: string): boolean {
    const index = this.templates.findIndex(t => t.id === id)
    if (index === -1 || this.templates[index].isBuiltIn) {
      return false
    }
    
    this.templates.splice(index, 1)
    this.saveUserTemplates()
    
    return true
  }

  /**
   * Increment template usage count
   */
  incrementUsage(id: string): void {
    const template = this.templates.find(t => t.id === id)
    if (template) {
      template.usageCount++
      this.saveUserTemplates()
    }
  }

  /**
   * Rate a template
   */
  rateTemplate(id: string, rating: number): boolean {
    const template = this.templates.find(t => t.id === id)
    if (!template || rating < 1 || rating > 5) {
      return false
    }
    
    template.rating = rating
    this.saveUserTemplates()
    
    return true
  }

  /**
   * Export templates
   */
  exportTemplates(): string {
    const userTemplates = this.templates.filter(t => !t.isBuiltIn)
    return JSON.stringify(userTemplates, null, 2)
  }

  /**
   * Import templates
   */
  importTemplates(jsonData: string): boolean {
    try {
      const importedTemplates = JSON.parse(jsonData) as CodeTemplate[]
      
      // Validate imported data
      if (!Array.isArray(importedTemplates)) {
        throw new Error('Invalid format: expected array')
      }
      
      // Add imported templates
      importedTemplates.forEach(template => {
        const existingTemplate = this.templates.find(t => t.id === template.id)
        if (!existingTemplate) {
          template.isBuiltIn = false
          template.usageCount = template.usageCount || 0
          this.templates.push(template)
        }
      })
      
      this.saveUserTemplates()
      return true
    } catch (error) {
      console.error('Failed to import templates:', error)
      return false
    }
  }

  private generateId(): string {
    return `template_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  private saveUserTemplates(): void {
    try {
      const userTemplates = this.templates.filter(t => !t.isBuiltIn)
      localStorage.setItem(this.userTemplatesKey, JSON.stringify(userTemplates))
    } catch (error) {
      console.warn('Failed to save user templates:', error)
    }
  }

  private loadUserTemplates(): void {
    try {
      const stored = localStorage.getItem(this.userTemplatesKey)
      if (stored) {
        const userTemplates = JSON.parse(stored) as CodeTemplate[]
        this.templates.push(...userTemplates)
      }
    } catch (error) {
      console.warn('Failed to load user templates:', error)
    }
  }
}

// Export singleton instance
export const codeTemplates = CodeTemplatesService.getInstance()

// Export composable for Vue components
export function useCodeTemplates() {
  return {
    templates: computed(() => codeTemplates.getTemplates()),
    categories: computed(() => codeTemplates.getCategories()),
    popularTemplates: computed(() => codeTemplates.getPopularTemplates()),
    recentTemplates: computed(() => codeTemplates.getRecentTemplates()),
    getTemplate: codeTemplates.getTemplate.bind(codeTemplates),
    getTemplatesByCategory: codeTemplates.getTemplatesByCategory.bind(codeTemplates),
    getTemplatesByLanguage: codeTemplates.getTemplatesByLanguage.bind(codeTemplates),
    searchTemplates: codeTemplates.searchTemplates.bind(codeTemplates),
    addTemplate: codeTemplates.addTemplate.bind(codeTemplates),
    updateTemplate: codeTemplates.updateTemplate.bind(codeTemplates),
    deleteTemplate: codeTemplates.deleteTemplate.bind(codeTemplates),
    incrementUsage: codeTemplates.incrementUsage.bind(codeTemplates),
    rateTemplate: codeTemplates.rateTemplate.bind(codeTemplates),
    exportTemplates: codeTemplates.exportTemplates.bind(codeTemplates),
    importTemplates: codeTemplates.importTemplates.bind(codeTemplates)
  }
}