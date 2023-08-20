<script setup lang="ts">
  import { ref } from 'vue'
  import { IScriptVersion } from '../../../../../types/script';
  import { IPagination } from '../../../../../types/index';
  import ScriptCited from '../list/script-cited.vue';

  const STATUS_MAP = {
    'not_deployed': '未上线',
    'deployed': '已上线',
    'shutdown': '已下线',
  }

  const props = defineProps<{
    scriptId: number;
    list: IScriptVersion[];
    pagination: IPagination;
  }>()

  const showCiteSlider = ref(false)
  const versionId = ref(0)

  const emits = defineEmits(['view', 'pageChange', 'pageLimitChange'])

  const handleOpenCitedSlider = (version: IScriptVersion) => {
    versionId.value = version.attachment.hook_id
    showCiteSlider.value = true
  }

</script>
<template>
  <bk-table
    :border="['outer']"
    :data="props.list"
    :pagination="pagination"
    @page-limit-change="emits('pageLimitChange', $event)"
    @page-change="emits('pageChange', $event)">
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
    <bk-table-column label="被引用" prop="spec.publish_num" :width="80">
      <template #default="{ row }">
        <template v-if="row.spec">
          <bk-button v-if="row.spec.publish_num > 0" text theme="primary" @click="handleOpenCitedSlider(row)">{{ row.spec.publish_num }}</bk-button>
          <span v-else>0</span>
        </template>
      </template>
    </bk-table-column>
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
</style>
