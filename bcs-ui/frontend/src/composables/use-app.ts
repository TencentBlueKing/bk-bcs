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
  const clusterList = computed<any[]>(() => ($store.state as any).cluster.clusterList || []);

  return {
    curCluster,
    curClusterId,
    isSharedCluster,
    clusterList,
  };
}
