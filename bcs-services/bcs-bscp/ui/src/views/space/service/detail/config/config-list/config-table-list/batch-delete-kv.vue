<template>
  <bk-button
    class="batch-delete-btn"
    :disabled="props.selectedIds.length === 0"
    @click="isBatchDeleteDialogShow = true">
    {{ t('批量删除') }}
  </bk-button>
  <DeleteConfirmDialog
    v-model:isShow="isBatchDeleteDialogShow"
    :title="t('确认删除所选的 {n} 项配置项？', { n: props.selectedIds.length })"
    @confirm="handleBatchDeleteConfirm">
    <div>{{ t('如果已生成版本中删除，该操作将无法撤销，请谨慎操作') }}</div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import Message from 'bkui-vue/lib/message';
  import { batchDeleteKv } from '../../../../../../../api/config';
  import DeleteConfirmDialog from '../../../../../../../components/delete-confirm-dialog.vue';

  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    selectedIds: number[];
  }>();

  const batchDeletePending = ref(false);
  const isBatchDeleteDialogShow = ref(false);

  const handleBatchDeleteConfirm = async () => {
    batchDeletePending.value = true;
    await batchDeleteKv(props.bkBizId, props.appId, props.selectedIds);
    Message({
      theme: 'success',
      message: t('批量删除配置项成功'),
    });
    isBatchDeleteDialogShow.value = false;
  };
</script>
