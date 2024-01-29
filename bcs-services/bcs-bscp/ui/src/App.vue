<template>
  <div class="page-content-container">
    <notice-component v-if="showNotice" :api-url="noticeApiURL" @show-alert-change="showNotice = $event" />
    <Header></Header>
    <div :class="['content', { 'show-notice': showNotice }]">
      <router-view></router-view>
      <permission-dialog :show="showApplyPermDialog"></permission-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue';
import { storeToRefs } from 'pinia';
import useGlobalStore from './store/global';
import useUserStore from './store/user';
import isCrossOriginIFrame from './utils/is-cross-origin-iframe';
import NoticeComponent from '@blueking/notice-component'
import '@blueking/notice-component/dist/style.css'
import Header from './components/head.vue';
import PermissionDialog from './components/permission/apply-dialog.vue';

const userStore = useUserStore();
const globalStore = useGlobalStore();
const { showLoginModal } = storeToRefs(userStore);
const { showApplyPermDialog, showNotice } = storeToRefs(globalStore);

// @ts-ignore
const noticeApiURL = `${window.BK_BCS_BSCP_API}/api/v1/announcements`

watch(
  () => showLoginModal.value,
  (val) => {
    if (val) {
      const topWindow = isCrossOriginIFrame() ? window : window.top;
      // @ts-ignore
      topWindow.BLUEKING.corefunc.open_login_dialog(userStore.loginUrl);
    }
  },
);
</script>

<style scoped lang="scss">
.page-content-container {
  min-width: 1366px;
  overflow: auto;
}
.content {
  height: calc(100vh - 52px);
  &.show-notice {
    height: calc(100vh - 92px);
  }
}
</style>
