<template>
  <div v-bkloading="{ isLoading: loading }">
    <Header>
      <div class="flex items-center">
        {{ $t('日志采集') }}
        <ClusterSelect
          size="small"
          class="small-select"
          cluster-type="all"
          v-model="clusterId" />
        <!-- 组件版本和更新 -->
        <bcs-badge
          dot
          theme="danger"
          :visible="showUpdateBtn"
          v-if="!['', 'failed-upgrade', 'failed-install', 'failed'].includes(onsData.status || '')">
          <span class="flex items-center h-[24px] bg-[#EAEBF0] px-[8px] ml-[8px] text-[12px] rounded-sm">
            {{ onsData.currentVersion || '--' }}
            <LoadingIcon
              class="ml-[8px]"
              v-if="runningStatus.includes(onsData.status || '')" />
            <bcs-button
              text
              class="text-[12px] ml-[8px]"
              v-else-if="showUpdateBtn"
              @click="handleShowUpdateDialog">
              {{ $t('更新') }}
            </bcs-button>
            <bcs-popover
              class="ml-[4px]"
              theme="light"
              v-if="onsData.message &&
                [
                  'failed',
                  'failed-install',
                  'failed-upgrade',
                  'failed-rollback',
                  'failed-uninstall'
                ].includes(onsData.status || '')">
              <span class="text-[#FF5656] relative top-[-1px]">
                <i class="bcs-icon bcs-icon-info-circle-shape"></i>
              </span>
              <template #content>
                <div>{{ onsData.message }}</div>
              </template>
            </bcs-popover>
          </span>
        </bcs-badge>
      </div>
      <bcs-dialog
        :title="$t('组件更新')"
        header-position="left"
        :ok-text="$t('立即更新')"
        :loading="updateLoading"
        v-model="showUpdateDialog"
        @confirm="confirmUpdate">
        <bk-form :label-width="120">
          <bk-form-item :label="$t('当前版本号')">
            <div class="flex items-end">
              {{ onsData.currentVersion }}
            </div>
          </bk-form-item>
          <bk-form-item class="!mt-[0px]" :label="$t('最新版本号')">
            <div class="flex items-end">
              {{ onsData.version }}
            </div>
          </bk-form-item>
          <bk-form-item class="!mt-[0px]" :label="$t('最新版本描述')">
            <div class="flex items-end">
              {{ onsData.description }}
            </div>
          </bk-form-item>
        </bk-form>
      </bcs-dialog>
    </Header>
    <bcs-exception type="empty" v-if="!onsData.status && !loading">
      <div>{{ isSharedOrVirtual ? $t('当前集群日志采集组件未启用，请联系集群管理员处理') : $t('当前集群日志采集组件未启用') }}</div>
      <!-- shared 和 virtual集群不支持启用 -->
      <bcs-button
        theme="primary"
        class="w-[88px] mt-[16px]"
        v-if="!isSharedOrVirtual"
        @click="confirmUpdate">
        {{ $t('启用') }}
      </bcs-button>
    </bcs-exception>
    <bcs-exception
      type="empty"
      v-else-if="['failed-upgrade', 'failed-install','failed'].includes(onsData.status || '') && !loading">
      <div>{{ $t('当前集群日志采集组件安装失败') }}</div>
      <div class="text-[#979BA5] mt-[16px]" v-if="onsData.message">{{ onsData.message }}</div>
      <bcs-button theme="primary" class="w-[88px] mt-[16px]" @click="confirmUpdate">{{ $t('重新安装') }}</bcs-button>
    </bcs-exception>
    <div class="pb-[16px] overflow-hidden" :key="clusterId" v-else>
      <bcs-alert type="info" closable>
        <template #title>
          <div class="flex items-center h-[16px]">
            {{ $t('支持容器标准输出日志和文件路径日志的采集设置。日志采集服务已升级，旧版本规则建议「生成新规则」并及时清理。') }}
            <bcs-button text class="text-[12px]" @click="openLink(PROJECT_CONFIG.rule)">{{ $t('了解更多') }}</bcs-button>
          </div>
        </template>
      </bcs-alert>
      <Row class="mt-[16px] px-[24px]">
        <template #left>
          <bcs-button theme="primary" icon="plus" @click="handleCreateRule">{{ $t('新建规则') }}</bcs-button>
        </template>
        <template #right>
          <bcs-input
            class="w-[320px]"
            :placeholder="$t('规则名称/命名空间/更新人')"
            right-icon="bk-icon icon-search"
            v-model="searchValue"
            clearable>
          </bcs-input>
        </template>
      </Row>
      <div class="mt-[16px] px-[24px]" v-bkloading="{ isLoading: ruleListLoading }">
        <bcs-table
          :data="curPageData"
          :pagination="pagination"
          @page-change="pageChange"
          @page-limit-change="pageSizeChange"
          v-if="!showEditStatus">
          <bcs-table-column :label="$t('规则名称')" prop="name" show-overflow-tooltip sortable>
            <template #default="{ row }">
              <div class="flex">
                <!-- 是否旧规则 -->
                <span
                  class="inline-flex items-center justify-center w-[24px] relative
                  h-[24px] bg-[#F0F1F5] rounded-sm text-[#979BA5] mr-[8px]"
                  v-bk-tooltips="{
                    content: $t('已生成新规则: <br/> {0}<br/>此规则自动清理期限:<br/>{1}', [
                      row.new_rule_name,
                      moment(row.updated_at).add(30, 'days').format('YYYY-MM-DD')
                    ]),
                    disabled: !row.new_rule_id
                  }"
                  v-if="row.old">
                  {{ $t('旧') }}
                  <!-- 是否转换过 -->
                  <i
                    class="absolute bottom-[0px] right-[-6px] text-[#87cfab] bcs-icon bcs-icon-check-circle-shape"
                    v-if="row.new_rule_id">
                  </i>
                </span>
                <bcs-button
                  class="flex-1"
                  :disabled="['PENDING', 'RUNNING'].includes(row.status)"
                  text
                  @click="handleGotoDetail(row)">
                  <span class="bcs-ellipsis">{{ row.name }}</span>
                </bcs-button>
              </div>
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('命名空间')" show-overflow-tooltip>
            <template #default="{ row }">
              {{ getNs(row) }}
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('备注')" show-overflow-tooltip prop="description">
            <template #default="{ row }">
              {{ row.description || '--' }}
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('更新人')" prop="updator"></bcs-table-column>
          <bcs-table-column :label="$t('更新时间')" sortable prop="updated_at" width="180"></bcs-table-column>
          <bcs-table-column :label="$t('状态')" width="120">
            <template #default="{ row }">
              <StatusIcon
                :status-color-map="statusColorMap"
                :status-text-map="statusTextMap"
                :status="row.status"
                :pending="['PENDING', 'RUNNING'].includes(row.status)">
                <span
                  v-bk-tooltips="{
                    content: row.message,
                    disabled: !row.message || row.status !== 'FAILED',
                    theme: 'bcs-tippy'
                  }"
                  :class="row.message && row.status === 'FAILED' ? 'border-dashed border-0 border-b' : ''">
                  {{statusTextMap[row.status] || $t('未知状态')}}
                </span>
              </StatusIcon>
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('操作')" width="280">
            <template #default="{ row }">
              <template v-if="row.status === 'TERMINATED'">
                <bcs-button text class="text-[12px] mr-[10px]" @click="handleToggleRule(row)">
                  {{ $t('启用') }}
                </bcs-button>
                <bcs-button text class="text-[12px] mr-[10px]" @click="handleDeleteRule(row)">
                  {{ $t('删除') }}
                </bcs-button>
              </template>
              <template v-if="row.status === 'SUCCESS'">
                <bcs-button
                  text
                  class="text-[12px] mr-[10px]"
                  v-if="row.entrypoint.file_log_url && row.rule.config.paths.length"
                  @click="openLink(row.entrypoint.file_log_url)">
                  {{ $t('文件日志') }}
                </bcs-button>
                <bcs-button
                  text
                  class="text-[12px] mr-[10px]"
                  v-if="row.entrypoint.std_log_url && row.rule.config.enable_stdout"
                  @click="openLink(row.entrypoint.std_log_url)">
                  {{ $t('标准输出日志') }}
                </bcs-button>
              </template>
              <bcs-button
                text
                class="text-[12px] mr-[10px]"
                :disabled="row.status !== 'FAILED'"
                @click="handleRetryRule(row)"
                v-if="row.status === 'FAILED'">
                {{ $t('重试') }}
              </bcs-button>
              <!-- 旧规则不允许编辑 -->
              <bcs-button
                text
                class="text-[12px] mr-[10px]"
                v-if="row.old"
                @click="handleCreateNewRule(row)">
                {{ $t('生成新规则') }}
              </bcs-button>
              <bcs-button
                text
                class="text-[12px] mr-[10px]"
                :disabled="['PENDING', 'RUNNING'].includes(row.status)"
                v-else-if="row.status !== 'TERMINATED'"
                @click="handleEditRule(row)">
                {{ $t('编辑') }}
              </bcs-button>

              <PopoverSelector
                :disabled="['PENDING', 'RUNNING'].includes(row.status)"
                v-if="row.status !== 'TERMINATED'">
                <span :class="['bcs-icon-more-btn', { disabled: ['PENDING', 'RUNNING'].includes(row.status) }]">
                  <i class="bcs-icon bcs-icon-more text-[18px] relative top-[1px]"></i>
                </span>
                <template #content>
                  <li
                    class="bcs-dropdown-item"
                    v-if="['SUCCESS', 'TERMINATED'].includes(row.status) && !row.old"
                    @click="handleToggleRule(row)">
                    {{ row.status === 'TERMINATED' ? $t('启用') : $t('停用') }}
                  </li>
                  <li class="bcs-dropdown-item" @click="handleDeleteRule(row)">{{ $t('删除') }}</li>
                </template>
              </PopoverSelector>
            </template>
          </bcs-table-column>
          <template #empty>
            <BcsEmptyTableStatus :type="searchValue.length ? 'search-empty' : 'empty'" @clear="searchValue = ''" />
          </template>
        </bcs-table>
        <EditLogCollector
          :cluster-id="clusterId"
          :data="tableDataMatchSearch"
          :status-color-map="statusColorMap"
          :status-text-map="statusTextMap"
          :id="curRow ? curRow.id : ''"
          :edit="edit"
          v-bkloading="{ isLoading: ruleListLoading }"
          ref="editLogRef"
          @show-list="handleShowList"
          @delete="handleDeleteRule"
          @refresh="handleGetLogCollectorRules"
          @toggle-rule="handleToggleRule"
          @update-create-id="handleUpdateCreatingID"
          v-else />
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onBeforeMount, ref, watch } from 'vue';
import Header from '@/components/layout/Header.vue';
import Row from '@/components/layout/Row.vue';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import StatusIcon from '@/components/status-icon';
import PopoverSelector from '@/components/popover-selector.vue';
import EditLogCollector from './edit-log-collector.vue';
import useLog, { IRuleData } from './use-log';
import { useCluster } from '@/composables/use-app';
import { LOG_COLLECTOR } from '@/common/constant';
import useInterval from '@/composables/use-interval';
import usePageConf from '@/composables/use-page';
import useTableSearch from '@/composables/use-search';
import $i18n from '@/i18n/i18n-setup';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import $bkMessage from '@/common/bkmagic';
import $router from '@/router';
import moment from 'moment';

