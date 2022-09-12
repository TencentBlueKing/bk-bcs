<template>
  <bk-table :data="data">
    <bk-table-column :label="$t('步骤')" prop="taskName" width="160">
      <template #default="{ row }">
        <span class="task-name">
          <span class="bcs-ellipsis">
            {{row.taskName}}
          </span>
          <i
            class="bcs-icon bcs-icon-fenxiang"
            v-if="row.params && row.params.taskUrl"
            @click="handleGotoSops(row)"></i>
        </span>
      </template>
    </bk-table-column>
    <bk-table-column :label="$t('状态')" prop="status" width="120">
      <template #default="{ row }">
        <LoadingIcon v-if="row.status === 'RUNNING'">
          <span class="bcs-ellipsis">{{ $t('运行中') }}</span>
        </LoadingIcon>
        <StatusIcon :status="row.status" :status-color-map="taskStatusColorMap" v-else>
          {{ taskStatusTextMap[row.status.toLowerCase()] }}
        </StatusIcon>
      </template>
    </bk-table-column>
    <bk-table-column :label="$t('执行时间')" width="220">
      <template #default="{ row }">
        {{row.start}}
      </template>
    </bk-table-column>
    <bk-table-column :label="$t('总耗时')">
      <template #default="{ row }">
        {{timeDelta(row.start, row.end) || '--'}}
      </template>
    </bk-table-column>
    <bk-table-column min-width="120" :label="$t('内容')" prop="message"></bk-table-column>
  </bk-table>
</template>
<script lang="ts">
import { defineComponent } from '@vue/composition-api';
import { taskStatusTextMap, taskStatusColorMap } from '@/common/constant';
import {  timeDelta } from '@/common/util';
import StatusIcon from '@/views/dashboard/common/status-icon';
import LoadingIcon from '@/components/loading-icon.vue';

export default defineComponent({
  name: 'TaskList',
  components: {
    StatusIcon,
    LoadingIcon,
  },
  props: {
    data: {
      type: Array,
      default: () => [],
    },
  },
  setup() {
    // 跳转标准运维
    const handleGotoSops = (row) => {
      window.open(row.params.taskUrl);
    };
    return {
      taskStatusColorMap,
      taskStatusTextMap,
      timeDelta,
      handleGotoSops,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .task-name {
    display: flex;
    align-items: center;
    justify-content: space-between;
    i {
        color: #3a84ff;
        cursor: pointer;
    }
}
</style>

