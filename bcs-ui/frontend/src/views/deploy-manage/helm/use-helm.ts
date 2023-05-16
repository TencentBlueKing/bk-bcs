import {
  reposList,
  createRepo,
  repoCharts,
  deleteRepoChart,
  repoChartVersions,
  repoChartVersionDetail,
  deleteRepoChartVersion,
  releaseDetail,
  deleteRelease,
  releaseChart,
  updateRelease,
  releaseHistory,
  previewRelease,
  rollbackRelease,
  releaseStatus,
  releasesList,
  downloadChartUrl,
  chartDetail,
  chartReleases,
} from '@/api/modules/helm';
import { parseUrl } from '@/api/request';
import Vue, { ref } from 'vue';
import $i18n from '@/i18n/i18n-setup';

const { $bkMessage } = Vue.prototype;

interface IParams {
  $clusterId: string
  $namespaceId: string
  $releaseName: string
  version: string
  repository: string
  chart: string
  values: string[]
  revision: number
}
export default function useHelm() {
  const loading = ref(false);
  const repos = ref<any[]>([]);
  const handleGetReposList = async () => {
    loading.value = true;
    repos.value = await reposList().catch(() => []);
    loading.value = false;
    return repos.value;
  };

  const handleCreateRepo  = async (params: {
    name: string
    type: string
  }) => {
    const result = await createRepo(params).then(() => true)
      .catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: $i18n.t('创建成功'),
    });
    return result;
  };
  const handleGetRepoCharts = async ($repoName: string, page: number, size: number, name = '') => {
    const data = await repoCharts({
      $repoName,
      name,
      page,
      size,
    }).catch(() => ({ data: [], total: 0 }));
    return data;
  };
  const handleDeleteRepoChart = async ($repoName: string, $chartName: string) => {
    const result = await deleteRepoChart({
      $repoName,
      $chartName,
    }).then(() => true)
      .catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: $i18n.t('删除成功'),
    });
    return result;
  };
  const handleGetRepoChartVersions = async ($repoName: string, $chartName: string) => {
    const data = await repoChartVersions({
      $repoName,
      $chartName,
    }).catch(() => ({ data: [], total: 0 }));
    return data;
  };
  const handleGetRepoChartVersionDetail = async ($repoName: string, $chartName: string, $version: string) => {
    const data = repoChartVersionDetail({
      $repoName,
      $chartName,
      $version,
    }).catch(() => ({}));
    return data;
  };
  const handleDeleteRepoChartVersion = async ($repoName: string, $chartName: string, $version: string) => {
    const result = deleteRepoChartVersion({
      $repoName,
      $chartName,
      $version,
    }).then(() => true)
      .catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: $i18n.t('删除成功'),
    });
    return result;
  };
  const handleGetReleaseDetail = async ($clusterId: string, $namespaceId: string, $releaseName: string) => {
    const data = releaseDetail({
      $clusterId,
      $namespaceId,
      $releaseName,
    }).catch(() => ({}));
    return data;
  };
  const handleDeleteRelease = async ($clusterId: string, $namespaceId: string, $releaseName: string) => {
    const result = deleteRelease({
      $clusterId,
      $namespaceId,
      $releaseName,
    }).then(() => true)
      .catch(() => false);

    result && $bkMessage({
      theme: 'success',
      message: $i18n.t('删除成功'),
    });
    return result;
  };
  const handleReleaseChart = async (params: Omit<IParams, 'revision'>) => {
    const result = await releaseChart(params).then(() => true)
      .catch(() => false);

    result && $bkMessage({
      theme: 'success',
      message: $i18n.t('部署成功'),
    });
    return result;
  };
  const handleUpdateRelease = async (params: Omit<IParams, 'revision'>) => {
    const result = await updateRelease(params).then(() => true)
      .catch(() => false);

    result && $bkMessage({
      theme: 'success',
      message: $i18n.t('更新成功'),
    });
    return result;
  };
  const handleGetReleaseHistory = async (
    $clusterId: string,
    $namespaceId: string,
    $releaseName: string,
    filter?: string,
  ) => {
    const data = await releaseHistory({
      $clusterId,
      $namespaceId,
      $releaseName,
      filter,
    }).catch(() => []);
    return data;
  };
  const handlePreviewRelease = async (params:
  Omit<IParams, 'revision'> | Omit<IParams, 'version' | 'repository' | 'values' | 'chart'>) => {
    const data = await previewRelease(params).catch(() => ({}));
    return data;
  };
  const handleRollbackRelease = async (params: {
    $clusterId: string
    $namespaceId: string
    $releaseName: string
    revision: string
  }) => {
    const result = await rollbackRelease(params).then(() => true)
      .catch(() => false);
    result && $bkMessage({
      theme: 'success',
      message: $i18n.t('回滚成功'),
    });
    return result;
  };
  const handleGetReleaseStatus = async (params: {
    $clusterId: string
    $namespaceId: string
    $releaseName: string
  }) => {
    const data = await releaseStatus(params).catch(() => []);
    return data;
  };
  const handleGetReleasesList = async (params: {
    $clusterId: string
    namespace?: string
    name?: string
    page?: number
    size?: number
  }) => {
    const data = await releasesList(params, { needRes: true }).catch(() => ({
      data: { data: [], total: 0 },
      web_annotations: {},
    }));
    return data;
  };
  const handleDownloadChart = async ($repoName: string, $chartName: string, $version: string) => {
    const { url } =  parseUrl('get', downloadChartUrl, {
      $repoName,
      $chartName,
      $version,
    });
    window.open(url);
  };
  const handleGetChartDetail = async ($repoName, $chartName) => {
    const data =  await chartDetail({
      $repoName,
      $chartName,
    }).catch(() => ({}));
    return data;
  };
  const handleGetChartReleases = async (params: {
    $repoName: string
    $chartName: string
    versions?: string[]
  }) => {
    const data = await chartReleases(params).catch(() => []);
    return data;
  };

  return {
    loading,
    repos,
    handleGetReposList,
    handleGetRepoCharts,
    handleCreateRepo,
    handleDeleteRepoChart,
    handleGetRepoChartVersions,
    handleGetRepoChartVersionDetail,
    handleDeleteRepoChartVersion,
    handleGetReleaseDetail,
    handleDeleteRelease,
    handleReleaseChart,
    handleUpdateRelease,
    handleGetReleaseHistory,
    handlePreviewRelease,
    handleRollbackRelease,
    handleGetReleaseStatus,
    handleGetReleasesList,
    handleDownloadChart,
    handleGetChartDetail,
    handleGetChartReleases,
  };
}
