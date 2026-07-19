<template>
  <div class="markdown-content" v-html="renderedMarkdown"></div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'

// 引入代码高亮样式
import 'highlight.js/styles/github.css'

interface Props {
  content: string
}

const props = defineProps<Props>()

// 配置 markdown-it 实例
const md: MarkdownIt = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  highlight: function (str: string, lang: string): string {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return (
          '<pre class="hljs"><code>' +
          hljs.highlight(str, { language: lang, ignoreIllegals: true }).value +
          '</code></pre>'
        )
      } catch {
        // 忽略错误，使用默认处理
      }
    }

    return '<pre class="hljs"><code>' + md.utils.escapeHtml(str) + '</code></pre>'
  },
})

// 预处理内容，确保 HTML 块级元素不被 MarkdownIt 包裹在 <p> 标签内
const preprocessContent = (content: string): string => {
  // 将 <div class="tool-history"> 和 <details> 标签前后添加足够的空行
  // 确保 MarkdownIt 将它们视为块级 HTML 元素
  let processed = content

  // 处理 tool-history div 标签
  processed = processed.replace(/(<div class="tool-history[^>]*>)/g, '\n\n$1')
  processed = processed.replace(/(<\/div>)/g, '$1\n\n')

  // 处理 details 标签
  processed = processed.replace(/(<details>)/g, '\n\n$1')
  processed = processed.replace(/(<\/details>)/g, '$1\n\n')

  return processed
}

// 计算渲染后的 Markdown
const renderedMarkdown = computed(() => {
  const processed = preprocessContent(props.content)
  return md.render(processed)
})
</script>

<style scoped>
.markdown-content {
  line-height: 1.6;
  color: #333;
  word-wrap: break-word;
}

/* 全局样式，影响 v-html 内容 */
.markdown-content :deep(h1),
.markdown-content :deep(h2),
.markdown-content :deep(h3),
.markdown-content :deep(h4),
.markdown-content :deep(h5),
.markdown-content :deep(h6) {
  margin: 1.5em 0 0.5em 0;
  font-weight: 600;
  line-height: 1.25;
}

.markdown-content :deep(h1) {
  font-size: 1.5em;
  border-bottom: 1px solid #eee;
  padding-bottom: 0.3em;
}

.markdown-content :deep(h2) {
  font-size: 1.3em;
  border-bottom: 1px solid #eee;
  padding-bottom: 0.3em;
}

.markdown-content :deep(h3) {
  font-size: 1.1em;
}

.markdown-content :deep(p) {
  margin: 0.8em 0;
}

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  margin: 0.8em 0;
  padding-left: 1.5em;
}

.markdown-content :deep(li) {
  margin: 0.3em 0;
}

.markdown-content :deep(blockquote) {
  margin: 1em 0;
  padding: 0.5em 1em;
  border-left: 4px solid #ddd;
  background-color: #f9f9f9;
  color: #666;
}

.markdown-content :deep(code) {
  background-color: #f1f1f1;
  padding: 0.2em 0.4em;
  border-radius: 3px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.9em;
}

.markdown-content :deep(pre) {
  background-color: #f8f8f8;
  border: 1px solid #e1e1e1;
  border-radius: 6px;
  padding: 1em;
  overflow-x: auto;
  margin: 1em 0;
}

.markdown-content :deep(pre code) {
  background-color: transparent;
  padding: 0;
  border-radius: 0;
  font-size: 0.9em;
  line-height: 1.4;
}

.markdown-content :deep(table) {
  border-collapse: collapse;
  margin: 1em 0;
  width: 100%;
}

.markdown-content :deep(table th),
.markdown-content :deep(table td) {
  border: 1px solid #ddd;
  padding: 0.5em 0.8em;
  text-align: left;
}

.markdown-content :deep(table th) {
  background-color: #f5f5f5;
  font-weight: 600;
}

.markdown-content :deep(table tr:nth-child(even)) {
  background-color: #f9f9f9;
}

.markdown-content :deep(a) {
  color: #1890ff;
  text-decoration: none;
}

.markdown-content :deep(a:hover) {
  text-decoration: underline;
}

.markdown-content :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 4px;
  margin: 0.5em 0;
}

