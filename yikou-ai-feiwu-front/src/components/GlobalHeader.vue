<template>
  <div class="global-header">
    <div class="header-left">
      <div class="logo-container">
        <img alt="Logo" class="logo" src="@/assets/logo.png" />
        <span class="title">易扣</span>
      </div>

      <a-menu v-model:selectedKeys="selectedKeys" mode="horizontal" class="nav-menu">
        <a-menu-item key="主页">
          <router-link to="/">首页</router-link>
        </a-menu-item>
        <a-menu-item key="用户管理" v-if="loginUserStore.loginUser?.userRole === 'admin'">
          <router-link to="/admin/userManage">用户管理</router-link>
        </a-menu-item>
        <a-menu-item key="应用管理" v-if="loginUserStore.loginUser?.userRole === 'admin'">
          <router-link to="/admin/appManage">应用管理</router-link>
        </a-menu-item>
        <a-menu-item key="聊天历史管理" v-if="loginUserStore.loginUser?.userRole === 'admin'">
          <router-link to="/admin/chatManage">聊天历史管理</router-link>
        </a-menu-item>
      </a-menu>
    </div>
    <div class="header-right">
      <!-- 登录按钮，后续替换为用户头像和昵称 -->
      <div v-if="loginUserStore.loginUser.id">
        <a-dropdown>
          <a-space>
            <a-avatar :src="loginUserStore.loginUser.userAvatar" />
            {{ loginUserStore.loginUser.userName ?? '无名' }}
          </a-space>
          <template #overlay>
            <a-menu>
              <a-menu-item @click="goToUserCenter">
                <UserOutlined />
                个人中心
              </a-menu-item>
              <a-menu-item @click="doLogout">
                <LogoutOutlined />
                退出登录
              </a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </div>
      <div v-else>
        <a-button type="primary" href="/user/login">登录</a-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useLoginUserStore } from '@/stores/loginUser.ts'
import { LogoutOutlined, UserOutlined } from '@ant-design/icons-vue'
import { userLogout } from '@/api/userController.ts'
import { message } from 'ant-design-vue'

const loginUserStore = useLoginUserStore()
loginUserStore.fetchLoginUser()

const router = useRouter()
const selectedKeys = ref<string[]>([])

router.afterEach((to, from, next) => {
  selectedKeys.value = [to.name as string]
})

// 用户注销
const doLogout = async () => {
  const res = await userLogout()
  if (res.data.code === 0) {
    loginUserStore.setLoginUser({
      userName: '未登录',
    })
    message.success('退出登录成功')
    await router.push('/user/login')
  } else {
    message.error('退出登录失败，' + res.data.message)
  }
}

// 跳转到个人中心
const goToUserCenter = () => {
  router.push('/user/center')
}
</script>

<style scoped>
.global-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 48px;
  height: 64px;
}

.header-left {
  display: flex;
  align-items: center;
  flex: 1;
}

.logo-container {
  display: flex;
  align-items: center;
  margin-right: 24px;
}

.logo {
  height: 32px;
  width: 32px;
  margin-right: 12px;
}

.title {
  font-weight: 600;
  font-size: 18px;
  color: #000;
}

.nav-menu {
  border: none;
  background: transparent;
  flex: 1;
}

.header-right {
  display: flex;
  align-items: center;
}
</style>
