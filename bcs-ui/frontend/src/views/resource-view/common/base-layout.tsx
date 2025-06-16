/* eslint-disable camelcase */
import yamljs from 'js-yaml';
import jp from 'jsonpath';
import { isEqual } from 'lodash';
import { computed, defineComponent, onBeforeUnmount, PropType, provide, reactive, ref, toRef, toRefs, watch } from 'vue';

import NSSelect from '../view-manage/ns-select.vue';
import useViewConfig from '../view-manage/use-view-config';
import Rollback from '../workload/rollback.vue';

import ConfirmContent from './confirm-content.vue';
import CreateResource from './create-resource.vue';
import useSearch from './use-search';
import { ISubscribeData } from './use-subscribe';
import useTableData from './use-table-data';

import { restartGameWorkloads, restartWorkloads } from '@/api/modules/cluster-resource';
import $bkMessage from '@/common/bkmagic';
import { bus } from '@/common/bus';
import ContentHeader from '@/components/layout/Header.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import { useCluster } from '@/composables/use-app';
import useInterval from '@/composables/use-interval';
import fullScreen from '@/directives/full-screen';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import { INamespace, isEntry, useNamespace } from '@/views/cluster-manage/namespace/use-namespace';

export default defineComponent({
  name: 'BaseLayout',
  components: { CodeEditor },
  directives: {
    'full-screen': fullScreen,
  },
  props: {
    title: {
      type: String,
      default: '',
      required: true,
    },
    // 父分类（crd类型的需要特殊处理），eg: workloads、networks（注意复数）
    type: {
      type: String,
      default: '',
      required: true,
    },
    // 子分类，eg: deployments、ingresses
    category: {
      type: String,
      default: '',
    },
    // 轮询时类型（type为crd时，kind仅作为资源详情展示的title用），eg: Deployment、Ingress（注意首字母大写）
    kind: {
      type: String,
      default: '',
      required: true,
    },
    // 是否显示创建资源按钮
    showCreate: {
      type: Boolean,
      default: true,
    },
    // 默认CRD值
    crd: {
      type: String,
      default: '',
    },
    // CRD资源的作用域
    scope: {
      type: String as PropType<'Namespaced'|'Cluster'>,
      default: '',
    },
    // 是否显示总览和yaml切换的tab
    showDetailTab: {
      type: Boolean,
      default: true,
    },
    // 默认展示详情标签
    defaultActiveDetailType: {
      type: String,
      default: 'overview',
    },
  },
  setup(props) {
    const {
      type,
      category,
      kind,
      defaultActiveDetailType,
      crd,
      scope,
    } = toRefs(props);
    const { clusterNameMap } = useCluster();
    const isViewConfigShow = computed(() => $store.state.isViewConfigShow);
    const { curViewData, isViewEditable, isClusterMode, dashboardViewID } = useViewConfig();
    const {
      searchSelectData,
      searchSelectChange,
      searchSelectValue,
      searchSelectKey,
    } = useSearch();

    const updateStrategyMap = ref({
      RollingUpdate: $i18n.t('k8s.updateStrategy.rollingUpdate'),
      InplaceUpdate: $i18n.t('k8s.updateStrategy.inplaceUpdate'),
      OnDelete: $i18n.t('k8s.updateStrategy.onDelete'),
      Recreate: $i18n.t('k8s.updateStrategy.reCreate'),
    });

    // 来源类型
    const sourceTypeMap = ref({
      Template: {
        iconClass: 'bcs-icon bcs-icon-templete',
        iconText: 'Template',
      },
      Helm: {
        iconClass: 'bcs-icon bcs-icon-helm',
        iconText: 'Helm',
      },
      Client: {
        iconClass: 'bcs-icon bcs-icon-client',
        iconText: 'Client',
      },
      Web: {
        iconClass: 'bcs-icon bcs-icon-web',
        iconText: 'Web',
      },
    });

    const renderCrdHeader = (h, { column }) => {
      const additionalData = additionalColumns.value.find(item => item.name === column.label);
      return h('span', {
        directives: [
          {
            name: 'bk-tooltips',
            value: {
              content: additionalData?.description || column.label,
              placement: 'top',
              boundary: 'window',
            },
          },
        ],
      }, [column.label]);
    };
    const getJsonPathValue = (row, path: string) => {
      try {
        return jp.value(row, path?.indexOf('$') === 0 ? path : `$.${path}`);
      } catch (_) {
        return undefined;
      }
    };
    // 状态
    const statusMap = {
      normal: $i18n.t('generic.status.ready'),
      creating: $i18n.t('generic.status.creating'),
      updating: $i18n.t('generic.status.updating'),
      deleting: $i18n.t('generic.status.deleting'),
      restarting: $i18n.t('generic.status.restarting'),
    };
    const statusFilters = computed(() => Object.keys(statusMap).map(key => ({
      text: statusMap[key],
      value: key,
    })));

    const pageConf = ref({
      current: 1,
      limit: $store.state.globalPageSize,
      showTotalCount: true,
      count: 0,
    });
    const handlePageChange = (page: number) => {
      pageConf.value.current = page;
      handleGetTableData();
    };
    const handlePageSizeChange = (size: number) => {
      pageConf.value.current = 1;
      pageConf.value.limit = size;
      $store.commit('updatePageSize', size);
      handleGetTableData();
    };
    // 排序
    const sortData = ref<{
      sortBy: string
      order: 'desc' | 'asc' | ''
    }>({
      sortBy: '',
      order: '',
    });
    const propMap = {
      'metadata.name': 'name',
      'metadata.namespace': 'namespace',
      createTime: 'age',
    };
    const handleSortChange = ({ prop, order }) => {
      sortData.value = {
        sortBy: propMap[prop] || prop,
        order: order === 'ascending' ? 'asc' : 'desc',
      };
      handleGetTableData();
    };
    // 表头过滤
    const customFilters = ref<Record<string, string[]>>({});
    const filters = ref<Record<string, string[]>>({});
    const handleFilterChange = (data) => {
      filters.value = data;
      handleGetTableData();
    };

    // 集群级别CRD
    const isClusterScopeCRD = computed(() => type.value === 'crd' && scope.value !== 'Namespaced');
    // 集群ID（集群视图下才会有）
    const clusterID = computed(() => $router.currentRoute?.params?.clusterId);
    // 显示过滤条件
    const showFilter = computed(() => !isClusterScopeCRD.value && isClusterMode.value);
    // 命名空间变更
    const curNsList = computed(() => $store.state.viewNsList);
    const handleNsChange = (nsList: string[]) => {
      if (!isClusterMode.value) return; // 自定义视图模式下命名空间只能在配置面板修改

      $store.commit('updateViewNsList', nsList);
    };

    // 表格数据
    const data = ref<ISubscribeData>({
      manifestExt: {},
      manifest: {},
      total: 0,
    });
    const applyURL = computed(() => data.value.perms?.applyURL);
    const {
      isLoading,
      webAnnotations,
      getMultiClusterResources,
      getMultiClusterResourcesCRD,
    } = useTableData();
    const curPageData = computed(() => data.value.manifest?.items || []);
    // 动态表格字段
    const additionalColumns = computed(() => webAnnotations.value.additionalColumns || []);
    // 获取表格数据
    const handleGetTableData = async (loading = true) => {
      if (!curViewData.value) return;

      isLoading.value = loading;
      let resourceData: ISubscribeData;

      if (type.value === 'crd') {
        // 自定义资源
        resourceData = await getMultiClusterResourcesCRD({
          ...curViewData.value,
          ...sortData.value,
          status: filters.value.status || [],
          $crd: crd.value,
          offset: (pageConf.value.current - 1) * pageConf.value.limit,
          limit: pageConf.value.limit,
        });
        // 设置资源数量（批量获取的接口比较慢，这里单个资源先出来就先回显数量）
        if (['GameDeployment', 'GameStatefulSet', 'HookTemplate'].includes(kind.value)) {
          bus.$emit('set-resource-count', kind.value, resourceData.total);
        }
      } else {
        // 普通资源
        resourceData = await getMultiClusterResources({
          ...curViewData.value,
          ...sortData.value,
          status: filters.value.status || [],
          ip: customFilters.value.ip?.join(','),
          $kind: kind.value,
          offset: (pageConf.value.current - 1) * pageConf.value.limit,
          limit: pageConf.value.limit,
        });
        // 设置资源数量（批量获取的接口比较慢，这里单个资源先出来就先回显数量）
        bus.$emit('set-resource-count', kind.value, resourceData.total);
      }
      pageConf.value.count = resourceData.total;
      data.value = resourceData;
      isLoading.value = false;
    };
    // 重新搜索
    watch(curViewData, (newValue, oldValue) => {
      if (!curViewData.value || isEqual(newValue, oldValue)) return;
      pageConf.value.current = 1;
      handleGetTableData();
    }, { deep: true });

    // 获取额外字段方法
    const handleGetExtData = (uid: string, ext?: string, defaultData?: any) => {
      const extData = data.value.manifestExt?.[uid] || {};
      return ext ? (extData[ext] || defaultData) : extData;
    };

    // 跳转详情界面
    const gotoDetail = ($event, url, row) => {
      if (isViewEditable.value) return;
      // 检测是否按下了 Ctrl 键
      if ($event.ctrlKey || $event.metaKey) {
        window.open(url, '_blank', 'noopener,noreferrer');
      } else {
        $router.push({
          name: 'dashboardWorkloadDetail',
          params: {
            category: category.value,
            name: row.metadata.name,
            namespace: row.metadata.namespace,
            clusterId: handleGetExtData(row.metadata.uid, 'clusterID'),
          },
          query: {
            kind: kind.value,
            crd: crd.value,
            viewID: dashboardViewID.value,
          },
        });
      }
    };

    const resolveLink = (row) => {
      if (isViewEditable.value) return 'javascript:void(0)';
      const { href } = $router.resolve({
        name: 'dashboardWorkloadDetail',
        params: {
          category: category.value,
          name: row.metadata.name,
          namespace: row.metadata.namespace,
          clusterId: handleGetExtData(row.metadata.uid, 'clusterID'),
        },
        query: {
          kind: kind.value,
          crd: crd.value,
          viewID: dashboardViewID.value,
        },
      });
      return href;
    };

    // 跳转命名空间
    const goNamespace = (row) => {
      const { href } = $router.resolve({
        name: 'clusterMain',
        // params: {
        //   namespace: row?.metadata?.namespace,
        // },
        query: {
          clusterId: handleGetExtData(row?.metadata?.uid, 'clusterID'),
          active: 'namespace',
          namespace: row?.metadata?.namespace,
        },
      });
      window.open(href);
    };

    // 详情侧栏
    const showDetailPanel = ref(false);
    // 当前详情行数据
    const curDetailRow = ref<{
      data: any
      extData: any
    }>({
      data: {},
      extData: {},
    });
    // 侧栏展示类型
    const detailType = ref({
      active: defaultActiveDetailType.value,
      list: [
        {
          id: 'overview',
          name: $i18n.t('dashboard.title.overview'),
        },
        {
          id: 'yaml',
          name: 'YAML',
        },
      ],
    });
    // 资源详情(列表数据不全)
    const detailLoading = ref(false);
    const handleGetResourceDetail = async ({ namespace, name, clusterID }) => {
      detailLoading.value = true;
      const res = await $store.dispatch('dashboard/getResourceDetail', {
        $namespaceId: namespace,
        $category: props.category,
        $name: name,
        $type: props.type,
        $clusterId: clusterID,
      });
      detailLoading.value = false;
      return res.data?.manifest;
    };
    // 自定义资源详情
    const handleGetCustomObjectDetail = async ({ namespace, name, clusterID }) => {
      detailLoading.value = true;
      const res = await $store.dispatch('dashboard/getCustomObjectResourceDetail', {
        $crdName: crd.value,
        $namespaceId: namespace,
        $name: name,
        $clusterId: clusterID,
      });
      detailLoading.value = false;
      return res.data?.manifest;
    };

    // 显示侧栏详情
    const handleShowDetail = async (row) => {
      curDetailRow.value.extData = handleGetExtData(row.metadata.uid);
      curDetailRow.value.data = row;// 先设置当前行数据（防止详情页时data为空）
      showDetailPanel.value = true;

      // 从详情接口中获取全量数据
      if (category.value === 'custom_objects') {
        const namespace = scope.value === 'Namespaced' ? row?.metadata?.namespace : '';
        curDetailRow.value.data = await handleGetCustomObjectDetail({
          name: row?.metadata?.name,
          namespace,
          clusterID: curDetailRow.value.extData?.clusterID,
        });
      } else {
        curDetailRow.value.data = await handleGetResourceDetail({
          name: row?.metadata?.name,
          namespace: row?.metadata?.namespace,
          clusterID: curDetailRow.value.extData?.clusterID,
        });
      }
    };

    const showCapacityDialog = ref(false);
    const replicas = ref(0);
    // 显示扩缩容弹框
    const handleEnlargeCapacity = (row) => {
      curDetailRow.value.data = row;
      replicas.value = row.spec.replicas;
      showCapacityDialog.value = true;
    };
    // 确定扩缩容
    const handleConfirmChangeCapacity = async () => {
      let result = false;
      const { name, namespace, uid } = curDetailRow.value.data?.metadata || {};
      if (type.value === 'crd') {
        result = await $store.dispatch('dashboard/crdEnlargeCapacityChange', {
          $crdName: crd.value,
          $cobjName: name,
          $clusterId: handleGetExtData(uid, 'clusterID'),
          replicas: replicas.value,
          namespace,
        });
      } else {
        result = await $store.dispatch('dashboard/enlargeCapacityChange', {
          $namespace: namespace,
          $category: category.value,
          $name: name,
          $clusterId: handleGetExtData(uid, 'clusterID'),
          replicas: replicas.value,
        });
      }
      result && $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.modify'),
      });
      handleGetTableData();
    };

    // 切换详情类型
    const handleChangeDetailType = (type) => {
      detailType.value.active = type;
    };
    // 重置详情类型
    watch(showDetailPanel, () => {
      handleChangeDetailType(defaultActiveDetailType.value);
    });
    // yaml内容
    const yaml = computed(() => {
      // 特殊处理-> apiVersion、kind、metadata强制排序在前三位
      const newDetailRow = {
        apiVersion: curDetailRow.value.data.apiVersion,
        kind: curDetailRow.value.data.kind,
        metadata: curDetailRow.value.data.metadata,
        ...curDetailRow.value.data,
      };
      return yamljs.dump(newDetailRow || {});
    });
    // 创建资源
    const showCreateDialog = ref(false);
    const formUpdate = computed(() => webAnnotations.value?.featureFlag?.FORM_CREATE);

    // 更新资源
    const handleUpdateResource = (row) => {
      const { name, namespace, uid } = row.metadata || {};
      const editMode = handleGetExtData(uid, 'editMode');
      if (editMode === 'yaml') {
        $router.push({
          name: 'dashboardResourceUpdate',
          params: {
            namespace,
            name,
            clusterId: handleGetExtData(uid, 'clusterID'),
          },
          query: {
            type: type.value,
            category: category.value,
            kind: kind.value,
            crd: crd.value,
          },
        });
      } else {
        $router.push({
          name: 'dashboardFormResourceUpdate',
          params: {
            namespace,
            name,
            clusterId: handleGetExtData(uid, 'clusterID'),
          },
          query: {
            type: type.value,
            category: category.value,
            kind: kind.value,
            crd: crd.value,
            formUpdate: webAnnotations.value?.featureFlag?.FORM_CREATE,
          },
        });
      }
    };

    // 确认对话框
    const confirmDialog = ref({
      show: false,
      title: '',
      type: '',
      loading: false,
      confirmText: '',
    });
    const confirmFn = async () => {
      confirmDialog.value.loading = true;
      if (confirmDialog.value.type === 'delete') {
        await confirmDelete();
      } else if (confirmDialog.value.type === 'restart') {
        await confirmRestart();
      }
      confirmDialog.value.loading = false;
      confirmDialog.value.show = false;
    };

    // 删除资源
    const handleDeleteResource = (row) => {
      curDetailRow.value.data = row;
      confirmDialog.value.title = $i18n.t('dashboard.title.confirmDelete');
      confirmDialog.value.confirmText = $i18n.t('generic.button.delete');
      confirmDialog.value.type = 'delete';
      confirmDialog.value.show = true;
    };
    const confirmDelete = async () => {
      const { name, namespace, uid } = curDetailRow.value.data?.metadata || {};
      let result = false;
      if (type.value === 'crd') {
        result = await $store.dispatch('dashboard/customResourceDelete', {
          namespace,
          $crd: crd.value,
          $category: category.value,
          $clusterId: handleGetExtData(uid, 'clusterID'),
          $name: name,
        });
      } else {
        result = await $store.dispatch('dashboard/resourceDelete', {
          $namespaceId: namespace,
          $type: type.value,
          $category: category.value,
          $clusterId: handleGetExtData(uid, 'clusterID'),
          $name: name,
        });
      };
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.delete'),
        });
        handleGetTableData();
      }
    };
    // 更新记录
    const handleGotoUpdateRecord = (row) => {
      const { name, namespace, uid } = row.metadata || {};
      $router.push({
        name: 'workloadRecord',
        params: {
          name,
          namespace,
          category: category.value,
          clusterId: handleGetExtData(uid, 'clusterID'),
        },
        query: {
          kind: kind.value,
          crd: crd.value,
        },
      });
    };
    // 回滚
    const showRollbackSideslider = ref(false);
    const curRow = ref({
      metadata: {
        name: '',
        namespace: '',
        uid: '',
      },
    });
    const handleRollback = (row) => {
      curRow.value = row;
      showRollbackSideslider.value = true;
    };
    const handleRollbackSidesilderHide = () => {
      showRollbackSideslider.value = false;
    };

    // 滚动重启
    const handleRestart = (row, actionName: string) => {
      curDetailRow.value.data = row;
      confirmDialog.value.title = $i18n.t('dashboard.title.confirmSchedule', { action: actionName ||  $i18n.t('dashboard.workload.button.restart') });
      confirmDialog.value.confirmText = actionName;
      confirmDialog.value.type = 'restart';
      confirmDialog.value.show = true;
    };
    const confirmRestart = async () => {
      const { name, namespace, uid } = curDetailRow.value.data?.metadata || {};
      let result = false;
      if (category.value === 'custom_objects') {
        result = await restartGameWorkloads({
          $crd: crd.value,
          $type: type.value,
          $category: category.value,
          $clusterId: handleGetExtData(uid, 'clusterID'),
          $name: name,
          namespace,
        }).then(() => true)
          .catch(() => false);
      } else {
        result = await restartWorkloads({
          $namespaceId: namespace,
          $type: type.value,
          $category: category.value,
          $clusterId: handleGetExtData(uid, 'clusterID'),
          $name: name,
        }).then(() => true)
          .catch(() => false);
      }

      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.restart'),
        });
        handleGetTableData();
      }
    };

    // 显示视图配置面板
    const handleShowViewConfig = () => {
      bus.$emit('toggle-show-view-config');
    };

    const { start, stop } = useInterval(() => handleGetTableData(false), 5000);

    // 通过provide暴露方法
    provide('handleGetExtData', handleGetExtData);

    // 命名空间逻辑
    const { clusterList, curClusterId } = useCluster();
    const currentRoute = computed(() => toRef(reactive($router), 'currentRoute').value);
    const { getNamespaceData } = useNamespace();
    const nsList = ref<Array<INamespace>>([]);
    const handleGetNsData = async () => {
      const exist = clusterList.value.find(item => item.clusterID === curClusterId.value);
      if (!exist) return;
      nsList.value = await getNamespaceData({ $clusterId: curClusterId.value });
    };
    watch(curClusterId, async () => {
      isLoading.value = true;
      await handleGetNsData();
      // 首次加载
      if (!isEntry.value && nsList.value[0]?.name && !currentRoute.value?.query?.namespace && !curNsList.value.length) {
        $store.commit('updateViewNsList', [nsList.value[0].name]);
      } else {
        // 切换集群，判断当前选中的命名空间是否在新的命名空间列表中
        const newNs = curNsList.value?.reduce<string[]>((acc, cur) => {
          const ns = nsList.value.find(item => item.name === cur);
          if (ns) {
            acc.push(cur);
          }
          return acc;
        }, []);
        $store.commit('updateViewNsList', newNs);
      }
      isEntry.value = true;

      if (isClosed.value) {
        isLoading.value = false;
        return;
      }
      // isLoading.value = true;
      await handleGetTableData();
      isLoading.value = false;
      // 轮询资源，频繁切换资源，即使页面卸载，这个回调也会执行，导致存在多个定时器，使用isClosed来控制
      start();
    }, { immediate: true });

    // 同步命名空间到query
    watch(curNsList, () => {
      const urlQuery = $router.currentRoute?.query || {};
      // 不是集群模式 / 命名空间未改变 / 清空命名空间，直接返回
      if (!isClusterMode.value
        || urlQuery.namespace === curNsList.value.join(',')
        || (!urlQuery.namespace && !curNsList.value.join(','))) return;
      const data = {
        ...urlQuery,
        namespace: curNsList.value.join(','),
      };
      // 删除值为空的参数
      Object.keys(data).forEach((key) => {
        if (!data[key]) {
          delete data[key];
        }
      });
      $router.replace({
        query: {
          ...data,
        },
      });
    });

    const isClosed = ref(false);
    onBeforeUnmount(() => {
      stop();
      isClosed.value = true;
    });

    return {
      customFilters, // hack 暴露过滤参数，其他组件会手动修改这个值
      clusterNameMap,
      updateStrategyMap,
      showDetailPanel,
      curDetailRow,
      replicas,
      yaml,
      detailType,
      isLoading,
      pageConf,
      curPageData,
      additionalColumns,
      webAnnotations,
      statusMap,
      showCapacityDialog,
      getJsonPathValue,
      renderCrdHeader,
      handlePageChange,
      handlePageSizeChange,
      handleGetExtData,
      handleSortChange,
      handleFilterChange,
      gotoDetail,
      handleShowDetail,
      handleChangeDetailType,
      handleUpdateResource,
      handleDeleteResource,
      handleEnlargeCapacity,
      handleConfirmChangeCapacity,
      statusFilters,
      handleShowViewConfig,
      handleGotoUpdateRecord,
      handleRestart,
      showRollbackSideslider,
      curRow,
      handleRollback,
      handleRollbackSidesilderHide,
      goNamespace,
      showCreateDialog,
      formUpdate,
      handleGetTableData,
      confirmDialog,
      confirmFn,
      isViewEditable,
      showFilter,
      clusterID,
      curNsList,
      handleNsChange,
      isViewConfigShow,
      isClusterMode,
      searchSelectData,
      searchSelectChange,
      searchSelectValue,
      searchSelectKey,
      detailLoading,
      sourceTypeMap,
      applyURL,
      resolveLink,
    };
  },
  render() {
    // 渲染筛选条件
    const renderSearch = () => {
      if (this.isViewEditable) return undefined;
      if (this.$scopedSlots?.search) return this.$scopedSlots?.search({
        handleShowViewConfig: this.handleShowViewConfig,
        clusterID: this.clusterID,
        handleNsChange: this.handleNsChange,
        curNsList: this.curNsList,
        showFilter: this.showFilter,
        isViewConfigShow: this.isViewConfigShow,
      });
      if (this.showFilter) {
        return (
          <div class="flex items-start justify-end pl-[24px] flex-1 text-[12px] h-[32px] z-10">
            <span class={[
              'inline-flex items-center justify-center bg-[#fff] w-[32px] h-[32px] mr-[8px]',
              'border border-solid border-[#C4C6CC] rounded-sm cursor-pointer',
              this.isViewConfigShow ? '!border-[#3a84ff] text-[#3a84ff]' : 'text-[#979BA5] hover:!border-[#979BA5]',
            ]}
            v-bk-trace_click={{
              module: 'view',
              operation: 'filter2',
              desc: '视图筛选按钮2',
              username: $store.state.user.username,
              projectCode: $store.getters.curProjectCode,
            }}
            onClick={this.handleShowViewConfig}>
              <i class="bk-icon icon-funnel text-[14px]"></i>
            </span>
            <span class={[
              'inline-flex items-center justify-center h-[32px] px-[8px]',
              'border border-solid border-[#C4C6CC] rounded-l-sm bg-[#FAFBFD] mr-[-1px]',
            ]}>
              {this.$t('k8s.namespace')}
            </span>
            <NSSelect
              value={this.curNsList}
              clusterId={this.clusterID}
              class="flex-1 bg-[#fff] max-w-[240px] mr-[8px]"
              displayTag={true}
              { ...{ on: { change: this.handleNsChange } }}/>
            <bcs-search-select
              class="flex-1 bg-[#fff] max-w-[460px]"
              clearable
              show-condition={false}
              show-popover-tag-change
              data={this.searchSelectData}
              values={this.searchSelectValue}
              placeholder={this.$t('view.placeholder.searchNameOrCreator')}
              key={this.searchSelectKey}
              onChange={this.searchSelectChange}
              onClear={() => this.searchSelectChange()} />
          </div>
        );
      }
      return undefined;
    };

    return (
      <div class="flex flex-col relative h-full"
        v-bkloading={{ isLoading: this.isLoading, opacity: 1, color: '#f5f7fa' }}>
        <ContentHeader
          class="flex-[0_0_auto] !h-[66px] !border-b-0 !shadow-none !bg-inherit"
          {
            ...{
              scopedSlots: {
                right: () => renderSearch(),
              },
            }
          }>
          <div class="flex items-center">
            <span class="text-[16px] text-[#313238] font-bold leading-[18px]">{this.title}</span>
            {
              !this.isViewEditable ? (
                <span class="inline-flex items-center">
                  <bcs-divider direction="vertical"></bcs-divider>
                  <bk-button text onClick={() => this.showCreateDialog = true}>
                    <span class="flex items-center">
                      <i class="flex items-center justify-center w-[14px] !leading-[16px] !top-0 bk-icon icon-plus-circle text-[#3a84ff] text-[14px]"></i>
                      <span class="flex !leading-[16px] text-[12px] ml-[4px]">{this.$t('generic.button.create')}</span>
                    </span>
                  </bk-button>
                </span>
              ) : null
            }
          </div>
        </ContentHeader>
        {
          this.applyURL ? (
            <bk-alert
              class="mx-[24px] mb-[16px]"
              type="warning"
              {
                ...{
                  scopedSlots: {
                    title: () => (<i18n path="dashboard.workload.tips.noPerms">
                      <span class="text-[#3a84ff] ml-[2px] cursor-pointer" onClick={() => window.open(this.applyURL)}>
                        {this.$t('dashboard.workload.button.applyURL')}
                      </span>
                    </i18n>),
                  },
                }
              }>
            </bk-alert>
          ) : null
        }
        <div class="dashboard-content flex-1 px-[24px] pb-[16px] overflow-auto">
          {
              this.$scopedSlots.default?.({
                clusterNameMap: this.clusterNameMap,
                isLoading: this.isLoading,
                pageConf: this.pageConf,
                curPageData: this.curPageData,
                statusMap: this.statusMap,
                handlePageChange: this.handlePageChange,
                handlePageSizeChange: this.handlePageSizeChange,
                handleGetExtData: this.handleGetExtData,
                handleSortChange: this.handleSortChange,
                handleFilterChange: this.handleFilterChange,
                gotoDetail: this.gotoDetail,
                handleShowDetail: this.handleShowDetail,
                handleEnlargeCapacity: this.handleEnlargeCapacity,
                handleUpdateResource: this.handleUpdateResource,
                handleDeleteResource: this.handleDeleteResource,
                getJsonPathValue: this.getJsonPathValue,
                renderCrdHeader: this.renderCrdHeader,
                additionalColumns: this.additionalColumns,
                webAnnotations: this.webAnnotations,
                updateStrategyMap: this.updateStrategyMap,
                statusFilters: this.statusFilters,
                handleShowViewConfig: this.handleShowViewConfig,
                handleGotoUpdateRecord: this.handleGotoUpdateRecord,
                handleRestart: this.handleRestart,
                handleRollback: this.handleRollback,
                goNamespace: this.goNamespace,
                isViewEditable: this.isViewEditable,
                isClusterMode: this.isClusterMode,
                sourceTypeMap: this.sourceTypeMap,
                resolveLink: this.resolveLink,
              })
          }
        </div>
        <bcs-sideslider
            quick-close
            isShow={this.showDetailPanel}
            width={800}
            {
            ...{
              on: {
                'update:isShow': (show: boolean) => {
                  this.showDetailPanel = show;
                },
              },
              scopedSlots: {
                header: () => (
                  <div class="flex items-center justify-between pr-[30px]">
                      <span>{this.curDetailRow.data?.metadata?.name}</span>
                      {
                        this.showDetailTab
                          ? (<div class="bk-button-group">
                                {
                                  this.detailType.list.map(item => (
                                    <bk-button class={{ 'is-selected': this.detailType.active === item.id }}
                                        onClick={() => {
                                          this.handleChangeDetailType(item.id);
                                        }}>
                                        {item.name}
                                    </bk-button>
                                  ))
                                }
                            </div>)
                          : null
                      }
                  </div>
                ),
                content: () => <div class="h-[calc(100vh-52px)] overflow-auto" v-bkloading={{ isLoading: this.detailLoading }}>
                  {
                    (this.detailType.active === 'overview'
                      ? (this.$scopedSlots.detail?.({
                        ...this.curDetailRow,
                      }))
                      : <CodeEditor
                      v-full-screen={{ tools: ['fullscreen', 'copy'], content: this.yaml }}
                      options={{
                        roundedSelection: false,
                        scrollBeyondLastLine: false,
                        renderLineHighlight: 'none',
                      }}
                      width="100%" height="100%" lang="yaml"
                      readonly={true} value={this.yaml} />)
                  }
                </div>,
              },
            }
            }>
        </bcs-sideslider>
        <bcs-dialog
          v-model={this.showCapacityDialog}
          mask-close={false}
          title={this.$t('deploy.templateset.scale')}
          header-position="left"
          on-confirm={this.handleConfirmChangeCapacity}
          width={480}
        >
          <bk-form label-width={100}>
            <div class="bg-[#F5F7FA] py-[8px]">
              <bk-form-item label={this.$t('cluster.labels.name')}>
                {this.clusterNameMap[this.handleGetExtData(this.curDetailRow?.data?.metadata?.uid, 'clusterID')]}
              </bk-form-item>
              <bk-form-item label={this.$t('k8s.namespace')} class="!mt-0">
                {this.curDetailRow?.data?.metadata?.namespace}
              </bk-form-item>
              <bk-form-item label={this.$t('view.labels.resourceName')} class="!mt-0">
                {this.curDetailRow?.data?.metadata?.name}
              </bk-form-item>
            </div>
            <bk-form-item label={this.$t('dashboard.workload.label.scaleNum')} class="mt-[16px]" required>
              <bk-input v-model={this.replicas} type="number" class="w-[100px]" min={0}></bk-input>
            </bk-form-item>
          </bk-form>
        </bcs-dialog>
        <bcs-dialog
          v-model={this.confirmDialog.show}
          show-footer={false}
          width={480}
          render-directive="if">
          <ConfirmContent
            title={this.confirmDialog.title}
            loading={this.confirmDialog.loading}
            confirmText={this.confirmDialog.confirmText}
            confirm={() => this.confirmFn()}
            cancel={() => this.confirmDialog.show = false}>
            <bk-form label-width={100} class="mt-[16px]">
              <div class="bg-[#F5F7FA] py-[8px]">
                <bk-form-item label={this.$t('cluster.labels.name')}>
                  {this.clusterNameMap[this.handleGetExtData(this.curDetailRow?.data?.metadata?.uid, 'clusterID')]}
                </bk-form-item>
                <bk-form-item label={this.$t('k8s.namespace')} class="!mt-0">
                  {this.curDetailRow?.data?.metadata?.namespace}
                </bk-form-item>
                <bk-form-item label={this.$t('view.labels.resourceName')} class="!mt-0">
                  {this.curDetailRow?.data?.metadata?.name}
                </bk-form-item>
              </div>
            </bk-form>
          </ConfirmContent>
        </bcs-dialog>
        <Rollback
          name={this.curRow.metadata.name}
          namespace={this.curRow.metadata.namespace}
          category={this.category}
          cluster-id={this.handleGetExtData(this.curRow.metadata.uid, 'clusterID')}
          revision={''}
          crd={this.crd}
          value={this.showRollbackSideslider}
          rollback={true}
          on-hidden={this.handleRollbackSidesilderHide}
          on-rollback-success={this.handleRollbackSidesilderHide}/>
        <CreateResource
          show={this.showCreateDialog}
          type={this.type}
          category={this.category}
          kind={this.kind}
          crd={this.crd}
          scope={this.scope}
          formUpdate={this.formUpdate}
          cancel={() => this.showCreateDialog = false} />
      </div>
    );
  },
});
