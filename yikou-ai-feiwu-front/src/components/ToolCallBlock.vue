<template>
  <div :class="['tool-call-block', { 'tool-completed': completed }]">
    <div class="tool-header">
      <div class="tool-header-left">
        <a-spin v-if="!completed" size="small" class="tool-spin" />
        <span v-else class="tool-icon">✓</span>
        <span class="tool-name">{{ displayName }}</span>
        <span v-if="toolPath" class="tool-path">{{ toolPath }}</span>
      </div>
      <span :class="['tool-status', completed ? 'status-done' : 'status-running']">
        {{ completed ? '完成' : '执行中...' }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  toolName: string
  displayName: string
  arguments: string
  completed: boolean
}

const props = defineProps<Props>()

const toolNameMap: Record<string, string> = {
  writeFile: '写入文件',
  readFile: '读取文件',
  modifyFile: '修改文件',
  readDir: '读取目录',
  deleteFile: '删除文件',
  exit: '退出工具调用',
}

const displayName = computed(() => {
  return toolNameMap[props.toolName] || props.toolName
})

// 从参数中解析路径
const toolPath = computed(() => {
  if (!props.arguments) return ''
  try {
    const args = JSON.parse(props.arguments)
    // 不同工具的路径字段名
    const pathFields = ['relative_path', 'relativeFilePath', 'path']
    for (const field of pathFields) {
      if (args[field]) {
        return args[field]
      }
    }
    return ''
  } catch {
    return ''
  }
})
</script>

<style scoped>
.tool-call-block {
  background: #1e1e1e;
  border-radius: 8px;
  margin: 8px 0;
  overflow: hidden;
  font-size: 13px;
  transition: all 0.3s ease;
}

.tool-call-block:not(.tool-completed) {
  border-left: 3px solid #1890ff;
}

.tool-call-block.tool-completed {
  border-left: 3px solid #52c41a;
  opacity: 0.85;
}

.tool-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
}

.tool-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tool-icon {
  color: #52c41a;
  font-weight: bold;
}

.tool-spin {
  display: inline-flex;
}

.tool-spin :deep(.ant-spin-dot-item) {
  background: #1890ff;
}

.tool-name {
  color: #e0e0e0;
  font-weight: 500;
}

.tool-path {
  color: #ce9178;
  font-size: 12px;
  margin-left: 4px;
}

.tool-status {
  font-size: 12px;
  padding: 1px 8px;
  border-radius: 10px;
}

.status-running {
  color: #1890ff;
  background: rgba(24, 144, 255, 0.15);
}

.status-done {
  color: #52c41a;
  background: rgba(82, 196, 26, 0.15);
}
</style>