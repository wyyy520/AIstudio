<template>
  <LogCenter />
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import LogCenter from '@/components/logs/LogCenter.vue'
import { useLogStore } from '@/stores/log'

const route = useRoute()
const store = useLogStore()

// 监听路由参数变化，自动选择任务
watch(
  () => route.query.taskId,
  (taskId) => {
    if (taskId && typeof taskId === 'string') {
      store.selectTask(taskId)
    }
  },
  { immediate: true }
)

onMounted(() => {
  // 页面加载时检查是否有taskId参数
  const taskId = route.query.taskId
  if (taskId && typeof taskId === 'string') {
    store.selectTask(taskId)
  }
})
</script>