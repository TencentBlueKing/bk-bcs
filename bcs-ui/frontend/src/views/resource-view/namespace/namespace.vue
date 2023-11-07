<template>
  <LayoutContent
    hide-back
    :title="$t('k8s.namespace')"
    v-bkloading="{ isLoading: namespaceLoading }">
    <div class="wrapper flex flex-col place-content-between">
      <div>
        <LayoutRow class="mb15">
          <template #left>
            <!-- 权限详情接口目前暂不支持多个actionID -->
            <bcs-button
              theme="primary"
              icon="plus"
              v-authority="{
                clickable: true,
                actionId: 'namespace_create',
                autoUpdatePerms: true,
                permCtx: {
                  resource_type: 'cluster',
                  project_id: projectID,
                  cluster_id: clusterID
                }
              }"
              @click="handleToCreatedPage">
              {{ $t('generic.button.create') }}
            </bcs-button>
          </template>
          <template #right>
            <ClusterSelect
              v-if="!curClusterId"
              v-model="clusterID"
              class="mr-[10px]"
              searchable
              :disabled="!!curClusterId">
            </ClusterSelect>
            <bcs-input
              class="search-input"
              right-icon="bk-icon icon-search"
              :placeholder="$t('dashboard.placeholder.search1')"
              clearable
              v-model="searchValue">
            </bcs-input>
          </template>
        </LayoutRow>
        <bcs-table
          :data="curPageData"
          :pagination="pagination"
          size="medium"
          @page-change="pageChange"
          @page-limit-change="pageSizeChange">
          <bcs-table-column :label="$t('generic.label.name')" prop="name" min-width="200" show-overflow-tooltip>
            <template #default="{ row }">
              <bk-button
                class="bcs-button-ellipsis"
                text
                v-authority="{
                  clickable: webAnnotations.perms[row.name]
                    && webAnnotations.perms[row.name].namespace_view,
                  actionId: 'namespace_view',
                  resourceName: row.name,
                  disablePerms: true,
                  permCtx: {
                    project_id: projectID,
                    cluster_id: clusterID,
                    name: row.name
                  }
                }"
                @click="showDetail(row)">
                <span class="bcs-ellipsis">{{ row.name }}</span>
              </bk-button>
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('generic.label.status')">
            <template #default="{ row }">
              <div v-if="!isSharedCluster">
                <StatusIcon
                  :status-color-map="{
                    'Active': 'green',
                  }"
                  :status="row.status"
                  :pending="['Terminating'].includes(row.status)">
                  {{ row.status || '--' }}
                </StatusIcon>
              </div>
              <div v-else>
                <span
                  v-if="row.itsmTicketURL"
                  class="text-[#3a84ff] cursor-pointer"
                  @click="handleToITSM(row.itsmTicketURL)">
                  {{ $t('dashboard.ns.status.waitingApproval') }}（{{ itsmTicketTypeMap[row.itsmTicketType] }})
                </span>
                <span v-else>{{ $t('generic.status.ready') }}</span>
              </div>
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('metrics.cpuUsage')" prop="cpuUseRate" :render-header="renderHeader">
            <template #default="{ row }">
              <bcs-round-progress
                v-if="row.quota"
                ext-cls="biz-cluster-ring"
                width="50px"
                :percent="row.cpuUseRate"
                :config="{
                  strokeWidth: 10,
                  bgColor: '#f0f1f5',
                  activeColor: '#3a84ff'
                }"
                :num-style="{
                  fontSize: '12px',
                  width: '100%'
                }"
                v-bk-tooltips="{
                  content: `${$t('dashboard.ns.quota.cpuUsageRatio', {
                    used: row.used ? row.used.cpuLimits : 0,
                    total: row.quota.cpuLimits,
                  })}`
                }"
              ></bcs-round-progress>
              <span
                class="ml-[16px]"
                v-else
                v-bk-tooltips="{ content: $t('dashboard.ns.tips.notEnabledNamespaceQuota') }">--</span>
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('metrics.memUsage')" prop="memoryUseRate" :render-header="renderHeader">
            <template #default="{ row }">
              <bcs-round-progress
                v-if="row.quota"
                ext-cls="biz-cluster-ring"
                width="50px"
                :percent="row.memoryUseRate"
                :config="{
                  strokeWidth: 10,
                  bgColor: '#f0f1f5',
                  activeColor: '#3a84ff'
                }"
                :num-style="{
                  fontSize: '12px',
                  width: '100%'
                }"
                v-bk-tooltips="{
                  content: `${$t('dashboard.ns.quota.usageRatio', {
                    used: row.used ? `${unitConvert(row.used.memoryLimits, 'Gi', 'mem')}GiB` : 0,
                    total: `${row.quota.memoryLimits}B`,
                  })}`
                }"
              ></bcs-round-progress>
              <span
                class="ml-[16px]"
                v-else
                v-bk-tooltips="{ content: $t('dashboard.ns.tips.notEnabledNamespace') }">--</span>
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('cluster.labels.createdAt')">
            <template #default="{ row }">
              {{ row.createTime ? timeZoneTransForm(row.createTime, false) : '--' }}
            </template>
          </bcs-table-column>
          <bcs-table-column :label="$t('generic.label.action')" width="200">
            <template #default="{ row }">
              <bk-button
                text
                class="mr-[10px]"
                :disabled="applyMap(row.itsmTicketType).setQuota"
                v-authority="{
                  clickable: webAnnotations.perms[row.name]
                    && webAnnotations.perms[row.name].namespace_update,
                  actionId: 'namespace_update',
                  resourceName: row.name,
                  disablePerms: true,
                  permCtx: {
                    project_id: projectID,
                    cluster_id: clusterID,
                    name: row.name
                  }
                }"
                @click="showSetQuota(row)">
                {{ $t('cluster.detail.title.quota') }}
              </bk-button>
              <bk-button
                text
                class="mr-[10px]"
                :disabled="applyMap(row.itsmTicketType).setVar"
                @click="showSetVariable(row)">
                {{ $t('dashboard.ns.action.setEnv') }}
              </bk-button>
              <bk-popover
                placement="bottom"
                theme="light dropdown"
                :arrow="false"
                trigger="click">
                <span class="bcs-icon-more-btn"><i class="bcs-icon bcs-icon-more"></i></span>
                <template #content>
                  <ul class="bcs-dropdown-list">
                    <template v-if="!isSharedCluster">
                      <li
                        class="bcs-dropdown-item"
                        v-authority="{
                          clickable: webAnnotations.perms[row.name]
                            && webAnnotations.perms[row.name].namespace_update,
                          actionId: 'namespace_update',
                          resourceName: row.name,
                          disablePerms: true,
                          permCtx: {
                            project_id: projectID,
                            cluster_id: clusterID,
                            name: row.name
                          }
                        }"
                        @click="handleSetLabel(row)">
                        {{ $t('cluster.nodeList.button.setLabel') }}
                      </li>
                      <li
                        class="bcs-dropdown-item"
                        v-authority="{
                          clickable: webAnnotations.perms[row.name]
                            && webAnnotations.perms[row.name].namespace_update,
                          actionId: 'namespace_update',
                          resourceName: row.name,
                          disablePerms: true,
                          permCtx: {
                            project_id: projectID,
                            cluster_id: clusterID,
                            name: row.name
                          }
                        }"
                        @click="handleSetAnnotations(row)">
                        {{ $t('dashboard.ns.action.setAnnotation') }}
                      </li>
                    </template>
                    <li
                      v-if="row.itsmTicketURL"
                      class="bcs-dropdown-item w-[80px]"
                      @click="withdrawNamespace(row)">
                      {{ $t('dashboard.ns.action.undo') }}
                    </li>
                    <li
                      v-else
                      class="bcs-dropdown-item w-[80px]"
                      :disabled="applyMap(row.itsmTicketType).delete"
                      v-authority="{
                        clickable: webAnnotations.perms[row.name]
                          && webAnnotations.perms[row.name].namespace_delete,
                        actionId: 'namespace_delete',
                        resourceName: row.name,
                        disablePerms: true,
                        permCtx: {
                          project_id: projectID,
                          cluster_id: clusterID,
                          name: row.name
                        }
                      }"
                      @click="handleDeleteNamespace(row)">
                      {{ $t('generic.button.delete') }}
                    </li>
                  </ul>
                </template>
              </bk-popover>
            </template>
          </bcs-table-column>
          <template #empty>
            <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="searchValue = ''" />
          </template>
        </bcs-table>
      </div>
      <AppFooter />
    </div>
    <!-- 设置变量 -->
    <bcs-sideslider
      :is-show.sync="setVariableConf.isShow"
      :title="setVariableConf.title"
      :width="600"
      :before-close="handleBeforeClose"
      quick-close
      @hidden="hideSetVariable">
      <div slot="content" class="py-5 px-5" v-bkloading="{ isLoading: variableLoading }">
        <div>
          <div class="bk-form-item text-[14px]">
            {{$t('generic.label.var1')}}
          </div>
          <div class="bk-form-item text-[12px]">
            <i18n path="dashboard.ns.label.createMoreNamespaceVars">
              <button place="action" class="bk-text-button" @click="handleGoVar">{{$t('deploy.variable.env')}}</button>
            </i18n>
          </div>
          <template v-if="variablesList.length">
            <div class="bk-form-item">
              <div class="bk-form-content">
                <div class="flex items-center mb-[10px]" v-for="(variable, index) in variablesList" :key="index">
                  <div
                    class="flex-1"
                    v-bk-tooltips="{
                      content: `${$t('cluster.nodeTemplate.variable.label.var.text')}: ${variable.name}`
                    }">
                    <bk-input disabled v-model="variable.key"></bk-input>
                  </div>
                  <span class="px-[5px]">=</span>
                  <bk-input
                    class="flex-1"
                    :placeholder="$t('generic.label.value')"
                    v-model="variable.value"
                    @change="setChanged(true)">
                  </bk-input>
                </div>
              </div>
            </div>
            <div class="mt-[20px]">
              <bk-button type="primary" :loading="variableLoading" @click="updateVariable">
                {{$t('generic.button.save')}}
              </bk-button>
              <bk-button class="ml-[5px]" :loading="variableLoading" @click="hideSetVariable">
                {{$t('generic.button.cancel')}}
              </bk-button>
            </div>
          </template>
          <template v-else>
            <div class="h-[100px] text-center leading-[100px]">{{ $t('dashboard.ns.label.noNamespaceVariables') }}</div>
          </template>
        </div>
      </div>
    </bcs-sideslider>
    <!-- 配额管理 -->
    <bcs-dialog
      v-model="setQuotaConf.isShow"
      :width="650"
      :title="$t('dashboard.ns.title.quotaManagement', { nsName: setQuotaConf.namespace })">
      <bk-form
        ref="setQuotaForm"
        :label-width="120"
        v-bkloading="{ isLoading: setQuotaConf.loading }">
        <bk-form-item :label="$t('dashboard.ns.create.quota')">
          <bk-switcher
            v-model="showQuota"
            class="ml-[10px]"
            :disabled="isSharedCluster && showQuota"
            size="small"
            :selected="showQuota"
            :key="showQuota"
            @change="toggleShowQuota">
          </bk-switcher>
        </bk-form-item>
        <bk-form-item
          v-if="showQuota"
          :rules="quotaRules"
          label="CPU"
          property="quota"
          error-display-type="normal">
          <div class="flex mr-[20px]">
            <bcs-input
              v-model="setQuotaConf.quota.cpuRequests"
              class="w-[200px]"
              type="number"
              :min="1"
              :max="512000"
              :precision="0">
              <div class="group-text" slot="append">{{ $t('units.suffix.cores') }}</div>
            </bcs-input>
            <span class="mx-[10px]">Mem</span>
            <bcs-input
              v-model="setQuotaConf.quota.memoryRequests"
              class="w-[200px]"
              type="number"
              :min="1"
              :max="1024000"
              :precision="0">
              <div class="group-text" slot="append">GiB</div>
            </bcs-input>
          </div>
        </bk-form-item>
      </bk-form>
      <div slot="footer">
        <bcs-button
          theme="primary"
          :loading="setQuotaConf.loading"
          class="mr5"
          @click="updateNamespace"
        >{{ $t('generic.button.confirm') }}</bcs-button>
        <bcs-button
          :disabled="setQuotaConf.loading"
          @click="cancelUpdateNamespace"
        >{{ $t('generic.button.cancel') }}</bcs-button>
      </div>
    </bcs-dialog>
    <!-- 命名空间Detail -->
    <bk-sideslider
      :is-show.sync="showNamespaceDetail"
      :title="namespaceInfo.name"
      :width="800"
      quick-close>
      <div slot="content">
        <Detail
          :data="namespaceInfo">
        </Detail>
      </div>
    </bk-sideslider>
    <!-- 设置标签 -->
    <bk-sideslider
      :is-show.sync="showSetLabel"
      :title="$t('cluster.nodeList.button.setLabel')"
      :width="800"
      :before-close="handleBeforeClose"
      quick-close>
      <div slot="content">
        <KeyValue
          class="key-value-content"
          :model-value="curNamespaceData.labels"
          :loading="updateBtnLoading"
          :min-items="0"
          :key-rules="[
            {
              message: $i18n.t('generic.validate.labelKey1'),
              validator: KEY_REGEXP
            }
          ]"
          :value-rules="[
            {
              message: $i18n.t('generic.validate.labelKey1'),
              validator: VALUE_REGEXP
            }
          ]"
          @data-change="setChanged(true)"
          @cancel="handleLabelEditCancel"
          @confirm="handleLabelEditConfirm"
        ></KeyValue>
      </div>
    </bk-sideslider>
    <!-- 设置注解 -->
    <bk-sideslider
      :is-show.sync="showSetAnnotations"
      :title="$t('dashboard.ns.action.setAnnotation')"
      :width="800"
      :before-close="handleBeforeClose"
      quick-close>
      <div slot="content">
        <KeyValue
          class="key-value-content"
          :model-value="curNamespaceData.annotations"
          :loading="updateBtnLoading"
          :min-items="0"
          :key-rules="[
            {
              message: $i18n.t('generic.validate.labelKey1'),
              validator: KEY_REGEXP
            }
          ]"
          :value-rules="[
            {
              message: $i18n.t('generic.validate.labelKey1'),
              validator: VALUE_REGEXP
            }
          ]"
          @data-change="setChanged(true)"
          @cancel="handleAnnotationsEditCancel"
          @confirm="handleAnnotationsEditConfirm"
        ></KeyValue>
      </div>
    </bk-sideslider>
  </LayoutContent>
