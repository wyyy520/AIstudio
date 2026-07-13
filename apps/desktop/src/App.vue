<template>
  <AppLayout>
    <router-view v-slot="{ Component }">
      <Transition name="fade" mode="out-in">
        <component :is="Component" />
      </Transition>
    </router-view>
  </AppLayout>
</template>

<script setup lang="ts">
import AppLayout from '@/layouts/AppLayout.vue'
</script>

<style>
#app {
  width: 100%;
  height: 100%;
}
</style>