const { curClusterId, clusterList } = useCluster();

const clusterId = ref(curClusterId.value);
const curCluster = computed(() => clusterList.value?.find(item => item.clusterID === clusterId.value) || {});
const isSharedOrVirtual = computed(() => curCluster.value?.is_shared || curCluster.value?.clusterType === 'virtual');
// 显示编辑态
const showEditStatus = ref(false);
const handleShowList = () => {
  showEditStatus.value = false;
  // handleGetLogCollectorRules();
};

watch(clusterId, () => {
  handleInitData();
});

const {
  onsData,
  getOnsDetail,
  updateLoading,
  updateOns,
  ruleList,
  logCollectorRules,
  retryLogCollectorRule,
  enableLogCollector,
  disableLogCollector,
  deleteLogCollectorRule,
} = useLog();

// 组件
const runningStatus = ref([
  'uninstalling',
  'pending-install',
  'pending-upgrade',
  'pending-rollback',
]);
const handleGetOnsData = async () => {
  const data = await getOnsDetail({ $clusterId: clusterId.value, $name: LOG_COLLECTOR });
  if (runningStatus.value.includes(data.status || '')) {
    start();
  } else {
    stop();
    handleGetLogCollectorRules();
  }
};
const { start, stop } = useInterval(handleGetOnsData, 5000);

