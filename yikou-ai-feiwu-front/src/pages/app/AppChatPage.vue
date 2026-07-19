<template>
  <div id="appChatPage">
    <!-- 顶部栏 -->
    <a-layout-header class="header">
      <div class="header-content">
        <div class="app-name">
          {{ appInfo.appName || '应用对话' }}
          <a-tag v-if="appInfo.codeGenType" color="blue" class="code-gen-type-tag">
            {{ formatCodeGenType(appInfo.codeGenType) }}
          </a-tag>
        </div>
        <div class="buttons">
          <a-button type="default" @click="showAppDetail">
            <template #icon>
              <InfoCircleOutlined />
            </template>
            应用详情
          </a-button>
          <a-button type="primary" @click="doDeploy" :loading="deployLoading">
            {{
              deployLoading
                ? '部署中...'
                : appInfo.deployKey !== null && appInfo.deployKey != ''
                  ? '取消部署'
                  : '部署'
            }}
          </a-button>
          <a-button
            type="primary"
            ghost
            @click="downloadCode"
            :loading="downloading"
            :disabled="!isOwner"
          >
            <template #icon>
              <DownloadOutlined />
            </template>
            下载代码
          </a-button>
          <a-button
            v-if="isOwner && previewUrl"
            type="link"
            :danger="isEditMode"
            @click="toggleEditMode"
            :class="{ 'edit-mode-active': isEditMode }"
            style="padding: 0; height: auto; margin-right: 12px"
          >
            <template #icon>
              <EditOutlined />
            </template>
            {{ isEditMode ? '退出编辑' : '编辑模式' }}
          </a-button>
        </div>
      </div>
    </a-layout-header>
    <a-layout class="main-layout">
      <!-- 左侧对话区域 -->
      <a-layout-content class="chat-container">
        <!-- 消息区域 -->
        <div
          class="messages-container"
          ref="messagesContainerRef"
        >
          <div class="load-more-container" v-if="hasMoreMessages && messages.length > 0">
            <a-button
              type="link"
              @click="loadChatHistory"
              :loading="loadingHistory"
              :disabled="loadingHistory"
            >
              加载更多
            </a-button>
          </div>
          <div
            v-for="msg in messages"
            :key="msg.id"
            :class="['message', msg.role]"
            :data-id="msg.id"
          >
            <div class="message-content">
              <div class="avatar">
                <a-avatar
                  v-if="msg.role === 'user'"
                  :src="loginUserStore.loginUser.userAvatar"
                  :alt="loginUserStore.loginUser.userName"
                />
                <a-avatar v-else src="/src/assets/logo.png" alt="AI助手" />
              </div>
              <div class="content">
                <template v-if="msg.role === 'ai'">
                  <template v-for="(seg, idx) in msg.segments" :key="idx">
                    <MarkdownRenderer v-if="seg.type === 'text'" :content="seg.data" />
                    <ToolCallBlock
                      v-else-if="seg.type === 'tool_request'"
                      :tool-name="seg.name"
                      :display-name="seg.displayName"
                      :arguments="seg.arguments"
                      :completed="false"
                    />
                    <ToolCallBlock
                      v-else-if="seg.type === 'tool_executed'"
                      :tool-name="seg.name"
                      :display-name="seg.displayName"
                      :arguments="seg.arguments"
                      :completed="true"
                    />
                    <div v-else-if="seg.type === 'reasoning'" class="reasoning-block">
                      <details>
                        <summary>
                          <span class="summary-left">深度思考</span>
                          <span class="summary-arrow">▼</span>
                        </summary>
                        <div class="reasoning-content">
                          <MarkdownRenderer :content="seg.data" />
                        </div>
                      </details>
                    </div>
                  </template>
                </template>
                <div v-else-if="msg.role === 'user'" class="user-content">{{ msg.content }}</div>
              </div>
            </div>
          </div>
        </div>
        <!-- 选中元素信息展示 -->
        <a-alert
          v-if="selectedElementInfo"
          class="selected-element-alert"
          type="info"
          closable
          @close="clearSelectedElement"
        >
          <template #message>
            <div class="selected-element-info">
              <div class="element-header">
                <span class="element-tag">
                  选中元素：{{ selectedElementInfo.tagName.toLowerCase() }}
                </span>
                <span v-if="selectedElementInfo.id" class="element-id">
                  #{{ selectedElementInfo.id }}
                </span>
                <span v-if="selectedElementInfo.className" class="element-class">
                  .{{ selectedElementInfo.className.split(' ').join('.') }}
                </span>
              </div>
              <div class="element-details">
                <div v-if="selectedElementInfo.textContent" class="element-item">
                  内容: {{ selectedElementInfo.textContent.substring(0, 50) }}
                  {{ selectedElementInfo.textContent.length > 50 ? '...' : '' }}
                </div>
                <div v-if="selectedElementInfo.pagePath" class="element-item">
                  页面路径: {{ selectedElementInfo.pagePath }}
                </div>
                <div class="element-item">
                  选择器:
                  <code class="element-selector-code">{{ selectedElementInfo.selector }}</code>
                </div>
              </div>
            </div>
          </template>
        </a-alert>
        <!-- 输入框 -->
        <div class="input-container">
          <a-tooltip :title="isInputDisabled ? '无法在别人的作品下对话哦~' : ''" placement="top">
            <a-input
              v-model:value="inputMessage"
              :placeholder="getInputPlaceholder()"
              @pressEnter="sendMessage"
              :disabled="sending || isInputDisabled"
            />
          </a-tooltip>
          <a-button
            v-if="!sending"
            type="primary"
            @click="sendMessage"
            :loading="sending"
            :disabled="!inputMessage.trim() || isInputDisabled"
          >
            发送
          </a-button>
          <a-button v-else type="primary" @click="handleStop" danger>
            <template #icon>
              <PauseOutlined />
            </template>
          </a-button>
        </div>
      </a-layout-content>
      <!-- 右侧网页展示区域 -->
      <a-layout-sider width="60%" class="preview-container">
        <div v-if="streamLoading" class="loading-container">
          <a-spin size="large" />
          <p>AI正在生成中...</p>
        </div>
        <iframe
          v-else-if="previewUrl"
          :src="previewUrl"
          class="preview-frame"
          frameborder="0"
          @load="onIframeLoad"
        ></iframe>
        <div v-else class="empty-preview">
          <p>暂无预览内容</p>
        </div>
      </a-layout-sider>
    </a-layout>
    <!-- 应用详情弹窗 -->
    <AppDetailModal
      v-model:open="appDetailVisible"
      :app="appInfo"
      :show-actions="isOwner || isAdmin"
      @edit="editApp"
      @delete="deleteApp"
    />
    <!-- 部署成功弹窗 -->
    <DeploySuccessModal
      v-model:open="deployModalVisible"
      :deploy-url="deployUrl"
      @open-site="openDeployedSite"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, nextTick, computed } from 'vue'
