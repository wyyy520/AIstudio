<template>
  <div class="empty-layout">
    <TitleBar />
    <div class="empty-content">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import TitleBar from '@/components/titlebar/TitleBar.vue'
</script>

<style scoped>
.empty-layout {
  width: 100%;
  height: 100%;
  overflow: hidden;
  background: var(--bg-primary);
}

.empty-content {
  width: 100%;
  height: calc(100% - var(--titlebar-height));
  overflow: hidden;
}
</style>