.markdown-content :deep(hr) {
  border: none;
  border-top: 1px solid #eee;
  margin: 1.5em 0;
}

/* 代码高亮样式优化 */
.markdown-content :deep(.hljs) {
  background-color: #f8f8f8 !important;
  border-radius: 6px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.9em;
  line-height: 1.4;
}

/* 特定语言的代码块样式 */
.markdown-content :deep(.hljs-keyword) {
  color: #d73a49;
  font-weight: 600;
}

.markdown-content :deep(.hljs-string) {
  color: #032f62;
}

.markdown-content :deep(.hljs-comment) {
  color: #6a737d;
  font-style: italic;
}

.markdown-content :deep(.hljs-number) {
  color: #005cc5;
}

.markdown-content :deep(.hljs-function) {
  color: #6f42c1;
}

.markdown-content :deep(.hljs-tag) {
  color: #22863a;
}

.markdown-content :deep(.hljs-attr) {
  color: #6f42c1;
}

.markdown-content :deep(.hljs-title) {
  color: #6f42c1;
  font-weight: 600;
}

/* 深度思考 details 标签样式 - 折叠框效果 */
.markdown-content :deep(details) {
  margin: 8px 0;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
  overflow: hidden;
}

.markdown-content :deep(summary) {
  padding: 8px 12px;
  cursor: pointer;
  font-size: 13px;
  color: #666;
  user-select: none;
  background: #f0f1f2;
  display: flex;
  justify-content: space-between;
  align-items: center;
  list-style: none;
}

.markdown-content :deep(summary::-webkit-details-marker) {
  display: none;
}

.markdown-content :deep(summary::marker) {
  display: none;
}

.markdown-content :deep(summary:hover) {
  background: #e8e9ea;
}

/* 箭头指示器 - 收起时向下 */
.markdown-content :deep(summary)::after {
  content: '▼';
  font-size: 10px;
  color: #999;
}

/* 展开时箭头旋转向上 */
.markdown-content :deep(details[open] > summary)::after {
  transform: rotate(180deg);
}

.markdown-content :deep(details[open] > summary) {
  border-bottom: 1px solid #e8e8e8;
}

.markdown-content :deep(details > *:not(summary)) {
  padding: 8px 12px;
  font-size: 13px;
  color: #555;
  max-height: 300px;
  overflow-y: auto;
}

/* 工具调用文本样式 - 黑色条框 */
.markdown-content :deep(p) {
  margin: 0.5em 0;
}

/* 工具调用文本中的代码块 */
.markdown-content :deep(pre) {
  margin: 0.5em 0;
}

/* 历史工具调用卡片样式 - 黑色条框 */
.markdown-content :deep(.tool-history) {
  background: #1e1e1e;
  border-radius: 8px;
  margin: 8px 0;
  overflow: hidden;
  font-size: 13px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
}

.markdown-content :deep(.tool-history.tool-pending) {
  border-left: 3px solid #1890ff;
}

.markdown-content :deep(.tool-history.tool-done) {
  border-left: 3px solid #52c41a;
  opacity: 0.85;
}

.markdown-content :deep(.tool-history .tool-name) {
  color: #e0e0e0;
  font-weight: 500;
}

.markdown-content :deep(.tool-history .tool-path) {
  color: #ce9178;
  font-size: 12px;
  margin-left: 8px;
}

.markdown-content :deep(.tool-history .tool-status) {
  font-size: 12px;
  padding: 1px 8px;
  border-radius: 10px;
}

.markdown-content :deep(.tool-history.tool-pending .tool-status) {
  color: #1890ff;
  background: rgba(24, 144, 255, 0.15);
}

.markdown-content :deep(.tool-history.tool-done .tool-status) {
  color: #52c41a;
  background: rgba(82, 196, 26, 0.15);
}
</style>
