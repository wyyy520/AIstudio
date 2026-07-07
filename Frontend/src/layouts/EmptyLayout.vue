<template>
  <div class="empty-layout">
    <router-view />
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.empty-layout {
  width: 100%;
  height: 100%;
  overflow: hidden;
  background: var(--bg-primary);
}
</style>