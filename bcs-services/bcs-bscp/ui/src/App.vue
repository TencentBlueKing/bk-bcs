<template>
  <div class="page-content-container">
    <notice-component v-if="enableNotice" :api-url="noticeApiURL" @show-alert-change="showNotice = $event" />
    <Header></Header>
    <div :class="['content', { 'show-notice': showNotice }]">
      <router-view></router-view>
      <permission-dialog :show="showApplyPermDialog"></permission-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { storeToRefs } from 'pinia';
  import useGlobalStore from './store/global';
  import NoticeComponent from '@blueking/notice-component';
  import '@blueking/notice-component/dist/style.css';
  import Header from './components/head.vue';
  import PermissionDialog from './components/permission/apply-dialog.vue';

  const globalStore = useGlobalStore();
  const { showApplyPermDialog, showNotice } = storeToRefs(globalStore);

  // @ts-ignore
  const noticeApiURL = `${window.BK_BCS_BSCP_API}/api/v1/announcements`;
  // @ts-ignore
  const enableNotice = window.ENABLE_BK_NOTICE === 'true';
</script>

<style scoped lang="scss">
  .page-content-container {
    min-width: 1366px;
    overflow: auto;
  }
  .content {
    height: calc(100vh - 52px);
    margin-top: 52px;
    &.show-notice {
      margin-top: 92px;
      height: calc(100vh - 92px);
    }
  }
</style>
