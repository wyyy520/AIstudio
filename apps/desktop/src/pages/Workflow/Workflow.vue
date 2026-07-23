<template>
  <WorkflowEditor
    :project-id="projectId"
    :workflow-name="projectId ? 'Project Workflow' : 'New Workflow'"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import WorkflowEditor from './components/WorkflowEditor.vue'

const route = useRoute()

// Support both /project/:projectId/workflow and /workflow/:id routes
const projectId = computed(() => {
  return (route.params.projectId as string) || undefined
})
</script>
