<template>
  <CommonConfigTable
    ref="configTable"
    v-model:selectedConfigs="selectedConfigs"
    :current-template-space="currentTemplateSpace"
    current-pkg="no_specified"
    :space-id="spaceId"
    :get-config-list="getConfigList">
    <template #tableOperations>
      <BatchAddTo :configs="selectedConfigs" @refresh="refreshConfigList" />
      <DeleteConfigs
        :space-id="spaceId"
        :current-template-space="currentTemplateSpace"
        :configs="selectedConfigs"
        @deleted="handleConfigsDeleted" />
    </template>
    <template #columnOperations="{ config }">
      <bk-button theme="primary" text @click="handleAddToPkgsClick(config)">{{ t('添加至') }}</bk-button>
      <bk-button class="delete-btn" theme="primary" text @click="handleDeleteClick(config)">{{ t('删除') }}</bk-button>
    </template>
  </CommonConfigTable>
  <AddToDialog v-model:show="isAddToPkgsDialogShow" :value="selectedConfigs" @added="refreshConfigList" />
  <DeleteConfigDialog
    v-model:show="isDeleteConfigDialogShow"
    :configs="selectedConfigs"
    @deleted="handleConfigsDeleted" />
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../../../../store/global';
  import useTemplateStore from '../../../../../../store/template';
  import { ICommonQuery } from '../../../../../../../types/index';
  import { ITemplateConfigItem } from '../../../../../../../types/template';
  import { getTemplatesWithNoSpecifiedPackage } from '../../../../../../api/template';
  import CommonConfigTable from './common-config-table.vue';
  import BatchAddTo from '../operations/add-to-pkgs/add-to-button.vue';
  import AddToDialog from '../operations/add-to-pkgs/add-to-dialog.vue';
  import DeleteConfigs from '../operations/delete-configs/delete-button.vue';
  import DeleteConfigDialog from '../operations/delete-configs/delete-config-dialog.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const templateStore = useTemplateStore();
  const { currentTemplateSpace } = storeToRefs(templateStore);
  const { t } = useI18n();

  const configTable = ref();
  const selectedConfigs = ref<ITemplateConfigItem[]>([]);
  const isAddToPkgsDialogShow = ref(false);
  const isDeleteConfigDialogShow = ref(false);

  const getConfigList = (params: ICommonQuery) => {
    const res = getTemplatesWithNoSpecifiedPackage(spaceId.value, currentTemplateSpace.value, params);
    return res;
  };

  const refreshConfigList = () => {
    configTable.value.refreshList();
    updateRefreshFlag();
  };

  const handleAddToPkgsClick = (config: ITemplateConfigItem) => {
    isAddToPkgsDialogShow.value = true;
    selectedConfigs.value = [config];
  };

  const handleDeleteClick = async (config: ITemplateConfigItem) => {
    isDeleteConfigDialogShow.value = true;
    selectedConfigs.value = [config];
  };

  const handleConfigsDeleted = () => {
    configTable.value.refreshListAfterDeleted(selectedConfigs.value.length);
    selectedConfigs.value = [];
    updateRefreshFlag();
  };

  const updateRefreshFlag = () => {
    templateStore.$patch((state) => {
      state.needRefreshMenuFlag = true;
    });
  };
</script>
<style lang="scss" scoped>
  .delete-btn {
    margin-left: 16px;
  }
</style>
