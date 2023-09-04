<script lang="ts" setup>
  import { ref } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../store/template'
  import { ICommonQuery } from '../../../../../../../types/index';
  import { ITemplateConfigItem } from '../../../../../../../types/template';
  import { getTemplatesByPackageId } from '../../../../../../api/template';
  import CommonConfigTable from './common-config-table.vue'
  import AddConfigs from '../operations/add-configs/add-button.vue'
  import BatchAddTo from '../operations/add-to-pkgs/add-to-button.vue'
  import BatchMoveOutFromPkg from '../operations/move-out-from-pkg/batch-move-out-button.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const templateStore = useTemplateStore()
  const { currentTemplateSpace, currentPkg } = storeToRefs(templateStore)

  const configTable = ref()
  const selectedConfigs = ref<ITemplateConfigItem[]>([])

  const getConfigList = (params: ICommonQuery) => {
    console.log('Package Config List Loading', currentTemplateSpace.value)
    return getTemplatesByPackageId(spaceId.value, currentTemplateSpace.value, <number>currentPkg.value, params)
  }

  const handleMovedOut = () => {
    configTable.value.refreshListAfterDeleted(selectedConfigs.value.length)
    selectedConfigs.value = []
    templateStore.$patch(state => {
      state.needRefreshMenuFlag = true
    })
  }

  const refreshConfigList = () => {
    configTable.value.refreshList()
  }

</script>
<template>
  <CommonConfigTable
    v-model:selectedConfigs="selectedConfigs"
    ref="configTable"
    :current-template-space="currentTemplateSpace"
    :key="currentPkg"
    :show-cited-by-pkgs-col="true"
    :show-bound-by-apps-col="true"
    :current-pkg="currentPkg"
    :get-config-list="getConfigList">
    <template #tableOperations>
      <AddConfigs :show-add-existing-config-option="true" @added="refreshConfigList" />
      <BatchAddTo :configs="selectedConfigs" @added="refreshConfigList" />
      <BatchMoveOutFromPkg :configs="selectedConfigs" :current-pkg="(currentPkg as number)" @movedOut="handleMovedOut" />
    </template>
  </CommonConfigTable>
</template>
<style lang="scss" scoped>

</style>