import { message } from 'ant-design-vue'
import { useRoute, useRouter } from 'vue-router'
import {
  getAppVoById,
  deployApp,
  deployAppCancel,
  deleteApp as deleteAppApi,
  stopToGenCode,
} from '@/api/appController.ts'
import { listAppChatHistory } from '@/api/chatHistoryController.ts'

import 'highlight.js/styles/github.css'
import { useLoginUserStore } from '@/stores/loginUser.ts'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import ToolCallBlock from '@/components/ToolCallBlock.vue'
import DeploySuccessModal from '@/components/DeploySuccessModal.vue'
import {
  InfoCircleOutlined,
  DownloadOutlined,
  EditOutlined,
  PauseOutlined,
} from '@ant-design/icons-vue'
import AppDetailModal from '@/components/AppDetailModal.vue'
import { getStaticPreviewUrl } from '@/config/env.ts'
import request from '@/request.ts'
import { formatCodeGenType } from '@/utils/codeGenTypes'
import { type ElementInfo, VisualEditor } from '@/utils/visualEditor.ts'

const loginUserStore = useLoginUserStore()

const route = useRoute()

// 应用信息
const appInfo = ref<API.AppVO>({})
const appId = ref<string>()

// 消息相关
interface MessageSegment {
  type: 'text' | 'tool_request' | 'tool_executed' | 'reasoning'
  data: string
  name?: string
  displayName?: string
  arguments?: string
  id?: string
}

interface ChatMessage {
  id: number
  role: string
  content: string
  segments: MessageSegment[]
}

const messages = ref<ChatMessage[]>([])
const inputMessage = ref('')
const sending = ref(false)
const messagesContainerRef = ref<HTMLDivElement | null>(null)

// SSE流相关
let eventSource: EventSource | null = null
const messagesContainerHeight = ref(0)

// 部署相关
const deployLoading = ref(false)
const previewUrl = ref<string>('')

