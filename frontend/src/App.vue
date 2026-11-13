<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import AppShell from "@/components/AppShell.vue";
import ToastHost from "@/components/ToastHost.vue";

const route = useRoute();
const isPublic = computed(() => route.meta.public === true);
</script>

<template>
  <div class="app-root">
    <router-view v-if="isPublic" v-slot="{ Component }">
      <transition name="page" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>

    <AppShell v-else>
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" :key="route.path" />
        </transition>
      </router-view>
    </AppShell>

    <ToastHost />
  </div>
</template>

<style>
.app-root {
  height: 100%;
  position: relative;
  z-index: 1;
}
</style>
