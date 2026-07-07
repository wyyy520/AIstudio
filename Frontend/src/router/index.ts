import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'
import EmptyLayout from '@/layouts/EmptyLayout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: MainLayout,
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/pages/Dashboard/Dashboard.vue'),
        meta: { title: '仪表盘' },
      },
      {
        path: 'chat',
        name: 'AIChat',
        component: () => import('@/pages/AIChat/AIChat.vue'),
        meta: { title: 'AI 对话' },
      },
      {
        path: 'plugins',
        name: 'PluginStore',
        component: () => import('@/pages/PluginStore/PluginStore.vue'),
        meta: { title: '插件市场' },
      },
      {
        path: 'projects',
        name: 'Project',
        component: () => import('@/pages/Project/Project.vue'),
        meta: { title: '项目管理' },
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('@/pages/Logs/Logs.vue'),
        meta: { title: '日志' },
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/pages/Settings/Settings.vue'),
        meta: { title: '设置' },
      },
    ],
  },
  {
    path: '/workflow',
    component: EmptyLayout,
    children: [
      {
        path: '',
        name: 'WorkflowList',
        component: () => import('@/pages/Workflow/Workflow.vue'),
        meta: { title: '工作流' },
      },
      {
        path: ':id',
        name: 'WorkflowEditor',
        component: () => import('@/pages/Workflow/Workflow.vue'),
        meta: { title: '工作流编辑器' },
        props: true,
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0, left: 0 }),
})

router.beforeEach((to, _from, next) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} - AIStudio` : 'AIStudio'
  next()
})

export default router