<template>
  <a-card hoverable class="app-card" :key="appData.id">
    <!-- 下拉菜单，移动到右下角 -->
    <div
      v-if="(!isGoodApp || isAdmin) && showExpandIcon"
      class="dropdown-container"
    >
      <a-dropdown>
        <div class="more-icon">⋮</div>
        <template #overlay>
          <a-menu>
            <a-menu-item @click="handleAction('edit', appData)">
              <div class="dropdown-item">编辑应用</div>
            </a-menu-item>
            <a-menu-item @click="handleAction('delete', appData)">
              <div class="dropdown-item">删除应用</div>
            </a-menu-item>
          </a-menu>
        </template>
      </a-dropdown>
    </div>
    <template #cover>
      <div
        class="cover-container"
        @mouseenter="showOverlay = true"
        @mouseleave="showOverlay = false"
      >
        <img :src="appData.cover || defaultCover" alt="应用封面" class="app-cover" />
        <!-- 覆盖层 -->
        <div v-if="showOverlay" class="cover-overlay">
          <div class="action-buttons-container">
            <!-- 精选应用：显示预览和查看作品（如果有deployKey） -->
            <div v-if="isGoodApp" class="capsule-buttons-horizontal">
              <div
                class="capsule-button"
                @click="handleAction('continue', appData)"
              >
                <span class="action-text">继续创作</span>
              </div>
              <div
                v-if="appData.deployKey"
                class="capsule-button"
                @click="handleViewWork(appData)"
              >
                <span class="action-text">查看作品</span>
              </div>
            </div>
            <!-- 我的应用：显示所有操作 -->
            <div v-else class="capsule-buttons-horizontal">
              <div
                class="capsule-button"
                @click="handleAction('continue', appData)"
              >
                <span class="action-text">继续创作</span>
              </div>
              <div
                v-if="appData.deployKey"
                class="capsule-button"
                @click="handleViewWork(appData)"
              >
                <span class="action-text">查看作品</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
    <div class="app-info">
      <span class="app-name">{{ appData.appName }}</span>
      <div class="user-info" v-if="showUserInfo">
        <a-avatar :src="appData.user?.userAvatar" size="small" />
        <span class="app-user-name">{{ appData.user?.userName }}</span>
      </div>
    </div>
  </a-card>
</template>

<script setup lang="ts">
import { defineProps, defineEmits, ref, computed } from 'vue'
import { Dropdown, Menu } from 'ant-design-vue'
import defaultCover from '@/assets/logo.png'
import { getDeployUrl } from '@/config/env.ts'
import { useLoginUserStore } from '@/stores/loginUser.ts'

// 注册组件
const ADropdown = Dropdown
const AMenu = Menu
const AMenuItem = Menu.Item

interface AppCardProps {
  appData: API.AppVO
  actions?: Array<{
    key: string
    label: string
  }>
  showUserInfo?: boolean
  isGoodApp?: boolean // 是否为精选应用
}

const props = withDefaults(defineProps<AppCardProps>(), {
  showUserInfo: false,
  isGoodApp: false,
})

const emit = defineEmits<{
  action: [actionKey: string, appData: API.AppVO]
}>()

const showOverlay = ref(false)
const loginUserStore = useLoginUserStore()

// 检查是否为管理员
const isAdmin = computed(() => {
  return loginUserStore.loginUser?.userRole === 'admin'
})

// 是否显示展开图标
const showExpandIcon = computed(() => {
  // 精选应用只有管理员才能看到展开图标，普通应用始终显示
  return !props.isGoodApp || isAdmin.value
})

const handleAction = (actionKey: string, appData: API.AppVO) => {
  emit('action', actionKey, appData)
}

// 查看作品
const handleViewWork = (appData: API.AppVO) => {
  if (appData.deployKey) {
    // 在新窗口中打开部署地址
    const deployUrl = getDeployUrl(appData.deployKey)
    window.open(deployUrl, '_blank')
  }
}
</script>

<style scoped>
.app-card {
  width: 300px;
  margin: 10px;
  position: relative;
}

/* Tooltip样式 */
:deep(.ant-tooltip-inner) {
  background-color: #fff;
  color: #333;
  border: 1px solid #ddd;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

:deep(.ant-tooltip-arrow::before) {
  background-color: #fff;
}

.cover-container {
  position: relative;
  overflow: hidden;
}

.app-cover {
  height: 150px;
  object-fit: cover;
  width: 100%;
  transition: all 0.3s ease;
}

.cover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  opacity: 0;
  animation: fadeIn 0.3s ease forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.action-buttons-container {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 10px;
  gap: 10px;
}

.capsule-buttons-horizontal {
  display: flex;
  align-items: center;
  gap: 10px;
}

.capsule-button {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 5px 20px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 15px;
  cursor: pointer;
  transition: all 0.3s ease;
  min-width: 70px;
  text-align: center;
  color: #333;
  border: 2px solid #1890ff; /* 蓝色边框 */
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
}

.capsule-button:hover {
  background: #1890ff; /* 蓝色背景 */
  color: white; /* 白色文字 */
  transform: translateY(-2px); /* 向上轻微移动 */
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2); /* 增强阴影 */
}

.action-text {
  font-size: 11px;
  font-weight: 500;
}

.dropdown-container {
  position: absolute;
  bottom: 5px;
  right: 10px;
  z-index: 10;
}

.more-icon {
  width: 45px;
  height: 45px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  cursor: pointer;
}

.dropdown-item {
  padding: 8px 15px;
  cursor: pointer;
  font-size: 14px;
  color: #333;
  text-align: center;
}

.dropdown-item:hover {
  background: #f5f5f5;
}

.app-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  padding: 8px;
}

.app-name {
  font-size: 16px;
  font-weight: bold;
  color: #333;
  margin-bottom: 3px;
}

.user-info {
  display: flex;
  align-items: center;
  margin-top: 3px;
}

.app-user-name {
  margin-left: 8px;
  font-size: 12px;
  color: #666;
}

:deep(.ant-card-body) {
  padding: 5px;
}

@media (max-width: 768px) {
  .app-card {
    width: 100%;
  }

  .capsule-buttons-horizontal {
    gap: 4px;
  }

  .capsule-button {
    min-width: 60px;
    padding: 0 20px;
  }

  .action-text {
    font-size: 10px;
  }

  .action-buttons-container {
    margin-bottom: 15px;
  }

  .app-info {
    padding: 6px;
  }

  .app-name {
    margin-bottom: 4px;
  }

  .user-info {
    margin-top: 4px;
  }
}

@media (max-width: 480px) {
  .capsule-buttons-horizontal {
    flex-direction: column;
    gap: 4px;
  }

  .capsule-button {
    min-width: 100px;
    padding: 0 20px;
  }

  .action-buttons-container {
    margin-bottom: 25px;
  }

  .app-info {
    padding: 6px;
  }

  .app-name {
    margin-bottom: 4px;
  }

  .user-info {
    margin-top: 4px;
  }
}
</style>
