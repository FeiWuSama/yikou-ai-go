<template>
  <div id="userManagePage">
    <div class="page-header">
      <h1>用户管理</h1>
    </div>
    <a-card :loading="loading">
      <a-form layout="inline" :model="searchParams" @finish="doSearch">
        <a-form-item label="用户名">
          <a-input v-model:value="searchParams.userName" placeholder="输入用户名" />
        </a-form-item>
        <a-form-item label="账号">
          <a-input v-model:value="searchParams.userAccount" placeholder="输入账号" />
        </a-form-item>
        <a-form-item label="角色">
          <a-select
            v-model:value="searchParams.userRole"
            placeholder="选择角色"
            style="width: 150px"
          >
            <a-select-option value="admin">管理员</a-select-option>
            <a-select-option value="user">普通用户</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit">搜索</a-button>
        </a-form-item>
      </a-form>
      <a-divider />
      <a-table
        :columns="columns"
        :data-source="dataSource"
        :pagination="pagination"
        @change="handleTableChange"
        :row-key="(record: API.UserVO) => record.id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="handleEdit(record)"> 编辑</a-button>
              <a-button type="link" size="small" danger @click="handleDelete(record)">
                删除
              </a-button>
            </a-space>
          </template>
          <template v-else-if="column.key === 'userAvatar'">
            <a-avatar :src="record.userAvatar" />
          </template>
          <template v-else-if="column.key === 'createTime'">
            {{ formatTime(record.createTime) }}
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- 编辑用户弹窗 -->
    <a-modal
      v-model:open="editModalVisible"
      title="编辑用户"
      @ok="handleSaveEdit"
      :confirm-loading="submitting"
    >
      <a-form :model="editForm" :rules="rules" ref="editFormRef">
        <a-form-item label="用户名" name="userName">
          <a-input v-model:value="editForm.userName" placeholder="输入用户名" />
        </a-form-item>
        <a-form-item label="角色" name="userRole">
          <a-select v-model:value="editForm.userRole" placeholder="选择角色">
            <a-select-option value="admin">管理员</a-select-option>
            <a-select-option value="user">普通用户</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { listUserVoByPage, updateUser, deleteUser } from '@/api/userController'
import type { FormInstance } from 'ant-design-vue'
import { formatTime } from '@/utils/time'

const loading = ref(false)
const submitting = ref(false)
const dataSource = ref<API.UserVO[]>([])
const total = ref(0)
const editModalVisible = ref(false)
const editFormRef = ref<FormInstance>()

// 搜索参数
const searchParams = reactive<API.UserQueryDto>({
  pageNum: 1,
  pageSize: 10,
})

// 编辑表单
const editForm = reactive<API.UserUpdateDto>({
  id: 0,
  userName: '',
  userRole: 'user',
})

// 表单验证规则
const rules = {
  userName: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  userRole: [{ required: true, message: '请选择角色', trigger: 'change' }],
}

// 表格列配置
const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '头像', dataIndex: 'userAvatar', key: 'userAvatar', width: 80 },
  { title: '用户名', dataIndex: 'userName', key: 'userName' },
  { title: '账号', dataIndex: 'userAccount', key: 'userAccount' },
  { title: '角色', dataIndex: 'userRole', key: 'userRole' },
  { title: '创建时间', dataIndex: 'createTime', key: 'createTime' },
  { title: '操作', key: 'action', width: 150 },
]

// 分页配置
const pagination = computed(() => ({
  current: searchParams.pageNum,
  pageSize: searchParams.pageSize,
  total: total.value,
  showTotal: (total: number) => `共 ${total} 条`,
}))

// 获取用户列表
const fetchUserList = async () => {
  loading.value = true
  try {
    const res = await listUserVoByPage({
      ...searchParams,
    })
    if (res.data.code === 0 && res.data.data) {
      dataSource.value = res.data.data.records || []
      total.value = res.data.data.totalRow || 0
    } else {
      message.error('获取用户列表失败：' + res.data.message)
    }
  } catch (error) {
    console.error('获取用户列表失败：', error)
    message.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const doSearch = () => {
  searchParams.pageNum = 1
  fetchUserList()
}

// 表格分页变化
const handleTableChange = (pagination: any) => {
  searchParams.pageNum = pagination.current
  searchParams.pageSize = pagination.pageSize
  fetchUserList()
}

// 编辑用户
const handleEdit = (record: API.UserVO) => {
  editForm.id = record.id
  editForm.userName = record.userName
  editForm.userRole = record.userRole
  editModalVisible.value = true
}

// 保存编辑
const handleSaveEdit = async () => {
  if (!editFormRef.value) return
  try {
    await editFormRef.value.validate()
    submitting.value = true
    const res = await updateUser({
      id: editForm.id,
      userName: editForm.userName,
      userRole: editForm.userRole,
    })
    if (res.data.code === 0) {
      message.success('编辑成功')
      editModalVisible.value = false
      fetchUserList()
    } else {
      message.error('编辑失败：' + res.data.message)
    }
  } catch (error) {
    console.error('编辑失败：', error)
    message.error('编辑失败')
  } finally {
    submitting.value = false
  }
}

// 删除用户
const handleDelete = (record: API.UserVO) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除用户 ${record.userName} 吗？`,
    onOk: async () => {
      try {
        const res = await deleteUser({ id: record.id })
        if (res.data.code === 0) {
          message.success('删除成功')
          fetchUserList()
        } else {
          message.error('删除失败：' + res.data.message)
        }
      } catch (error) {
        console.error('删除失败：', error)
        message.error('删除失败')
      }
    },
  })
}

// 页面加载时获取用户列表
onMounted(() => {
  fetchUserList()
})
</script>

<style scoped>
#userManagePage {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}
</style>
