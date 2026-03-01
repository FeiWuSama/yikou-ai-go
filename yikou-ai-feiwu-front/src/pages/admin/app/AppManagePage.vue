<template>
  <div id="appManagePage">
    <!-- 搜索表单 -->
    <a-form layout="inline" :model="searchParams" @finish="doSearch">
      <a-form-item label="应用名称">
        <a-input v-model:value="searchParams.appName" placeholder="输入应用名称" />
      </a-form-item>
      <a-form-item label="生成类型">
        <a-select v-model:value="searchParams.codeGenType" placeholder="选择生成类型" style="width: 150px">
          <a-select-option
            v-for="option in CODE_GEN_TYPE_OPTIONS"
            :key="option.value"
            :value="option.value"
            >{{ option.label }}
          </a-select-option>
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
        <template v-if="column.dataIndex === 'cover'">
          <a-image :src="record.cover || defaultCover" :width="120" />
        </template>
        <template v-else-if="column.dataIndex === 'initPrompt'">
          <span :title="record.initPrompt">
            {{ record.initPrompt }}
          </span>
        </template>
        <template v-else-if="column.dataIndex === 'priority'">
          <a-tag :color="record.priority && record.priority >= 99 ? 'green' : 'default'">
            {{ record.priority === 99 ? '精选' : '普通' }}
          </a-tag>
        </template>
        <template v-else-if="column.dataIndex === 'createTime'">
          {{ dayjs(record.createTime).format('YYYY-MM-DD HH:mm:ss') }}
        </template>
        <template v-else-if="column.dataIndex === 'user'">
          <div class="user-info">
            <a-avatar :src="record.user?.userAvatar" size="small" />
            <span class="user-name">{{ record.user?.userName || '未知用户' }}</span>
          </div>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-button
            type="primary"
            @click="doEdit(record)"
            style="margin-right: 8px; margin-bottom: 8px"
            >编辑
          </a-button>
          <a-button @click="doSetGood(record)" style="margin-right: 8px">
            {{ record.priority && record.priority >= 99 ? '取消精选' : '设为精选' }}
          </a-button>
          <a-button danger @click="doDelete(record.id)">删除</a-button>
        </template>
      </template>
    </a-table>
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
        <a-form-item label="应用封面">
          <a-input v-model:value="editForm.cover" placeholder="请输入封面URL" />
        </a-form-item>
        <a-form-item label="优先级">
          <a-input-number v-model:value="editForm.priority" :min="0" :max="100" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  deleteAppByAdmin,
  listAppVoByPageByAdmin,
  updateAppByAdmin,
  getAppVoByIdByAdmin,
} from '@/api/appController.ts'
import dayjs from 'dayjs'
import defaultCover from '@/assets/logo.png'
import { useRouter } from 'vue-router'
import { CODE_GEN_TYPE_OPTIONS } from '@/utils/constants.ts'

const router = useRouter()

const columns = [
  {
    title: 'id',
    dataIndex: 'id',
  },
  {
    title: '应用名称',
    dataIndex: 'appName',
  },
  {
    title: '封面',
    dataIndex: 'cover',
  },
  {
    title: '提示词',
    dataIndex: 'initPrompt',
  },
  {
    title: '创建用户',
    dataIndex: 'user',
  },
  {
    title: '优先级',
    dataIndex: 'priority',
  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
  },
  {
    title: '操作',
    key: 'action',
  },
]

// 数据
const data = ref<API.AppVO[]>([])
const total = ref(0)

// 编辑模态框相关
const editModalVisible = ref(false)
const editConfirmLoading = ref(false)
const editForm = reactive<API.AppAdminUpdateDto>({
  id: undefined,
  appName: '',
  cover: '',
  priority: 0,
})

// 搜索条件
const searchParams = reactive<API.AppQueryDto>({
  pageNum: 1,
  pageSize: 20,
})

// 获取数据
const fetchData = async () => {
  const res = await listAppVoByPageByAdmin({
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

// 编辑数据
const doEdit = (record: API.AppVO) => {
  // 直接使用record中的数据填充editForm
  editForm.id = record.id
  editForm.appName = record.appName ?? ''
  editForm.cover = record.cover ?? ''
  editForm.priority = record.priority ?? 0
  editModalVisible.value = true
}

// 处理编辑确认
const handleEditOk = async () => {
  editConfirmLoading.value = true
  try {
    const res = await updateAppByAdmin(editForm)
    if (res.data.code === 0) {
      message.success('编辑成功')
      editModalVisible.value = false
      // 刷新数据
      fetchData()
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

// 删除数据
const doDelete = async (id: number) => {
  if (!id) {
    return
  }
  // 添加确认对话框
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该应用吗？此操作不可恢复。',
    okText: '确认',
    cancelText: '取消',
    onOk: async () => {
      const res = await deleteAppByAdmin({ id })
      if (res.data.code === 0) {
        message.success('删除成功')
        fetchData()
      } else {
        message.error('删除失败，' + (res.data.message || '未知错误'))
      }
    },
  })
}

// 设置精选
const doSetGood = async (record: API.AppVO) => {
  if (!record.id) {
    return
  }

  // 获取应用详细信息
  const res = await getAppVoByIdByAdmin({ id: record.id })
  if (res.data.code !== 0) {
    message.error('获取应用信息失败，' + res.data.message)
    return
  }

  const appInfo = res.data.data
  if (!appInfo) {
    message.error('应用信息不存在')
    return
  }

  // 设置优先级为99或0
  const newPriority = appInfo.priority && appInfo.priority >= 99 ? 0 : 99

  const updateRes = await updateAppByAdmin({
    id: appInfo.id,
    appName: appInfo.appName,
    cover: appInfo.cover,
    priority: newPriority,
  })

  if (updateRes.data.code === 0) {
    message.success(newPriority >= 99 ? '设置精选成功' : '取消精选成功')
    fetchData()
  } else {
    message.error('操作失败，' + updateRes.data.message)
  }
}

// 页面加载时请求一次
onMounted(() => {
  fetchData()
})
</script>

<style scoped>
#appManagePage {
  max-width: 1200px;
  margin: 0 auto;
}

.user-info {
  width: 80px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-name {
  font-size: 14px;
}
</style>
