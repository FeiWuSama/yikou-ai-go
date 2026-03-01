<template>
  <div id="homePage">
    <h2 class="title">易扣 AI 应用生成平台</h2>
    <div class="desc">零代码生成完整应用</div>

    <!-- 创建应用表单 -->
    <a-form :model="createForm" @finish="handleCreateApp">
      <a-form-item name="initPrompt">
        <div class="prompt-input-container">
          <a-textarea
            v-model:value="createForm.initPrompt"
            placeholder="请输入你的应用需求，如：做一个个人博客网站，包含文章列表和详情页面"
            :rows="3"
            class="prompt-textarea"
          />
          <a-button
            type="primary"
            shape="circle"
            class="send-button"
            @click="handleCreateApp(createForm)"
            :loading="createLoading"
            :disabled="!createForm.initPrompt?.trim()"
          >
            <template #icon>
              <SendOutlined />
            </template>
          </a-button>
        </div>
      </a-form-item>
    </a-form>

    <!-- 快捷提示词 -->
    <div class="quick-prompts">
      <div class="quick-prompts-list">
        <a-tag
          v-for="prompt in quickPrompts"
          :key="prompt"
          class="quick-prompt-tag"
          @click="useQuickPrompt(prompt)"
        >
          {{ prompt }}
        </a-tag>
      </div>
    </div>

    <div class="app-sections">
      <!-- 精选应用 -->
      <div class="section-title">精选应用</div>
      <!-- 应用列表 -->
      <div class="app-list">
        <div class="app-cards">
          <AppCard
            v-for="item in goodAppData"
            :key="item.id"
            :app-data="item"
            :show-user-info="true"
            :is-good-app="true"
            @action="handleCardAction"
          />
        </div>
        <!-- 分页 -->
        <div class="pagination-container" v-if="goodTotal > goodSearchParams.pageSize">
          <a-pagination
            :current="goodPagination.current"
            :page-size="goodPagination.pageSize"
            :total="goodPagination.total"
            :show-size-changer="goodPagination.showSizeChanger"
            :show-total="goodPagination.showTotal"
            @change="handleGoodPageChange"
            @show-size-change="handleGoodPageSizeChange"
          />
        </div>
      </div>

      <!-- 我的应用 -->
      <div v-if="isLoggedIn" class="section-title">我的应用</div>
      <!-- 应用列表 -->
      <div v-if="isLoggedIn" class="app-list">
        <div class="app-cards">
          <AppCard
            v-for="item in myAppData"
            :key="item.id"
            :app-data="item"
            :is-good-app="false"
            @action="handleCardAction"
          />
        </div>
        <!-- 分页 -->
        <div class="pagination-container" v-if="myTotal > 6">
          <a-pagination
            :current="myPagination.current"
            :page-size="myPagination.pageSize"
            :total="myPagination.total"
            :show-size-changer="myPagination.showSizeChanger"
            :show-total="myPagination.showTotal"
            @change="handleMyPageChange"
          />
        </div>
      </div>

      <!-- 未登录提示 -->
      <div v-if="!isLoggedIn" class="login-prompt">
        <div class="login-prompt-content">
          <div class="login-prompt-text">登录后查看和管理您的应用</div>
          <a-button type="primary" href="/user/login">立即登录</a-button>
        </div>
      </div>
    </div>

    <!-- 编辑模态框 -->
    <a-modal
      v-model:open="editModalVisible"
      title="编辑应用"
      @ok="handleEditOk"
      @cancel="handleEditCancel"
      :confirm-loading="editConfirmLoading"
      ok-text="确认"
      cancel-text="取消"
    >
      <a-form :model="editForm" layout="vertical">
        <a-form-item label="应用名称">
          <a-input v-model:value="editForm.appName" placeholder="请输入应用名称" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { useRouter } from 'vue-router'
import { useLoginUserStore } from '@/stores/loginUser.ts'
import {
  addApp,
  deleteApp,
  listMyAppVoByPage,
  listGoodAppVoByPage,
  updateApp,
} from '@/api/appController.ts'
import defaultCover from '@/assets/logo.png'
import { SendOutlined } from '@ant-design/icons-vue'
import AppCard from '@/components/AppCard.vue'

const router = useRouter()
const loginUserStore = useLoginUserStore()

// 判断用户是否已登录
const isLoggedIn = computed(() => {
  const user = loginUserStore.loginUser
  return user && user.userRole && user.userRole !== 'notLogin'
})

// 快捷提示词 - 改为生成类型词
const quickPrompts = ref([
  '个人博客网站',
  '在线商城',
  '任务管理系统',
  '企业官网',
  '在线教育平台',
])

// 创建应用表单
const createForm = reactive<API.AppAddDto>({
  initPrompt: '',
})
const createLoading = ref(false)

