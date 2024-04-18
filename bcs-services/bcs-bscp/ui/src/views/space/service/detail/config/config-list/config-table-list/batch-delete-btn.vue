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
    :pending="batchDeletePending"
    @confirm="handleBatchDeleteConfirm">
    <div>
      {{
        t('已生成版本中存在的配置项，可以通过恢复按钮撤销删除，新增且未生成版本的配置项，将无法撤销删除，请谨慎操作。')
      }}
    </div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import Message from 'bkui-vue/lib/message';
  import { batchDeleteServiceConfigs, batchDeleteKv } from '../../../../../../../api/config';
  import DeleteConfirmDialog from '../../../../../../../components/delete-confirm-dialog.vue';

  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    selectedIds: number[];
    isFileType: boolean; // 是否为文件型配置
  }>();

  const emits = defineEmits(['deleted']);

  const batchDeletePending = ref(false);
  const isBatchDeleteDialogShow = ref(false);

  const handleBatchDeleteConfirm = async () => {
    batchDeletePending.value = true;
    if (props.isFileType) {
      await batchDeleteServiceConfigs(props.bkBizId, props.appId, props.selectedIds);
    } else {
      await batchDeleteKv(props.bkBizId, props.appId, props.selectedIds);
    }
    Message({
      theme: 'success',
      message: t('批量删除配置项成功'),
    });
    batchDeletePending.value = false;
    isBatchDeleteDialogShow.value = false;
    emits('deleted');
  };
</script>
