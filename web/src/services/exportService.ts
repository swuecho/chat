/**
 * Export Service
 * Handles exporting artifacts, code, and execution results in various formats
 */

import JSZip from 'jszip'
import { saveAs } from 'file-saver'
import type { ExecutionResult } from './codeRunner'
import type { ExecutionHistoryEntry } from './executionHistory'
import type { CodeTemplate } from './codeTemplates'

export interface ExportOptions {
  format: 'single' | 'zip' | 'json' | 'html' | 'pdf'
  includeHistory?: boolean
  includeResults?: boolean
  includeMetadata?: boolean
  includeTimestamps?: boolean
  template?: 'minimal' | 'detailed' | 'presentation'
}

export interface ShareOptions {
  title?: string
  description?: string
  language?: string
  tags?: string[]
  public?: boolean
  expiration?: Date
}

export interface ExportableArtifact {
  id: string
  title: string
  content: string
  type: string
  language?: string
  createdAt: string
  updatedAt?: string
  tags?: string[]
  results?: ExecutionResult[]
  history?: ExecutionHistoryEntry[]
  metadata?: Record<string, any>
}

class ExportService {
  private static instance: ExportService

  private constructor() {}

  static getInstance(): ExportService {
    if (!ExportService.instance) {
      ExportService.instance = new ExportService()
    }
    return ExportService.instance
  }

  /**
   * Export a single artifact
   */
  async exportArtifact(artifact: ExportableArtifact, options: ExportOptions = { format: 'single' }): Promise<void> {
    switch (options.format) {
      case 'single':
        this.exportSingleFile(artifact, options)
        break
      case 'zip':
        await this.exportAsZip([artifact], options)
        break
      case 'json':
        this.exportAsJson([artifact], options)
        break
      case 'html':
        this.exportAsHtml([artifact], options)
        break
      case 'pdf':
        await this.exportAsPdf([artifact], options)
        break
    }
  }

  /**
   * Export multiple artifacts
   */
  async exportArtifacts(artifacts: ExportableArtifact[], options: ExportOptions = { format: 'zip' }): Promise<void> {
    switch (options.format) {
      case 'zip':
        await this.exportAsZip(artifacts, options)
        break
      case 'json':
        this.exportAsJson(artifacts, options)
        break
      case 'html':
        this.exportAsHtml(artifacts, options)
        break
      case 'pdf':
        await this.exportAsPdf(artifacts, options)
        break
      default:
        throw new Error('Multiple artifacts require zip, json, html, or pdf format')
    }
  }

  /**
   * Export execution history
   */
  async exportHistory(history: ExecutionHistoryEntry[], options: ExportOptions = { format: 'json' }): Promise<void> {
    const exportData = {
      type: 'execution_history',
      exportedAt: new Date().toISOString(),
      count: history.length,
      history: history.map(entry => ({
        id: entry.id,
        timestamp: entry.timestamp,
        code: entry.code,
        language: entry.language,
        success: entry.success,
        executionTime: entry.executionTime,
        tags: entry.tags,
        notes: entry.notes,
        ...(options.includeResults && { results: entry.results })
      }))
    }

    switch (options.format) {
      case 'json':
        this.downloadJson(exportData, 'execution_history.json')
        break
      case 'html':
        this.exportHistoryAsHtml(exportData, options)
        break
      case 'zip':
        await this.exportHistoryAsZip(exportData, options)
        break
      default:
        throw new Error('Unsupported format for history export')
    }
  }

  /**
   * Export code templates
   */
  async exportTemplates(templates: CodeTemplate[], options: ExportOptions = { format: 'json' }): Promise<void> {
    const exportData = {
      type: 'code_templates',
      exportedAt: new Date().toISOString(),
      count: templates.length,
      templates: templates.map(template => ({
        ...template,
        ...(options.includeMetadata && {
          metadata: {
            usageCount: template.usageCount,
            rating: template.rating,
            isBuiltIn: template.isBuiltIn
          }
        })
      }))
    }

    switch (options.format) {
      case 'json':
        this.downloadJson(exportData, 'code_templates.json')
        break
      case 'zip':
        await this.exportTemplatesAsZip(exportData, options)
        break
      default:
        throw new Error('Unsupported format for template export')
    }
  }

