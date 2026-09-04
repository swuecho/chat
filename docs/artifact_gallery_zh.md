# Artifact 画廊

Artifact 画廊集中展示聊天会话中创建的结构化内容。Artifact 可用于查看、复制、编辑、整理和下载；应用不会运行 Artifact 中的代码。

## 支持的类型

- 代码：带语言标识的静态源代码
- HTML：经过清理并在沙盒中显示的静态预览
- SVG：经过清理的图片预览
- Mermaid：使用严格安全模式生成的图表预览
- JSON：格式化的数据预览
- Markdown：经过清理的渲染预览

## 创建 Artifact

在会话设置中启用 Artifacts，然后让助手生成结构化内容。代码围栏的起始行需包含标记：

````markdown
```javascript <!-- artifact: 描述性标题 -->
export function example() {
  return '仅展示源代码'
}
```
````

仅支持 `artifact` 标记；可执行标记会被忽略。

## 画廊操作

画廊支持搜索，并可按类型、语言和会话筛选。每个 Artifact 都可预览、编辑、复制或删除，也可将当前结果页导出为 JSON。

HTML 预览不允许运行脚本、内联事件处理器、嵌入式框架、对象或 `javascript:` URL。代码 Artifact 始终作为文本处理。
