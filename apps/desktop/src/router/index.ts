import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/pages/Dashboard/Dashboard.vue'),
  },
  {
    path: '/chat',
    name: 'AIChat',
    component: () => import('@/pages/AIChat/AIChat.vue'),
  },
  {
    path: '/project',
    name: 'Project',
    component: () => import('@/pages/Project/Project.vue'),
  },
  {
    path: '/projects',
    redirect: '/project',
  },
  {
    path: '/workflow',
    name: 'Workflow',
    component: () => import('@/pages/Workflow/Workflow.vue'),
  },
  {
    path: '/workflow/:id',
    name: 'WorkflowEdit',
    component: () => import('@/pages/Workflow/Workflow.vue'),
  },
  {
    path: '/project/:projectId/workflow',
    name: 'ProjectWorkflow',
    component: () => import('@/pages/Workflow/Workflow.vue'),
  },
  {
    path: '/project/:projectId/workflow/:workflowId',
    name: 'ProjectWorkflowEdit',
    component: () => import('@/pages/Workflow/Workflow.vue'),
  },
  {
    path: '/plugins',
    name: 'PluginStore',
    component: () => import('@/pages/PluginStore/PluginStore.vue'),
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/pages/Settings/Settings.vue'),
  },
  {
    path: '/logs',
    name: 'Logs',
    component: () => import('@/pages/Logs/Logs.vue'),
  },
  {
    path: '/logs/:taskId',
    name: 'TaskLogs',
    component: () => import('@/pages/Logs/Logs.vue'),
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router