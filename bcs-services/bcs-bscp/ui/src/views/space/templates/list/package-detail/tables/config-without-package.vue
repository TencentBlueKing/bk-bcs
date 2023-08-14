<script lang="ts" setup>
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../store/template'
  import { ICommonQuery } from '../../../../../../../types/index';
  import { getTemplatesWithNoSpecifiedPackage } from '../../../../../../api/template';
  import CommonConfigTable from './common-config-table.vue'
  import BatchAddTo from '../operations/batch-add-to/add-to-button.vue'
  import BatchDelete from '../operations/batch-delete/delete-button.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const { currentTemplateSpace } = storeToRefs(useTemplateStore())

  const getConfigList = (params: ICommonQuery) => {
    console.log('Config Without Package List Loading')
    return getTemplatesWithNoSpecifiedPackage(spaceId.value, currentTemplateSpace.value, params)
  }

</script>
<template>
  <CommonConfigTable current-pkg="no_specified"  :get-config-list="getConfigList">
    <template #tableOperations>
      <BatchAddTo />
      <BatchDelete />
    </template>
  </CommonConfigTable>
</template>
<style lang="scss" scoped>

</style>