// 组件更新
const showUpdateDialog = ref(false);
const showUpdateBtn = computed(() => onsData.value.currentVersion !== onsData.value.version
&& !isSharedOrVirtual.value);
const handleShowUpdateDialog = () => {
  showUpdateDialog.value = true;
};
const confirmUpdate = async () => {
  const result = await updateOns({
    $clusterId: clusterId.value,
    $name: LOG_COLLECTOR,
    version: ['failed-upgrade', 'failed-install', 'failed'].includes(onsData.value.status || '')
      ? onsData.value.currentVersion || ''
      : onsData.value.version || '',
    values: (!onsData.value.status ? onsData.value.defaultValues : onsData.value.currentValues) || '',
  });
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('更新任务提交成功'),
    });
    showUpdateDialog.value = false;
    handleGetOnsData();
  }
};

// 规则列表
const statusColorMap = ref({
  FAILED: 'red',
  SUCCESS: 'green',
  TERMINATED: 'gray',
});
const statusTextMap = ref({
  PENDING: $i18n.t('下发中'),
  RUNNING: $i18n.t('下发中'),
  FAILED: $i18n.t('下发失败'),
  SUCCESS: $i18n.t('正常'),
  TERMINATED: $i18n.t('已停用'),
});
const searchKeys = ref(['name', 'status', 'updator', 'rule.config.namespaces']);
const { searchValue, tableDataMatchSearch } = useTableSearch(ruleList, searchKeys);
const {
  curPageData,
  pageChange,
  pageSizeChange,
  pagination,
} = usePageConf(tableDataMatchSearch);
const ruleListLoading = ref(false);