// 部署成功弹窗相关
const deployModalVisible = ref(false)
const deployUrl = ref('')

// 流式响应相关
const streamLoading = ref(false)

// 对话历史相关
const lastCreateTime = ref<string | null>(null)
const hasMoreMessages = ref(true)
const loadingHistory = ref(false)

// 权限相关
// 判断是否为自己的应用
const isMyApp = computed(() => {
  return appInfo.value.userId === loginUserStore.loginUser.id
})

// 可视化编辑相关
const isEditMode = ref(false)
const selectedElementInfo = ref<ElementInfo | null>(null)
const visualEditor = new VisualEditor({
  onElementSelected: (elementInfo: ElementInfo) => {
    selectedElementInfo.value = elementInfo
  },
})

// 绑定消息处理函数（保持同一引用以便移除监听器）
const handleIframeMessage = visualEditor.handleIframeMessage.bind(visualEditor)

// 输入框禁用状态
const isInputDisabled = computed(() => {
  // 如果应用信息还未加载，不禁用
  if (!appInfo.value.id) {
    return false
  }
  // 如果是自己的应用，不禁用
  if (isMyApp.value) {
    return false
  }
  // 如果不是自己的应用，禁用输入框
  return true
})

// 可视化编辑相关函数
const toggleEditMode = () => {
  // 检查 iframe 是否已经加载
  const iframe = document.querySelector('.preview-frame') as HTMLIFrameElement
  if (!iframe) {
    message.warning('请等待页面加载完成')
    return
  }
  // 确保 visualEditor 已初始化
  if (!previewUrl.value) {
    message.warning('请等待页面加载完成')
    return
  }
  // 初始化 visualEditor 的 iframe 引用
  visualEditor.init(iframe)
  const newEditMode = visualEditor.toggleEditMode()
  isEditMode.value = newEditMode
}

const clearSelectedElement = () => {
  selectedElementInfo.value = null
  visualEditor.clearSelection()
}

// iframe 加载完成回调
const onIframeLoad = () => {
  const iframe = document.querySelector('.preview-frame') as HTMLIFrameElement
  if (iframe) {
    visualEditor.init(iframe)
    visualEditor.onIframeLoad()
  }
}

// 获取应用信息
const fetchAppInfo = async () => {
  const id = route.params.id as string
  if (!id) {
    message.error('应用ID不存在')
    return
  }

  appId.value = id

  const res = await getAppVoById({ id: id })
  if (res.data.code === 0 && res.data.data) {
    appInfo.value = res.data.data
    // 加载对话历史
    await loadChatHistory()
    // 检查是否需要自动发送初始提示词
    checkAndSendInitialMessage()
  } else {
    message.error('获取应用信息失败，' + res.data.message)
  }
}

const getInputPlaceholder = () => {
  if (selectedElementInfo.value) {
    return `正在编辑 ${selectedElementInfo.value.tagName.toLowerCase()} 元素，描述您想要的修改...`
  }
  return '请描述你想生成的网站，越详细效果越好哦'
}

// 工具名称映射
const toolNameMap: Record<string, string> = {
  writeFile: '写入文件',
  readFile: '读取文件',
  modifyFile: '修改文件',
  readDir: '读取目录',
  deleteFile: '删除文件',
  exit: '退出工具调用',
}

