<script setup lang="ts">
  import { ref } from 'vue'
  import { IScriptVersionListItem } from '../../../../../types/script';
  import { IPagination } from '../../../../../types/index';
  import { datetimeFormat } from '../../../../utils/index';
  import ScriptCited from '../list/script-cited.vue';

  const STATUS_MAP = {
    'not_deployed': '未上线',
    'deployed': '已上线',
    'shutdown': '已下线',
  }

  const props = defineProps<{
    scriptId: number;
    list: IScriptVersionListItem[];
    pagination: IPagination;
  }>()

  const showCiteSlider = ref(false)
  const versionId = ref(0)

  const emits = defineEmits(['view', 'pageChange', 'pageLimitChange'])

  const handleOpenCitedSlider = (version: IScriptVersionListItem) => {
    versionId.value = version.hook_revision.id
    showCiteSlider.value = true
  }

</script>
<template>
  <bk-table
    :border="['outer']"
    :data="props.list">
    <bk-table-column label="版本号" prop="spec.name" show-overflow-tooltip>
      <template #default="{ row }">
        <div v-if="row.hook_revision" class="version-name" @click="emits('view', row)">{{ row.hook_revision.spec.name }}</div>
      </template>
    </bk-table-column>
    <bk-table-column label="版本说明">
      <template #default="{ row }">
        <span>{{ (row.hook_revision && row.hook_revision.spec.memo) || '--' }}</span>
      </template>
    </bk-table-column>
    <bk-table-column label="被引用" prop="bound_num" :width="80">
      <template #default="{ row }">
        <bk-button v-if="row.bound_num > 0" text theme="primary" @click="handleOpenCitedSlider(row)">{{ row.bound_num }}</bk-button>
        <span v-else>0</span>
      </template>
    </bk-table-column>
    <bk-table-column label="更新人" prop="hook_revision.revision.reviser"></bk-table-column>
    <bk-table-column label="更新时间" width="220">
      <template #default="{ row }">
        <span v-if="row.hook_revision">{{ datetimeFormat(row.hook_revision.revision.update_at) }}</span>
      </template>
    </bk-table-column>
    <bk-table-column label="状态">
      <template #default="{ row }">
        <span v-if="row.hook_revision">
          <span :class="['status-dot', row.hook_revision.spec.state]"></span>
          {{ STATUS_MAP[row.hook_revision.spec.state as keyof typeof STATUS_MAP] }}
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
  <ScriptCited v-model:show="showCiteSlider" :id="props.scriptId" :version-id="versionId" />
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
