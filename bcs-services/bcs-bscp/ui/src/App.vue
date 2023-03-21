<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useUserStore } from './store/user'
import isCrossOriginIFrame from './utils/is-cross-origin-iframe'

import Header from "./components/head.vue";

const store = useUserStore()
const { getUserInfo } = store

watch(() => store.showLoginModal, (val) => {
  if (val) {
    const topWindow = isCrossOriginIFrame() ? window : window.top
    // @ts-ignore
    topWindow.BLUEKING.corefunc.open_login_dialog(store.loginUrl)
  }
})

onMounted(() => {
  getUserInfo()
});


</script>

<template>
  <div>
    <Header></Header>
    <div class="content">
      <router-view></router-view>
    </div>
  </div>
</template>


<style scoped>
.content {
  height: calc(100vh - 52px);
}
</style>