// 解析历史消息中的 JSON 流式数据，构建 segments
const parseHistoryMessage = (rawMessage: string): MessageSegment[] => {
  const segments: MessageSegment[] = []

  // 逐字符解析，找到完整的 JSON 对象
  // 这样可以处理包含嵌套 JSON 的字段（如 arguments）
  const extractJsonObjects = (str: string): string[] => {
    const result: string[] = []
    let depth = 0
    let start = -1

    for (let i = 0; i < str.length; i++) {
      const char = str[i]
      if (char === '{') {
        if (depth === 0) {
          start = i
        }
        depth++
      } else if (char === '}') {
        depth--
        if (depth === 0 && start !== -1) {
          result.push(str.substring(start, i + 1))
          start = -1
        }
      }
    }

    return result
  }

  const jsonObjects = extractJsonObjects(rawMessage)

  if (jsonObjects.length === 0) {
    // 如果没有解析到 JSON 格式，直接作为文本返回
    return [{ type: 'text', data: rawMessage }]
  }

  // 处理每个 JSON 对象
  for (const jsonStr of jsonObjects) {
    try {
      const jsonObj = JSON.parse(jsonStr)
      const msgType = jsonObj.type

      if (msgType === 'ai_response') {
        // 连续的 ai_response 需要合并
        const lastSeg = segments[segments.length - 1]
        if (lastSeg && lastSeg.type === 'text') {
          lastSeg.data += jsonObj.data || ''
        } else {
          segments.push({ type: 'text', data: jsonObj.data || '' })
        }
      } else if (msgType === 'reasoning') {
        // 连续的 reasoning 需要合并
        const lastSeg = segments[segments.length - 1]
        if (lastSeg && lastSeg.type === 'reasoning') {
          lastSeg.data += jsonObj.data || ''
        } else {
          segments.push({ type: 'reasoning', data: jsonObj.data || '' })
        }
      } else if (msgType === 'tool_request') {
        const displayName = toolNameMap[jsonObj.name] || jsonObj.name
        segments.push({
          type: 'tool_request',
          data: '',
          name: jsonObj.name,
          displayName: displayName,
          arguments: jsonObj.arguments,
          id: jsonObj.id,
        })
      } else if (msgType === 'tool_executed') {
        const displayName = toolNameMap[jsonObj.name] || jsonObj.name
        segments.push({
          type: 'tool_executed',
          data: '',
          name: jsonObj.name,
          displayName: displayName,
          arguments: jsonObj.arguments,
          id: jsonObj.id,
        })
      }
    } catch {
      // JSON 解析失败，忽略这个片段
    }
  }

  // 筛选工具调用：只保留 tool_executed，过滤掉对应的 tool_request
  // 因为历史消息中工具已完成，不需要显示"执行中"状态
  const toolRequestIds = new Set<string>()
  segments.forEach((seg) => {
    if (seg.type === 'tool_request' && seg.id) {
      toolRequestIds.add(seg.id)
    }
  })

  const filteredSegments = segments.filter((seg) => {
    // 如果是 tool_request，且有对应的 tool_executed（相同 id），则过滤掉
    if (seg.type === 'tool_request' && seg.id) {
      // 检查是否存在相同 id 的 tool_executed
      const hasExecuted = segments.some(
        (s) => s.type === 'tool_executed' && s.id === seg.id
      )
      if (hasExecuted) {
        return false // 过滤掉重复的 tool_request
      }
    }
    return true
  })

  return filteredSegments.length > 0 ? filteredSegments : [{ type: 'text', data: rawMessage }]
}

// 加载对话历史
const loadChatHistory = async () => {
  if (!appId.value || !hasMoreMessages.value) {
    return
  }

  loadingHistory.value = true
  try {
    const res = await listAppChatHistory({
      appId: appId.value,
      pageSize: 10,
      ...(lastCreateTime.value ? { lastCreateTime: lastCreateTime.value } : {}),
    })

    if (res.data.code === 0 && res.data.data) {
      const records = res.data.data.records || []

      // 如果没有更多消息，设置标志位
      if (records.length < 10) {
        hasMoreMessages.value = false
      }

      // 如果有记录，更新最后创建时间游标
      if (records.length > 0) {
        const lastRecord = records[records.length - 1]
        lastCreateTime.value = lastRecord.createTime || null
      }

      // 将历史消息转换为前端格式并添加到消息列表开头
      const historyMessages = records
        .map((record) => {
          if (record.messageType === 'user') {
            return {
              id: record.id || Date.now() + Math.random(),
              role: 'user',
              content: record.message || '',
              segments: [],
            }
          } else {
            // AI消息需要解析JSON流式数据
            const segments = parseHistoryMessage(record.message || '')
            return {
              id: record.id || Date.now() + Math.random(),
              role: 'ai',
              content: record.message || '',
              segments: segments,
            }
          }
        })
        .reverse() // 按时间升序排列

      // 添加到消息列表开头
      messages.value = [...historyMessages, ...messages.value]

      // 如果是第一次加载且消息数量大于等于2，显示预览
      if (messages.value.length >= 2) {
        showPreview()
      }
    } else {
      message.error('获取对话历史失败，' + res.data.message)
    }
  } catch (error) {
    message.error('获取对话历史失败，请重试')
  } finally {
    loadingHistory.value = false
  }
}

// 检查并发送初始消息
const checkAndSendInitialMessage = () => {
  // 只有当是自己的应用且没有对话历史时才自动发送初始提示词
  if (isMyApp.value && messages.value.length === 0 && appInfo.value.initPrompt) {
    sendInitialMessage(appInfo.value.initPrompt)
  }
}

