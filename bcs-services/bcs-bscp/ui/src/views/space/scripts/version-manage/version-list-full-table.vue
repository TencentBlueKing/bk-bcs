<script setup lang="ts">
  import { IScriptVersion } from '../../../../../types/script';
  import { IPagination } from '../../../../../types/index';

  const STATUS_MAP = {
    'not_deployed': '未上线',
    'deployed': '已上线',
    'shutdown': '已下线',
  }

  const props = defineProps<{
    list: IScriptVersion[];
    pagination: IPagination;
  }>()

  const emits = defineEmits(['view', 'pageChange', 'pageLimitChange'])
</script>
<template>
  <bk-table
    :border="['outer']"
    :data="props.list">
    <bk-table-column label="版本号" prop="spec.name" show-overflow-tooltip>
      <template #default="{ row }">
        <div v-if="row.spec" class="version-name" @click="emits('view', row)">{{ row.spec.name }}</div>
      </template>
    </bk-table-column>
    <bk-table-column label="版本说明">
      <template #default="{ row }">
        <span>{{ (row.spec && row.spec.memo) || '--' }}</span>
      </template>
    </bk-table-column>
    <bk-table-column label="被引用" prop="spec.publish_num" :width="80"></bk-table-column>
    <bk-table-column label="更新人" prop="revision.reviser"></bk-table-column>
    <bk-table-column label="更新时间" prop="revision.update_at"></bk-table-column>
    <bk-table-column label="状态">
      <template #default="{ row }">
        <span v-if="row.spec">
          <span :class="['status-dot', row.spec.state]"></span>
          {{ STATUS_MAP[row.spec.state as keyof typeof STATUS_MAP] }}
        </span>
      </template>
    </bk-table-column>
    <bk-table-column label="操作" width="240">
      <template #default="{ row }">
        <slot name="operations" :data="row"></slot>
      </template>
    </bk-table-column>
  </bk-table>
  <bk-pagination
    :model-value="props.pagination.current"
    class="table-list-pagination"
    location="left"
    :layout="['total', 'limit', 'list']"
    :count="props.pagination.count"
    :limit="props.pagination.limit"
    @change="emits('pageChange', $event)"
    @limit-change="emits('pageLimitChange', $event)" />
</template>
<style scoped lang="scss">
  .version-name {
    font-size: 12px;
    color: #3a84ff;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    cursor: pointer;
  }
  .status-dot {
    display: inline-block;
    margin-right: 6px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    border: 1px solid #c4c6cc;
    background: #f0f1f5;
    &.deployed {
      border: 1px solid #3fc06d;
      background: #e5f6ea;
    }
    &.not_deployed {
      border: 1px solid #ff9c01;
      background: #ffe8c3;
    }
  }
  .table-list-pagination {
    padding: 12px;
    background: #ffffff;
    border: 1px solid #dcdee5;
    border-top: none;
    border-radius: 0 0 2px 2px;
    :deep(.bk-pagination-list.is-last) {
      margin-left: auto;
    }
  }
</style>