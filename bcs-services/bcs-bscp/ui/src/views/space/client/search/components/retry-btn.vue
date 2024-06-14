<template>
  <bk-button theme="primary" text @click="isRetryOpen = true">
    {{ t('重试') }}
  </bk-button>
  <RetryDialog
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    :is-show="isRetryOpen"
    :is-batch="false"
    :selections="selections"
    @close="isRetryOpen = false"
    @retried="emits('retried', $event)" />
</template>
<script lang="ts" setup>
  import { ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import RetryDialog from './retry-dialog.vue';

  interface IClient {
    id: number;
    spec: {
      current_release_name: string;
      target_release_name: string;
    };
    attachment: {
      uid: string;
    };
  }

  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    client: IClient;
  }>();

  const emits = defineEmits(['retried']);

  const isRetryOpen = ref(false);

  const selections = computed(() => {
    return [
      {
        id: props.client.id,
        uid: props.client.attachment.uid,
        current_release_name: props.client.spec.current_release_name,
        target_release_name: props.client.spec.target_release_name,
      },
    ];
  });
</script>
<style lang="scss" scoped>
  .bk-button {
    margin-left: 8px;
  }
</style>
