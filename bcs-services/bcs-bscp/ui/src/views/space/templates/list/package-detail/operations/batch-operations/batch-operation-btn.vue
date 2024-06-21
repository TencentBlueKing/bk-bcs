<template>
  <bk-popover
    ref="buttonRef"
    theme="light batch-operation-button-popover"
    placement="bottom-end"
    trigger="click"
    width="108"
    :arrow="false"
    @after-show="isPopoverOpen = true"
    @after-hidden="isPopoverOpen = false">
    <bk-button :disabled="props.configs.length === 0" :class="['batch-set-btn', { 'popover-open': isPopoverOpen }]">
      {{ t('批量操作') }}
      <AngleDown class="angle-icon" />
    </bk-button>
    <template #content>
      <div class="operation-item" @click="handleOpenAddToDialog">
        {{ t('批量添加至') }}
      </div>
      <div class="operation-item" @click="handleOpenBantchEditPerm">
        {{ t('批量修改权限') }}
      </div>
      <div v-if="props.pkgType === 'without'" class="operation-item" @click="handleOpenBantchDelet">
        {{ t('批量删除') }}
      </div>
      <div v-if="props.pkgType === 'pkg'" class="operation-item" @click="handleOpenBatchMoveOut">
        {{ t('批量移出套餐') }}
      </div>
    </template>
  </bk-popover>
  <AddToDialog v-model:show="isAddToDialogShow" :value="props.configs" @added="emits('refresh')" />
  <EditPermissionDialg
    v-model:show="isEditPermissionShow"
    :loading="editLoading"
    :configs-length="props.configs.length"
    @confirm="handleConfirmEditPermission" />
  <BatchMoveOutFromPkgDialog
    v-model:show="isBatchMoveDialogShow"
    :current-pkg="props.currentPkg as number"
    :value="props.configs"
    @moved-out="emits('movedOut')" />
  <MoveOutFromPkgsDialog
    v-model:show="isSingleMoveDialogShow"
    :id="props.configs.length > 0 ? props.configs[0].id : 0"
    :name="props.configs.length > 0 ? props.configs[0].spec.name : ''"
    :current-pkg="props.currentPkg as number"
    @moved-out="emits('movedOut')" />
  <DeleteConfigDialog v-model:show="isDeleteConfigDialogShow" :configs="props.configs" @deleted="emits('deleted')" />
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { AngleDown } from 'bkui-vue/lib/icon';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import Message from 'bkui-vue/lib/message';
  import { ITemplateConfigItem } from '../../../../../../../../types/template';
  import { batchEditTemplatePermission } from '../../../../../../../api/template';
  import AddToDialog from '../add-to-pkgs/add-to-dialog.vue';
  import EditPermissionDialg from '../edit-permission/edit-permission-dialog.vue';
  import BatchMoveOutFromPkgDialog from '../move-out-from-pkg/batch-move-out-from-pkg-dialog.vue';
  import MoveOutFromPkgsDialog from '../move-out-from-pkg/move-out-from-pkgs-dialog.vue';
  import DeleteConfigDialog from '../delete-configs/delete-config-dialog.vue';

  interface IPermissionType {
    privilege: string;
    user: string;
    user_group: string;
  }

  const { t } = useI18n();
  const route = useRoute();

  const props = defineProps<{
    spaceId: string;
    currentTemplateSpace: number;
    configs: ITemplateConfigItem[];
    pkgType: string;
    currentPkg?: number;
  }>();

  const emits = defineEmits(['refresh', 'movedOut', 'deleted']);
  const bkBizId = ref(String(route.params.spaceId));

  const isDeleteConfigDialogShow = ref(false);
  const isEditPermissionShow = ref(false);
  const isPopoverOpen = ref(false);
  const buttonRef = ref();
  const isAddToDialogShow = ref(false);
  const isBatchMoveDialogShow = ref(false);
  const isSingleMoveDialogShow = ref(false);
  const editLoading = ref(false);

  const handleOpenAddToDialog = () => {
    buttonRef.value.hide();
    isAddToDialogShow.value = true;
  };

  const handleOpenBantchEditPerm = () => {
    buttonRef.value.hide();
    isEditPermissionShow.value = true;
  };

  const handleOpenBantchDelet = () => {
    buttonRef.value.hide();
    isDeleteConfigDialogShow.value = true;
  };

  const handleOpenBatchMoveOut = () => {
    buttonRef.value.hide();
    if (props.configs.length === 1) {
      isSingleMoveDialogShow.value = true;
    } else {
      isBatchMoveDialogShow.value = true;
    }
  };

  const handleConfirmEditPermission = async (permission: IPermissionType) => {
    try {
      editLoading.value = true;
      const { privilege, user, user_group } = permission;
      const query = {
        privilege,
        user,
        user_group,
        template_ids: props.configs.map((item) => item.id),
      };
      await batchEditTemplatePermission(bkBizId.value, query);
      Message({
        theme: 'success',
        message: t('配置文件权限批量修改成功'),
      });
      isEditPermissionShow.value = false;
      setTimeout(() => {
        emits('refresh');
      }, 300);
    } catch (error) {
      console.error(error);
    } finally {
      editLoading.value = false;
    }
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
</style>
