<script setup lang="ts">
import { watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useGlobalStore } from './store/global'
import { useUserStore } from './store/user'
import isCrossOriginIFrame from './utils/is-cross-origin-iframe'
import Header from "./components/head.vue";
import PermissionDialog from './components/permission/apply-dialog.vue'

const userStore = useUserStore()
const globalStore = useGlobalStore()
const { showApplyPermDialog } = storeToRefs(globalStore)

watch(() => userStore.showLoginModal, (val) => {
  if (val) {
    const topWindow = isCrossOriginIFrame() ? window : window.top
    // @ts-ignore
    topWindow.BLUEKING.corefunc.open_login_dialog(userStore.loginUrl)
  }
})

</script>

<template>
  <div class="page-content-container">
    <Header></Header>
    <div class="content">
      <router-view></router-view>
      <permission-dialog :show="showApplyPermDialog"></permission-dialog>
    </div>
  </div>
</template>

<style scoped>
.page-content-container {
  min-width: 1366px;
  overflow: auto;
}
.content {
  height: calc(100vh - 52px);
}
</style>
