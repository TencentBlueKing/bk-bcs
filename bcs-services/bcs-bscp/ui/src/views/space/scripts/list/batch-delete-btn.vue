<template>
  <bk-button
    class="batch-delete-btn"
    :disabled="props.selectedIds.length === 0 && !isAcrossChecked"
    @click="isBatchDeleteDialogShow = true">
    {{ t('批量删除') }}
  </bk-button>
  <DeleteConfirmDialog
    v-model:isShow="isBatchDeleteDialogShow"
    :title="
      t('确认删除所选的 {n} 项脚本？', {
        n: isAcrossChecked ? dataCount - props.selectedIds.length : props.selectedIds.length,
      })
    "
    :pending="batchDeletePending"
    @confirm="handleBatchDeleteConfirm">
    <div>
      {{ t('一旦删除，该操作将无法撤销，服务配置的未命名版本中引用该脚本也将清除，请谨慎操作。') }}
    </div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import Message from 'bkui-vue/lib/message';
  import { batchDeleteScript } from '../../../../api/script';
  import DeleteConfirmDialog from '../../../../components/delete-confirm-dialog.vue';

  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    selectedIds: number[];
    isAcrossChecked: boolean;
    dataCount: number;
  }>();

  const emits = defineEmits(['deleted']);

  const batchDeletePending = ref(false);
  const isBatchDeleteDialogShow = ref(false);

  const handleBatchDeleteConfirm = async () => {
    batchDeletePending.value = true;
    await batchDeleteScript(props.bkBizId, props.selectedIds, props.isAcrossChecked);
    Message({
      theme: 'success',
      message: t('批量删除脚本成功'),
    });
    batchDeletePending.value = false;
    isBatchDeleteDialogShow.value = false;
    emits('deleted');
  };
</script>
