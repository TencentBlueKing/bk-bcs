<template>
  <CommonConfigTable
    v-model:selected-configs="selectedConfigs"
    ref="configTable"
    current-pkg="all"
    :show-cited-by-pkgs-col="true"
    :show-bound-by-apps-col="true"
    :get-config-list="getConfigList"
    :is-across-checked="acrossCheckedType.isAcrossChecked"
    :data-count="acrossCheckedType.dataCount"
    @send-across-checked-type="
      (checked, dataCount) => {
        acrossCheckedType.isAcrossChecked = checked;
        acrossCheckedType.dataCount = dataCount;
      }
    ">
    <template #tableOperations>
      <AddConfigs @refresh="refreshConfigList" />
      <BatchOperationButton
        :space-id="spaceId"
        :configs="selectedConfigs"
        :current-template-space="currentTemplateSpace"
        pkg-type="all"
        :is-across-checked="acrossCheckedType.isAcrossChecked"
        :data-count="acrossCheckedType.dataCount"
        @refresh="refreshConfigList" />
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
  import BatchOperationButton from '../operations/batch-operations/batch-operation-btn.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const templateStore = useTemplateStore();
  const { currentTemplateSpace } = storeToRefs(templateStore);

  const configTable = ref();
  const selectedConfigs = ref<ITemplateConfigItem[]>([]);
  const acrossCheckedType = ref<{ isAcrossChecked: boolean; dataCount: number }>({
    isAcrossChecked: false,
    dataCount: 0,
  });

  const getConfigList = (params: ICommonQuery) => {
    console.log('All Config List Loading');
    return getTemplatesBySpaceId(spaceId.value, currentTemplateSpace.value, params);
  };

  const refreshConfigList = (createConfig = false) => {
    if (createConfig) {
      configTable.value.refreshList(1, createConfig);
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
