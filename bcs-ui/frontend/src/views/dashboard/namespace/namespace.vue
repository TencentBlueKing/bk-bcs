<template>
  <LayoutContent
    hide-back
    :title="$t('命名空间')"
    v-bkloading="{ isLoading: namespaceLoading }">
    <LayoutRow class="mb15">
      <template #left>
        <bcs-button
          class="w-[100px]"
          theme="primary"
          icon="plus"
          @click="handleToCreatedPage">
          {{ $t('创建') }}
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
          :placeholder="$t('搜索名称')"
          v-model="searchValue">
        </bcs-input>
      </template>
    </LayoutRow>
    <bcs-table
      :data="curPageData"
      :pagination="pagination"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <bcs-table-column :label="$t('名称')" prop="name" min-width="200" show-overflow-tooltip>
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
            {{ row.name }}
          </bk-button>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('状态')">
        <template #default="{ row }">
          <div v-if="!isSharedCluster">
            {{ row.status || '--' }}
          </div>
          <div v-else>
            <span
              v-if="row.itsmTicketURL"
              class="text-[#3a84ff] cursor-pointer"
              @click="handleToITSM(row.itsmTicketURL)">
              {{ $t('待审批') }}（{{ itsmTicketTypeMap[row.itsmTicketType] }})
            </span>
            <span v-else>{{ $t('正常') }}</span>
          </div>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('CPU使用率')" prop="cpuUseRate" :render-header="renderHeader">
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
              content: `${$t('{used}核 / {total}核 (已使用/总量)', {
                used: row.used ? row.used.cpuLimits : 0,
                total: row.quota.cpuLimits,
              })}`
            }"
          ></bcs-round-progress>
          <span class="ml-[16px]" v-else v-bk-tooltips="{ content: $t('未开启命名空间配额') }">--</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('内存使用率')" prop="memoryUseRate" :render-header="renderHeader">
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
              content: `${$t('{used} / {total} (已使用/总量)', {
                used: row.used ? `${unitConvert(row.used.memoryLimits, 'Gi', 'mem')}Gi` : 0,
                total: row.quota.memoryLimits,
              })}`
            }"
          ></bcs-round-progress>
          <span class="ml-[16px]" v-else v-bk-tooltips="{ content: $t('未开启命名空间') }">--</span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('创建时间')">
        <template #default="{ row }">
          {{ row.createTime ? timeZoneTransForm(row.createTime, false) : '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('操作')" width="220">
        <template #default="{ row }">
          <bk-button
            text
            class="mr-[10px]"
            :disabled="applyMap(row.itsmTicketType).setVar"
            @click="showSetVariable(row)">
            {{ $t('设置变量值') }}
          </bk-button>
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
            {{ $t('配额管理') }}
          </bk-button>
          <bk-button
            text
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
            {{ $t('删除') }}
          </bk-button>
        </template>
      </bcs-table-column>
    </bcs-table>
    <!-- 设置变量 -->
    <bcs-sideslider
      :is-show.sync="setVariableConf.isShow"
      :title="setVariableConf.title"
      :width="600"
      :quick-close="false"
      @hidden="hideSetVariable">
      <div slot="content" class="py-5 px-5" v-bkloading="{ isLoading: variableLoading }">
        <template v-if="variablesList.length">
          <div>
            <div class="bk-form-item text-[14px]">
              {{$t('变量：')}}
            </div>
            <div class="bk-form-item text-[12px]">
              <i18n path="可通过 {action} 创建更多作用在命名空间的变量">
                <button place="action" class="bk-text-button" @click="handleGoVar">{{$t('变量管理')}}</button>
              </i18n>
            </div>
            <div class="bk-form-item">
              <div class="bk-form-content">
                <div class="flex items-center mb-[10px]" v-for="(variable, index) in variablesList" :key="index">
                  <div class="flex-1" v-bk-tooltips="{ content: `${$t('变量名')}: ${variable.name}` }">
                    <bk-input disabled v-model="variable.key"></bk-input>
                  </div>
                  <span class="px-[5px]">=</span>
                  <bk-input class="flex-1" :placeholder="$t('值')" v-model="variable.value"></bk-input>
                </div>
              </div>
            </div>
            <div class="mt-[20px]">
              <bk-button type="primary" :loading="variableLoading" @click="updateVariable">
                {{$t('保存')}}
              </bk-button>
              <bk-button class="ml-[5px]" :loading="variableLoading" @click="hideSetVariable">
                {{$t('取消')}}
              </bk-button>
            </div>
          </div>
        </template>
      </div>
    </bcs-sideslider>
    <!-- 配额管理 -->
    <bcs-dialog
      v-model="setQuotaConf.isShow"
      :width="650"
      :title="$t('配额管理：{nsName}', { nsName: setQuotaConf.namespace })"
      @confirm="updateNamespace"
      @cancel="cancelUpdateNamespace">
      <bcs-form :label-width="200" form-type="vertical" v-bkloading="{ isLoading: setQuotaConf.loading }">
        <div class="mb-[20px]">
          {{$t('配额设置')}}
          <bk-switcher
            v-model="showQuota"
            class="ml-[10px]"
            :disabled="isSharedCluster"
            size="small"
            :selected="showQuota"
            :key="showQuota"
            @change="toggleShowQuota">
          </bk-switcher>
        </div>
        <bcs-form-item v-if="showQuota">
          <div class="flex mr-[20px]">
            <span class="mr-[10px]">CPU</span>
            <bcs-input
              v-model="setQuotaConf.quota.cpuRequests"
              class="w-[200px]"
              type="number"
              :min="1"
              :precision="0">
              <div class="group-text" slot="append">{{ $t('核') }}</div>
            </bcs-input>
            <span class="mx-[10px]">MEN</span>
            <bcs-input
              v-model="setQuotaConf.quota.memoryRequests"
              class="w-[200px]"
              type="number"
              :min="1"
              :precision="0">
              <div class="group-text" slot="append">Gi</div>
            </bcs-input>
          </div>
        </bcs-form-item>
      </bcs-form>
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
  </LayoutContent>
</template>

<script lang="ts">
import { defineComponent, computed, watch, ref } from '@vue/composition-api';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import LayoutContent from '@/components/layout/Content.vue';
import LayoutRow from '@/components/layout/Row.vue';
import Detail from './detail.vue';
import usePage from '../common/use-page';
import useSearch from '../common/use-search';
import { useCluster, useProject } from '@/common/use-app';
import { useNamespace } from './use-namespace';
import { timeZoneTransForm } from '@/common/util';
import { BCS_CLUSTER } from '@/common/constant';
import { CreateElement } from 'vue';
import useInterval from '@/views/dashboard/common/use-interval';

export default defineComponent({
  name: 'NamespaceList',
  components: {
    ClusterSelect,
    LayoutContent,
    LayoutRow,
    Detail,
  },
  setup(props, ctx) {
    const { $store, $bkInfo, $i18n, $bkMessage, $router, $route } = ctx.root;

    const { projectID } = useProject();
    const { isSharedCluster } = useCluster();

    const viewMode = computed(() => $store.state.viewMode);

    const curClusterId = computed(() => $store.state.curClusterId);

    const clusterID = ref(curClusterId.value);

    const keys = ref(['name']);

    const showQuota = ref(true);

    const toggleShowQuota = () => {
      showQuota.value = !showQuota.value;
    };

    const {
      variablesList,
      variableLoading,
      namespaceLoading,
      namespaceData,
      webAnnotations,
      handleGetVariablesList,
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
            content: data.column.property === 'cpuUseRate' ? $i18n.t('所有容器CPU limits总和 / CPU配额') : $i18n.t('所有容器内存 limits总和 / 内存配额'),
          },
        },
      ],
    }, [data.column.label]);

    const clusterList = computed(() => $store.state.cluster.clusterList || []);
    const projectCode = computed(() => $route.params.projectCode);
    const curCluster = computed(() => clusterList.value.find(item => item.clusterID === clusterID.value));

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
        title: $i18n.t('确认删除当前命名空间'),
        subTitle: $i18n.t('删除Namespace将销毁Namespace下的所有资源，销毁后所有数据将被清除且不可恢复，请提前备份好数据。'),
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
              message: $i18n.t('删除成功'),
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
      setVariableConf.value.title = $i18n.t('设置变量值：') + namespace;
      await handleGetVariablesList({
        $clusterId: clusterID.value,
        $namespace: namespace,
      });
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
          message: $i18n.t('保存成功'),
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

    const updateNamespace = async () => {
      const { namespace, labels, annotations, quota } = setQuotaConf.value;
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
          message: $i18n.t('更新成功'),
        });
        getNamespaceData({
          $clusterId: clusterID.value,
        });
        cancelUpdateNamespace();
      };
    };

    const cancelUpdateNamespace = () => {
      setQuotaConf.value = { ...initQuotaConf };
    };

    const handleToCreatedPage = () => {
      if (viewMode.value === 'dashboard') {
        $router.push({
          name: 'dashboardNamespaceCreate',
          query: {
            kind: 'Namespace',
          },
        });
      } else {
        $router.push({
          name: 'namespaceCreate',
          params: {
            clusterId: clusterID.value,
          },
        });
      }
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
      CREATE: $i18n.t('创建命名空间'),
      UPDATE: $i18n.t('配额调整'),
      DELETE: $i18n.t('删除命名空间'),
    };

    const handleToITSM = (link) => {
      window.open(link, '_blank');
    };

    // namespace itsmTicketType -> 操作按钮禁用状态控制
    const applyMap = (type: ('CREATE' | 'UPDATE' | 'DELETE' | '')) => {
      const typeMap = {
        CREATE: {
          setVar: true,
          setQuota: true,
          delete: true,
        },
        UPDATE: {
          setVar: false,
          setQuota: true,
          delete: true,
        },
        DELETE: {
          setVar: true,
          setQuota: true,
          delete: true,
        },
        '': {
          setVar: false,
          setQuota: false,
          delete: false,
        },
      };
      return typeMap[type];
    };

    const updateViewMode = () => {
      localStorage.setItem('FEATURE_CLUSTER', 'done');
      localStorage.setItem(BCS_CLUSTER, curCluster.value.cluster_id);
      sessionStorage.setItem(BCS_CLUSTER, curCluster.value.cluster_id);
      $store.commit('cluster/forceUpdateCurCluster', curCluster.value.cluster_id ? curCluster.value : {});
      $store.commit('updateCurClusterId', curCluster.value.cluster_id);
      $store.commit('updateViewMode', 'cluster');
      $store.dispatch('getFeatureFlag');
    };

    const handleGoVar = () => {
      if (viewMode.value === 'dashboard') {
        updateViewMode();
      };
      setVariableConf.value.isShow = false;
      $router.push({
        name: 'var',
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
    };
  },
});
</script>

<style lang="postcss" scoped>
  @import './namespace.css';
  >>> .custom-header-cell {
    text-decoration: underline;
    text-decoration-style: dashed;
    text-underline-position: under;
}
</style>