// 发送初始消息
const sendInitialMessage = async (prompt: string) => {
  // 添加用户消息
  const userMsg = {
    id: Date.now(),
    role: 'user',
    content: prompt,
    segments: [],
  }
  messages.value.push(userMsg)

  // 添加AI消息占位符
  const aiMsg = {
    id: Date.now() + 1,
    role: 'ai',
    content: '',
    segments: [],
  }
  messages.value.push(aiMsg)
  const aiMsgIndex = messages.value.length - 1

  // 滚动到底部
  await nextTick()
  scrollToBottom()

  // 发送请求
  await sendSSEMessage(prompt, aiMsgIndex)
}

// 发送消息
const sendMessage = async () => {
  if (!inputMessage.value.trim() || sending.value) {
    return
  }

  let userMsgContent = inputMessage.value
  inputMessage.value = ''
  // 如果有选中的元素，将元素信息添加到提示词中
  if (selectedElementInfo.value) {
    let elementContext = `\n\n选中元素信息：`
    if (selectedElementInfo.value.pagePath) {
      elementContext += `\n- 页面路径: ${selectedElementInfo.value.pagePath}`
    }
    elementContext += `\n- 标签: ${selectedElementInfo.value.tagName.toLowerCase()}\n- 选择器: ${selectedElementInfo.value.selector}`
    if (selectedElementInfo.value.textContent) {
      elementContext += `\n- 当前内容: ${selectedElementInfo.value.textContent.substring(0, 100)}`
    }
    userMsgContent += elementContext
  }

  // 添加用户消息
  const userMsg = {
    id: Date.now(),
    role: 'user',
    content: userMsgContent,
    segments: [],
  }
  messages.value.push(userMsg)

  // 添加AI消息占位符
  const aiMsg = {
    id: Date.now() + 1,
    role: 'ai',
    content: '',
    segments: [],
  }
  messages.value.push(aiMsg)
  const aiMsgIndex = messages.value.length - 1

  // 滚动到底部
  await nextTick()
  scrollToBottom()

  // 发送请求
  await sendSSEMessage(userMsgContent, aiMsgIndex)
}

