<template>
  <div id="appUpdatePage">
    <h2 class="title">修改应用信息</h2>
    <a-form :model="form" @finish="handleSubmit" :label-col="{ span: 4 }" :wrapper-col="{ span: 14 }">
      <a-form-item label="应用名称" name="appName" :rules="[{ required: true, message: '请输入应用名称' }]">
        <a-input v-model:value="form.appName" placeholder="请输入应用名称" />
      </a-form-item>
      <a-form-item :wrapper-col="{ offset: 4, span: 14 }">
        <a-button type="primary" html-type="submit" :loading="loading">更新</a-button>
        <a-button style="margin-left: 10px" @click="goBack">返回</a-button>
      </a-form-item>
    </a-form>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { message } from 'ant-design-vue'
import { useRoute, useRouter } from 'vue-router'
import { getAppVoById, updateApp, getAppVoByIdByAdmin, updateAppByAdmin } from '@/api/appController.ts'
import { useLoginUserStore } from '@/stores/loginUser.ts'

const route = useRoute()
const router = useRouter()
const loginUserStore = useLoginUserStore()

const form = reactive<API.AppUpdateDto>({
  id: undefined,
  appName: ''
})

const loading = ref(false)

// 获取应用信息
const fetchAppInfo = async () => {
  const appId = route.params.id as string
  if (!appId) {
    message.error('应用ID不存在')
    return
  }
  
  form.id = parseInt(appId)
  
  // 普通用户只能获取自己的应用信息
  // 管理员可以获取任意应用信息
  let res
  if (loginUserStore.loginUser.userRole === 'admin') {
    res = await getAppVoByIdByAdmin({ id: form.id })
  } else {
    res = await getAppVoById({ id: form.id })
  }
  
  if (res.data.code === 0 && res.data.data) {
    const appInfo = res.data.data
    // 检查权限：普通用户只能编辑自己的应用
    if (loginUserStore.loginUser.userRole !== 'admin' && appInfo.userId !== loginUserStore.loginUser.id) {
      message.error('您没有权限编辑此应用')
      router.back()
      return
    }
    
    form.appName = appInfo.appName ?? ''
  } else {
    message.error('获取应用信息失败，' + res.data.message)
  }
}

// 提交表单
const handleSubmit = async (values: any) => {
  loading.value = true
  try {
    // 普通用户使用普通更新接口
    // 管理员使用管理员更新接口
    let res
    if (loginUserStore.loginUser.userRole === 'admin') {
      res = await updateAppByAdmin(values)
    } else {
      res = await updateApp(values)
    }
    
    if (res.data.code === 0) {
      message.success('更新成功')
      router.back()
    } else {
      message.error('更新失败，' + res.data.message)
    }
  } catch (error) {
    message.error('更新失败，请重试')
  } finally {
    loading.value = false
  }
}

// 返回
const goBack = () => {
  router.back()
}

// 页面加载时获取应用信息
onMounted(() => {
  fetchAppInfo()
})
</script>

<style scoped>
#appUpdatePage {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.title {
  text-align: center;
  margin-bottom: 32px;
}
</style>