<script setup lang="ts">
  import { ref } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useConfigStore } from '../../../../../../../store/config'
  import SearchInput from '../../../../../../../components/search-input.vue';
  import CreateConfig from './create-config/index.vue'
  import TableWithTemplates from './tables/table-with-templates.vue';
  import TableWithPagination from './tables/table-with-pagination.vue';

  const configStore = useConfigStore()
  const { versionData } = storeToRefs(configStore)

  const props = defineProps<{
    bkBizId: string,
    appId: number,
  }>()

  const tableRef = ref()
  const searchStr = ref('')
  const useTemplate = ref(true)

  const refreshConfigList = () => {
    tableRef.value.refresh()
  }

  defineExpose({
    refreshConfigList
  })
</script>
<template>
  <section class="config-list-wrapper">
    <div class="operate-area">
      <CreateConfig
        v-if="versionData.status.publish_status === 'editing'"
        :bk-biz-id="props.bkBizId"
        :app-id="props.appId"
        @created="refreshConfigList"
        @imported="refreshConfigList" />
      <div class="groups-info" v-if="versionData.status.released_groups.length > 0">
        <div v-for="group in versionData.status.released_groups" class="group-item" :key="group.id">
          {{ group.name }}
        </div>
      </div>
      <SearchInput
        v-model="searchStr"
        class="config-search-input"
        placeholder="配置文件名/创建人/修改人"
        :width="280"
        @search="refreshConfigList" />
    </div>
    <section class="config-list-table">
      <TableWithTemplates v-if="useTemplate" ref="tableRef" :bk-biz-id="props.bkBizId" :app-id="props.appId" :search-str="searchStr" />
      <TableWithPagination v-else ref="tableRef" :bk-biz-id="props.bkBizId" :app-id="props.appId" :search-str="searchStr" />
    </section>
  </section>
</template>
<style lang="scss" scoped>
  .config-list-wrapper {
    position: relative;
    padding: 0 24px;
  }
  .operate-area {
    display: flex;
    align-items: center;
    padding: 16px 0;
    .groups-info {
      display: flex;
      align-items: center;
      .group-item {
        padding: 0 8px;
        line-height: 22px;
        color: #63656e;
        font-size: 12px;
        background: #f0f1f5;
        border-radius: 2px;
        &:not(:last-of-type) {
          margin-right: 8px;
        }
      }
    }
    .config-search-input {
      margin-left: auto;
    }
  }
</style>
