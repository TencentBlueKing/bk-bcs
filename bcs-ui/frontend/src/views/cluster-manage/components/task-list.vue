<template>
  <div class="flex h-[calc(100vh-108px)]">
    <bk-table
      :data="data"
      :row-class-name="rowClass"
      @row-click="handleRowClick">
      <bk-table-column :label="$t('步骤')" prop="taskName" min-width="155">
        <template #default="{ row }">
          <span class="task-name">
            <span class="bcs-ellipsis">
              {{row.taskName}}
            </span>
            <i
              class="bcs-icon bcs-icon-fenxiang"
              v-if="row.params && row.params.showUrl === 'true' && row.params.taskUrl"
              @click="handleGotoSops(row)"></i>
          </span>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('状态')" prop="status" width="90">
        <template #default="{ row }">
          <LoadingIcon v-if="row.status === 'RUNNING'">
            <span class="bcs-ellipsis">{{ $t('运行中') }}</span>
          </LoadingIcon>
          <StatusIcon type="result" :status="row.status" :status-color-map="taskStatusColorMap" v-else>
            {{ taskStatusTextMap[row.status.toLowerCase()] }}
          </StatusIcon>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('执行时间')" width="165">
        <template #default="{ row }">
          <template v-if="row.status.toLowerCase() === 'notstarted'">
            --
          </template>
          <template v-else>
            {{timeFormat(row.start)}}
          </template>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('耗时')">
        <template #default="{ row }">
          <template v-if="row.status === 'RUNNING'">
            {{timeDelta(row.start, new Date()) || '--'}}
          </template>
          <template v-else-if="row.start && row.end">
            {{timeDelta(row.start, row.end) || '1s'}}
          </template>
          <template v-else>
            --
          </template>
        </template>
      </bk-table-column>
    </bk-table>
    <div class="bg-[#F5F7FA] p-[16px] w-[420px] text-[#313238] text-[12px] task-message">
      {{ curTaskRow.message }}
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref, computed } from 'vue';
import { timeDelta, timeFormat } from '@/common/util';
import StatusIcon from '@/components/status-icon';
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
  setup(props) {
    const taskStatusTextMap = {
      initialzing: window.i18n.t('初始化中'),
      running: window.i18n.t('运行中'),
      success: window.i18n.t('成功'),
      failure: window.i18n.t('失败'),
      timeout: window.i18n.t('超时'),
      notstarted: window.i18n.t('未执行'),
    };
    const taskStatusColorMap = {
      initialzing: 'blue',
      running: 'blue',
      success: 'green',
      failure: 'red',
      timeout: 'red',
      notstarted: 'gray',
    };
    const activeIndex = ref(0);
    const curTaskRow = computed<Record<string, any>>(() => props.data?.[activeIndex.value] || {});

    // 跳转标准运维
    const handleGotoSops = (row) => {
      window.open(row.params.taskUrl);
    };
    const handleRowClick = (row, event, column, rowIndex) => {
      activeIndex.value = rowIndex;
    };
    const rowClass = ({ rowIndex }) => (activeIndex.value === rowIndex ? 'active-row' : 'normal-row');

    return {
      taskStatusColorMap,
      taskStatusTextMap,
      timeFormat,
      timeDelta,
      curTaskRow,
      handleGotoSops,
      handleRowClick,
      rowClass,
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
.task-message {
    border: 1px solid #DCDEE5;
    border-left: none;
    word-break: break-word;
}
>>> .bk-table .bk-table-header-wrapper {
    border-top: none;
}
>>> .active-row {
  background-color: #E1ECFF;
  cursor: pointer;
}
>>> .normal-row td {
  background-color: #fff !important;
  cursor: pointer;
}
</style>