// 我的应用数据
const myAppData = ref<API.AppVO[]>([])
const myTotal = ref(0)
const mySearchParams = reactive<API.AppQueryDto>({
  pageNum: 1,
  pageSize: 6,
})

// 精选应用数据
const goodAppData = ref<API.AppVO[]>([])
const goodTotal = ref(0)
const goodSearchParams = reactive<API.AppQueryDto>({
  pageNum: 1,
  pageSize: 20,
  sortField: 'createTime',
  sortOrder: 'desc',
})

// 编辑模态框相关
const editModalVisible = ref(false)
const editConfirmLoading = ref(false)
const editForm = reactive<API.AppUpdateDto>({
  id: undefined,
  appName: '',
})

// 卡片操作处理
const handleCardAction = (actionKey: string, appData: API.AppVO) => {
  switch (actionKey) {
    case 'experience':
    case 'chat':
    case 'preview': // 新增预览操作
    case 'continue': // 新增继续创作操作
      goToChat(appData.id!.toString())
      break
    case 'edit':
      doEdit(appData)
      break
    case 'delete':
      doDelete(appData.id!)
      break
  }
}

// 获取我的应用数据
const fetchMyAppData = async () => {
  // 未登录时不请求我的应用数据
  if (!isLoggedIn.value) {
    myAppData.value = []
    myTotal.value = 0
    return
  }

  const res = await listMyAppVoByPage({
    ...mySearchParams,
  })
  if (res.data.code === 0 && res.data.data) {
    myAppData.value = res.data.data.records ?? []
    myTotal.value = res.data.data.totalRow ?? 0
  } else {
    message.error('获取我的应用失败，' + res.data.message)
  }
}

// 获取精选应用数据
const fetchGoodAppData = async () => {
  const res = await listGoodAppVoByPage({
    ...goodSearchParams,
  })
  if (res.data.code === 0 && res.data.data) {
    goodAppData.value = res.data.data.records ?? []
    goodTotal.value = res.data.data.totalRow ?? 0
  } else {
    message.error('获取精选应用失败，' + res.data.message)
  }
}

// 我的应用分页参数
const myPagination = computed(() => {
  return {
    current: mySearchParams.pageNum ?? 1,
    pageSize: mySearchParams.pageSize ?? 6,
    total: Number(myTotal.value),
    showSizeChanger: false,
    showTotal: (total: number) => `共 ${total} 条`,
  }
})

// 精选应用分页参数
const goodPagination = computed(() => {
  return {
    current: goodSearchParams.pageNum ?? 1,
    pageSize: goodSearchParams.pageSize ?? 20,
    total: Number(goodTotal.value),
    showSizeChanger: true,
    showTotal: (total: number) => `共 ${total} 条`,
  }
})

// 我的应用分页变化处理
const handleMyPageChange = (page: number, pageSize: number) => {
  mySearchParams.pageNum = page
  mySearchParams.pageSize = pageSize
  fetchMyAppData()
}

// 精选应用分页变化处理
const handleGoodPageChange = (page: number, pageSize: number) => {
  goodSearchParams.pageNum = page
  goodSearchParams.pageSize = pageSize
  fetchGoodAppData()
}

// 精选应用页面大小变化处理
const handleGoodPageSizeChange = (page: number, pageSize: number) => {
  goodSearchParams.pageNum = 1
  goodSearchParams.pageSize = pageSize
  fetchGoodAppData()
}

// 精选应用搜索
const doGoodAppSearch = () => {
  goodSearchParams.pageNum = 1
  fetchGoodAppData()
}

// 创建应用
const handleCreateApp = async (values: any) => {
  // 未登录时提示登录
  if (!isLoggedIn.value) {
    message.warning('请先登录后再创建应用')
    router.push('/user/login')
    return
  }

  createLoading.value = true
  try {
    const res = await addApp(values)
    if (res.data.code === 0 && res.data.data) {
      message.success('创建应用成功')
      // 跳转到对话页面
      router.push(`/app/chat/${res.data.data}`)
    } else {
      message.error('创建应用失败，' + res.data.message)
    }
  } catch (error) {
    message.error('创建应用失败，请重试')
  } finally {
    createLoading.value = false
  }
}

// 编辑应用
const doEdit = (record: API.AppVO) => {
  editForm.id = record.id
  editForm.appName = record.appName ?? ''
  editModalVisible.value = true
}

// 处理编辑确认
const handleEditOk = async () => {
  editConfirmLoading.value = true
  try {
    const res = await updateApp(editForm)
    if (res.data.code === 0) {
      message.success('编辑成功')
      editModalVisible.value = false
      // 刷新数据
      fetchMyAppData()
      fetchGoodAppData()
    } else {
      message.error('编辑失败，' + res.data.message)
    }
  } catch (error) {
    message.error('编辑失败，请重试')
  } finally {
    editConfirmLoading.value = false
  }
}

// 处理编辑取消
const handleEditCancel = () => {
  editModalVisible.value = false
}

