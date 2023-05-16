import { computed, ref } from '@vue/composition-api';
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


export interface IProject {
  name: string
  businessID: string
  businessName: string
  projectID: string
  projectCode: string
  description: string
  kind: string
  project_name: string // 兼容旧版数据
  project_id: string // 兼容旧版数据
}
/**
 * 获取项目相关配置
 */
export function useProject() {
  const curProject = computed(() => $store.state.curProject);
  const projectID = computed<string>(() => curProject.value?.project_id);
  // todo 详情接口会丢失project_code
  const projectCode = computed<string>(() => curProject.value?.project_code || $store.getters.curProjectCode);
  const projectList = computed<any[]>(() => $store.state.projectList || []);

  return {
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
  const curCluster = computed(() => $store.state.curCluster || {});
  const curClusterId = computed<string>(() => $store.getters.curClusterId);
  const isSharedCluster = computed<boolean>(() => $store.state.curCluster?.is_shared);
  const clusterList = computed<any[]>(() => $store.state.cluster.clusterList || []);


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
