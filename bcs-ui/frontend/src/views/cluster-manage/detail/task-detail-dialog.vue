<template>
  <bcs-dialog
    class="task-dialog"
    :value="value"
    width="70%"
    :show-footer="false"
    render-directive="if"
    @value-change="handleValueChange">
    <TaskLog
      :status="status"
      :title="title"
      :type="type"
      :data="data"
      :loading="loading"
      :height="height"
      :enable-statistics="enableStatistics"
      :statistics="statistics"
      :rolling-loading="rollingLoading"
      :enable-rolling-loading="enableRollingLoading"
      @refresh="refresh"
      @auto-refresh="autoRefresh"
      @download="download" />
  </bcs-dialog>
</template>
<script lang="ts" setup>
import { PropType } from 'vue';

import { IStatisticsItem, IStep, LogType } from '@blueking/task-log/src/lib/@types/index';
import TaskLog from '@blueking/task-log/vue2';

import '@blueking/task-log/vue2/vue2.css';

defineProps({
  value: {
    type: Boolean,
    default: false,
  },
  // 状态
  status: {
    type: String,
    default: '',
  },
  // 标题
  title: {
    type: String,
    default: '',
  },
  // 自定义模块
  // modules: {
  //   type: Array as PropType<IModule[]>,
  //   default: () => [],
  // },
  // 当前模块
  // activeModule: {
  //   type: [String, Number],
  //   default: BK_TASK_LOG_MODULE,
  // },
  // 日志类型
  type: {
    type: String as PropType<LogType>,
    default: 'default',
  },
  // 日志数据
  data: {
    type: [Array, Object] as PropType<IStep | IStep[]>,
    default: () => [],
  },
  // 日志加载loading
  loading: {
    type: Boolean,
    default: false,
  },
  // 高度
  height: {
    type: [String, Number],
    default: 560,
  },
  // 开启右侧统计状态栏
  enableStatistics: {
    type: Boolean,
    default: true,
  },
  // 自定义统计规则
  statistics: {
    type: Array as PropType<IStatisticsItem[]>,
    default: () => [],
  },
  // 滚动加载loading
  rollingLoading: {
    type: Boolean,
    default: false,
  },
  // 是否开启滚动加载
  enableRollingLoading: {
    type: [Boolean, String] as PropType<'manual' | boolean>,
    default: false,
  },
});

const emits = defineEmits(['download', 'refresh', 'auto-refresh', 'value-change']);

function handleValueChange(v) {
  emits('value-change', v);
}

function refresh() {
  emits('refresh');
}

function autoRefresh(v: boolean) {
  emits('auto-refresh', v);
}

function download() {
  emits('download');
}
</script>
<style lang="postcss" scoped>
.task-dialog {
  /deep/ .bk-dialog-tool {
      display: none;
  }
  /deep/ .bk-dialog-close {
      display: none;
  }
  /deep/ .bk-dialog-body {
      padding: 0;
      height: 100%;
  }
  /deep/ .bk-dialog-wrapper .bk-dialog-content {
      border-radius: 6px;
      height: 100%;
      width: 100% !important;
  }
}

</style>
