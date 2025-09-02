import { ref } from 'vue';

import { meshClusters, meshConfig, meshDelete, meshDetail, meshList, meshUpdate } from '@/api/modules/mesh-manager';
import { useProject } from '@/composables/use-app';

export interface IMesh {
  projectID: string,
  projectCode: string,
  name: string,
  version: string,
  revision: string,
  controlPlaneMode: string,
  clusterMode: string,
  description: string,
  primaryClusters: string[],
  remoteClusters: Partial<MeshCluster>[],
  differentNetwork: boolean,
  status?: string,
  sidecarResourceConfig: ISidecar,
  multiClusterEnabled: boolean,
  clbID: string,
  highAvailability: {
    autoscaleEnabled: boolean,
    autoscaleMin: number,
    autoscaleMax: number,
    replicaCount: number,
    targetCPUAverageUtilizationPercent: number,
    resourceConfig: ISidecar,
    dedicatedNode: {
      enabled?: boolean,
      nodeLabels?: {
        [key: string]: string
      }
    }
  },
  observabilityConfig: {
    metricsConfig: {
      metricsEnabled?: boolean,
      controlPlaneMetricsEnabled?: boolean,
      dataPlaneMetricsEnabled?: boolean
    },
    logCollectorConfig: {
      enabled?: boolean,
      accessLogEncoding?: string,
      accessLogFormat?: string
    },
    tracingConfig: {
      enabled?: boolean,
      traceSamplingPercent?: number,
      endpoint?: string,
      bkToken?: string
    }
  },
  featureConfigs: {
    [key: string]: {
      name?: string,
      description?: string,
      value: string,
      defaultValue?: string,
      availableValues?: string[],
      supportVersions?: string[]
    },
  }
}

export interface ISidecar {
  cpuRequest: string,
  cpuLimit: string,
  memoryRequest: string,
  memoryLimit: string
}

export type MeshCluster = {
  clusterID: string,
  clusterName: string,
  clusterType: string,
  isShared: boolean,
  isInstalled: boolean
  status: string,
  version: string,
  region: string,
  joinTime: string,
};

export const configData = ref<Partial<Pick<IMesh, 'featureConfigs' | 'highAvailability' | 'sidecarResourceConfig' | 'observabilityConfig'>> & { istioVersions?: any[] }>({});

export const web_annotations = ref({ perms: {} });

export const clusterList = ref<MeshCluster[]>([]);
export const clusterLoading = ref<boolean>(false);

// 提取字符串中的数字
export function extractNumbers(str?: string): number {
  if (!str) return 0;
  const matched = str.match(/\d+/g);
  return matched ? parseInt(matched[0], 10) : 0;
}

export default function useMesh() {
  const { projectCode } = useProject();
  const meshData = ref<IMesh[]>([]);
  const total = ref(0);
  const loading = ref(false);
  // 获取列表
  const fetchMeshData = async (params) => {
    loading.value = true;
    const res = await meshList(
      { projectCode: projectCode.value, ...params },
      { needRes: true },
    );
    meshData.value = res?.data?.items || [];
    total.value = res?.data?.total || 0;
    web_annotations.value = res?.web_annotations || { perms: {} };
    loading.value = false;
  };

  // 删除
  async function handleDelete({ meshID, projectCode }) {
    if (!meshID || !projectCode) return false;
    const result = await meshDelete({
      $meshID: meshID,
      projectCode,
    }).then(() => true)
      .catch(() => false);

    return result;
  }

  // 更新
  async function handleUpdateMesh(data) {
    if (!data.meshID) return false;
    const result = await meshUpdate({
      $meshID: data.meshID,
      ...data,
    })
      .then(() => true)
      .catch(() => false);

    return result;
  }

  // 获取详情
  async function handleGetMeshDetail({ meshID, projectCode }) {
    if (!meshID || !projectCode) return {};
    const result = await meshDetail({
      $meshID: meshID,
      projectCode,
    }).catch(() => {});

    return result;
  }

  // 获取默认配置
  async function handleGetConfig() {
    loading.value = true;
    configData.value = await meshConfig().catch(() => {});
    loading.value = false;
  }

  // 获取集群列表
  async function getClusterList() {
    clusterLoading.value = true;
    const res = await meshClusters({
      projectCode: projectCode.value,
    }).catch(() => ({ clusters: [] }));
    clusterList.value = res.clusters;
    clusterLoading.value = false;
  }

  return {
    loading,
    total,
    meshData,
    configData,
    web_annotations,
    clusterList,
    clusterLoading,
    fetchMeshData,
    handleDelete,
    handleGetMeshDetail,
    handleUpdateMesh,
    handleGetConfig,
    extractNumbers,
    getClusterList,
  };
}
