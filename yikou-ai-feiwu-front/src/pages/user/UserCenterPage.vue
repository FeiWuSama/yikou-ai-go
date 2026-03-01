<template>
  <div id="userCenterPage">
    <div class="page-header">
      <h1>个人中心</h1>
    </div>
    <a-card :loading="loading">
      <a-form :model="formData" :rules="rules" ref="formRef" layout="vertical">
        <a-form-item label="用户名" name="userName">
          <a-input v-model:value="formData.userName" placeholder="输入用户名" />
        </a-form-item>
        <a-form-item label="简介" name="userProfile">
          <a-textarea v-model:value="formData.userProfile" placeholder="输入简介" :rows="3" />
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" html-type="submit" @click="handleSave" :loading="submitting">
              保存修改
            </a-button>
            <a-button @click="resetForm">重置</a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { useLoginUserStore } from '@/stores/loginUser'
import { updateUser, getLoginUser } from '@/api/userController'
import type { FormInstance } from 'ant-design-vue'

const loading = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const loginUserStore = useLoginUserStore()

// 表单数据
const formData = reactive<API.UserUpdateDto>({
  id: 0,
  userName: '',
  userProfile: '',
})

// 表单验证规则
const rules = {
  userName: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  userProfile: [{ max: 200, message: '简介长度不超过200个字符', trigger: 'blur' }],
}

// 获取用户信息
const fetchUserInfo = async () => {
  loading.value = true
  try {
    const res = await getLoginUser()
    if (res.data.code === 0 && res.data.data) {
      const userInfo = res.data.data
      formData.id = userInfo.id
      formData.userName = userInfo.userName
      formData.userProfile = userInfo.userProfile
    } else {
      message.error('获取用户信息失败：' + res.data.message)
    }
  } catch (error) {
    console.error('获取用户信息失败：', error)
    message.error('获取用户信息失败')
  } finally {
    loading.value = false
  }
}

// 保存修改
const handleSave = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
    submitting.value = true
    const res = await updateUser({
      id: formData.id,
      userName: formData.userName,
      userProfile: formData.userProfile,
    })
    if (res.data.code === 0) {
      message.success('保存成功')
      // 更新登录用户信息
      await loginUserStore.fetchLoginUser()
    } else {
      message.error('保存失败：' + res.data.message)
    }
  } catch (error) {
    console.error('保存失败：', error)
    message.error('保存失败')
  } finally {
    submitting.value = false
  }
}

// 重置表单
const resetForm = () => {
  fetchUserInfo()
  formRef.value?.clearValidate()
}

// 页面加载时获取用户信息
onMounted(() => {
  fetchUserInfo()
})
</script>

<style scoped>
#userCenterPage {
  max-width: 800px;
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
