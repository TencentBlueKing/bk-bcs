<template>
  <div>
    <Row>
      <template #left>
        <DatePicker
          :placeholder="$t('generic.placeholder.searchDate')"
          class="bg-[#fff]"
          :model-value="zoneDate"
          :timezone="timezone"
          @update:modelValue="handleValueChange"
          @update:timezone="handleTimezoneChange"
        />
      </template>
      <template #right>
        <bcs-search-select
          clearable
          class="bg-[#fff] min-w-[360px] ml-[10px]"
          :data="searchSelectData"
          :show-condition="false"
          :show-popover-tag-change="false"
          :placeholder="$t('cluster.operateRecord.placeholder.searchRecord')"
          default-focus
          ref="searchSelect"
          v-model="searchSelectValue"
          @change="searchSelectChange"
          @clear="handleClear">
        </bcs-search-select>
      </template>
    </Row>
    <bcs-table
      v-bkloading="{ isLoading }"
      :data="list"
      :pagination="pagination"
      class="mt-[16px]"
      ref="tableRef"
      @page-change="pageChange"
      @page-limit-change="pageLimitChange"
      @header-dragend="handleHeaderDragend(tableRef)">
      <bk-table-column :label="$t('generic.label.time')" prop="createTime" width="200" fixed resizable>
        <template #default="{ row }">
          {{ formatTimeWithTimezone(row.createTime) }}
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('generic.label.resourceName')" prop="resourceName" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.resourceName || '--' }}
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('generic.label.resourceID')" prop="resourceID" width="180"></bk-table-column>
      <bk-table-column :label="$t('generic.label.status')" width="120">
        <template #default="{ row }">
          <StatusIcon
            :status="row.status"
            :status-color-map="statusColorMap"
            :status-text-map="statusTextMap"
            :pending="row.status === 'RUNNING'">
          </StatusIcon>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('projects.operateAudit.operator')" prop="opUser" width="120" show-overflow-tooltip>
        <template #default="{ row }">
          <bk-user-display-name :user-id="row.opUser"></bk-user-display-name>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('cluster.create.label.desc')" prop="message" show-overflow-tooltip min-width="160">
      </bk-table-column>
      <bcs-table-column :label="$t('generic.label.action')" width="100" fixed="right" :resizable="false">
        <template #default="{ row }">
          <template v-if="row.taskID">
            <bcs-button
              text
              v-if="row.taskID"
              @click="showTaskDetail(row)">{{ $t("generic.button.detail") }}</bcs-button>
            <bcs-button
              text
              v-if="row.taskID && !hideRetryStatus.includes(row.status) && row.allowRetry"
              class="ml-[8px]"
              @click="handleRetryTask(row)">
              {{ $t("generic.button.retry") }}
            </bcs-button>
          </template>
          <span v-else>--</span>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus
          :type="searchSelectValue?.length ? 'search-empty' : 'empty'"
          @clear="handleClear" />
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
          :height="'calc(100vh - 92px)'"
          :enable-statistics="true"
          :rolling-loading="false"
          :enable-rolling-loading="false"
          :show-step-retry-fn="handleShowStepRetry"
          :show-step-skip-fn="handleShowStepSkip"
          @refresh="showTaskDetail(curRow)"
          @auto-refresh="handleAutoRefresh"
          @download="getDownloadTaskRecords"
          @retry="handleRetry"
          @skip="handleSkip" />
      </template>
    </bcs-sideslider>
  </div>
</template>
<script lang="ts" setup>
import { computed, onBeforeMount, onBeforeUnmount, ref, watch } from 'vue';

import DatePicker from '@blueking/date-picker/vue2';
import TaskLog from '@blueking/task-log/vue2';

import { useTask } from '../cluster/use-cluster';

import '@blueking/task-log/vue2/vue2.css';
import '@blueking/date-picker/vue2/vue2.css';
import { clusterOperationLogs, clusterTaskRecords, taskLogsDownloadURL, taskRetry } from '@/api/modules/cluster-manager';
import { parseUrl } from '@/api/request';
import $bkMessage from '@/common/bkmagic';
import { formatTimeWithTimezone, getBrowserTimezoneId, getDateInTimezone, timezoneToUTC } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import Row from '@/components/layout/Row.vue';
import StatusIcon from '@/components/status-icon';
import useInterval from '@/composables/use-interval';
import { handleHeaderDragend, setTableColWByMemory } from '@/composables/use-table-col-w-memory';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';

interface Props {
  clusterId: string
}
const props = defineProps<Props>();

const hideRetryStatus = ref(['SUCCESS', 'RUNNING']);

const timeRange = ref<Date[]>([]);

const isLoading = ref(false);
const list = ref<any[]>([]);
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
  RUNNING: $i18n.t('generic.status.loading'),
  INITIALIZING: $i18n.t('generic.status.loading'),
  PART_FAILURE: $i18n.t('generic.status.halfSuccess'),
  NOTSTARTED: $i18n.t('generic.status.waiting'),
  TIMEOUT: $i18n.t('generic.status.terminate'),
};

