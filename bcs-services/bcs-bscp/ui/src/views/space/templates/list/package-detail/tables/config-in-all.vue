<template>
  <CommonConfigTable
    v-model:selectedConfigs="selectedConfigs"
    ref="configTable"
    current-pkg="all"
    :show-cited-by-pkgs-col="true"
    :show-bound-by-apps-col="true"
    :current-template-space="currentTemplateSpace"
    :get-config-list="getConfigList"
  >
    <template #tableOperations>
      <AddConfigs @refresh="refreshConfigList" />
      <BatchAddTo :configs="selectedConfigs" @refresh="refreshConfigList" />
    </template>
  </CommonConfigTable>
</template>
<script lang="ts" setup>
import { storeToRefs } from 'pinia';
import { ref } from 'vue';
import useGlobalStore from '../../../../../../store/global';
import useTemplateStore from '../../../../../../store/template';
import { ICommonQuery } from '../../../../../../../types/index';
import { ITemplateConfigItem } from '../../../../../../../types/template';
import { getTemplatesBySpaceId } from '../../../../../../api/template';
import CommonConfigTable from './common-config-table.vue';
import AddConfigs from '../operations/add-configs/add-button.vue';
import BatchAddTo from '../operations/add-to-pkgs/add-to-button.vue';

const { spaceId } = storeToRefs(useGlobalStore());
const templateStore = useTemplateStore();
const { currentTemplateSpace } = storeToRefs(templateStore);

const configTable = ref();
const selectedConfigs = ref<ITemplateConfigItem[]>([]);

const getConfigList = (params: ICommonQuery) => {
  console.log('All Config List Loading');
  return getTemplatesBySpaceId(spaceId.value, currentTemplateSpace.value, params);
};

const refreshConfigList = (isBatchUpload = false) => {
  if (isBatchUpload) {
    configTable.value.refreshList(1, isBatchUpload);
  } else {
    configTable.value.refreshList();
  }
  updateRefreshFlag();
};

const updateRefreshFlag = () => {
  templateStore.$patch((state) => {
    state.needRefreshMenuFlag = true;
  });
};
</script>
<style lang="scss" scoped></style>
