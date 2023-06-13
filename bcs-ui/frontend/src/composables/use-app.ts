import { computed, ref } from 'vue';
import $store from '@/store';
/**
 * 获取项目文档配置信息
 * @returns
 */
export function useConfig() {
  // 当前版本的文档链接信息
  const PROJECT_CONFIG = ref(window.BCS_CONFIG);
  // 是否是内部版
  const _INTERNAL_ = computed(() => !['ce', 'ee'].includes(window.REGION));
  return {
    PROJECT_CONFIG,
    _INTERNAL_,
  };
}

// todo 完善类型
export interface IProject {
  name: string
  businessID: string
  businessName: string
  projectID: string
  projectCode: string
  description: string
  kind: string
  project_name: string // 兼容旧版数据（不要再使用）
  project_id: string // 兼容旧版数据
}
// todo 完善类型
export interface ICluster {
  clusterID: string
  clusterName: string
  status: 'INITIALIZATION' | 'DELETING'
  clusterCategory: string
  providerType: string
  networkSettings: any
  master: any
  clusterBasicSettings: any
  manageType: 'INDEPENDENT_CLUSTER' | 'MANAGED_CLUSTER'
  provider: string
  is_shared: boolean
  cluster_id: string // 兼容旧版数据（不要再使用）
}
/**
 * 获取项目相关配置
 */
export function useProject() {
  const curProject = computed<IProject>(() => $store.state.curProject as any);
  const projectID = computed<string>(() => curProject.value?.projectID);
  // todo 详情接口会丢失project_code
  const projectCode = computed<string>(() => curProject.value?.projectCode || $store.getters.curProjectCode);
  const projectList = computed<any[]>(() => $store.state.projectList || []);
  const isMaintainer = computed(() => ($store.state.cluster.maintainers as string[])
    .includes($store.state.user.username));

  return {
    isMaintainer,
    curProject,
    projectID,
    projectCode,
    projectList,
  };
}

/**
 * 获取集群相关信息
 */
export function useCluster() {
  const curCluster = computed<Partial<ICluster>>(() => $store.state.curCluster || {});
  const curClusterId = computed<string>(() => $store.getters.curClusterId);
  const isSharedCluster = computed<boolean>(() => $store.state.curCluster?.is_shared);
  const clusterList = computed<Partial<ICluster>[]>(() => $store.state.cluster.clusterList || []);

  const { projectID } = useProject();
  const terminalWins = ref<Window | null>(null);
  const handleGotoConsole = ({ clusterID, clusterName }: {
    clusterID: string
    clusterName?: string
  }) => {
    const url = `${window.DEVOPS_BCS_API_URL}/web_console/projects/${projectID.value}/mgr/#cluster=${clusterID}`;
    // 缓存当前窗口，再次打开时重新进入
    if (terminalWins.value) {
      if (!terminalWins.value.closed) {
        terminalWins.value.postMessage({
          clusterId: clusterID,
          clusterName,
        }, location.origin);
        terminalWins.value.focus();
      } else {
        terminalWins.value = window.open(url, '');
      }
    } else {
      terminalWins.value = window.open(url, '');
    }
  };

  return {
    curCluster,
    curClusterId,
    isSharedCluster,
    clusterList,
    handleGotoConsole,
  };
}
