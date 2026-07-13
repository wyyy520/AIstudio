<template>
  <WorkflowEditor
    :workflow-id="workflowId"
    :workflow-name="workflowId ? '编辑工作流' : '新建工作流'"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import WorkflowEditor from './components/WorkflowEditor.vue'

const route = useRoute()

const workflowId = computed(() => {
  return (route.params.id as string) || undefined
})
</script>