// 状态
const status = Object.keys(statusTextMap).map(item => ({
  id: item,
  name: statusTextMap[item],
  text: statusTextMap[item],
  value: item,
}));
// searchSelect数据源配置
const searchSelectDataSource = [
  {
    name: $i18n.t('generic.label.resourceName'),
    id: 'resourceName',
    placeholder: $i18n.t('generic.placeholder.input'),
  },
  {
    name: $i18n.t('generic.label.resourceID'),
    id: 'resourceID',
    placeholder: $i18n.t('generic.placeholder.input'),
  },
  {
    name: $i18n.t('generic.label.status'),
    id: 'status',
    children: status,
  },
  {
    name: $i18n.t('projects.operateAudit.operator'),
    id: 'opUser',
    placeholder: $i18n.t('generic.placeholder.input'),
  },
];
// 检索参数
const searchParams = ref({});
function searchSelectChange(inputList) {
  const params = {};
  inputList.forEach((item) => {
    params[item.id] = item.values && item.values.length > 0 ? item?.values[0]?.id ?? '' : '';
  });
  searchParams.value = params;
}
function handleClear() {
  searchSelectValue.value = [];
  searchParams.value = {};
}
const searchSelectValue = ref<any[]>([]);
// 搜索项有值后就不展示了
const searchSelectData = ref<any[]>([]);
const searchSelect = ref();

// 获取集群操作日志
async function getOperationLogs() {
  if (!props.clusterId) return;

  isLoading.value = true;
  const { results, count } = await clusterOperationLogs({
    startTime: timezoneToUTC(timeRange.value[0], timezone.value),
    endTime: timezoneToUTC(timeRange.value[1], timezone.value),
    clusterID: props.clusterId,
    limit: pagination.value.limit,
    page: pagination.value.current,
    v2: true,
    taskIDNull: true,
    ...searchParams.value,
  }).catch(() => []);
  isLoading.value = false;
  list.value = results;
  pagination.value.count = count;
}

// 下载集群操作日志
async function getDownloadTaskRecords() {
  if (!props.clusterId) return;

  const { url } = parseUrl('get', taskLogsDownloadURL, {
    startTime: timezoneToUTC(timeRange.value[0], timezone.value),
    endTime: timezoneToUTC(timeRange.value[1], timezone.value),
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

const tableRef = ref();
watch(tableRef, () => setTableColWByMemory(tableRef.value));

// 显示任务详情
const curRow = ref<Record<string, any>>({});
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
  await handleGetStepGroupData(row);
  taskConfig.value.loading = false;
}
// 获取任务详情数据
async function handleGetStepGroupData(row) {
  if (!row.taskID) return;

  const { status, step } = await clusterTaskRecords({
    taskID: row.taskID,
  }).catch(() => ({ status: '', step: [] }));
  taskConfig.value.data = step;
  taskConfig.value.status = status;
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

const { skipTask, retryTask } = useTask();

// 任务重试
function handleShowStepRetry(item) {
  return item?.step?.status === 'FAILED' && item?.step?.allowRetry;
}
function handleRetry(data) {
  if (!data?.step?.allowRetry || !curRow.value.taskID) {
    $bkMessage({
      theme: 'warning',
      message: $i18n.t('cluster.title.allowRetry'),
    });
    return;
  }
  $bkInfo({
    type: 'warning',
    title: $i18n.t('cluster.title.retryTask'),
    clsName: 'custom-info-confirm default-info',
    subTitle: data?.step?.name,
    confirmFn: async () => {
      taskConfig.value.loading = true;
      const result = await retryTask(curRow.value.taskID);
      if (result) {
        handleGetStepGroupData(curRow.value);
        getOperationLogs();
      }
      taskConfig.value.loading = false;
    },
  });
}

// 任务跳过
function handleShowStepSkip(item) {
  return item?.step?.status === 'FAILED' && item?.step?.allowSkip;
}
function handleSkip(data) {
  if (!data?.step?.allowSkip || !curRow.value.taskID) {
    $bkMessage({
      theme: 'warning',
      message: $i18n.t('cluster.title.cantSkip'),
    });
    return;
  }
  $bkInfo({
    type: 'warning',
    title: $i18n.t('cluster.title.skipTask'),
    clsName: 'custom-info-confirm default-info',
    subTitle: data?.step?.name,
    confirmFn: async () => {
      taskConfig.value.loading = true;
      const result = await skipTask(curRow.value.taskID);
      if (result) {
        handleGetStepGroupData(curRow.value);
        getOperationLogs();
      }
      taskConfig.value.loading = false;
    },
  });
}

// 时间选择
const zoneDate = ref([]);
const timezone = ref(user.value?.time_zone || getBrowserTimezoneId());
function handleValueChange(v, info) {
  if (v.length === 0) {
    return;
  }
  const [start, end] = info;
  timeRange.value = [start.formatText, end.formatText];
  zoneDate.value = v;
}
function handleTimezoneChange(value) {
  zoneDate.value = [];
  timezone.value = value;
}

watch(
  () => [
    timeRange.value,
    searchParams.value,
  ],
  () => {
    pagination.value.current = 1;
    getOperationLogs();
    searchSelect.value?.hidePopper();
  },
);

watch(searchSelectValue, () => {
  const ids = searchSelectValue.value.map(item => item.id);
  searchSelectData.value = searchSelectDataSource.filter(item => !ids.includes(item.id));
}, { immediate: true, deep: true });

onBeforeMount(() => {
  // 初始化默认时间
  const end = getDateInTimezone(timezone.value);
  const start = getDateInTimezone(timezone.value);
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
