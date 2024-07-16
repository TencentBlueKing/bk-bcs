import { computed, onBeforeMount, Ref, ref, watch } from 'vue';

import { customResourceCreate, customResourceDelete, customResourceList, customResourceUpdate } from '@/api/modules/cluster-resource';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import { useCluster } from '@/composables/use-app';
import usePage from '@/composables/use-page';
import useSearch from '@/composables/use-search';
import $i18n from '@/i18n/i18n-setup';

interface IConfig<T> {
  $crd: string
  $apiVersion: string
  $kind: string
  clusterId: string
  formData: Ref<T>
  initFormData: T
  getParams?: Function
  $namespace?: string
}

export default function useCustomCrdList<
T extends { metadata: { name: string, namespace: string } }
>(config: IConfig<T>) {
  const { $crd, $kind, $apiVersion, clusterId, formData, initFormData, getParams, $namespace } = config;

  const { clusterList } = useCluster();
  const curCluster = computed(() => clusterList.value.find(item => item.clusterID === clusterId));

  const crdList = ref([]);
  const crdManifestExt = ref({});

  // 搜索
  const keys = ref(['metadata.name']);
  const { searchValue, tableDataMatchSearch } = useSearch(crdList, keys);
  watch(searchValue, () => {
    pagination.value.current = 1;
  });
  const handleClearSearchData = () => {
    searchValue.value = '';
  };

  // 前端分页
  const { curPageData, pagination, pageChange, pageSizeChange } = usePage(tableDataMatchSearch);

  // 获取表格数据
  const tableLoading = ref(false);
  const handleGetCrdList = async (params) => {
    tableLoading.value = true;
    const { manifest, manifestExt } = await customResourceList(params).catch(() => ({ manifest: {} }));
    crdList.value = manifest.items || [];
    crdManifestExt.value = manifestExt;
    tableLoading.value = false;
  };

  // 获取ext data属性
  const handleGetExtData = (uid: string, key: string) => crdManifestExt.value[uid]?.[key];

  // 当前操作行
  const currentRow = ref();

  const title = ref();// 侧滑标题
  const isShowCreate = ref(false);

  // 创建crd
  const showCreateCrdSideslider = () => {
    currentRow.value = undefined;
    formData.value = JSON.parse(JSON.stringify(initFormData));
    title.value = $i18n.t('plugin.tools.add');
    isShowCreate.value = true;
  };
  // 更新crd配置
  const showUpdateCrdSideslider = (row) => {
    currentRow.value = JSON.parse(JSON.stringify(row));
    formData.value = currentRow.value;
    title.value = row.metadata.name;
    isShowCreate.value = true;
  };
  const formRef = ref();
  const saving = ref(false);
  const createOrUpdateCrd = async () => {
    const isValidate = await formRef.value?.validate().catch(() => false);
    if (!isValidate) return;

    saving.value = true;
    let result = false;
    const params = getParams ? getParams() : formData.value;
    if (!currentRow.value) {
      // 创建配置
      result = await customResourceCreate({
        $crd,
        $clusterId: clusterId,
        $category: 'custom_objects',
        format: 'manifest',
        rawData: {
          apiVersion: $apiVersion,
          kind: $kind,
          ...params,
        },
      }).then(() => true)
        .catch(() => false);
    } else {
      // 更新配置
      result = await customResourceUpdate({
        $crd,
        $clusterId: clusterId,
        $category: 'custom_objects',
        $name: formData.value.metadata.name,
        format: 'manifest',
        rawData: {
          apiVersion: $apiVersion,
          kind: $kind,
          ...params,
        },
        namespace: formData.value.metadata.namespace,
      }).then(() => true)
        .catch(() => false);
    }
    if (result) {
      $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.ok'),
      });
      isShowCreate.value = false;
      handleGetCrdList({
        $crd,
        $clusterId: clusterId,
        $category: 'custom_objects',
        namespace: $namespace,
      });
    }
    saving.value = false;
  };

  // 删除crd
  const deleteCrd = (row) => {
    const { name, namespace } = row.metadata || {};
    if (!name || !namespace) return;

    $bkInfo({
      type: 'warning',
      clsName: 'custom-info-confirm',
      title: `${$i18n.t('generic.title.confirmDelete')} ${name}`,
      defaultInfo: true,
      okText: $i18n.t('plugin.tools.confirmDelete'),
      confirmFn: async () => {
        const result = await customResourceDelete({
          $crd,
          $clusterId: clusterId,
          $category: 'custom_objects',
          $name: name,
          namespace,
        }).then(() => true)
          .catch(() => false);
        if (result) {
          $bkMessage({
            theme: 'success',
            message: $i18n.t('generic.msg.success.delete'),
          });
          handleGetCrdList({
            $crd,
            $clusterId: clusterId,
            $category: 'custom_objects',
            namespace: $namespace,
          });
        }
      },
    });
  };


  onBeforeMount(() => {
    handleGetCrdList({
      $crd,
      $clusterId: clusterId,
      $category: 'custom_objects',
      namespace: $namespace,
    });
  });

  return {
    tableLoading,
    saving,
    currentRow,
    curCluster,
    curPageData,
    pagination,
    searchValue,
    isShowCreate,
    title,
    formRef,
    pageChange,
    pageSizeChange,
    handleGetExtData,
    showCreateCrdSideslider,
    showUpdateCrdSideslider,
    createOrUpdateCrd,
    deleteCrd,
    handleClearSearchData,
    handleGetCrdList,
  };
}
