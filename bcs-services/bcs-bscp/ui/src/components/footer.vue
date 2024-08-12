<template>
  <div class="app-footer">
    <div class="links" v-html="appGlobalConfig.i18n.footerInfoHTML"></div>
    <p>{{ copyRight }}</p>
  </div>
</template>
<script lang="ts" setup>
  import { computed } from 'vue';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../store/global';
  const { bscpVersion, appGlobalConfig } = storeToRefs(useGlobalStore());

  const copyRight = computed(() => {
    return (
      appGlobalConfig.value.footerCopyrightContent ||
      `Copyright © 2012 Tencent BlueKing. All Rights Reserved. ${bscpVersion.value}`
    );
  });
</script>
<style lang="scss" scoped>
  .app-footer {
    font-size: 12px;
    color: #b7c0ca;
    width: 100%;
    text-align: center;
    line-height: 20px;
    padding-top: 25px;
    .links :deep(.link-item) {
      color: #3a84ff;
      &:hover {
        color: #3a84ff;
      }
    }
  }
</style>
