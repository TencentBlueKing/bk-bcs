import { cloneDeep, get } from 'lodash';
import { computed } from 'vue';

import { MultiClusterResourcesType } from '../common/use-table-data';

import {
  createViewConfig as createViewConfigApi,
  deleteViewConfig as deleteViewConfigApi,
  labelSuggest as labelSuggestApi,
  updateViewConfig as updateViewConfigApi,
  valuesSuggest as valuesSuggestApi,
  viewConfigDetail,
  viewConfigList,
  viewConfigRename as viewConfigRenameApi } from '@/api/modules/cluster-resource';
import { isDeepEmpty } from '@/common/util';
import { useCluster } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

export default function () {
  const { clusterNameMap } = useCluster();
  const clusterID = computed(() => $router.currentRoute?.params?.clusterId);
  const isViewEditable = computed(() => $store.state.isViewEditable);
  const viewList = computed<IViewData[]>(() => $store.state.dashboardViewList);
  const dashboardViewID = computed(() => $store.state.dashboardViewID);

  // 临时视图数据(集群视图只有临时数据)
  const curTmpViewData = computed(() => $store.state.tmpViewData);
  // 详情数据(原始数据、当前临时条件数据和编辑的数据)
  const viewDetailData = computed<{
    originViewData: IViewData|undefined
    tmpViewData: IViewData|undefined
    editViewData: IViewData|undefined
  }>(() => {
    // 原始数据
    const originViewData = viewList.value.find(item => item.id === dashboardViewID.value);

    // 临时条件数据
    const tmpViewData = cloneDeep(curTmpViewData.value);
    if (tmpViewData) {
      // hack 过滤掉空字段（接口不能传一空数据）
      tmpViewData.filter = Object.keys(tmpViewData.filter || {}).reduce((pre, key) => {
        const field = cloneDeep(tmpViewData.filter);
        if (isDeepEmpty(get(field, key))) return pre;

        if (key === 'labelSelector') {
          pre[key] = field[key]?.filter(item => !!item.key); // 过滤空的标签
        } else {
          pre[key] = field[key];
        }
        return pre;
      }, {});
    }

    // 当前编辑状态数据
    const editViewData = cloneDeep($store.state.editViewData);
    if (editViewData?.filter?.labelSelector?.length) { // 过滤空的标签
      editViewData.filter.labelSelector = editViewData.filter.labelSelector.filter(item => !!item.key);
    }

    return {
      originViewData, // 原始数据
      tmpViewData, // 当前临时数据
      editViewData, // 编辑态数据
    };
  });
  // 当前视图查询数据（用于接口搜索）
  const curViewData = computed<Partial<MultiClusterResourcesType>|undefined>(() => {
    let data: Partial<MultiClusterResourcesType> | undefined;
    if (isViewEditable.value) {
      // 编辑模式
      const clusterNamespaces = viewDetailData.value?.editViewData?.clusterNamespaces || [];
      data = clusterNamespaces.length
        ? {
          ...(viewDetailData.value?.editViewData?.filter || {}),
          viewID: '', // 编辑模式时，不传视图ID给后端
          clusterNamespaces,
        }
        : undefined;// 编辑数据未就绪
    } else if (dashboardViewID.value) {
      // 自定义视图查看模式
      data = viewList.value.length
        ? {
          ...(viewDetailData.value?.tmpViewData?.filter || {}), // 临时条件
          viewID: dashboardViewID.value, // 自定义视图查看模式时，需要把临时条件给后端
          clusterNamespaces: viewDetailData.value?.originViewData?.clusterNamespaces || [], // 集群跟命名空间不能临时编辑的
        }
        : undefined;// 视图列表数据未就绪
    } else if (!dashboardViewID.value && clusterID.value) {
      // 集群视图
      data = {
        ...(viewDetailData.value?.tmpViewData?.filter || {}), // 临时条件
        viewID: '',
        clusterNamespaces: [
          {
            clusterID: clusterID.value,
            namespaces: $store.state.viewNsList,
          },
        ],
      };
    }

    return data;
  });

  // 是否是集群模式
  const isClusterMode = computed(() => !viewDetailData.value?.originViewData?.id);
  // 当前视图名称
  const curViewName = computed(() => {
    if (isClusterMode.value) {
    // 集群模式显示集群名称
      return clusterNameMap.value[clusterID.value];
    }
    return viewDetailData.value?.originViewData?.name;
  });
  // 视图类型
  const curViewType = computed(() => {
    if (isClusterMode.value) {
    // 集群模式显示集群ID
      return clusterID.value;
    }
    return $i18n.t('view.labels.custom');
  });

  // 默认视图
  const isDefaultView = computed(() => !viewDetailData.value?.originViewData?.id);

  // 更新视图缓存ID
  const updateViewIDStore = async (id = '') => {
    $store.commit('updateDashboardViewID', id);// 更新当前视图ID
    $store.commit('updateTmpViewData', {});
  };

  const getViewConfigList = async () => {
    const data: IViewData[] = await viewConfigList({}, { cancelPrevious: true }).catch(() => []);
    $store.commit('updateDashboardViewList', data);
    return data;
  };

  const getViewConfigDetail = async ($id: string) => {
    const data: IViewData = await viewConfigDetail({ $id }).catch(() => ({}));
    return data;
  };

  const createViewConfig = async (params: Partial<Omit<IViewData, 'id'> & { saveAs?: boolean }>) => {
    const result = await createViewConfigApi(params).catch(() => ({ id: null }));
    return result;
  };

  const updateViewConfig = async (params: Pick<IViewData, 'id'|'filter'> & { $id: string }) => {
    const result = await updateViewConfigApi(params).then(() => true)
      .catch(() => false);
    return result;
  };

  const deleteViewConfig = async (params: { $id: string }) => {
    const result = await deleteViewConfigApi(params).then(() => true)
      .catch(() => false);
    return result;
  };

  const viewConfigRename = async (params: {
    $id: string
    name: string
  }) => {
    const result = await viewConfigRenameApi(params).then(() => true)
      .catch(() => false);
    return result;
  };

  const labelSuggest = async (params: { clusterNamespaces: IClusterNamespace[]}) => {
    if (!params.clusterNamespaces?.length) return [];
    const data = await labelSuggestApi(params).catch(() => ({ values: [] }));
    return data.values;
  };

  const valuesSuggest = async (params: {
    label: string
    clusterNamespaces: IClusterNamespace[]
  }) => {
    if (!params.clusterNamespaces?.length || !params.label) return [];
    const data = await valuesSuggestApi(params).catch(() => ({ values: [] }));
    return data.values;
  };

  return {
    viewList,
    isDefaultView,
    dashboardViewID,
    isClusterMode,
    curTmpViewData,
    curViewData,
    curViewName,
    curViewType,
    isViewEditable,
    getViewConfigList,
    getViewConfigDetail,
    createViewConfig,
    updateViewConfig,
    deleteViewConfig,
    viewConfigRename,
    updateViewIDStore,
    labelSuggest,
    valuesSuggest,
  };
}