// createID(产品要求把刚创建的排在第一个)
const creatingID = ref('');
const handleUpdateCreatingID = (id) => {
  creatingID.value = id;
};
// 获取规则列表
const handleGetLogCollectorRules = async (loading = true) => {
  if (!onsData.value.status) return;
  ruleListLoading.value = loading;
  const data = await logCollectorRules({ $clusterId: clusterId.value });
  ruleList.value = data.sort((pre) => {
    if (pre.id === creatingID.value) return -1;

    return 0;
  });
  if (data.some(item => ['PENDING', 'RUNNING'].includes(item.status || ''))) {
    startLoopRuleList();
  } else {
    stopLoopRuleList();
  }
  ruleListLoading.value = false;
};
const { start: startLoopRuleList, stop: stopLoopRuleList } = useInterval(() => handleGetLogCollectorRules(false), 5000);

const getNs = row => (row?.rule?.config?.namespaces?.length
  ? row.rule.config.namespaces.join(';')
  : $i18n.t('全部'));

// 规则 - 操作
const edit = ref(false);
const curRow = ref<IRuleData|null>();

// 打开日志链接
const openLink = (link) => {
  if (!link) return;
  window.open(link);
};
// 任务重试
const handleRetryRule = (row) => {
  $bkInfo({
    type: 'warning',
    title: $i18n.t('确定重试规则 {0}', [row.name]),
    clsName: 'custom-info-confirm default-info',
    okText: $i18n.t('重试'),
    cancelText: $i18n.t('取消'),
    async confirmFn() {
      const result = await retryLogCollectorRule({
        $clusterId: clusterId.value,
        $ID: row.id,
      });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('重试任务提交成功'),
        });
        handleGetLogCollectorRules();
      }
    },
  });
};
// 启用或停用规则
const handleToggleRule = (row) => {
  $bkInfo({
    type: 'warning',
    title: row.status === 'TERMINATED' ? $i18n.t('确定启用规则 {0}', [row.name]) : $i18n.t('确定停用规则 {0}', [row.name]),
    clsName: 'custom-info-confirm default-info',
    okText: row.status === 'TERMINATED' ? $i18n.t('启用') : $i18n.t('停用'),
    cancelText: $i18n.t('取消'),
    async confirmFn() {
      const params = {
        $clusterId: clusterId.value,
        $ID: row.id,
      };
      const result = row.status === 'TERMINATED'
        ? await enableLogCollector(params)
        : await disableLogCollector(params);
      if (result) {
        $bkMessage({
          theme: 'success',
          message: row.status === 'TERMINATED' ? $i18n.t('启用成功') :  $i18n.t('停用成功'),
        });
        handleGetLogCollectorRules();
        editLogRef.value?.handleGetDetail();
      }
    },
  });
};
// 删除规则
const handleDeleteRule = (row) => {
  $bkInfo({
    type: 'warning',
    title: $i18n.t('确定删除规则 {0}', [row.name]),
    clsName: 'custom-info-confirm default-info',
    okText: $i18n.t('删除'),
    cancelText: $i18n.t('取消'),
    async confirmFn() {
      const result = await deleteLogCollectorRule({
        $clusterId: clusterId.value,
        $ID: row.id,
      });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('删除成功'),
        });
        showEditStatus.value = false;
        handleGetLogCollectorRules();
      }
    },
  });
};

