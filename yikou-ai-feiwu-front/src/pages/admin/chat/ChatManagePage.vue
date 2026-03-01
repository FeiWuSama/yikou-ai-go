<template>
  <div id="chatManagePage">
    <!-- 搜索表单 -->
    <a-form layout="inline" :model="searchParams" @finish="doSearch">
      <a-form-item label="应用ID">
        <a-input-number v-model:value="searchParams.appId" placeholder="输入应用ID" />
      </a-form-item>
      <a-form-item label="消息类型">
        <a-select v-model:value="searchParams.messageType" placeholder="选择消息类型" style="width: 150px">
          <a-select-option value="user">用户消息</a-select-option>
          <a-select-option value="ai">AI消息</a-select-option>
        </a-select>
      </a-form-item>
      <a-form-item>
        <a-button type="primary" html-type="submit">搜索</a-button>
      </a-form-item>
    </a-form>
    <a-divider />
    <!-- 表格 -->
    <a-table
      :columns="columns"
      :data-source="data"
      :pagination="pagination"
      @change="doTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.dataIndex === 'message'">
          <span :title="record.message">
            {{ record.message }}
          </span>
        </template>
        <template v-else-if="column.dataIndex === 'messageType'">
          <a-tag :color="record.messageType === 'user' ? 'blue' : 'green'">
            {{ record.messageType === 'user' ? '用户消息' : 'AI消息' }}
          </a-tag>
        </template>
        <template v-else-if="column.dataIndex === 'appId'">
          <a @click="goToAppChat(record.appId)">{{ record.appId }}</a>
        </template>
        <template v-else-if="column.dataIndex === 'userId'">
          <div class="user-info" v-if="record.user">
            <a-avatar :src="record.user.userAvatar" :size="24" />
            <span class="user-name">{{ record.user.userName }}</span>
          </div>
          <span v-else>{{ record.userId }}</span>
        </template>
        <template v-else-if="column.dataIndex === 'createTime'">
          <span>{{ formatTime(record.createTime) }}</span>
        </template>
        
      </template>
    </a-table>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { message } from 'ant-design-vue'
import { listAllChatHistoryByPageForAdmin } from '@/api/chatHistoryController.ts'
import { formatTime } from '@/utils/time.ts'
import { useRouter } from 'vue-router'

const router = useRouter()

const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
  },
  {
    title: '应用ID',
    dataIndex: 'appId',
  },
  {
    title: '消息内容',
    dataIndex: 'message',
  },
  {
    title: '消息类型',
    dataIndex: 'messageType',
  },
  {
    title: '用户',
    dataIndex: 'userId',
  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
  },
]

// 数据
const data = ref<API.ChatHistory[]>([])
const total = ref(0)

// 搜索条件
const searchParams = reactive<API.ChatHistoryQueryDto>({
  pageNum: 1,
  pageSize: 20,
})

// 获取数据
const fetchData = async () => {
  const res = await listAllChatHistoryByPageForAdmin({
    ...searchParams,
  })
  if (res.data.code === 0 && res.data.data) {
    data.value = res.data.data.records ?? []
    total.value = res.data.data.totalRow ?? 0
  } else {
    message.error('获取数据失败，' + res.data.message)
  }
}

// 分页参数
const pagination = computed(() => {
  return {
    current: searchParams.pageNum ?? 1,
    pageSize: searchParams.pageSize ?? 20,
    total: Number(total.value),
    showSizeChanger: true,
    showTotal: (total: number) => `共 ${total} 条`,
  }
})

// 表格变化处理
const doTableChange = (page: any) => {
  searchParams.pageNum = page.current
  searchParams.pageSize = page.pageSize
  fetchData()
}

// 获取数据
const doSearch = () => {
  // 重置页码
  searchParams.pageNum = 1
  fetchData()
}



// 跳转到应用对话页面
const goToAppChat = (appId: number) => {
  router.push(`/app/chat/${appId}`)
}

// 页面加载时请求一次
onMounted(() => {
  fetchData()
})
</script>

<style scoped>
#chatManagePage {
  max-width: 1200px;
  margin: 0 auto;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-name {
  font-size: 14px;
}
</style>