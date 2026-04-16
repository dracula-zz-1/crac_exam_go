<template>
  <div id="app">
    <router-view v-slot="{ Component, route }">
      <keep-alive include="Home">
        <component :is="Component" :key="route.fullPath" v-if="route.name !== 'Login'" />
      </keep-alive>
      <component :is="Component" v-if="route.name === 'Login'" />
    </router-view>
  </div>
</template>

<script setup lang="ts">
// 应用根组件
</script>

<style>
#app {
  width: 100%;
  height: 100vh;
  margin: 0;
  padding: 0;
  overflow: hidden !important;
}

body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  overflow: hidden !important;
}

html {
  overflow: hidden !important;
}

/* 全局移除滚动条 */
* {
  scrollbar-width: none !important; /* Firefox */
  -ms-overflow-style: none !important; /* IE/Edge */
}

*::-webkit-scrollbar {
  display: none !important; /* Chrome/Safari/Opera */
}

/* 允许内部容器滚动但不显示滚动条 */
.question-list,
.el-textarea__inner,
.el-select-dropdown,
.el-scrollbar__wrap {
  overflow-y: auto !important;
  scrollbar-width: none !important;
  -ms-overflow-style: none !important;
}

.question-list::-webkit-scrollbar,
.el-textarea__inner::-webkit-scrollbar,
.el-select-dropdown::-webkit-scrollbar,
.el-scrollbar__wrap::-webkit-scrollbar {
  display: none !important;
}
</style>
