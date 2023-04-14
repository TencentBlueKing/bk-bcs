<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useUserStore } from './store/user'
import { useGlobalStore } from './store/global'
import isCrossOriginIFrame from './utils/is-cross-origin-iframe'

import Header from "./components/head.vue";
import PermissionDialog from './components/permission/apply-dialog.vue'

const userStore = useUserStore()
const globalStore = useGlobalStore()
const { getUserInfo } = userStore
const { showApplyPermDialog } = storeToRefs(globalStore)

watch(() => userStore.showLoginModal, (val) => {
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
      <permission-dialog :show="showApplyPermDialog"></permission-dialog>
    </div>
  </div>
</template>


<style scoped>
.content {
  height: calc(100vh - 52px);
}
</style>
