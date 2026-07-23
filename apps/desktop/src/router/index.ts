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
  {
    path: '/compiler',
    name: 'Compiler',
    component: () => import('@/pages/Compiler/Compiler.vue'),
  },
  {
    path: '/generator',
    name: 'Generator',
    component: () => import('@/pages/Generator/Generator.vue'),
  },
  {
    path: '/runtime',
    name: 'Runtime',
    component: () => import('@/pages/Runtime/Runtime.vue'),
  },
  {
    path: '/diagnose',
    name: 'Diagnose',
    component: () => import('@/pages/Diagnose/Diagnose.vue'),
  },
  {
    path: '/skills',
    name: 'SkillCenter',
    component: () => import('@/pages/SkillCenter/SkillCenter.vue'),
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router