  /**
   * Share artifact via URL
   */
  async shareArtifact(artifact: ExportableArtifact, options: ShareOptions = {}): Promise<string> {
    const shareData = {
      title: options.title || artifact.title,
      description: options.description || `Shared artifact: ${artifact.title}`,
      content: artifact.content,
      language: options.language || artifact.language,
      type: artifact.type,
      tags: options.tags || artifact.tags,
      createdAt: new Date().toISOString(),
      public: options.public ?? true,
      expiration: options.expiration?.toISOString()
    }

    // In a real implementation, this would POST to a sharing service
    // For now, we'll create a data URL
    const dataUrl = this.createDataUrl(shareData)
    
    // Copy to clipboard
    try {
      await navigator.clipboard.writeText(dataUrl)
      return dataUrl
    } catch (error) {
      console.error('Failed to copy share URL:', error)
      throw new Error('Failed to create share URL')
    }
  }

  /**
   * Export project (multiple files)
   */
  async exportProject(artifacts: ExportableArtifact[], projectName: string, options: ExportOptions = { format: 'zip' }): Promise<void> {
    const zip = new JSZip()
    const projectFolder = zip.folder(projectName)
    
    if (!projectFolder) {
      throw new Error('Failed to create project folder')
    }

    // Add README
    const readme = this.generateReadme(artifacts, projectName)
    projectFolder.file('README.md', readme)

    // Add package.json for JavaScript projects
    const hasJavaScript = artifacts.some(a => a.language === 'javascript' || a.language === 'typescript')
    if (hasJavaScript) {
      const packageJson = this.generatePackageJson(artifacts, projectName)
      projectFolder.file('package.json', JSON.stringify(packageJson, null, 2))
    }

    // Add requirements.txt for Python projects
    const hasPython = artifacts.some(a => a.language === 'python')
    if (hasPython) {
      const requirements = this.generateRequirements(artifacts)
      projectFolder.file('requirements.txt', requirements)
    }

    // Add artifacts
    artifacts.forEach(artifact => {
      const filename = this.generateFileName(artifact)
      const folder = this.getArtifactFolder(artifact, projectFolder)
      folder.file(filename, artifact.content)

      // Add metadata if requested
      if (options.includeMetadata) {
        const metadata = this.generateMetadata(artifact, options)
        folder.file(`${filename}.meta.json`, JSON.stringify(metadata, null, 2))
      }

      // Add execution results if requested
      if (options.includeResults && artifact.results) {
        const resultsFile = this.generateResultsFile(artifact.results)
        folder.file(`${filename}.results.json`, resultsFile)
      }
    })

    // Generate and download zip
    const content = await zip.generateAsync({ type: 'blob' })
    saveAs(content, `${projectName}.zip`)
  }

  /**
   * Export as single file
   */
  private exportSingleFile(artifact: ExportableArtifact, options: ExportOptions): void {
    const filename = this.generateFileName(artifact)
    const content = this.prepareContent(artifact, options)
    
    const blob = new Blob([content], { type: 'text/plain' })
    saveAs(blob, filename)
  }

  /**
   * Export as ZIP
   */
  private async exportAsZip(artifacts: ExportableArtifact[], options: ExportOptions): Promise<void> {
    const zip = new JSZip()
    const timestamp = new Date().toISOString().split('T')[0]
    
    artifacts.forEach((artifact, index) => {
      const filename = this.generateFileName(artifact, index)
      const content = this.prepareContent(artifact, options)
      zip.file(filename, content)

      if (options.includeMetadata) {
        const metadata = this.generateMetadata(artifact, options)
        zip.file(`${filename}.meta.json`, JSON.stringify(metadata, null, 2))
      }
    })

    // Add manifest
    const manifest = this.generateManifest(artifacts, options)
    zip.file('manifest.json', JSON.stringify(manifest, null, 2))

    const content = await zip.generateAsync({ type: 'blob' })
    saveAs(content, `artifacts_${timestamp}.zip`)
  }

