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
  const $INTERNAL = computed(() => !['ce', 'ee'].includes(window.REGION));
  return {
    PROJECT_CONFIG,
    $INTERNAL,
  };
}


/**
 * 获取项目相关配置
 */
export function useProject() {
  const curProject = computed(() => $store.state.curProject);
  const projectID = computed(() => curProject.value?.project_id);
  const projectCode = computed(() => curProject.value?.project_code);

  return {
    curProject,
    projectID,
    projectCode,
  };
}

/**
 * 获取集群相关信息
 */
export function useCluster() {
  const curClusterId = computed(() => $store.state.curClusterId);

  return {
    curClusterId,
  };
}

export default {
  ...useCluster(),
  ...useProject(),
  ...useConfig(),
};
