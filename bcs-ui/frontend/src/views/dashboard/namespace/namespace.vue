<template>
  <LayoutContent hide-back :title="$tc('命名空间')"
    v-bkloading="{ isLoading: namespaceLoading }">
    <LayoutRow class="mb15">
      <template #left>
        <bcs-button class="w-[100px]" theme="primary" icon="plus" @click="handleToCreatedPage">{{ $t('创建') }}</bcs-button>
      </template>
      <template #right>
        <ClusterSelect
          v-if="!curCluesterId"
          v-model="searchScope"
          class="mr-[10px]"
          searchable
          :disabled="!!curCluesterId"
          @change="handleChangeCluester"
        ></ClusterSelect>
        <bcs-input class="search-input"
          right-icon="bk-icon icon-search"
          :placeholder="$t('搜索名称')"
          v-model="searchValue">
        </bcs-input>
      </template>
    </LayoutRow>
    <bcs-table :data="curPageData"
      :pagination="pagination"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange"
      @sort-change="handleSortChange">
      <bcs-table-column :label="$t('名称')" sortable prop="name">
        <template #default="{ row }">
          {{ row.name || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('状态')">
        <template #default="{ row }">
          {{ row.status || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('CPU使用率')" prop="cpuUseRate" :min-width="100">
        <template #default="{ row }">
          <bcs-round-progress
            ext-cls="biz-cluster-ring"
            width="50px"
            :percent="row.cpuUseRate"
            :config="{
              strokeWidth: 10,
              bgColor: '#f0f1f5',
              activeColor: '#3a84ff'
            }"
            :num-style="{
              fontSize: '12px'
            }"
          ></bcs-round-progress>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('内存使用率')" prop="memoryUseRate" :min-width="100">
        <template #default="{ row }">
          <bcs-round-progress
            ext-cls="biz-cluster-ring"
            width="50px"
            :percent="row.memoryUseRate"
            :config="{
              strokeWidth: 10,
              bgColor: '#f0f1f5',
              activeColor: '#3a84ff'
            }"
            :num-style="{
              fontSize: '12px'
            }"
          ></bcs-round-progress>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('操作')" width="220">
        <template #default="{ row }">
          <bk-button text class="mr-[10px]" @click="showSetVariable(row)">{{ $t('设置变量值') }}</bk-button>
          <bk-button text class="mr-[10px]" @click="showSetQuota(row)">{{ $t('配额管理') }}</bk-button>
          <bk-button text @click="handleDeleteNamespace(row)">{{ $t('删除') }}</bk-button>
        </template>
      </bcs-table-column>
    </bcs-table>
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
            <div class="bk-form-item">
              <div class="bk-form-content">
                <div class="flex items-center mb-[10px]" v-for="(variable, index) in variablesList" :key="index">
                  <bk-input :disabled="true" v-model="variable.key"></bk-input>
                  <span class="px-[5px]">=</span>
                  <bk-input :placeholder="$t('值')" v-model="variable.value"></bk-input>
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
    <bcs-dialog
      v-model="setQuotaConf.isShow"
      :width="650"
      :title="setQuotaConf.title"
      @confirm="updateNamespace"
      @cancel="cancelUpdateNamespace">
      <bcs-form :label-width="200" form-type="vertical">
        <bcs-form-item :label="$t('配额设置')" :required="true">
          <div class="flex">
            <div class="flex mr-[20px]">
              <span class="mr-[10px]">MEN</span>
              <bcs-input v-model="setQuotaConf.quota.memoryRequests" class="w-[200px]" type="number" :min="0">
                  <div class="group-text" slot="append">G</div>
              </bcs-input>
            </div>
            <div class="flex">
              <span class="mr-[10px]">CPU</span>
              <bcs-input v-model="setQuotaConf.quota.cpuRequests" class="w-[200px]" type="number" :min="0">
                  <div class="group-text" slot="append">{{ $t('核') }}</div>
              </bcs-input>
            </div>
          </div>
        </bcs-form-item>
      </bcs-form>
    </bcs-dialog>
  </LayoutContent>
  
