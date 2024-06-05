<template>
  <bk-button :disabled="props.selections.length === 0" @click="isRetryOpen = true">{{ t('批量重试') }}</bk-button>
  <RetryDialog
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    :is-show="isRetryOpen"
    :is-batch="true"
    :selections="props.selections"
    @close="isRetryOpen = false"
    @retried="emits('retried', $event)" />
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import RetryDialog from './retry-dialog.vue';

  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    selections: { id: number; uid: string; release_name: string }[];
  }>();

  const emits = defineEmits(['retried']);

  const isRetryOpen = ref(false);
</script>
<style lang="scss" scoped>
  .bk-button {
    margin-bottom: 16px;
  }
</style>
