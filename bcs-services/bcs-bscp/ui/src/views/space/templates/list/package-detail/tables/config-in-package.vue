<template>
  <CommonConfigTable
    v-model:selected-configs="selectedConfigs"
    ref="configTable"
    :key="currentPkg"
    :show-cited-by-pkgs-col="true"
    :show-bound-by-apps-col="true"
    :current-pkg="currentPkg"
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
      <AddConfigs :show-add-existing-config-option="true" @refresh="refreshConfigList" />
      <CountTips
        :max="spaceFeatureFlags.RESOURCE_LIMIT.TmplSetTmplCnt"
        :current="countOfTemplatesForCurrentPackage"
        :is-temp="true"
        :is-file-type="false" />
      <BatchOperationButton
        :space-id="spaceId"
        :configs="selectedConfigs"
        :current-template-space="currentTemplateSpace"
        pkg-type="pkg"
        :current-pkg="currentPkg as number"
        :is-across-checked="acrossCheckedType.isAcrossChecked"
        :data-count="acrossCheckedType.dataCount"
        @refresh="refreshConfigList"
        @moved-out="handleMovedOut" />
    </template>
  </CommonConfigTable>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../../../../store/global';
  import useTemplateStore from '../../../../../../store/template';
  import { ICommonQuery } from '../../../../../../../types/index';
  import { ITemplateConfigItem } from '../../../../../../../types/template';
  import { getTemplatesByPackageId } from '../../../../../../api/template';
  import CommonConfigTable from './common-config-table.vue';
  import AddConfigs from '../operations/add-configs/add-button.vue';
  import BatchOperationButton from '../operations/batch-operations/batch-operation-btn.vue';
  import CountTips from '../../../../service/detail/config/components/count-tips.vue';

  const { spaceId, spaceFeatureFlags } = storeToRefs(useGlobalStore());
  const templateStore = useTemplateStore();
  const { currentTemplateSpace, currentPkg, countOfTemplatesForCurrentPackage } = storeToRefs(templateStore);

  const configTable = ref();
  const selectedConfigs = ref<ITemplateConfigItem[]>([]);
  const acrossCheckedType = ref<{ isAcrossChecked: boolean; dataCount: number }>({
    isAcrossChecked: false,
    dataCount: 0,
  });

  const getConfigList = (params: ICommonQuery) => {
    console.log('Package Config List Loading', currentTemplateSpace.value);
    return getTemplatesByPackageId(spaceId.value, currentTemplateSpace.value, currentPkg.value as number, params);
  };

  const handleMovedOut = () => {
    configTable.value.refreshListAfterDeleted(selectedConfigs.value.length);
    selectedConfigs.value = [];
    updateRefreshFlag();
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
  watch(acrossCheckedType.value, () => {
    console.log(acrossCheckedType.value, '+++++++++++++++++++++++');
  });
</script>
<style lang="scss" scoped></style>
