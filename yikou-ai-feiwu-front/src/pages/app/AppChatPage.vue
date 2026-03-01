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
          :style="{ height: messagesContainerHeight + 'px' }"
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
                <MarkdownRenderer v-if="msg.role === 'ai'" :content="msg.content" />
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
import { onMounted, ref, nextTick, computed } from 'vue'
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
const messages = ref<Array<{ id: number; role: string; content: string }>>([])
const inputMessage = ref('')
const sending = ref(false)
const messagesContainerRef = ref<HTMLDivElement | null>(null)

// SSE流相关
let eventSource: EventSource | null = null
const messagesContainerHeight = ref(500)

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
  debugger
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
  const newEditMode = visualEditor.toggleEditMode()
  isEditMode.value = newEditMode
}

const clearSelectedElement = () => {
  selectedElementInfo.value = null
  visualEditor.clearSelection()
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
        .map((record) => ({
          id: record.id || Date.now() + Math.random(),
          role: record.messageType === 'user' ? 'user' : 'ai',
          content: record.message || '',
        }))
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
  }
  messages.value.push(userMsg)

  // 添加AI消息占位符
  const aiMsg = {
    id: Date.now() + 1,
    role: 'ai',
    content: '',
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
  }
  messages.value.push(userMsg)

  // 添加AI消息占位符
  const aiMsg = {
    id: Date.now() + 1,
    role: 'ai',
    content: '',
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
    eventSource.onmessage = (event) => {
      let processedData = ''
      try {
        // 解析JSON数据
        const jsonData = JSON.parse(event.data)
        // 提取d字段的内容
        processedData = jsonData.d || ''
      } catch (error) {
        // 如果JSON解析失败，直接使用原始数据
        console.warn('JSON解析失败，使用原始数据:', error)
        processedData = event.data
      }

      fullData += processedData
      // 根据索引更新AI消息内容
      if (messages.value[aiMsgIndex]) {
        messages.value[aiMsgIndex].content = fullData
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
      // 流结束
      eventSource.close()
      sending.value = false
      streamLoading.value = false
      // 显示部署预览
      showPreview()
    })

    eventSource.onerror = (error) => {
      if(eventSource.readyState!==0){
        eventSource?.close()
        sending.value = false
        streamLoading.value = false
        message.error('消息发送失败')
      }
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
    link.click()
    // 清理
    URL.revokeObjectURL(downloadUrl)
    message.success('代码下载成功')
  } catch (error) {
    console.error('下载失败：', error)
    message.error('下载失败，请重试')
  } finally {
    downloading.value = false
  }
}

// 应用详情相关
const appDetailVisible = ref(false)

const router = useRouter()

// 显示应用详情
const showAppDetail = () => {
  appDetailVisible.value = true
}
// 编辑应用
const editApp = () => {
  if (appInfo.value?.id) {
    router.push(`/app/edit/${appInfo.value.id}`)
  }
}

// 删除应用
const deleteApp = async () => {
  if (!appInfo.value?.id) return

  try {
    const res = await deleteAppApi({ id: appInfo.value.id })
    if (res.data.code === 0) {
      message.success('删除成功')
      appDetailVisible.value = false
      router.push('/')
    } else {
      message.error('删除失败：' + res.data.message)
    }
  } catch (error) {
    console.error('删除失败：', error)
    message.error('删除失败')
  }
}

//iframe加载完成
const onIframeLoad = () => {
  const iframe = document.querySelector('.preview-frame') as HTMLIFrameElement
  if (iframe) {
    visualEditor.init(iframe)
    visualEditor.onIframeLoad()
  }
}

// 页面加载时获取应用信息和用户信息
onMounted(async () => {
  // 获取用户信息
  await loginUserStore.fetchLoginUser()
  // 获取应用信息
  await fetchAppInfo()
  onIframeLoad()
  //监听iframe消息
  window.addEventListener('message', (event) => {
    visualEditor.handleIframeMessage(event)
  })
})
</script>

<style scoped>
#appChatPage {
  height: 85vh;
  display: flex;
  flex-direction: column;
  margin: 20px auto;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
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
}

.buttons {
  display: flex;
  justify-content: space-around;
  gap: 15px;
}

.chat-container {
  display: flex;
  flex-direction: column;
  height: calc(90vh - 66px - 40px);
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #f5f5f5;
  border-radius: 8px;
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
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0 8px 8px 0;
}

:deep(.ant-layout-sider-children) {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 60vw;
}

.preview-frame {
  width: 100%;
  height: 100%;
}

.loading-container {
  text-align: center;
  padding: 20px;
}

.empty-preview {
  text-align: center;
  color: #999;
}
</style>
