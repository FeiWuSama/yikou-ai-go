import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '@/pages/HomePage.vue'
import UserLoginPage from '@/pages/user/UserLoginPage.vue'
import UserRegisterPage from '@/pages/user/UserRegisterPage.vue'
import UserManagePage from '@/pages/admin/user/UserManagePage.vue'
import UserCenterPage from '@/pages/user/UserCenterPage.vue'
import AppManagePage from '@/pages/admin/app/AppManagePage.vue'
import ChatManagePage from '@/pages/admin/chat/ChatManagePage.vue'
import AppUpdatePage from '@/components/AppUpdatePage.vue'
import AppChatPage from '@/pages/app/AppChatPage.vue'
import accessEnum from '@/access/accessEnum.ts'
import AppEditPage from '@/pages/app/AppEditPage.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: '主页',
      component: HomePage,
    },
    {
      path: '/user/login',
      name: '用户登录',
      component: UserLoginPage,
    },
    {
      path: '/user/register',
      name: '用户注册',
      component: UserRegisterPage,
    },
    {
      path: '/admin/userManage',
      name: '用户管理',
      component: UserManagePage,
      meta: {
        access: accessEnum.ADMIN,
      },
    },
    {
      path: '/admin/appManage',
      name: '应用管理',
      component: AppManagePage,
      meta: {
        access: accessEnum.ADMIN,
      },
    },
    {
      path: '/admin/chatManage',
      name: '对话管理',
      component: ChatManagePage,
      meta: {
        access: accessEnum.ADMIN,
      },
    },
    {
      path: '/admin/app/update/:id',
      name: '应用信息修改',
      component: AppUpdatePage,
      meta: {
        access: accessEnum.USER,
      },
    },
    {
      path: '/app/chat/:id',
      name: '应用对话',
      component: AppChatPage,
      meta: {
        access: accessEnum.USER,
      },
    },
    {
      path: '/app/edit/:id',
      name: '编辑应用',
      component: AppEditPage,
    },
    {
      path: '/user/center',
      name: '个人中心',
      component: UserCenterPage,
      meta: {
        access: accessEnum.USER,
      },
    },
  ],
})

export default router
