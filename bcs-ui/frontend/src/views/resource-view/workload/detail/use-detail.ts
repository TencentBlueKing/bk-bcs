/* eslint-disable camelcase */
import yamljs from 'js-yaml';
import { computed, ref } from 'vue';

import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

export interface IWorkloadDetail {
  manifest: any;
  manifestExt: any;
  webAnnotations?: any;
}

export interface IDetailOptions {
  category: string;
  name: string;
  namespace: string;
  type: string;
  defaultActivePanel: string;
  clusterId: string;
  crd?: string;
}

export default function useDetail(options: IDetailOptions) {
  const isLoading = ref(false);
  const detail = ref<IWorkloadDetail|null>(null);
  const activePanel = ref(options.defaultActivePanel);
  const showYamlPanel = ref(false);

  // 标签数据
  const labels = computed(() => {
    const obj = detail.value?.manifest?.metadata?.labels || {};
    return Object.keys(obj).map(key => ({
      key,
      value: obj[key],
    }));
  });
  // 注解数据
  const annotations = computed(() => {
    const obj = detail.value?.manifest?.metadata?.annotations || {};
    return Object.keys(obj).map(key => ({
      key,
      value: obj[key],
    }));
  });
  // Selector数据
  const selectors = computed(() => {
    const obj = detail.value?.manifest?.spec?.selector?.matchLabels || {};
    return Object.keys(obj).map(key => ({
      key,
      value: obj[key],
    }));
  });
  // spec 数据
  const spec = computed(() => detail.value?.manifest.spec || {});
  // metadata 数据
  const metadata = computed(() => detail.value?.manifest?.metadata || {});
  // manifestExt 数据
  const manifestExt = computed(() => detail.value?.manifestExt || {});
  // updateStrategy数据(GameDeployments需要)
  const updateStrategy = computed(() => {
    if (options.category === 'deployments') {
      return detail.value?.manifest?.spec?.strategy || {};
    }
    return detail.value?.manifest?.spec?.updateStrategy || {};
  });
  // yaml数据
  const yaml = computed(() => yamljs.dump(detail.value?.manifest || {}));
  const webAnnotations = ref<any>({});
  const additionalColumns = ref<any>([]);
  // 界面权限
  const pagePerms = computed(() => ({
    create: webAnnotations.value?.perms?.page?.create_btn || {},
    delete: webAnnotations.value?.perms?.page?.delete_btn || {},
    update: webAnnotations.value?.perms?.page?.update_btn || {},
  }));

  const handleTabChange = (item) => {
    activePanel.value = item.name;
  };
  // 获取workload详情
  const handleGetDetail = async (loading = true) => {
    const { namespace, category, name, type, clusterId } = options;
    // workload详情
    isLoading.value = loading;
    const res = await $store.dispatch('dashboard/getResourceDetail', {
      $namespaceId: namespace,
      $category: category,
      $name: name,
      $type: type,
      $clusterId: clusterId,
    });
    detail.value = res.data;
    webAnnotations.value = res.webAnnotations || { perms: {} };
    isLoading.value = false;
    return detail.value;
  };

  const handleGetCustomObjectDetail = async (loading = true) => {
    const { name, crd, namespace, clusterId } = options;
    // workload详情
    isLoading.value = loading;
    const res = await $store.dispatch('dashboard/getCustomObjectResourceDetail', {
      $crdName: crd,
      $namespaceId: namespace,
      $name: name,
      $clusterId: clusterId,
    });
    detail.value = res.data;
    webAnnotations.value = res.webAnnotations || { perms: {} };
    additionalColumns.value = res.webAnnotations?.additionalColumns;
    isLoading.value = false;
    return detail.value;
  };

  const handleShowYamlPanel = () => {
    showYamlPanel.value = true;
  };

  // 更新资源
  const handleUpdateResource = () => {
    const kind = detail.value?.manifest?.kind;
    const editMode = detail.value?.manifestExt?.editMode;
    const { namespace, category, name, type } = options;
    if (editMode === 'form') {
      $router.push({
        name: 'dashboardFormResourceUpdate',
        params: {
          namespace,
          name,
        },
        query: {
          type,
          category,
          kind,
          crd: options.crd,
          formUpdate: webAnnotations.value?.featureFlag?.FORM_UPDATE,
        },
      });
    } else {
      $router.push({
        name: 'dashboardResourceUpdate',
        params: {
          namespace,
          name,
        },
        query: {
          type,
          category,
          kind,
          crd: options.crd,
        },
      });
    }
  };

  // 删除资源
  const handleDeleteResource = () => {
    const kind = detail.value?.manifest?.kind;
    const { namespace, category, name, type, crd, clusterId } = options;
    $bkInfo({
      type: 'warning',
      clsName: 'custom-info-confirm',
      title: $i18n.t('dashboard.title.confirmDelete'),
      subTitle: `${kind} ${name}`,
      defaultInfo: true,
      confirmFn: async () => {
        let result = false;
        if (type === 'crd') {
          result = await $store.dispatch('dashboard/customResourceDelete', {
            namespace,
            $crd: crd,
            $category: category,
            $name: name,
            $clusterId: clusterId,
          });
        } else {
          result = await $store.dispatch('dashboard/resourceDelete', {
            $namespaceId: namespace,
            $type: type,
            $category: category,
            $name: name,
            $clusterId: clusterId,
          });
        }
        result && $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.delete'),
        });
        $router.back();
      },
    });
  };

  const getJsonPathValue = (row, path: string) => {
    const keys = path.split('.').filter(str => !!str);
    const value = keys.reduce((data, key) => {
      if (typeof data === 'object') {
        return data?.[key];
      }
      return data;
    }, row);
    return value || value === 0 ? value : '--';
  };

  return {
    isLoading,
    detail,
    activePanel,
    labels,
    annotations,
    selectors,
    updateStrategy,
    spec,
    metadata,
    manifestExt,
    webAnnotations,
    additionalColumns,
    yaml,
    showYamlPanel,
    pagePerms,
    handleShowYamlPanel,
    handleTabChange,
    handleGetDetail,
    handleGetCustomObjectDetail,
    handleUpdateResource,
    handleDeleteResource,
    getJsonPathValue,
  };
}