// 发送SSE消息
const sendSSEMessage = async (content: string, aiMsgIndex: number) => {
  if (!appId.value) {
    message.error('应用ID不存在')
    return
  }

  sending.value = true
  streamLoading.value = true

  try {
    // 使用原生EventSource接收流式响应
    eventSource = new EventSource(
      `${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8123/api'}/app/chat/gen/code?appId=${appId.value}&message=${encodeURIComponent(content)}`,
    )

    // 处理数据
    let fullData = ''
    // 工具名称映射
    const toolNameMap: Record<string, string> = {
      writeFile: '写入文件',
      readFile: '读取文件',
      modifyFile: '修改文件',
      readDir: '读取目录',
      deleteFile: '删除文件',
      exit: '退出工具调用',
    }

    // 记录当前正在执行的工具请求，用于匹配对应的执行结果
    const pendingToolRequests: Record<string, { name: string; displayName: string; arguments: string }> = {}

    eventSource.onmessage = (event) => {
      let processedData = ''
      try {
        // 解析JSON数据
        const jsonData = JSON.parse(event.data)
        // 提取d字段的内容
        processedData = jsonData.d || ''
      } catch (error) {
        // 如果JSON解析失败，直接使用原始数据
        processedData = event.data
      }

      // 过滤心跳包内容
      if (processedData === 'heartBeat') {
        return
      }

      fullData += processedData

      // 尝试解析消息类型，构建分段内容
      if (messages.value[aiMsgIndex]) {
        const aiMsg = messages.value[aiMsgIndex]
        try {
          const streamMsg = JSON.parse(processedData)
          const msgType = streamMsg.type

          if (msgType === 'ai_response') {
            // AI文本响应，追加到最后一个text段或新建
            const lastSeg = aiMsg.segments[aiMsg.segments.length - 1]
            if (lastSeg && lastSeg.type === 'text') {
              lastSeg.data += streamMsg.data
            } else {
              aiMsg.segments.push({ type: 'text', data: streamMsg.data })
            }
          } else if (msgType === 'tool_request') {
            // 工具请求，添加工具调用段
            const displayName = toolNameMap[streamMsg.name] || streamMsg.name
            const seg: MessageSegment = {
              type: 'tool_request',
              data: '',
              name: streamMsg.name,
              displayName: displayName,
              arguments: streamMsg.arguments,
              id: streamMsg.id,
            }
            aiMsg.segments.push(seg)
            // 记录待处理的工具请求
            if (streamMsg.id) {
              pendingToolRequests[streamMsg.id] = {
                name: streamMsg.name,
                displayName: displayName,
                arguments: streamMsg.arguments,
              }
            }
          } else if (msgType === 'tool_executed') {
            // 工具执行完成，更新对应的工具请求段
            const toolId = streamMsg.id
            const displayName = toolNameMap[streamMsg.name] || streamMsg.name
            // 查找对应的tool_request段的索引并替换整个对象（确保Vue响应式更新）
            if (toolId) {
              const reqSegIndex = aiMsg.segments.findIndex(
                (s) => s.type === 'tool_request' && s.id === toolId
              )
              if (reqSegIndex !== -1) {
                // 替换整个segment对象，确保Vue响应式系统检测到变化
                aiMsg.segments[reqSegIndex] = {
                  ...aiMsg.segments[reqSegIndex],
                  type: 'tool_executed',
                }
              } else {
                // 没找到对应的请求段，新增一个完成段
                aiMsg.segments.push({
                  type: 'tool_executed',
                  data: '',
                  name: streamMsg.name,
                  displayName: displayName,
                  arguments: streamMsg.arguments,
                })
              }
              delete pendingToolRequests[toolId]
            } else {
              aiMsg.segments.push({
                type: 'tool_executed',
                data: '',
                name: streamMsg.name,
                displayName: displayName,
                arguments: streamMsg.arguments,
              })
            }
          } else if (msgType === 'reasoning') {
            // 深度思考内容
            const lastSeg = aiMsg.segments[aiMsg.segments.length - 1]
            if (lastSeg && lastSeg.type === 'reasoning') {
              lastSeg.data += streamMsg.data
            } else {
              aiMsg.segments.push({ type: 'reasoning', data: streamMsg.data })
            }
          } else {
            // 未知类型，当作纯文本处理
            const lastSeg = aiMsg.segments[aiMsg.segments.length - 1]
            if (lastSeg && lastSeg.type === 'text') {
              lastSeg.data += processedData
            } else {
              aiMsg.segments.push({ type: 'text', data: processedData })
            }
          }
        } catch {
          // JSON解析失败，当作纯文本处理
          const lastSeg = aiMsg.segments[aiMsg.segments.length - 1]
          if (lastSeg && lastSeg.type === 'text') {
            lastSeg.data += processedData
          } else {
            aiMsg.segments.push({ type: 'text', data: processedData })
          }
        }
        // 同步更新content（兼容历史存储）
        aiMsg.content = fullData
      }

      // 强制更新DOM并滚动到底部
      nextTick().then(() => {
        scrollToBottom()
      })
    }

    // 处理business-error事件（后端限流等错误）
    eventSource.addEventListener('business-error', function (event: MessageEvent) {
      try {
        const errorData = JSON.parse(event.data)
        console.error('SSE业务错误事件:', errorData)

        // 显示具体的错误信息
        const errorMessage = errorData.message || '生成过程中出现错误'
        messages.value[aiMsgIndex].content = `❌ ${errorMessage}`
        message.error(errorMessage)

        eventSource?.close()
        sending.value = false
        streamLoading.value = false
      } catch (parseError) {
        console.error('解析错误事件失败:', parseError, '原始数据:', event.data)
      }
    })

    // 监听特定事件类型
    eventSource.addEventListener('done', (event) => {
      // 流结束，将所有未完成的工具请求标记为已完成
      if (messages.value[aiMsgIndex]) {
        // 使用索引遍历并替换整个对象，确保Vue响应式更新
        for (let i = 0; i < messages.value[aiMsgIndex].segments.length; i++) {
          const seg = messages.value[aiMsgIndex].segments[i]
          if (seg.type === 'tool_request') {
            messages.value[aiMsgIndex].segments[i] = {
              ...seg,
              type: 'tool_executed',
            }
          }
        }
      }
      eventSource.close()
      sending.value = false
      streamLoading.value = false
      // 显示部署预览
      showPreview()
    })

    eventSource.onerror = (error) => {
      console.error('SSE error:', error.target)
            eventSource.close()
            sending.value = false
            streamLoading.value = false
            message.error('消息发送失败')
    }
  } catch (error) {
    sending.value = false
    streamLoading.value = false
    message.error('消息发送失败，请重试')
  }
}

