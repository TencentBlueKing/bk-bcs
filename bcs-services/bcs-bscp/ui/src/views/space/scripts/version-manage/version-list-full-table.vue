<template>
  <bk-table
    :border="['outer']"
    :data="props.list"
    :remote-pagination="true"
    :pagination="pagination"
    @page-limit-change="emits('pageLimitChange', $event)"
    @page-value-change="emits('pageChange', $event)"
  >
    <bk-table-column :label="t('版本号')" prop="spec.name" show-overflow-tooltip>
      <template #default="{ row }">
        <div v-if="row.hook_revision" class="version-name" @click="emits('view', row)">
          {{ row.hook_revision.spec.name }}
        </div>
      </template>
    </bk-table-column>
    <bk-table-column :label="t('版本说明')">
      <template #default="{ row }">
        <span>{{ (row.hook_revision && row.hook_revision.spec.memo) || '--' }}</span>
      </template>
    </bk-table-column>
    <bk-table-column :label="t('被引用')" prop="bound_num" :width="80">
      <template #default="{ row }">
        <bk-button v-if="row.bound_num > 0" text theme="primary" @click="handleOpenCitedSlider(row)">{{
          row.bound_num
        }}</bk-button>
        <span v-else>0</span>
      </template>
    </bk-table-column>
    <bk-table-column :label="t('更新人')" prop="hook_revision.revision.reviser"></bk-table-column>
    <bk-table-column :label="t('更新时间')" width="220">
      <template #default="{ row }">
        <span v-if="row.hook_revision">{{ datetimeFormat(row.hook_revision.revision.update_at) }}</span>
      </template>
    </bk-table-column>
    <bk-table-column :label="t('状态')">
      <template #default="{ row }">
        <span v-if="row.hook_revision">
          <span :class="['status-dot', row.hook_revision.spec.state]"></span>
          {{ STATUS_MAP[row.hook_revision.spec.state as keyof typeof STATUS_MAP] }}
        </span>
      </template>
    </bk-table-column>
    <bk-table-column :label="t('操作')" width="240">
      <template #default="{ row }">
        <slot name="operations" :data="row"></slot>
      </template>
    </bk-table-column>
    <template #empty>
      <tableEmpty :is-search-empty="isSearchEmpty" @clear="emits('clearStr')"></tableEmpty>
    </template>
  </bk-table>
  <ScriptCited v-model:show="showCiteSlider" :id="props.scriptId" :version-id="versionId" />
</template>
<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { IScriptVersionListItem } from '../../../../../types/script';
import { IPagination } from '../../../../../types/index';
import { datetimeFormat } from '../../../../utils/index';
import ScriptCited from '../list/script-cited.vue';
import tableEmpty from '../../../../components/table/table-empty.vue';

const { t } = useI18n();

const STATUS_MAP = computed(() => ({
  not_deployed: t('未上线'),
  deployed: t('已上线'),
  shutdown: t('已下线'),
}));

const props = defineProps<{
  scriptId: number;
  list: IScriptVersionListItem[];
  pagination: IPagination;
  isSearchEmpty: boolean;
}>();

const showCiteSlider = ref(false);
const versionId = ref(0);

const emits = defineEmits(['view', 'pageChange', 'pageLimitChange', 'clearStr']);

const handleOpenCitedSlider = (version: IScriptVersionListItem) => {
  versionId.value = version.hook_revision.id;
  showCiteSlider.value = true;
};
</script>
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
