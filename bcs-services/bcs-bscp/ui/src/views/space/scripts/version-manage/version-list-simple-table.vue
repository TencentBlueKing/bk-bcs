<script setup lang="ts">
  import { RightShape } from 'bkui-vue/lib/icon';
  import { IScriptVersion } from '../../../../../types/script';
  import { IPagination } from '../../../../../types/index';

  const props = defineProps<{
    list: IScriptVersion[];
    pagination: IPagination;
    versionId: number;
  }>()

  const emits = defineEmits(['pageChange', 'select'])

  const getRowCls = (data: IScriptVersion) => {
    if (data.id === props.versionId) {
      return 'selected'
    }
    return ''
  }

  const handleSelectVersion = (event: Event|undefined, data: IScriptVersion) => {
    emits('select', data)
  }

</script>
<template>
  <bk-table
    :border="['row']"
    :data="props.list"
    :row-class="getRowCls"
    @row-click="handleSelectVersion">
    <bk-table-column label="版本号" show-overflow-tooltip>
      <template #default="{ row }">
        <div v-if="row.spec" class="version-name-wrapper">
          <div class="name">{{ row.spec.name }}</div>
          <RightShape v-if="props.versionId === row.id" class="arrow-icon" />
        </div>
      </template>
    </bk-table-column>
  </bk-table>
  <bk-pagination
    :model-value="props.pagination.current"
    class="table-compact-pagination"
    small
    align="right"
    :show-limit="false"
    :show-total-count="false"
    @change="emits('pageChange', $event)" />
</template>
<style scoped lang="scss">
  .version-name-wrapper {
    position: relative;
    width: 100%;
    padding-right: 5px;
    .arrow-icon {
      position: absolute;
      top: 15px;
      right: -10px;
      font-size: 12px;
      color: #3a84ff;
    } 
  }

  .bk-table {
    :deep(.bk-table-body) {
      tr {
        cursor: pointer;
        &.selected td {
          background: #e1ecff !important;
        }
      }
    }
  }
  .table-compact-pagination {
    margin-top: 16px;
  }
</style>