<template>
  <bk-exception
    class="exception-apply"
    type="403">
    <div class="tips-content">
      <div class="title">{{ t('产品仅供内部体验') }}</div>
      <div class="desc">{{ t('如需体验，请联系') }} {{ helper }}</div>
    </div>
  </bk-exception>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
const helper = (window as any).HELPER;
</script>
<style lang="scss" scoped>
  .exception-apply {
    padding-top: 120px;
    .title {
      font-size: 14px;
      color: #63656e;
      line-height: 22px;
    }
    .desc {
      margin-top: 8px;
      font-size: 12px;
      color: #979ba5;
      line-height: 20px;
    }
  }
  :deep(.bk-exception-footer) {
    margin-top: 0;
  }
  :deep(.bk-exception-img) {
    width: 280px;
    height: auto;
  }
</style>