// 滚动到底部
const scrollToBottom = () => {
  if (messagesContainerRef.value) {
    messagesContainerRef.value.scrollTop = messagesContainerRef.value.scrollHeight
  }
}

// 显示预览
const showPreview = () => {
  if (appInfo.value.codeGenType && appInfo.value.id) {
    previewUrl.value = getStaticPreviewUrl(appInfo.value.codeGenType, appInfo.value.id.toString())
  }
}

// 部署应用
const doDeploy = async () => {
  if (!appId.value) {
    message.error('应用ID不存在')
    return
  }

  // 如果当前正在部署中，则取消部署
  if (appInfo.value.deployKey !== null && appInfo.value.deployKey !== '') {
    try {
      const res = await deployAppCancel({ appId: appId.value })
      if (res.data.code === 0 && res.data.data) {
        message.success('已取消部署')
        // 刷新应用信息
        await fetchAppInfo()
      } else {
        message.error('取消部署失败，' + res.data.message)
      }
    } catch (error) {
      message.error('取消部署失败，请重试')
    }
    return
  }

  deployLoading.value = true
  try {
    const res = await deployApp({ appId: appId.value })
    if (res.data.code === 0 && res.data.data) {
      // 显示部署成功弹窗
      deployUrl.value = res.data.data
      deployModalVisible.value = true
      // 刷新应用信息
      await fetchAppInfo()
    } else {
      message.error('部署失败，' + res.data.message)
    }
  } catch (error) {
    message.error('部署失败，请重试')
  } finally {
    deployLoading.value = false
  }
}

// 打开部署的网站
const openDeployedSite = () => {
  if (deployUrl.value) {
    window.open(deployUrl.value, '_blank')
  }
}

// 停止生成
const handleStop = async () => {
  if (!appId.value) {
    message.error('应用ID不存在')
    return
  }

  try {
    // 发送停止请求
    const res = await stopToGenCode({ appId: appId.value })
    if (res.data.code === 0 && res.data.data) {
      // 关闭SSE流
      if (eventSource) {
        eventSource.close()
        eventSource = null
      }
      // 更新状态
      sending.value = false
      streamLoading.value = false
      message.success('已停止生成')
    } else {
      message.error('停止生成失败：' + res.data.message)
    }
  } catch (error) {
    message.error('停止生成失败，请重试')
  }
}

// 权限相关
const isOwner = computed(() => {
  return appInfo.value?.userId === loginUserStore.loginUser.id
})

const isAdmin = computed(() => {
  return loginUserStore.loginUser.userRole === 'admin'
})

// 下载相关
const downloading = ref(false)

// 下载代码
const downloadCode = async () => {
  if (!appId.value) {
    message.error('应用ID不存在')
    return
  }
  downloading.value = true
  try {
    const API_BASE_URL = request.defaults.baseURL || ''
    const url = `${API_BASE_URL}/app/download/${appId.value}`
    const response = await fetch(url, {
      method: 'GET',
      credentials: 'include',
    })
    if (!response.ok) {
      throw new Error(`下载失败: ${response.status}`)
    }
    // 获取文件名
    const contentDisposition = response.headers.get('Content-Disposition')
    const fileName = contentDisposition?.match(/filename="(.+)"/)?.[1] || `app-${appId.value}.zip`
    // 下载文件
    const blob = await response.blob()
    const downloadUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = downloadUrl
    link.download = fileName
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(downloadUrl)
    message.success('下载成功')
  } catch (error) {
    message.error('下载失败，请重试')
  } finally {
    downloading.value = false
  }
}

// 页面加载时请求数据
onMounted(() => {
  fetchAppInfo()
  // 监听来自 iframe 的消息
  window.addEventListener('message', handleIframeMessage)
})

// 页面卸载时清理
onUnmounted(() => {
  // 移除消息监听器
  window.removeEventListener('message', handleIframeMessage)
})

// 应用详情相关
const appDetailVisible = ref(false)

const showAppDetail = () => {
  appDetailVisible.value = true
}

const editApp = () => {
  // 编辑应用逻辑
}

const deleteApp = () => {
  // 删除应用逻辑
}
</script>

<style scoped>
#appChatPage {
  height: 85vh;
  display: flex;
  flex-direction: column;
  margin: 20px auto;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.header {
  background: #fff;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  z-index: 1;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 64px;
}

