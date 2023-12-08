<template>
  <div v-bkloading="{ isLoading: loading }">
    <Header>
      <div class="flex items-center">
        {{ $t('nav.log') }}
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
              {{ $t('generic.button.update') }}
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
        :title="$t('logCollector.label.updateChart')"
        header-position="left"
        :ok-text="$t('logCollector.button.updateChart')"
        :loading="updateLoading"
        v-model="showUpdateDialog"
        @confirm="confirmUpdate">
        <bk-form :label-width="labelWidth" ref="formRef">
          <bk-form-item :label="$t('deploy.templateset.currentVersionNumber')">
            <div class="flex items-end">
              {{ onsData.currentVersion }}
            </div>
          </bk-form-item>
          <bk-form-item class="!mt-[0px]" :label="$t('logCollector.label.newChartVersion')">
            <div class="flex items-end">
              {{ onsData.version }}
            </div>
          </bk-form-item>
          <bk-form-item class="!mt-[0px]" :label="$t('logCollector.label.newChartDesc')">
            <div class="flex items-end">
              {{ onsData.description }}
            </div>
          </bk-form-item>
        </bk-form>
      </bcs-dialog>
    </Header>
    <div class="pb-[16px] max-h-[calc(100vh-104px)] overflow-auto" :key="clusterId">
      <bcs-exception type="empty" v-if="!onsData.status && !loading">
        <div>{{ isSharedOrVirtual ? $t('logCollector.msg.notEnable1') : $t('logCollector.msg.notEnable') }}</div>
        <!-- shared 和 virtual集群不支持启用 -->
        <bcs-button
          theme="primary"
          class="w-[88px] mt-[16px]"
          v-if="!isSharedOrVirtual"
          @click="confirmUpdate">
          {{ $t('logCollector.action.enable') }}
        </bcs-button>
      </bcs-exception>
      <bcs-exception
        type="empty"
        v-else-if="['failed-upgrade', 'failed-install','failed'].includes(onsData.status || '') && !loading">
        <div>{{ $t('logCollector.msg.installFailed') }}</div>
        <div class="text-[#979BA5] mt-[16px]" v-if="onsData.message">{{ onsData.message }}</div>
        <bcs-button
          theme="primary"
          class="w-[88px] mt-[16px]"
          @click="confirmUpdate">{{ $t('logCollector.action.reInstall') }}</bcs-button>
      </bcs-exception>
      <template v-else>
        <bcs-alert type="info" closable>
          <template #title>
            <div class="flex items-center">
              <span class="flex-1">
                {{ $t('logCollector.msg.logCollector') }}
              </span>
              <bcs-button
                text
                class="text-[12px]"
                @click="openLink(PROJECT_CONFIG.rule)">
                {{ $t('generic.button.learnMore') }}
              </bcs-button>
            </div>
          </template>
        </bcs-alert>
        <Row class="mt-[16px] px-[24px]">
          <template #left>
            <bcs-button
              theme="primary"
              icon="plus"
              @click="handleCreateRule">{{ $t('plugin.tools.create') }}</bcs-button>
          </template>
          <template #right>
            <bcs-input
              class="w-[320px]"
              :placeholder="$t('logCollector.placeholder.search')"
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
            <bcs-table-column :label="$t('plugin.tools.ruleName')" prop="name" sortable>
              <template #default="{ row }">
                <div class="flex">
                  <!-- 是否旧规则 -->
                  <span
                    class="inline-flex items-center justify-center w-[24px] relative
                  h-[24px] bg-[#F0F1F5] rounded-sm text-[#979BA5] mr-[8px]"
                    v-bk-tooltips="{
                      content: $t('logCollector.tips.oldRuleHasNewRule', [
                        row.new_rule_name,
                        moment(row.new_rule_created_at).add(30, 'days').format('YYYY-MM-DD')
                      ]),
                      disabled: !row.new_rule_id
                    }"
                    v-if="row.old">
                    {{ $t('logCollector.msg.oldRuleFlag') }}
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
                    <span class="bcs-ellipsis" v-bk-overflow-tips>{{ row.name }}</span>
                  </bcs-button>
                </div>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('k8s.namespace')" show-overflow-tooltip>
              <template #default="{ row }">
                {{ getNs(row) }}
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('generic.label.memo')" show-overflow-tooltip prop="description">
              <template #default="{ row }">
                {{ row.description || '--' }}
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('generic.label.updator')" prop="updator"></bcs-table-column>
            <bcs-table-column
              :label="$t('cluster.labels.updatedAt')"
              sortable
              prop="updated_at"
              width="180"></bcs-table-column>
            <bcs-table-column :label="$t('generic.label.status')" width="120">
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
                    {{statusTextMap[row.status] || $t('generic.status.unknown1')}}
                  </span>
                </StatusIcon>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('generic.label.action')" width="280">
              <template #default="{ row }">
                <template v-if="row.status === 'TERMINATED'">
                  <bcs-button text class="text-[12px] mr-[10px]" @click="handleToggleRule(row)">
                    {{ $t('logCollector.action.enable') }}
                  </bcs-button>
                  <bcs-button text class="text-[12px] mr-[10px]" @click="handleDeleteRule(row)">
                    {{ $t('generic.button.delete') }}
                  </bcs-button>
                </template>
                <template v-if="row.status === 'SUCCESS'">
                  <bcs-button
                    text
                    class="text-[12px] mr-[10px]"
                    v-if="row.entrypoint.file_log_url && row.rule.config.paths && row.rule.config.paths.length"
                    @click="openLink(row.entrypoint.file_log_url)">
                    {{ $t('logCollector.action.fileLog') }}
                  </bcs-button>
                  <bcs-button
                    text
                    class="text-[12px] mr-[10px]"
                    v-if="row.entrypoint.std_log_url && row.rule.config.enable_stdout"
                    @click="openLink(row.entrypoint.std_log_url)">
                    {{ $t('logCollector.action.stdLog') }}
                  </bcs-button>
                </template>
                <bcs-button
                  text
                  class="text-[12px] mr-[10px]"
                  :disabled="row.status !== 'FAILED'"
                  @click="handleRetryRule(row)"
                  v-if="row.status === 'FAILED'">
                  {{ $t('cluster.ca.nodePool.records.action.retry') }}
                </bcs-button>
                <!-- 旧规则不允许编辑 -->
                <bcs-button
                  text
                  class="text-[12px] mr-[10px]"
                  v-if="row.old"
                  @click="handleCreateNewRule(row)">
                  {{ $t('logCollector.action.newRule') }}
                </bcs-button>
                <bcs-button
                  text
                  class="text-[12px] mr-[10px]"
                  :disabled="['PENDING', 'RUNNING'].includes(row.status)"
                  v-else-if="row.status !== 'TERMINATED'"
                  @click="handleEditRule(row)">
                  {{ $t('generic.button.edit') }}
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
                      {{ row.status === 'TERMINATED'
                        ? $t('logCollector.action.enable') : $t('logCollector.action.stop') }}
                    </li>
                    <li class="bcs-dropdown-item" @click="handleDeleteRule(row)">{{ $t('generic.button.delete') }}</li>
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
      </template>
    </div>
  </div>