</template>

<script lang="ts">
import { defineComponent, computed, watch, ref, onMounted } from '@vue/composition-api';
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import LayoutContent from '@/components/layout/Content.vue';
import LayoutRow from '@/components/layout/Row.vue';
import usePage from '../common/use-page';
import useSearch from '../common/use-search';
import { useNamespace } from './use-namespace';
import { sort } from '@/common/util';

export default defineComponent({
  name: 'NamespaceList',
  components: {
    ClusterSelect,
    LayoutContent,
    LayoutRow
  },
  setup(props, ctx) {
    const { $store, $bkInfo, $i18n, $bkMessage, $router } = ctx.root;

    const clusterList = computed(() => {
      return $store.state.cluster.clusterList;
    });

    const viewMode = computed(() => {
      return $store.state.viewMode
    })

    const curCluesterId = computed(() => {
      return $store.state.curClusterId;
    });

    const searchScope = ref('');
    watch(curCluesterId, () => {
      if (clusterList.value.length) {
        const clusterIds = clusterList.value.map(item => item.cluster_id);
        if (sessionStorage['bcs-cluster'] && clusterIds.includes(sessionStorage['bcs-cluster'])) {
          searchScope.value = sessionStorage['bcs-cluster'];
        } else {
          searchScope.value = curCluesterId.value || clusterList.value[0].cluster_id;
        };
      };
    }, { immediate: true });

    const keys = ref(['name']);

    // 初始化集群列表信息
    // useCluster(ctx)
    // 获取命名空间
    const { variablesList, variableLoading, namespaceLoading, namespaceData,
      handleGetVariablesList, handleDeleteNameSpace, handleUpdateNameSpace,
      handleUpdateVariablesList, getNamespaceData, } = useNamespace()

    // 排序
    const sortData = ref({
      prop: '',
      order: '',
    });
    const handleSortChange = (data) => {
      sortData.value = {
        prop: data.prop,
        order: data.order,
      };
    };

    // 表格数据
    const tableData = computed(() => {
      const items = JSON.parse(JSON.stringify(namespaceData.value || []));
      const { prop, order } = sortData.value;
      return prop ? sort(items, prop, order) : items;
    });
    // 搜索功能
    const { tableDataMatchSearch, searchValue } = useSearch(tableData, keys);

    // 分页
    const { pagination, curPageData, pageConf, pageChange, pageSizeChange } = usePage(tableDataMatchSearch);
    // 搜索时重置分页
    watch(searchValue, () => {
      pageConf.current = 1;
    });

    // 删除命名空间
    const handleDeleteNamespace = (row) => {
      const namespace = row?.name;
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('确认删除当前命名空间'),
        subTitle: $i18n.t('删除Namespace将销毁Namespace下的所有资源，销毁后所有数据将被清除且不可恢复，请提前备份好数据。'),
        defaultInfo: true,
        confirmFn: async () => {
          const result = await handleDeleteNameSpace({
            $clusterId: searchScope.value,
            $namespace: namespace,
          })
          result && getNamespaceData({
            $clusterId: searchScope.value,
          })
          result && $bkMessage({
            theme: 'success',
            message: $i18n.t('删除成功'),
          });
        },
      });
    };

    const setVariableConf = ref({
      isShow: false,
      title: '',
      namespace: '',
    })

    // 设置变量值
    const showSetVariable = async (row) => {
      const namespace = row?.name;
      setVariableConf.value.isShow = true;
      setVariableConf.value.namespace = namespace;
      setVariableConf.value.title = $i18n.t('设置变量值：') + namespace;
      await handleGetVariablesList({
        $clusterId: searchScope.value,
        $namespace: namespace,
      })
    };

    const hideSetVariable = () => {
      variablesList.value = [];
      setVariableConf.value.isShow = false;
      setVariableConf.value.title = '';
    };

    // 更新变量值
    const updateVariable = async () => {
      const result = await handleUpdateVariablesList({
        $clusterId: searchScope.value,
        $namespace: setVariableConf.value.namespace,
        data: variablesList.value,
      });
      result && $bkMessage({
        theme: 'success',
        message: $i18n.t('保存成功'),
      });
      result && hideSetVariable()
    };

    const setQuotaConf = ref({
      isShow: false,
      namespace: '',
      title: '',
      labels: [],
      annotations: [],
      quota: {
        cpuLimits: '',
        cpuRequests: '',
        memoryLimits: '',
        memoryRequests: '',
      },
    });


    // 设置配额
    const showSetQuota = (row) => {
      setQuotaConf.value.title = $i18n.t('配额管理：') + row.name;
      setQuotaConf.value.isShow = true;
      setQuotaConf.value.namespace = row.name;
      setQuotaConf.value.labels = row.labels;
      setQuotaConf.value.annotations = row.annotations;
      if (row.quota) {
        setQuotaConf.value.quota = Object.assign({}, {
          cpuLimits: row.quota.cpuLimits,
          cpuRequests: uintConversion(row.quota.cpuRequests),
          memoryLimits: row.quota.memoryLimits,
          memoryRequests: uintConversion(row.quota.memoryRequests),
        });
      };
    };

    const updateNamespace = async () => {
      const { namespace, labels, annotations, quota } = setQuotaConf.value
      const result = await handleUpdateNameSpace({
        $clusterId: searchScope.value,
        $namespace: namespace,
        labels,
        annotations,
        quota: {
          cpuLimits: String(quota.cpuRequests),
          cpuRequests: String(quota.cpuRequests),
          memoryLimits: quota.memoryRequests + 'Gi',
          memoryRequests: quota.memoryRequests + 'Gi',
        },
      });
      result && $bkMessage({
        theme: 'success',
        message: $i18n.t('更新成功'),
      });
      result && getNamespaceData({
        $clusterId: searchScope.value,
      });
      result && cancelUpdateNamespace();
    };

    const cancelUpdateNamespace = () => {
      setQuotaConf.value.quota = {
        cpuLimits: '',
        cpuRequests: '',
        memoryLimits: '',
        memoryRequests: '',
      };
      setQuotaConf.value.labels = [];
      setQuotaConf.value.annotations = [];
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
            clusterId: searchScope.value,
          },
        })
      }
    };

    const handleChangeCluester = (id) => {
      searchScope.value = id;
      getNamespaceData({
        $clusterId: searchScope.value,
      });
    };

    const uintConversion = (val) => {
      const total = val.match(/\d+/gi)[0];
      const uint = val.match(/[a-z|A-Z]+/gi)[0] || '';
      let result = 0;
      switch (uint) {
        case 'm':
          result = total / Math.pow(10, 10);
          break;

        case 'k':
          result = total / Math.pow(10, 6);
          break;

        case 'M':
          result = total / Math.pow(10, 3);
          break;

        case 'Gi':
          result = total
          break;

        case 'T':
          result = total * Math.pow(10, 3);
          break;

        case 'P':
          result = total * Math.pow(10, 6);
          break;
      };
      return result + '';
    };

    onMounted(() => {
      getNamespaceData({
        $clusterId: searchScope.value,
      });
    });

    return {
      namespaceLoading,
      curCluesterId,
      searchScope,
      pagination,
      searchValue,
      curPageData,
      setVariableConf,
      setQuotaConf,
      variablesList,
      variableLoading,
      uintConversion,
      pageChange,
      pageSizeChange,
      handleSortChange,
      handleDeleteNamespace,
      showSetVariable,
      hideSetVariable,
      updateVariable,
      showSetQuota,
      updateNamespace,
      cancelUpdateNamespace,
      handleToCreatedPage,
      handleChangeCluester,
    };
  },
});
</script>

<style lang="postcss" scoped>
  @import './namespace.css';
</style>

