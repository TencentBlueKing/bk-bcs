<template>
  <bk-table :border="['row']" :data="props.list" :row-class="getRowCls" @row-click="handleSelectVersion">
    <bk-table-column :label="t('版本号')" show-overflow-tooltip>
      <template #default="{ row }">
        <div v-if="row.hook_revision" class="version-name-wrapper">
          <span :class="['status-dot', row.hook_revision.spec.state]"></span>
          <div class="name">{{ row.hook_revision.spec.name }}</div>
        </div>
      </template>
    </bk-table-column>
  </bk-table>
  <bk-pagination
    :model-value="props.pagination.current"
    class="table-compact-pagination"
    small
    align="right"
    :count="props.pagination.count"
    :limit="props.pagination.limit"
    :show-limit="false"
    :show-total-count="false"
    @change="emits('pageChange', $event)"/>
</template>
<script setup lang="ts">
import { IScriptVersionListItem } from '../../../../../types/script';
import { IPagination } from '../../../../../types/index';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const props = defineProps<{
  list: IScriptVersionListItem[];
  pagination: IPagination;
  versionId: number;
}>();

const emits = defineEmits(['pageChange', 'select']);

const getRowCls = (data: IScriptVersionListItem) => {
  if (data.hook_revision.id === props.versionId) {
    return 'selected';
  }
  return '';
};

const handleSelectVersion = (event: Event | undefined, data: IScriptVersionListItem) => {
  emits('select', data.hook_revision);
};
</script>
<style scoped lang="scss">
.version-name-wrapper {
  position: relative;
  width: 100%;
  padding-right: 5px;
  .status-dot {
    position: absolute;
    top: 17px;
    left: -12px;
    display: inline-block;
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
