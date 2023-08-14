<script lang="ts" setup>
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../store/template'
  import { ICommonQuery } from '../../../../../../../types/index';
  import { getTemplatesByPackageId } from '../../../../../../api/template';
  import CommonConfigTable from './common-config-table.vue'
  import AddConfigs from '../operations/add-configs/add-button.vue'
  import BatchAddTo from '../operations/batch-add-to/add-to-button.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const { currentTemplateSpace, currentPkg } = storeToRefs(useTemplateStore())

  const getConfigList = (params: ICommonQuery) => {
    console.log('Package Config List Loading')
    return getTemplatesByPackageId(spaceId.value, currentTemplateSpace.value, <number>currentPkg.value, params)
  }

</script>
<template>
  <CommonConfigTable :current-pkg="currentPkg" :get-config-list="getConfigList">
    <template #tableOperations>
      <AddConfigs :show-add-existing-config-option="true"/>
      <BatchAddTo />
    </template>
    <template #columns>
      <bk-table-column label="所在套餐"></bk-table-column>
      <bk-table-column label="被引用"></bk-table-column>
    </template>
  </CommonConfigTable>
</template>
<style lang="scss" scoped>

</style>