// 删除应用
const doDelete = async (id: number) => {
  if (!id) {
    return
  }
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该应用吗？此操作不可恢复。',
    okText: '确认',
    cancelText: '取消',
    onOk: async () => {
      const res = await deleteApp({ id })
      if (res.data.code === 0) {
        message.success('删除成功')
        // 刷新数据
        fetchMyAppData()
        fetchGoodAppData()
      } else {
        message.error('删除失败，' + (res.data.message || '未知错误'))
      }
    },
  })
}

// 跳转到对话页面
const goToChat = (id: string) => {
  router.push(`/app/chat/${id}?view=1`)
}

// 使用快捷提示词 - 补充完整提示词
const useQuickPrompt = (prompt: string) => {
  // 未登录时提示登录
  if (!isLoggedIn.value) {
    message.warning('请先登录后再创建应用')
    router.push('/user/login')
    return
  }

  // 根据类型词生成完整提示词
  const promptMap: Record<string, string> = {
    '个人博客网站': '做一个个人博客网站，包含文章列表和详情页面',
    '在线商城': '创建一个在线商城，包含商品展示和购物车功能',
    '任务管理系统': '开发一个任务管理系统，支持任务分配和进度跟踪',
    '企业官网': '制作一个企业官网，包含公司介绍和联系方式',
    '在线教育平台': '搭建一个在线教育平台，包含课程管理和学习记录',
  }

  // 获取完整提示词或使用默认格式
  const fullPrompt = promptMap[prompt] || `${prompt}`

  createForm.initPrompt = fullPrompt
  // // 自动提交表单
  // handleCreateApp(createForm)
}

// 页面加载时请求数据
onMounted(() => {
  fetchMyAppData()
  fetchGoodAppData()
})
</script>

<style scoped>
#homePage {
  margin: 0 auto;
  padding: 20px;
}

.title {
  text-align: center;
  margin-bottom: 16px;
  font-size: 2.5rem;
  font-weight: 700;
  background: linear-gradient(135deg, #0037ff 0%, #009dfe 50%, #00ff87 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  text-shadow: 0 0 20px rgba(0, 180, 255, 0.5);
  position: relative;
  animation: techGlow 3s ease-in-out infinite alternate;
}

@keyframes techGlow {
  0% {
    text-shadow: 0 0 20px rgba(0, 180, 255, 0.5), 0 0 40px rgba(0, 81, 255, 0.505);
  }
  100% {
    text-shadow: 0 0 30px rgba(0, 180, 255, 0.8), 0 0 60px rgba(0, 119, 255, 0.838);
  }
}

.desc {
  text-align: center;
  color: #9f9f9f;
  margin-bottom: 32px;
  font-size: 1.2rem;
  font-weight: 300;
  background: linear-gradient(135deg, #817df1 0%, #ffc8d9 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  text-shadow: 0 0 15px rgba(168, 237, 234, 0.4);
  animation: descGlow 4s ease-in-out infinite alternate;
}

@keyframes descGlow {
  0% {
    text-shadow: 0 0 15px rgba(168, 237, 234, 0.4);
  }
  100% {
    text-shadow: 0 0 25px rgba(168, 237, 234, 0.7), 0 0 35px rgba(168, 237, 234, 0.5);
  }
}

.prompt-input-container {
  display: flex;
  justify-content: center;
  position: relative;
  width: 800px;
  margin: 0 auto;
}

.prompt-textarea {
  padding: 15px 40px 15px 15px !important;
  width: 100%;
  resize: vertical;
  min-height: 80px;
}

.send-button {
  position: absolute;
  bottom: 10px;
  right: 10px;
  width: 32px;
  height: 32px;
  min-width: 32px;
  z-index: 10;
}

.quick-prompts {
  margin-top: 20px;
}

.quick-prompts-list {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 10px;
}

.quick-prompt-tag {
  border-radius: 20px;
  padding: 10px;
  background: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  transition: all 0.3s;
}

.quick-prompt-tag:hover {
  transform: scale(1.05);
}

.app-sections {
  display: flex;
  flex-direction: column;
  gap: 24px;
  margin-top: 50px;
  padding: 0 225px;
}

.section-title {
  font-size: 20px;
  font-weight: bold;
  margin-bottom: 5px;
  margin-left: 100px;
  color: #333;
}

.app-list {
  min-height: 100px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  padding: 0 80px;
}

.app-cards {
  margin: 0 auto;
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  justify-content: left;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
  margin-bottom: 20px;
}

.login-prompt {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
  background: #f8f9fa;
  border-radius: 8px;
  margin-top: 20px;
}

.login-prompt-content {
  text-align: center;
}

.login-prompt-text {
  font-size: 16px;
  color: #666;
  margin-bottom: 16px;
}

@media (max-width: 768px) {
  .app-card {
    width: 100%;
  }
}
</style>
