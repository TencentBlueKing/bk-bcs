<template>
  <CommonConfigTable
    ref="configTable"
    v-model:selectedConfigs="selectedConfigs"
    :current-template-space="currentTemplateSpace"
    current-pkg="no_specified"
    :space-id="spaceId"
    :get-config-list="getConfigList"
    :show-delete-action="true"
    :is-across-checked="acrossCheckedType.isAcrossChecked"
    :data-count="acrossCheckedType.dataCount"
    @send-across-checked-type="
      (checked, dataCount) => {
        acrossCheckedType.isAcrossChecked = checked;
        acrossCheckedType.dataCount = dataCount;
      }
    ">
    <template #tableOperations>
      <BatchOperationButton
        :space-id="spaceId"
        :configs="selectedConfigs"
        :current-template-space="currentTemplateSpace"
        pkg-type="without"
        :is-across-checked="acrossCheckedType.isAcrossChecked"
        :data-count="acrossCheckedType.dataCount"
        @refresh="refreshConfigList"
        @deleted="handleConfigsDeleted" />
    </template>
  </CommonConfigTable>
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../../../../store/global';
  import useTemplateStore from '../../../../../../store/template';
  import { ICommonQuery } from '../../../../../../../types/index';
  import { ITemplateConfigItem } from '../../../../../../../types/template';
  import { getTemplatesWithNoSpecifiedPackage } from '../../../../../../api/template';
  import CommonConfigTable from './common-config-table.vue';
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
    const res = getTemplatesWithNoSpecifiedPackage(spaceId.value, currentTemplateSpace.value, params);
    return res;
  };

  const refreshConfigList = () => {
    configTable.value.refreshList();
    updateRefreshFlag();
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
  .opt-btn:not(:first-child) {
    margin-left: 16px;
  }
</style>
