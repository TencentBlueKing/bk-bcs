<template>
  <bk-table :border="['row']" :data="tableList" :row-class="getRowCls" @row-click="handleSelectVersion">
    <bk-table-column :label="t('版本号')" show-overflow-tooltip>
      <template #default="{ row, index }">
        <div class="version-name-wrapper">
          <div class="name">{{ row.name }}</div>
          <bk-tag v-if="index === 0" theme="success"> latest </bk-tag>
          <RightShape v-if="props.id === row.id" class="arrow-icon" />
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
    @change="emits('pageChange', $event)" />
</template>
<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { RightShape } from 'bkui-vue/lib/icon';
  import { IPagination } from '../../../../../../types/index';
  import { ITemplateVersionItem } from '../../../../../../types/template';

  const { t } = useI18n();
  const props = defineProps<{
    list: ITemplateVersionItem[];
    pagination: IPagination;
    id: number;
  }>();

  const emits = defineEmits(['pageChange', 'select']);

  const tableList = computed(() => {
    const simpleList = props.list.map((item) => {
      const { id, spec } = item;
      return { id, name: spec.revision_name };
    });
    if (props.id === 0) {
      simpleList.unshift({ id: 0, name: t('新建版本') });
    }
    return simpleList;
  });

  const getRowCls = (data: { id: number; name: string }) => {
    if (data.id === props.id) {
      return 'selected';
    }
    return '';
  };

  const handleSelectVersion = (event: Event | undefined, data: { id: number; name: string }) => {
    emits('select', data.id);
  };
</script>
<style scoped lang="scss">
  .version-name-wrapper {
    display: flex;
    align-items: center;
    position: relative;
    width: 100%;
    padding-right: 5px;
    .name {
      color: #3a84ff;
      margin-right: 8px;
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