// 查看规则
const handleGotoDetail = (row) => {
  curRow.value = row;
  edit.value = false;
  showEditStatus.value = true;
};
// 创建规则
const editLogRef = ref();
const handleCreateRule = () => {
  curRow.value = null;
  showEditStatus.value = true;
  editLogRef.value?.setStatusOfCreateRule();
};
// 编辑规则
const handleEditRule = (row) => {
  curRow.value = row;
  edit.value = true;
  showEditStatus.value = true;
};
// 旧规则生成新规则
const createNewRule = (row) => {
  curRow.value = row;
  showEditStatus.value = true;
  edit.value = true;
};
const handleCreateNewRule = (row) => {
  if (row.new_rule_id) {
    $bkInfo({
      type: 'warning',
      title: $i18n.t('{0} 已生成新规则，是否再次生成？', [row.name]),
      clsName: 'custom-info-confirm default-info',
      okText: $i18n.t('继续生成'),
      cancelText: $i18n.t('取消'),
      async confirmFn() {
        createNewRule(row);
      },
    });
  } else {
    createNewRule(row);
  }
};

// 初始化数据
const loading = ref(false);
const handleInitData = async () => {
  if (!clusterId.value) return;
  showEditStatus.value = false;
  loading.value = true;
  await handleGetOnsData();
  await handleGetLogCollectorRules(false);
  loading.value = false;
  const { id } = $router.currentRoute.query || {};
  const activeRow = ruleList.value.find(item => item.id === id);
  if (activeRow) {
    curRow.value = activeRow;
    showEditStatus.value = true;
  }
};

onBeforeMount(() => {
  handleInitData();
});
</script>
<style lang="postcss" scoped>
.small-select {
  background: #EAEBF0 !important;
  border: none !important;
  box-shadow: none !important;
  width: 160px !important;
  margin-left: 24px;
}
</style>
