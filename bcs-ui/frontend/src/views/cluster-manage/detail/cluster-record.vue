<template>
  <div>
    <Row>
      <template #left>
        <bcs-date-picker
          :shortcuts="shortcuts"
          type="datetimerange"
          shortcut-close
          :use-shortcut-text="false"
          :clearable="false"
          v-model="timeRange">
        </bcs-date-picker>
      </template>
      <!-- <template #right>
        <bcs-input
          class="min-w-[360px]"
          right-icon="bk-icon icon-search"
          clearable
          v-model.trim="searchValue" />
      </template> -->
    </Row>
    <bcs-table
      v-bkloading="{ isLoading }"
      :data="list"
      :pagination="pagination"
      class="mt-[16px]"
      @page-change="pageChange"
      @page-limit-change="pageLimitChange">
      <bk-table-column :label="$t('generic.label.time')" prop="createTime" width="170"></bk-table-column>
      <bk-table-column :label="$t('generic.label.resourceName')" prop="resourceName" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.resourceName || '--' }}
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('generic.label.resourceID')" prop="resourceID" width="180"></bk-table-column>
      <bk-table-column :label="$t('generic.label.status')" width="120">
        <template #default="{ row }">
          <StatusIcon :status="row.status" :status-color-map="statusColorMap" :pending="row.status === 'RUNNING'">
            {{ statusTextMap[row.status] || '--' }}
          </StatusIcon>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('projects.operateAudit.operator')" prop="opUser" width="120" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.opUser || '--' }}
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('cluster.create.label.desc')" prop="message" show-overflow-tooltip min-width="160">
      </bk-table-column>
      <bcs-table-column :label="$t('generic.label.action')" width="100">
        <template #default="{ row }">
          <bcs-button text v-if="row.taskID" @click="showTaskDetail(row)">{{ $t("generic.button.detail") }}</bcs-button>
          <bcs-button
            text
            v-if="row.taskID && row.status !== 'SUCCESS'"
            class="ml-[8px]"
            @click="handleRetryTask(row)">
            {{ $t("generic.button.retry") }}
          </bcs-button>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="searchValue = ''" />
      </template>
    </bcs-table>
    <bcs-sideslider
      :title="$t('cluster.title.opRecord')"
      :is-show.sync="isShowTaskDetail"
      :width="960"
      quick-close
      transfer
      @hidden="isShowTaskDetail = false">
      <template #content>
        <TaskLog
          :status="taskConfig.status"
          :title="taskConfig.title"
          type="multi-task"
          :data="taskConfig.data"
          :loading="taskConfig.loading"
          :height="'calc(100vh - 100px)'"
          :enable-statistics="true"
          :rolling-loading="false"
          :enable-rolling-loading="false"
          @refresh="showTaskDetail(curRow)"
          @auto-refresh="handleAutoRefresh"
          @download="getDownloadTaskRecords" />
      </template>
    </bcs-sideslider>
  </div>
</template>
<script lang="ts" setup>
import { computed, onBeforeMount, onBeforeUnmount, ref, watch } from 'vue';

import TaskLog from '@blueking/task-log/vue2';

import '@blueking/task-log/vue2/vue2.css';
import { clusterOperationLogs, clusterTaskRecords, taskLogsDownloadURL, taskRetry } from '@/api/modules/cluster-manager';
import { parseUrl } from '@/api/request';
import Row from '@/components/layout/Row.vue';
import StatusIcon from '@/components/status-icon';
import useDebouncedRef from '@/composables/use-debounce';
import useInterval from '@/composables/use-interval';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';

interface Props {
  clusterId: string
}
const props = defineProps<Props>();

const timeRange = ref<Date[]>([]);
const searchValue = useDebouncedRef('');
// 快捷时间配置
const shortcuts = ref([
  {
    text: $i18n.t('units.time.today'),
    value() {
      const end = new Date();
      const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
      return [start, end];
    },
  },
  {
    text: $i18n.t('units.time.lastDays'),
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
      return [start, end];
    },
  },
  {
    text: $i18n.t('units.time.last15Days'),
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 15);
      return [start, end];
    },
  },
  {
    text: $i18n.t('units.time.last30Days'),
    value() {
      const end = new Date();
      const start = new Date();
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
      return [start, end];
    },
  },
]);

const isLoading = ref(false);
const list = ref([]);
const pagination = ref({
  current: 1,
  limit: 10,
  count: 0,
  showTotalCount: true,
});
const user = computed(() => $store.state.user);
// 规则列表
const statusColorMap = ref({
  FAILURE: 'red',
  SUCCESS: 'green',
});
const statusTextMap = {
  FAILURE: $i18n.t('generic.status.failed'),
  SUCCESS: $i18n.t('generic.status.success'),
  RUNNING: $i18n.t('generic.status.running'),
};

// 获取集群操作日志
async function getOperationLogs() {
  if (!props.clusterId) return;

  isLoading.value = true;
  const { results, count } = await clusterOperationLogs({
    startTime: Math.ceil(new Date(timeRange.value[0]).getTime() / 1000),
    endTime: Math.ceil(new Date(timeRange.value[1]).getTime() / 1000),
    clusterID: props.clusterId,
    limit: pagination.value.limit,
    page: pagination.value.current,
    v2: true,
  }).catch(() => []);
  isLoading.value = false;
  list.value = results;
  pagination.value.count = count;
}

// 下载集群操作日志
async function getDownloadTaskRecords() {
  if (!props.clusterId) return;

  const { url } = parseUrl('get', taskLogsDownloadURL, {
    startTime: Math.ceil(new Date(timeRange.value[0]).getTime() / 1000),
    endTime: Math.ceil(new Date(timeRange.value[1]).getTime() / 1000),
    clusterID: props.clusterId,
    limit: pagination.value.limit,
    page: pagination.value.current,
    v2: true,
  });
  url && window.open(url);
}

function pageChange(page) {
  pagination.value.current = page;
  getOperationLogs();
};

function  pageLimitChange(limit) {
  pagination.value.current = 1;
  pagination.value.limit = limit;
  getOperationLogs();
};

// 显示任务详情
const curRow = ref({});
const isShowTaskDetail = ref(false);
const taskConfig = ref({
  loading: false,
  data: [],
  status: '',
  title: '',
});
async function showTaskDetail(row, loading = true) {
  curRow.value = row;
  isShowTaskDetail.value = true;
  taskConfig.value.loading = loading;
  taskConfig.value.title = row.message;
  const { status, step } = await clusterTaskRecords({
    taskID: row.taskID,
  }).catch(() => ({ status: '', step: [] }));
  taskConfig.value.data = step;
  taskConfig.value.status = status;
  taskConfig.value.loading = false;
}

// 任务重试
async function handleRetryTask(row) {
  const result = await taskRetry({
    $taskId: row.taskID,
    updater: user.value.username,
  }).catch(() => false);
  result && getOperationLogs();
};

// 自动刷新
const { start, stop } = useInterval(async () => {
  await showTaskDetail(curRow.value, false);
}, 5000);
function handleAutoRefresh(v: boolean) {
  if (v) {
    start();
  } else {
    stop();
  }
}

watch(
  () => [
    timeRange.value,
    searchValue.value,
  ],
  () => {
    pagination.value.current = 1;
    getOperationLogs();
  },
);

onBeforeMount(() => {
  // 初始化默认时间
  const end = new Date();
  const start = new Date();
  start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
  timeRange.value = [
    start,
    end,
  ];
  getOperationLogs();
});

onBeforeUnmount(() => {
  stop();
});
</script>
