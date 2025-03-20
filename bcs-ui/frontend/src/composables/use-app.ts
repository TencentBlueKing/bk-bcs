import { computed, InjectionKey, ref } from 'vue';

import { featureFlags as featureFlagsApi } from '@/api/modules/project';
import { userInfo } from '@/api/modules/user-manager';
import { Preset } from '@/components/assistant/use-assistant-store';
import $store from '@/store';

// todo 完善类型
export interface IProject {
  name: string
  businessID: string
  businessName: string
  projectID: string
  projectCode: string
  description: string
  kind: string
  enableVcluster: boolean
  project_name: string // 兼容旧版数据（不要再使用）
  project_id: string // 兼容旧版数据
}
// todo 完善类型
export interface ICluster {
  region: string
  clusterID: string
  clusterName: string
  status: 'INITIALIZATION' | 'DELETING' | 'RUNNING'
  clusterCategory: string
  providerType: string
  networkSettings: {
    maxNodePodNum: number
    maxServiceNum: number
    clusterIPv4CIDR: string
    multiClusterCIDR: string[]
    cidrStep: number
    serviceIPv4CIDR: string
    isStaticIpMode: boolean
    eniSubnetIDs: string[]
    enableVPCCni: boolean
    status: string
    networkMode: 'tke-route-eni' | 'tke-direct-eni'
  }
  master: any
  provider: CloudID,
  clusterBasicSettings: any
  environment: 'stag'|'debug'|'prod'
  extraInfo?: Record<string, any>
  manageType: 'INDEPENDENT_CLUSTER' | 'MANAGED_CLUSTER'
  clusterType: string
  is_shared: boolean
  cluster_id: string // 兼容旧版数据（不要再使用）
  importCategory: string
  clusterAdvanceSettings: any
  systemID: string
  description: string
  creator: string
  createTime: string
  updateTime: string
  vpcID: string
  networkType: string
  cloudAccountID: string
  labels: Record<string, string>
  sharedRanges?: {
    bizs: string[],
    projectIdOrCodes: string[]
  }

  // 从clusterExtraInfo中merge过来的
  autoScale: boolean
}

/**
 * 获取项目相关配置
 */
export function useProject() {
  const curProject = computed<IProject>(() => $store.state.curProject as any);
  const projectID = computed<string>(() => curProject.value?.projectID);
  // todo 详情接口会丢失project_code
  const projectCode = computed<string>(() => curProject.value?.projectCode || $store.getters.curProjectCode);
  const isMaintainer = computed(() => ($store.state.cluster.maintainers as string[])
    .includes($store.state.user.username));

  return {
    isMaintainer,
    curProject,
    projectID,
    projectCode,
  };
}
/**
 * 获取集群相关信息
 */
export function useCluster() {
  const curCluster = computed<ICluster>(() => $store.state.curCluster || {});
  const curClusterId = computed<string>(() => $store.getters.curClusterId);
  const isSharedCluster = computed<boolean>(() => $store.state.curCluster?.is_shared);
  const clusterList = computed<ICluster[]>(() => $store.state.cluster.clusterList || []);
  const clusterNameMap = computed<Record<string, string>>(() => clusterList.value.reduce((pre, item) => {
    pre[item.clusterID] = item.clusterName;
    return pre;
  }, {}));
  const clusterMap = computed<Record<string, ICluster>>(() => clusterList.value.reduce((pre, item) => {
    pre[item.clusterID] = item;
    return pre;
  }, {}));

  const { projectCode } = useProject();
  // const terminalWins = ref<Window | null>(null);
  const handleGotoConsole = ({ clusterID, clusterName }: {
    clusterID: string
    clusterName?: string
  }) => {
    console.log(clusterName);
    const url = `${window.BCS_API_HOST}/bcsapi/v4/webconsole/projects/${projectCode.value}/mgr/#cluster=${clusterID}`;
    window.open(url, '');
    // 缓存当前窗口，再次打开时重新进入
    // if (terminalWins.value) {
    //   if (!terminalWins.value.closed) {
    //     terminalWins.value.postMessage({
    //       clusterId: clusterID,
    //       clusterName,
    //     }, location.origin);
    //     terminalWins.value.focus();
    //   } else {
    //     terminalWins.value = window.open(url, '');
    //   }
    // } else {
    //   terminalWins.value = window.open(url, '');
    // }
  };

  return {
    curCluster,
    curClusterId,
    isSharedCluster,
    clusterList,
    clusterNameMap,
    clusterMap,
    handleGotoConsole,
  };
}
/**
 * APP相关信息,eg: 用户, feature_flag, 文档配置等
 */
export function useAppData() {
  // 用户信息
  const user = computed(() => $store.state.user);
  async function getUserInfo() {
    const data = await userInfo().catch(() => ({}));
    $store.commit('updateUser', data);
    return data;
  }

  // 特性开关(菜单等)
  const flagsMap = computed(() => $store.state.featureFlags);
  const defaultFlags = {
    CLOUDTOKEN: false,
    PROJECT_LIST: false,
    AZURECLOUD: true,
    IMPORTSOPSCLUSTER: true,
  };
  async function getFeatureFlags(params: { projectCode: string }) {
    const data = await featureFlagsApi(params).catch(() => ({}));
    $store.commit('updateFeatureFlags', Object.assign(defaultFlags, data));
    return data;
  }

  // 当前版本的文档链接信息
  const linksMap = ref(window.BCS_CONFIG);

  // 是否是内部版
  const _INTERNAL_ = computed(() => !['ce', 'ee'].includes(window.REGION));

  return {
    user,
    flagsMap,
    linksMap,
    _INTERNAL_,
    getUserInfo,
    getFeatureFlags,
  };
}

// inject key
export const AiSendMsgFnInjectKey: InjectionKey<(msg: string, preset?: Preset) => void> = Symbol('ai-send-message');