</template>
<script setup lang="ts">
import moment from 'moment';
import { computed, onBeforeMount, ref, watch } from 'vue';

import EditLogCollector from './edit-log-collector.vue';
import useLog, { IRuleData } from './use-log';

import $bkMessage from '@/common/bkmagic';
import { LOG_COLLECTOR } from '@/common/constant';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import Header from '@/components/layout/Header.vue';
import Row from '@/components/layout/Row.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import StatusIcon from '@/components/status-icon';
import { useCluster } from '@/composables/use-app';
import useFormLabel from '@/composables/use-form-label';
import useInterval from '@/composables/use-interval';
import usePageConf from '@/composables/use-page';
import useTableSearch from '@/composables/use-search';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

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
const { labelWidth, initFormLabelWidth } = useFormLabel();
const formRef = ref();
const showUpdateDialog = ref(false);
const showUpdateBtn = computed(() => onsData.value.currentVersion !== onsData.value.version
&& !isSharedOrVirtual.value);
const handleShowUpdateDialog = () => {
  showUpdateDialog.value = true;
  setTimeout(() => {
    initFormLabelWidth(formRef.value);
  }, 0);
};
const confirmUpdate = async () => {
  ruleList.value = []; // 清空上一次列表数据
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
      message: $i18n.t('deploy.helm.upgradeTaskSubmit'),
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
  PENDING: $i18n.t('logCollector.status.pending'),
  RUNNING: $i18n.t('logCollector.status.pending'),
  FAILED: $i18n.t('logCollector.status.failed'),
  SUCCESS: $i18n.t('generic.status.ready'),
  TERMINATED: $i18n.t('generic.status.terminated'),
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
  : $i18n.t('generic.label.total'));

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
    title: $i18n.t('logCollector.title.confirmRetry', [row.name]),
    clsName: 'custom-info-confirm default-info',
    okText: $i18n.t('cluster.ca.nodePool.records.action.retry'),
    cancelText: $i18n.t('generic.button.cancel'),
    async confirmFn() {
      const result = await retryLogCollectorRule({
        $clusterId: clusterId.value,
        $ID: row.id,
      });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('logCollector.msg.success.retry'),
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
    title: row.status === 'TERMINATED' ? $i18n.t('logCollector.title.confirmEnable', [row.name]) : $i18n.t('logCollector.title.confirmStop', [row.name]),
    clsName: 'custom-info-confirm default-info',
    okText: row.status === 'TERMINATED' ? $i18n.t('logCollector.action.enable') : $i18n.t('logCollector.action.stop'),
    cancelText: $i18n.t('generic.button.cancel'),
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
          message: row.status === 'TERMINATED' ? $i18n.t('logCollector.msg.success.enable') :  $i18n.t('logCollector.msg.success.stop'),
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
    title: $i18n.t('logCollector.title.confirmDelete', [row.name]),
    clsName: 'custom-info-confirm default-info',
    okText: $i18n.t('generic.button.delete'),
    cancelText: $i18n.t('generic.button.cancel'),
    async confirmFn() {
      const result = await deleteLogCollectorRule({
        $clusterId: clusterId.value,
        $ID: row.id,
      });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.delete'),
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
      title: $i18n.t('logCollector.title.exitNewRule', [row.name]),
      clsName: 'custom-info-confirm default-info',
      okText: $i18n.t('logCollector.button.continueCreateNewRule'),
      cancelText: $i18n.t('generic.button.cancel'),
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
  // await handleGetLogCollectorRules(false);
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
