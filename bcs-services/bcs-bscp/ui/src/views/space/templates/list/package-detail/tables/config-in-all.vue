<script lang="ts" setup>
  import { storeToRefs } from 'pinia'
  import { ref } from 'vue'
  import { useGlobalStore } from '../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../store/template'
  import { ICommonQuery } from '../../../../../../../types/index';
  import { ITemplateConfigItem } from '../../../../../../../types/template';
  import { getTemplatesBySpaceId } from '../../../../../../api/template';
  import CommonConfigTable from './common-config-table.vue'
  import AddConfigs from '../operations/add-configs/add-button.vue'
  import BatchAddTo from '../operations/add-to-pkgs/add-to-button.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const { currentTemplateSpace, currentPkg } = storeToRefs(useTemplateStore())

  const configTable = ref()
  const selectedConfigs = ref<ITemplateConfigItem[]>([])

  const getConfigList = (params: ICommonQuery) => {
    console.log('All Config List Loading')
    return getTemplatesBySpaceId(spaceId.value, currentTemplateSpace.value, params)
  }

  const handleAdded = () => {
    configTable.value.refreshList()
  }

</script>
<template>
  <CommonConfigTable
    ref="configTable"
    current-pkg="all"
    :current-template-space="currentTemplateSpace"
    :get-config-list="getConfigList"
    v-model:selectedConfigs="selectedConfigs">
    <template #tableOperations>
      <AddConfigs @added="handleAdded" />
      <BatchAddTo :configs="selectedConfigs" />
    </template>
    <template #columns>
      <bk-table-column label="所在套餐"></bk-table-column>
      <bk-table-column label="被引用"></bk-table-column>
    </template>
  </CommonConfigTable>
</template>
<style lang="scss" scoped>

</style>