  /**
   * Export as JSON
   */
  private exportAsJson(artifacts: ExportableArtifact[], options: ExportOptions): void {
    const exportData = {
      type: 'artifacts',
      exportedAt: new Date().toISOString(),
      count: artifacts.length,
      options,
      artifacts: artifacts.map(artifact => ({
        ...artifact,
        ...(options.includeHistory && { history: artifact.history }),
        ...(options.includeResults && { results: artifact.results }),
        ...(options.includeMetadata && { metadata: artifact.metadata })
      }))
    }

    this.downloadJson(exportData, `artifacts_${new Date().toISOString().split('T')[0]}.json`)
  }

  /**
   * Export as HTML
   */
  private exportAsHtml(artifacts: ExportableArtifact[], options: ExportOptions): void {
    const html = this.generateHtmlReport(artifacts, options)
    const blob = new Blob([html], { type: 'text/html' })
    saveAs(blob, `artifacts_${new Date().toISOString().split('T')[0]}.html`)
  }

  /**
   * Export as PDF
   */
  private async exportAsPdf(artifacts: ExportableArtifact[], options: ExportOptions): Promise<void> {
    // This would require a PDF library like jsPDF
    // For now, we'll export as HTML with print-friendly styling
    const html = this.generatePrintableHtml(artifacts, options)
    const blob = new Blob([html], { type: 'text/html' })
    saveAs(blob, `artifacts_${new Date().toISOString().split('T')[0]}_printable.html`)
  }