</template>

<script lang="ts">
import { computed, CreateElement, defineComponent, reactive, ref, toRef, watch } from 'vue';

import StatusIcon from '../../../components/status-icon';
import usePage from '../../../composables/use-page';
import useSearch from '../../../composables/use-search';

import Detail from './detail.vue';
import { useNamespace } from './use-namespace';

import $bkMessage from '@/common/bkmagic';
import { KEY_REGEXP, VALUE_REGEXP } from '@/common/constant';
import { timeZoneTransForm } from '@/common/util';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import KeyValue from '@/components/key-value.vue';
import LayoutContent from '@/components/layout/Content.vue';
import LayoutRow from '@/components/layout/Row.vue';
import { useCluster, useProject } from '@/composables/use-app';
import useInterval from '@/composables/use-interval';
import useSideslider from '@/composables/use-sideslider';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import AppFooter from '@/views/app/app-footer.vue';

export default defineComponent({
  name: 'NamespaceList',
  components: {
    ClusterSelect,
    LayoutContent,
    LayoutRow,
    Detail,
    KeyValue,
    StatusIcon,
    AppFooter,
  },
  setup() {
    const $route = computed(() => toRef(reactive($router), 'currentRoute').value);

    const { handleBeforeClose, reset, setChanged } = useSideslider();
    const { projectID } = useProject();
    const { isSharedCluster } = useCluster();

    const curClusterId = computed(() => $store.getters.curClusterId);

    const clusterID = ref(curClusterId.value);

    const keys = ref(['name']);

    const showQuota = ref(true);

    const toggleShowQuota = () => {
      if (!setQuotaConf.value.quota.cpuRequests) {
        setQuotaConf.value.quota.cpuRequests = '1';
      }
      if (!setQuotaConf.value.quota.memoryRequests) {
        setQuotaConf.value.quota.memoryRequests = '1';
      }
      showQuota.value = !showQuota.value;
    };

    const quotaRules = [
      {
        validator() {
          return setQuotaConf.value.quota.cpuRequests && setQuotaConf.value.quota.memoryRequests
              && setQuotaConf.value.quota.cpuRequests !== 'NaN' && setQuotaConf.value.quota.memoryRequests !== 'NaN';
        },
        message: $i18n.t('dashboard.ns.validate.setMinMaxMemCpu'),
        trigger: 'blur',
      },
    ];

    const {
      variablesList,
      variableLoading,
      namespaceLoading,
      namespaceData,
      webAnnotations,
      handleGetVariablesList,
      handleWithdrawNamespace,
      handleDeleteNameSpace,
      handleUpdateNameSpace,
      handleUpdateVariablesList,
      getNamespaceData,
      getNamespaceInfo,
    } = useNamespace();

    const { start, stop } = useInterval(() => getNamespaceData({ $clusterId: clusterID.value }, false));
    watch(namespaceData, () => {
      if (namespaceData.value.some(item => ['Terminating'].includes(item.status))) {
        start();
      } else {
        stop();
      }
    });
    // 渲染表头
    const renderHeader = (h: CreateElement, data) => h('span', {
      class: 'custom-header-cell',
      directives: [
        {
          name: 'bkTooltips',
          value: {
            content: data.column.property === 'cpuUseRate' ? $i18n.t('dashboard.ns.tips.totalCpuLimitsQuota') : $i18n.t('dashboard.ns.tips.totalMemoryLimitsQuota'),
          },
        },
      ],
    }, [data.column.label]);

    const projectCode = computed(() => $route.value.params.projectCode);

    // 搜索功能
    const { tableDataMatchSearch, searchValue } = useSearch(namespaceData, keys);

    // 分页
    const { pagination, curPageData, pageConf, pageChange, pageSizeChange } = usePage(tableDataMatchSearch);

    // 搜索时重置分页
    watch(searchValue, () => {
      pageConf.current = 1;
    });

    // 删除命名空间
    const handleDeleteNamespace = (row) => {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('generic.title.confirmDelete1', { name: row.name }),
        subTitle: $i18n.t('dashboard.ns.title.deleteNamespaceWarning'),
        defaultInfo: true,
        confirmFn: async () => {
          const result = await handleDeleteNameSpace({
            $clusterId: clusterID.value,
            $namespace: row?.name,
          });
          if (result) {
            getNamespaceData({
              $clusterId: clusterID.value,
            });
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.delete'),
            });
          }
        },
      });
    };

    // 设置变量值
    const setVariableConf = ref({
      isShow: false,
      title: '',
      namespace: '',
    });
    const showSetVariable = async (row) => {
      const namespace = row?.name;
      setVariableConf.value.isShow = true;
      setVariableConf.value.namespace = namespace;
      setVariableConf.value.title = $i18n.t('generic.title.setVar') + namespace;
      await handleGetVariablesList({
        $clusterId: clusterID.value,
        $namespace: namespace,
      });
      reset();
    };

    const hideSetVariable = () => {
      variablesList.value = [];
      setVariableConf.value.isShow = false;
      setVariableConf.value.title = '';
    };

    // 更新变量值
    const updateVariable = async () => {
      variableLoading.value = true;
      const result = await handleUpdateVariablesList({
        $clusterId: clusterID.value,
        $namespace: setVariableConf.value.namespace,
        data: variablesList.value,
      });
      variableLoading.value = false;
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.save'),
        });
        hideSetVariable();
        getNamespaceData({
          $clusterId: clusterID.value,
        });
      }
    };

    // 设置配额
    const initQuotaConf = {
      loading: false,
      isShow: false,
      namespace: '',
      labels: [],
      annotations: [],
      quota: {
        cpuLimits: '',
        cpuRequests: '',
        memoryLimits: '',
        memoryRequests: '',
      },
    };
    const setQuotaConf = ref({ ...initQuotaConf });

    const curNamespace = ref<any>({});
    const showSetQuota = async (row) => {
      setQuotaConf.value.isShow = true;
      setQuotaConf.value.namespace = row.name;
      setQuotaConf.value.loading = true;
      curNamespace.value = await getNamespaceInfo({
        $clusterId: clusterID.value,
        $name: row.name,
      });
      setQuotaConf.value.loading = false;

      showQuota.value = !!curNamespace.value.quota;
      const { labels, annotations, quota = {} } = curNamespace.value;
      setQuotaConf.value.labels = labels;
      setQuotaConf.value.annotations = annotations;
      if (quota) {
        setQuotaConf.value.quota = {
          cpuLimits: quota.cpuLimits || '',
          cpuRequests: unitConvert(quota.cpuRequests, '', 'cpu'),
          memoryLimits: quota.memoryLimits || '',
          memoryRequests: unitConvert(quota.memoryRequests, 'Gi', 'mem'),
        };
      }
    };

    const setQuotaForm = ref();
    const updateNamespace = async () => {
      setQuotaForm.value.validate().then(async () => {
        const { namespace, labels, annotations, quota } = setQuotaConf.value;
        setQuotaConf.value.loading = true;
        const result = await handleUpdateNameSpace({
          $clusterId: clusterID.value,
          $namespace: namespace,
          labels,
          annotations,
          quota: showQuota.value ? {
            cpuLimits: String(quota.cpuRequests),
            cpuRequests: String(quota.cpuRequests),
            memoryLimits: `${quota.memoryRequests}Gi`,
            memoryRequests: `${quota.memoryRequests}Gi`,
          } : null,
        });
        if (result) {
          $bkMessage({
            theme: 'success',
            message: $i18n.t('generic.msg.success.update'),
          });
          getNamespaceData({
            $clusterId: clusterID.value,
          });
          cancelUpdateNamespace();
        };
        setQuotaConf.value.loading = false;
      });
    };

    const cancelUpdateNamespace = () => {
      setQuotaConf.value = { ...initQuotaConf };
    };

    const handleToCreatedPage = () => {
      $router.push({
        name: 'dashboardNamespaceCreate',
        query: {
          kind: 'Namespace',
        },
      });
    };

    const unitMap = {
      cpu: {
        list: ['m', '', 'k', 'M', 'G', 'T', 'P', 'E'],
        digit: 3,
        base: 10,
      },
      mem: {
        list: ['', 'Ki', 'Mi', 'Gi', 'Ti', 'Pi', 'Ei'],
        digit: 10,
        base: 2,
      },
    };
    const unitConvert = (val, toUnit = '', type: 'cpu' | 'mem') => {
      const { list, digit, base } = unitMap[type];
      const num = val.match(/\d+/gi)?.[0];
      const uint = val.match(/[a-z|A-Z]+/gi)?.[0] || '';

      const index = list.indexOf(uint);
      // 没有单位直接返回
      if (index === -1) return val;

      // 要转换成的单位
      const toUnitIndex = list.indexOf(toUnit) || -1;
      const factorial = index - toUnitIndex;
      if (factorial >= 0) {
        return num * (base ** (digit * factorial));
      }
      return num / (base ** (Math.abs(digit) * Math.abs(factorial)));
    };

    const itsmTicketTypeMap = {
      CREATE: $i18n.t('dashboard.ns.status.createNamespace'),
      UPDATE: $i18n.t('dashboard.ns.status.quotaAdjustment'),
      DELETE: $i18n.t('dashboard.ns.status.deleteNamespace'),
    };

    const handleToITSM = (link) => {
      window.open(link, '_blank');
    };

    // namespace itsmTicketType -> 操作按钮禁用状态控制
    const applyMap = (type: ('CREATE' | 'UPDATE' | 'DELETE' | '')) => {
      const typeMap = {
        CREATE: {
          setVar: true,
          setLabel: true,
          setAnnotations: true,
          setQuota: true,
          delete: true,
        },
        UPDATE: {
          setVar: false,
          setLabel: false,
          setAnnotations: false,
          setQuota: true,
          delete: true,
        },
        DELETE: {
          setVar: true,
          setLabel: true,
          setAnnotations: true,
          setQuota: true,
          delete: true,
        },
        '': {
          setVar: false,
          setLabel: false,
          setAnnotations: false,
          setQuota: false,
          delete: false,
        },
      };
      return typeMap[type];
    };

    const handleGoVar = () => {
      setVariableConf.value.isShow = false;
      $router.push({
        name: 'variable',
        params: {
          projectCode: projectCode.value,
        },
      });
    };
    const showNamespaceDetail = ref(false);
    const namespaceInfo = ref<any>({});
    const showDetail = (row) => {
      showNamespaceDetail.value = true;
      namespaceInfo.value = row;
    };

    // 设置标签
    const updateBtnLoading = ref(false);
    const showSetLabel = ref(false);
    const curNamespaceData = ref<any>({});
    const handleSetLabel = (row) => {
      showSetLabel.value = true;
      curNamespaceData.value = row;
      reset();
    };
    const handleLabelEditCancel = () => {
      showSetLabel.value = false;
    };
    const handleLabelEditConfirm = async (val) => {
      const labels = Object.keys(val).reduce((prev: any, cur) => {
        prev.push({
          key: cur,
          value: val[cur],
        });
        return prev;
      }, []);
      updateBtnLoading.value = true;
      const result = await handleUpdateNameSpace({
        $clusterId: clusterID.value,
        $namespace: curNamespaceData.value.name,
        ...curNamespaceData.value,
        labels,
      });
      updateBtnLoading.value = false;
      if (result) {
        handleLabelEditCancel();
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.save'),
        });
        getNamespaceData({
          $clusterId: clusterID.value,
        });
      }
    };

    // 设置注解
    const showSetAnnotations = ref(false);
    const handleSetAnnotations = (row) => {
      showSetAnnotations.value = true;
      curNamespaceData.value = row;
      reset();
    };

    const handleAnnotationsEditConfirm = async (val) => {
      const annotations = Object.keys(val).reduce((prev: any, cur) => {
        prev.push({
          key: cur,
          value: val[cur],
        });
        return prev;
      }, []);
      updateBtnLoading.value = true;
      const result = await handleUpdateNameSpace({
        $clusterId: clusterID.value,
        $namespace: curNamespaceData.value.name,
        ...curNamespaceData.value,
        annotations,
      });
      updateBtnLoading.value = false;
      if (result) {
        handleAnnotationsEditCancel();
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.save'),
        });
        getNamespaceData({
          $clusterId: clusterID.value,
        });
      }
    };
    const handleAnnotationsEditCancel = () => {
      showSetAnnotations.value = false;
    };

    const withdrawNamespace = async (row) => {
      namespaceLoading.value = true;
      const result = await handleWithdrawNamespace({
        $clusterId: clusterID.value,
        $namespace: row.name,
      });
      namespaceLoading.value = false;
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('dashboard.ns.msg.undoSuccess'),
        });
        getNamespaceData({
          $clusterId: clusterID.value,
        });
      }
    };

    watch(curClusterId, () => {
      if (!curClusterId.value) return;
      clusterID.value = curClusterId.value;
    });
    watch(clusterID, () => {
      pageConf.current = 1;
      getNamespaceData({
        $clusterId: clusterID.value,
      });
    });

    getNamespaceData({
      $clusterId: clusterID.value,
    });

    return {
      namespaceLoading,
      webAnnotations,
      curClusterId,
      isSharedCluster,
      clusterID,
      projectID,
      pagination,
      showQuota,
      searchValue,
      curPageData,
      setVariableConf,
      setQuotaConf,
      variablesList,
      variableLoading,
      itsmTicketTypeMap,
      namespaceInfo,
      showNamespaceDetail,
      showSetLabel,
      showSetAnnotations,
      curNamespaceData,
      updateBtnLoading,
      VALUE_REGEXP,
      KEY_REGEXP,
      quotaRules,
      setQuotaForm,
      toggleShowQuota,
      unitConvert,
      applyMap,
      pageChange,
      pageSizeChange,
      handleDeleteNamespace,
      showSetVariable,
      hideSetVariable,
      updateVariable,
      showSetQuota,
      updateNamespace,
      cancelUpdateNamespace,
      handleToCreatedPage,
      handleToITSM,
      timeZoneTransForm,
      handleGoVar,
      showDetail,
      renderHeader,
      handleSetLabel,
      handleSetAnnotations,
      handleLabelEditCancel,
      handleLabelEditConfirm,
      handleAnnotationsEditConfirm,
      handleAnnotationsEditCancel,
      withdrawNamespace,
      handleBeforeClose,
      setChanged,
    };
  },
});
</script>

<style lang="postcss" scoped>
  @import './namespace.css';
  >>> .custom-header-cell {
    display: inline-block;
    line-height: 18px;
    border-bottom: 1px dashed #979ba5;
  }
  .key-value-content {
    padding: 20px 30px;
  }
  ::v-deep .form-error-tip {
    text-align: left;
  }
  .wrapper {
    min-height: calc(100vh - 144px);
  }
</style>
