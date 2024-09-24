<template>
  <bk-popover
    v-if="isFileType"
    ref="buttonRef"
    theme="light batch-operation-button-popover"
    placement="bottom-end"
    trigger="click"
    width="108"
    :arrow="false"
    @after-show="isPopoverOpen = true"
    @after-hidden="isPopoverOpen = false">
    <bk-button :disabled="props.selectedIds.length === 0" :class="['batch-set-btn', { 'popover-open': isPopoverOpen }]">
      {{ t('批量操作') }}
      <AngleDown class="angle-icon" />
    </bk-button>
    <template #content>
      <div class="operation-item" @click="handleOpenBantchEditPerm">
        {{ t('批量修改权限') }}
      </div>
      <div class="operation-item" @click="handleOpenBantchDelet">
        {{ t('批量删除') }}
      </div>
    </template>
  </bk-popover>
  <bk-button
    v-else
    class="batch-delete-btn"
    :disabled="props.selectedIds.length === 0 && !isAcrossChecked"
    @click="isBatchDeleteDialogShow = true">
    {{ t('批量删除') }}
  </bk-button>
  <DeleteConfirmDialog
    v-model:is-show="isBatchDeleteDialogShow"
    :title="
      t('确认删除所选的 {n} 项配置项？', {
        n: isAcrossChecked ? dataCount - props.selectedIds.length : props.selectedIds.length,
      })
    "
    :pending="batchDeletePending"
    @confirm="handleBatchDeleteConfirm">
    <div>
      {{
        t('已生成版本中存在的配置项，可以通过恢复按钮撤销删除，新增且未生成版本的配置项，将无法撤销删除，请谨慎操作。')
      }}
    </div>
  </DeleteConfirmDialog>
  <EditPermissionDialog
    v-model:show="isBatchEditPermDialogShow"
    :loading="editLoading"
    :configs-length="props.selectedIds.length"
    @confirm="handleConfimEditPermission" />
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { AngleDown } from 'bkui-vue/lib/icon';
  import { useI18n } from 'vue-i18n';
  import Message from 'bkui-vue/lib/message';
  import { batchDeleteServiceConfigs, batchDeleteKv, batchAddConfigList } from '../../../../../../../api/config';
  import DeleteConfirmDialog from '../../../../../../../components/delete-confirm-dialog.vue';
  import EditPermissionDialog from '../../../../../templates/list/package-detail/operations/edit-permission/edit-permission-dialog.vue';
  import { IConfigItem } from '../../../../../../../../types/config';

  const { t } = useI18n();

  interface IPermissionType {
    privilege: string;
    user: string;
    user_group: string;
  }

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    selectedIds: number[];
    isFileType: boolean; // 是否为文件型配置
    selectedItems: IConfigItem[];
    isAcrossChecked: boolean;
    dataCount: number;
  }>();

  const emits = defineEmits(['deleted']);

  const batchDeletePending = ref(false);
  const isBatchDeleteDialogShow = ref(false);
  const isBatchEditPermDialogShow = ref(false);
  const isPopoverOpen = ref(false);
  const buttonRef = ref();
  const editLoading = ref(false);

  const handleBatchDeleteConfirm = async () => {
    batchDeletePending.value = true;
    if (props.isFileType) {
      await batchDeleteServiceConfigs(props.bkBizId, props.appId, props.selectedIds);
    } else {
      await batchDeleteKv(props.bkBizId, props.appId, props.selectedIds, props.isAcrossChecked);
    }
    Message({
      theme: 'success',
      message: props.isFileType ? t('批量删除配置文件成功') : t('批量删除配置项成功'),
    });
    isBatchDeleteDialogShow.value = false;
    setTimeout(() => {
      emits('deleted');
      batchDeletePending.value = false;
    }, 300);
  };

  const handleOpenBantchEditPerm = () => {
    buttonRef.value.hide();
    isBatchEditPermDialogShow.value = true;
  };

  const handleOpenBantchDelet = () => {
    buttonRef.value.hide();
    isBatchDeleteDialogShow.value = true;
  };

  const handleConfimEditPermission = async ({ permission }: { permission: IPermissionType }) => {
    try {
      editLoading.value = true;
      const { privilege, user, user_group } = permission;
      const editConfigList = props.selectedItems.map((item) => {
        const { id, spec, commit_spec } = item;
        return {
          id,
          ...spec,
          privilege: privilege || spec.permission.privilege,
          user: user || spec.permission.user,
          user_group: user_group || spec.permission.user_group,
          byte_size: commit_spec.content.byte_size,
          sign: commit_spec.content.signature,
        };
      });
      await batchAddConfigList(props.bkBizId, props.appId, { items: editConfigList });
      Message({
        theme: 'success',
        message: t('配置文件权限批量修改成功'),
      });
      isBatchEditPermDialogShow.value = false;
    } catch (error) {
      console.error(error);
    } finally {
      editLoading.value = false;
    }
    emits('deleted');
  };
</script>

<style lang="scss" scoped>
  .batch-set-btn {
    min-width: 108px;
    height: 32px;
    margin-left: 8px;
    &.popover-open {
      .angle-icon {
        transform: rotate(-180deg);
      }
    }
    .angle-icon {
      font-size: 20px;
      transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }
  }
  .user-settings {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
  }
  .perm-input {
    display: flex;
    align-items: center;
    width: 172px;
    :deep(.bk-input) {
      width: 140px;
      border-right: none;
      border-top-right-radius: 0;
      border-bottom-right-radius: 0;
      .bk-input--number-control {
        display: none;
      }
    }
    .perm-panel-trigger {
      width: 32px;
      height: 32px;
      text-align: center;
      background: #fafcfe;
      color: #3a84ff;
      border: 1px solid #3a84ff;
      cursor: pointer;
      &.disabled {
        color: #dcdee5;
        border-color: #dcdee5;
        cursor: not-allowed;
      }
    }
  }
  .privilege-select-panel {
    display: flex;
    align-items: top;
    border: 1px solid #dcdee5;
    .group-item {
      .header {
        padding: 0 16px;
        height: 42px;
        line-height: 42px;
        color: #313238;
        font-size: 12px;
        background: #fafbfd;
        border-bottom: 1px solid #dcdee5;
      }
      &:not(:last-of-type) {
        .header,
        .checkbox-area {
          border-right: 1px solid #dcdee5;
        }
      }
    }
    .checkbox-area {
      padding: 10px 16px 12px;
      background: #ffffff;
      &:not(:last-child) {
        border-right: 1px solid #dcdee5;
      }
    }
    .group-checkboxs {
      font-size: 12px;
      .bk-checkbox ~ .bk-checkbox {
        margin-left: 16px;
      }
      :deep(.bk-checkbox-label) {
        font-size: 12px;
      }
    }
  }
  .selected-tag {
    display: inline-block;
    height: 32px;
    background: #f0f1f5;
    line-height: 32px;
    border-radius: 16px;
    padding: 0 12px;
    margin: 8px 0px 16px;
    .count {
      color: #3a84ff;
    }
  }
</style>

<style lang="scss">
  .batch-operation-button-popover.bk-popover.bk-pop2-content {
    padding: 4px 0;
    border: 1px solid #dcdee5;
    box-shadow: 0 2px 6px 0 #0000001a;
    width: auto !important;
    .operation-item {
      padding: 0 12px;
      min-width: 58px;
      height: 32px;
      line-height: 32px;
      color: #63656e;
      font-size: 12px;
      cursor: pointer;
      &:hover {
        background: #f5f7fa;
      }
    }
  }
</style>