  /**
   * Generate HTML report
   */
  private generateHtmlReport(artifacts: ExportableArtifact[], options: ExportOptions): string {
    const timestamp = new Date().toISOString()
    const template = options.template || 'detailed'
    
    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Artifact Export Report</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            text-align: center;
            margin-bottom: 40px;
            padding-bottom: 20px;
            border-bottom: 2px solid #e0e0e0;
        }
        .artifact {
            margin-bottom: 40px;
            padding: 20px;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            background: #f9f9f9;
        }
        .artifact-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 16px;
        }
        .artifact-title {
            font-size: 20px;
            font-weight: 600;
            color: #2c3e50;
        }
        .artifact-meta {
            display: flex;
            gap: 16px;
            font-size: 14px;
            color: #666;
        }
        .artifact-content {
            background: white;
            padding: 16px;
            border-radius: 4px;
            overflow-x: auto;
        }
        .code-block {
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            white-space: pre-wrap;
            background: #f4f4f4;
            padding: 12px;
            border-radius: 4px;
            border-left: 4px solid #007acc;
        }
        .results {
            margin-top: 16px;
            padding: 12px;
            background: #f0f8ff;
            border-radius: 4px;
        }
        .tag {
            display: inline-block;
            padding: 2px 8px;
            background: #e1f5fe;
            border-radius: 12px;
            font-size: 12px;
            margin: 2px;
        }
        .footer {
            text-align: center;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #e0e0e0;
            color: #666;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Artifact Export Report</h1>
        <p>Generated on ${new Date(timestamp).toLocaleString()}</p>
        <p>Total artifacts: ${artifacts.length}</p>
    </div>

    ${artifacts.map(artifact => `
        <div class="artifact">
            <div class="artifact-header">
                <div class="artifact-title">${artifact.title}</div>
                <div class="artifact-meta">
                    <span>Type: ${artifact.type}</span>
                    ${artifact.language ? `<span>Language: ${artifact.language}</span>` : ''}
                    <span>Created: ${new Date(artifact.createdAt).toLocaleDateString()}</span>
                </div>
            </div>
            
            <div class="artifact-content">
                <div class="code-block">${this.escapeHtml(artifact.content)}</div>
                
                ${artifact.tags && artifact.tags.length > 0 ? `
                    <div style="margin-top: 12px;">
                        <strong>Tags:</strong>
                        ${artifact.tags.map(tag => `<span class="tag">${tag}</span>`).join('')}
                    </div>
                ` : ''}
                
                ${options.includeResults && artifact.results ? `
                    <div class="results">
                        <strong>Execution Results:</strong>
                        <pre>${JSON.stringify(artifact.results, null, 2)}</pre>
                    </div>
                ` : ''}
            </div>
        </div>
    `).join('')}

    <div class="footer">
        <p>Exported from Code Runner | ${artifacts.length} artifacts</p>
    </div>
</body>
</html>`
  }

  /**
   * Generate printable HTML
   */
  private generatePrintableHtml(artifacts: ExportableArtifact[], options: ExportOptions): string {
    const html = this.generateHtmlReport(artifacts, options)
    // Add print-specific styles
    return html.replace('<style>', `<style>
        @media print {
            body { font-size: 12px; }
            .artifact { page-break-inside: avoid; }
            .header { page-break-after: always; }
        }
    `)
  }

  /**
   * Generate metadata
   */
  private generateMetadata(artifact: ExportableArtifact, options: ExportOptions): Record<string, any> {
    return {
      id: artifact.id,
      title: artifact.title,
      type: artifact.type,
      language: artifact.language,
      createdAt: artifact.createdAt,
      updatedAt: artifact.updatedAt,
      tags: artifact.tags,
      contentLength: artifact.content.length,
      contentLines: artifact.content.split('\n').length,
      exportedAt: new Date().toISOString(),
      exportOptions: options,
      ...(options.includeMetadata && artifact.metadata)
    }
  }

  /**
   * Generate manifest
   */
  private generateManifest(artifacts: ExportableArtifact[], options: ExportOptions): Record<string, any> {
    return {
      type: 'artifact_export',
      version: '1.0.0',
      exportedAt: new Date().toISOString(),
      count: artifacts.length,
      options,
      artifacts: artifacts.map(artifact => ({
        id: artifact.id,
        title: artifact.title,
        type: artifact.type,
        language: artifact.language,
        filename: this.generateFileName(artifact),
        size: artifact.content.length
      }))
    }
  }

  /**
   * Generate filename
   */
  private generateFileName(artifact: ExportableArtifact, index?: number): string {
    const sanitizeTitle = artifact.title.replace(/[^a-zA-Z0-9\-_]/g, '_')
    const indexSuffix = index !== undefined ? `_${index + 1}` : ''
    const extension = this.getFileExtension(artifact.type, artifact.language)
    
    return `${sanitizeTitle}${indexSuffix}.${extension}`
  }

  /**
   * Get file extension
   */
  private getFileExtension(type: string, language?: string): string {
    if (language) {
      const extensions = {
        'javascript': 'js',
        'typescript': 'ts',
        'python': 'py',
        'html': 'html',
        'css': 'css',
        'json': 'json',
        'markdown': 'md'
      }
      return extensions[language] || 'txt'
    }
    
    const typeExtensions = {
      'html': 'html',
      'svg': 'svg',
      'json': 'json',
      'mermaid': 'mmd',
      'markdown': 'md'
    }
    return typeExtensions[type] || 'txt'
  }

  /**
   * Prepare content for export
   */
  private prepareContent(artifact: ExportableArtifact, options: ExportOptions): string {
    let content = artifact.content
    
    if (options.includeMetadata) {
      const metadata = this.generateMetadata(artifact, options)
      const metadataComment = this.generateMetadataComment(metadata, artifact.language)
      content = metadataComment + '\n\n' + content
    }
    
    return content
  }

  /**
   * Generate metadata comment
   */
  private generateMetadataComment(metadata: Record<string, any>, language?: string): string {
    const commentStyle = this.getCommentStyle(language)
    const metadataLines = Object.entries(metadata)
      .map(([key, value]) => `${key}: ${typeof value === 'string' ? value : JSON.stringify(value)}`)
      .map(line => `${commentStyle.line} ${line}`)
      .join('\n')
    
    return `${commentStyle.start}\n${metadataLines}\n${commentStyle.end}`
  }

  /**
   * Get comment style for language
   */
  private getCommentStyle(language?: string): { start: string; line: string; end: string } {
    const styles = {
      'javascript': { start: '/**', line: ' *', end: ' */' },
      'typescript': { start: '/**', line: ' *', end: ' */' },
      'python': { start: '"""', line: '', end: '"""' },
      'html': { start: '<!--', line: '', end: '-->' },
      'css': { start: '/**', line: ' *', end: ' */' }
    }
    
    return styles[language || 'javascript'] || { start: '/**', line: ' *', end: ' */' }
  }

  /**
   * Generate README
   */
  private generateReadme(artifacts: ExportableArtifact[], projectName: string): string {
    const languages = [...new Set(artifacts.map(a => a.language).filter(Boolean))]
    const types = [...new Set(artifacts.map(a => a.type))]
    
    return `# ${projectName}

Exported from Code Runner on ${new Date().toLocaleDateString()}

## Project Overview

This project contains ${artifacts.length} artifact(s) with the following characteristics:

- **Languages**: ${languages.join(', ') || 'None'}
- **Types**: ${types.join(', ')}
- **Total Files**: ${artifacts.length}

## Files

${artifacts.map(artifact => `
### ${artifact.title}
- **File**: ${this.generateFileName(artifact)}
- **Type**: ${artifact.type}
- **Language**: ${artifact.language || 'N/A'}
- **Created**: ${new Date(artifact.createdAt).toLocaleDateString()}
${artifact.tags && artifact.tags.length > 0 ? `- **Tags**: ${artifact.tags.join(', ')}` : ''}
`).join('\n')}

## Getting Started

${languages.includes('javascript') || languages.includes('typescript') ? `
### JavaScript/TypeScript
\`\`\`bash
npm install
npm start
\`\`\`
` : ''}

${languages.includes('python') ? `
### Python
\`\`\`bash
pip install -r requirements.txt
python main.py
\`\`\`
` : ''}

## Export Information

- **Generated**: ${new Date().toISOString()}
- **Source**: Code Runner
- **Version**: 1.0.0
`
  }

  /**
   * Generate package.json
   */
  private generatePackageJson(artifacts: ExportableArtifact[], projectName: string): Record<string, any> {
    return {
      name: projectName.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
      version: '1.0.0',
      description: `Exported project containing ${artifacts.length} artifacts`,
      main: 'index.js',
      scripts: {
        start: 'node index.js',
        dev: 'node --watch index.js'
      },
      dependencies: {},
      devDependencies: {},
      keywords: ['code-runner', 'export'],
      author: 'Code Runner',
      license: 'MIT'
    }
  }

  /**
   * Generate requirements.txt
   */
  private generateRequirements(artifacts: ExportableArtifact[]): string {
    const requirements = new Set<string>()
    
    artifacts.forEach(artifact => {
      if (artifact.language === 'python') {
        // Extract import statements and map to package names
        const imports = artifact.content.match(/^(?:from|import)\s+([a-zA-Z_][a-zA-Z0-9_]*)/gm) || []
        imports.forEach(importLine => {
          const packageName = importLine.replace(/^(?:from|import)\s+/, '').split('.')[0]
          const commonPackages = {
            'numpy': 'numpy',
            'pandas': 'pandas',
            'matplotlib': 'matplotlib',
            'scipy': 'scipy',
            'sklearn': 'scikit-learn',
            'requests': 'requests',
            'bs4': 'beautifulsoup4',
            'PIL': 'pillow',
            'sympy': 'sympy',
            'networkx': 'networkx',
            'seaborn': 'seaborn',
            'plotly': 'plotly',
            'bokeh': 'bokeh',
            'altair': 'altair'
          }
          
          if (commonPackages[packageName]) {
            requirements.add(commonPackages[packageName])
          }
        })
      }
    })
    
    return Array.from(requirements).sort().join('\n')
  }

  /**
   * Utility methods
   */
  private downloadJson(data: any, filename: string): void {
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    saveAs(blob, filename)
  }

  private createDataUrl(data: any): string {
    const jsonString = JSON.stringify(data)
    return `data:application/json;base64,${btoa(jsonString)}`
  }

  private escapeHtml(text: string): string {
    const div = document.createElement('div')
    div.textContent = text
    return div.innerHTML
  }

  private getArtifactFolder(artifact: ExportableArtifact, parentFolder: JSZip): JSZip {
    const folderName = artifact.type
    return parentFolder.folder(folderName) || parentFolder
  }

  private generateResultsFile(results: ExecutionResult[]): string {
    return JSON.stringify({
      type: 'execution_results',
      timestamp: new Date().toISOString(),
      results
    }, null, 2)
  }

  private async exportHistoryAsZip(exportData: any, options: ExportOptions): Promise<void> {
    const zip = new JSZip()
    
    // Add main history file
    zip.file('history.json', JSON.stringify(exportData, null, 2))
    
    // Add individual execution files
    exportData.history.forEach((entry: any, index: number) => {
      const filename = `execution_${index + 1}_${entry.id}.${this.getFileExtension('code', entry.language)}`
      zip.file(filename, entry.code)
    })
    
    const content = await zip.generateAsync({ type: 'blob' })
    saveAs(content, `execution_history_${new Date().toISOString().split('T')[0]}.zip`)
  }

  private exportHistoryAsHtml(exportData: any, options: ExportOptions): void {
    const html = `<!DOCTYPE html>
<html>
<head>
    <title>Execution History</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .entry { margin: 20px 0; padding: 15px; border: 1px solid #ddd; }
        .code { background: #f4f4f4; padding: 10px; font-family: monospace; }
    </style>
</head>
<body>
    <h1>Execution History</h1>
    <p>Exported: ${exportData.exportedAt}</p>
    <p>Total entries: ${exportData.count}</p>
    
    ${exportData.history.map((entry: any, index: number) => `
        <div class="entry">
            <h3>Execution ${index + 1}</h3>
            <p><strong>Date:</strong> ${new Date(entry.timestamp).toLocaleString()}</p>
            <p><strong>Language:</strong> ${entry.language}</p>
            <p><strong>Success:</strong> ${entry.success ? 'Yes' : 'No'}</p>
            <p><strong>Execution Time:</strong> ${entry.executionTime}ms</p>
            <div class="code"><pre>${this.escapeHtml(entry.code)}</pre></div>
        </div>
    `).join('')}
</body>
</html>`
    
    const blob = new Blob([html], { type: 'text/html' })
    saveAs(blob, `execution_history_${new Date().toISOString().split('T')[0]}.html`)
  }

  private async exportTemplatesAsZip(exportData: any, options: ExportOptions): Promise<void> {
    const zip = new JSZip()
    
    // Add main templates file
    zip.file('templates.json', JSON.stringify(exportData, null, 2))
    
    // Add individual template files
    exportData.templates.forEach((template: any) => {
      const filename = this.generateFileName(template)
      zip.file(filename, template.code)
    })
    
    const content = await zip.generateAsync({ type: 'blob' })
    saveAs(content, `code_templates_${new Date().toISOString().split('T')[0]}.zip`)
  }
}

// Export singleton instance
export const exportService = ExportService.getInstance()

// Export composable for Vue components
export function useExportService() {
  return {
    exportArtifact: exportService.exportArtifact.bind(exportService),
    exportArtifacts: exportService.exportArtifacts.bind(exportService),
    exportHistory: exportService.exportHistory.bind(exportService),
    exportTemplates: exportService.exportTemplates.bind(exportService),
    exportProject: exportService.exportProject.bind(exportService),
    shareArtifact: exportService.shareArtifact.bind(exportService)
  }
}