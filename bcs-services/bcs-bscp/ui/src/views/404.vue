<template>
  <bk-exception class="exception-wrap-item" type="404" :title="t('页面不存在')"> </bk-exception>
</template>
<script setup lang="ts">
  import { useI18n } from 'vue-i18n';
  const { t } = useI18n();
</script>