.app-name {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 18px;
  font-weight: bold;
}

.code-gen-type-tag {
  margin-top: 3px;
  font-size: 12px;
}

.main-layout {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.buttons {
  display: flex;
  justify-content: space-around;
  gap: 15px;
}

.chat-container {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #f5f5f5;
  min-height: 0;
}

.load-more-container {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.message {
  margin-bottom: 20px;
}

.message-content {
  display: flex;
  max-width: 100%;
}

.message.user .message-content {
  margin-left: auto;
  flex-direction: row-reverse;
  justify-content: flex-end;
}

.avatar {
  margin-right: 10px;
}

.message.user .avatar {
  margin-right: 0;
  margin-left: 10px;
}

.content {
  background: #fff;
  padding: 10px 15px;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
  max-width: 80%;
  word-wrap: break-word;
}

.selected-element-alert {
  margin: 5px 16px;
}

.selected-element-info {
  line-height: 1.4;
}

.element-header {
  margin-bottom: 8px;
}

.element-details {
  margin-top: 8px;
}

.element-item {
  margin-bottom: 4px;
  font-size: 12px;
}

.element-item:last-child {
  margin-bottom: 0;
}

.element-tag {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 14px;
  font-weight: 600;
  color: #007bff;
}

.element-id {
  color: #28a745;
  margin-left: 4px;
}

.element-class {
  color: #ffc107;
  margin-left: 4px;
}

.element-selector-code {
  font-family: 'Monaco', 'Menlo', monospace;
  background: #f6f8fa;
  padding: 2px 4px;
  border-radius: 3px;
  font-size: 12px;
  color: #d73a49;
  border: 1px solid #e1e4e8;
}

.message.user .content {
  background: #1890ff;
  color: #fff;
  margin-left: auto;
}

/* 深度思考折叠框样式 */
.reasoning-block {
  margin: 8px 0;
}

.reasoning-block details {
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
  overflow: hidden;
}

.reasoning-block summary {
  padding: 8px 12px;
  cursor: pointer;
  font-size: 13px;
  color: #666;
  user-select: none;
  list-style: none;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.reasoning-block summary::-webkit-details-marker {
  display: none;
}

.reasoning-block summary::marker {
  display: none;
}

.reasoning-block summary:hover {
  background: #f0f1f2;
}

.reasoning-block .summary-left {
  display: flex;
  align-items: center;
  gap: 6px;
}

.reasoning-block .summary-arrow {
  font-size: 10px;
  color: #999;
  transition: transform 0.2s;
}

.reasoning-block details[open] .summary-arrow {
  transform: rotate(180deg);
}

.reasoning-block .reasoning-content {
  padding: 8px 12px;
  font-size: 13px;
  color: #555;
  max-height: 300px;
  overflow-y: auto;
  border-top: 1px solid #e8e8e8;
}

.ai-content pre {
  background: #f8f8f8;
  border: 1px solid #e1e1e1;
  border-radius: 6px;
  padding: 16px;
  margin: 10px 0;
  overflow-x: auto;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.4;
}

.ai-content code {
  background: #f5f5f5;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.ai-content pre code {
  background: none;
  padding: 0;
  border-radius: 0;
}

.input-container {
  display: flex;
  padding: 20px;
  background: #fff;
  border-top: 1px solid #e8e8e8;
}

.input-container .ant-input {
  flex: 1;
  margin-right: 10px;
}

.preview-container {
  background: #fff;
  border-left: 1px solid #e8e8e8;
  flex: 1.5;
  min-width: 0;
}

:deep(.ant-layout-sider-children) {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
}

.preview-frame {
  flex: 1;
  width: 100%;
  border: none;
}

.loading-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 20px;
}

.empty-preview {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  color: #999;
}

/* 响应式设计 */
@media (max-width: 768px) {
  #appChatPage {
    height: 90vh;
    margin: 10px;
  }

  .main-layout {
    flex-direction: column;
  }

  .preview-container {
    border-left: none;
    border-top: 1px solid #e8e8e8;
    flex: 1;
  }

  .chat-container {
    flex: 1;
  }

  .header-content {
    flex-direction: column;
    height: auto;
    padding: 10px 0;
    gap: 10px;
  }

  .buttons {
    flex-wrap: wrap;
    justify-content: center;
  }

  .content {
    max-width: 90%;
  }
}
</style>
