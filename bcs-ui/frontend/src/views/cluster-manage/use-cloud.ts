import { computed, ref } from 'vue';

import { ICloudRegion, INodeManCloud, ISecurityGroup, IVpcItem } from './add/tencent/types';

import {
  cloudAccounts as cloudAccountsAPI,
  cloudRegionByAccount,
  cloudSecurityGroups,
  cloudVPC,
  clusterConnect as clusterConnectAPI,
  createCloudAccounts as createCloudAccountsAPI,
  deleteCloudAccounts as deleteCloudAccountsAPI,
  nodemanCloud,
  updateCloudAccounts as updateCloudAccountsAPI,
  validateCloudAccounts as validateCloudAccountsAPI,
} from '@/api/modules/cluster-manager';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store';

export interface IGoogleAccount {
  gkeProjectID?: string
  serviceAccountSecret: string
}

export interface ITencentAccount {
  secretID: string
  secretKey: string
}

export interface IAzureAccount {
  subscriptionID: string
  tenantID: string
  clientID: string
  clientSecret: string
}

export interface ICreateAccountParams<T> {
  $cloudId: CloudID
  accountName: string
  desc: string
  account: T
}

export interface IAccount {
  account: {
    secretID: string
    secretKey: string
    serviceAccountSecret?: string
  }
  accountID: string
  accountName: string
  desc: string
  creator: string
  creatTime: string
}

export interface ICloudAccount {
  account: IAccount
  clusters: string
}

export default function () {
  const curProject = computed(() => $store.state.curProject);
  const user = computed(() => $store.state.user);

  // 云服务器type列表
  const providerNameMap = {
    bluekingCloud: {
      label: $i18n.t('provider.blueKingyun'),
      className: '#bcs-icon-color-k8s',
    },
    tencentCloud: {
      label: $i18n.t('provider.tencentyun'),
      className: '#bcs-icon-color-tencentcloud',
    },
    huaweiCloud: {
      label: $i18n.t('provider.huaweiyun'),
      className: '#bcs-icon-color-huaweicloud',
    },
    awsCloud: {
      label: $i18n.t('provider.yamaxunyun'),
      className: '#bcs-icon-color-awscloud',
    },
    gcpCloud: {
      label: $i18n.t('provider.gugeyun'),
      className: '#bcs-icon-color-gcpcloud',
    },
    azureCloud: {
      label: $i18n.t('provider.weiruanyun'),
      className: '#bcs-icon-color-weiruanyun',
    },
    tencentPublicCloud: {
      label: $i18n.t('provider.tencentPublicCloud'),
      className: '#bcs-icon-color-publiccloud',
    },
  };

  // 云账号列表
  const cloudAccountList = ref<ICloudAccount[]>([]);
  const cloudAccountLoading = ref(false);
  const cloudAccounts = async ($cloudId: CloudID|undefined) => {
    if (!$cloudId) return { data: [], web_annotations: { perms: {} } };
    cloudAccountLoading.value = true;
    const res = await cloudAccountsAPI({
      $cloudId,
      projectID: curProject.value.projectID,
      operator: user.value.username,
    }, { needRes: true }).catch(() => []);
    cloudAccountList.value = res.data || [];
    cloudAccountLoading.value = false;
    return res as {
      data: ICloudAccount[]
      web_annotations: {
        perms: Record<string, any>
      }
    };
  };
  // 创建云账号
  const createCloudAccounts = async <T = IGoogleAccount>(params: ICreateAccountParams<T>) => {
    const result = await createCloudAccountsAPI({
      ...params,
      enable: true,
      creator: user.value.username,
      projectID: curProject.value.projectID,
    }).catch(() => false);
    return result;
  };
  // 删除云账号
  const deleteCloudAccounts = async (params: {
    $cloudId: CloudID
    $accountID: string
  }) => {
    const result = await deleteCloudAccountsAPI(params).then(() => true)
      .catch(() => false);
    return result;
  };
  // 更新云账号
  const updateCloudAccounts = async (params: {
    $cloudId: CloudID
    $accountID: string
    desc: string
  }) => {
    const result = await updateCloudAccountsAPI({
      ...params,
      projectID: curProject.value.projectID,
      updater: user.value.username,
    }).then(() => true)
      .catch(() => false);
    return result;
  };
  // 校验云账号
  const validateCloudAccounts = async (params: {
    $cloudId: CloudID
    account: IGoogleAccount | ITencentAccount | IAzureAccount
  }) => {
    const result = await validateCloudAccountsAPI(params).then(() => '')
      .catch(data => data?.response?.data?.message || data);

    return result;
  };
  // 集群联通性
  const clusterConnect = async (params: {
    $cloudId: CloudID
    $clusterID: string
    isExtranet: boolean
    accountID: string
    region: string
    resourceGroupName?: string
  }) => {
    const result = await clusterConnectAPI(params).then(() => '')
      .catch(data => data?.response?.data?.message || data);
    return result;
  };

  // 区域列表
  const regionLoading = ref(false);
  const regionList = ref<Array<ICloudRegion>>($store.state.cloudMetadata.regionList);
  const handleGetRegionList = async ({ cloudAccountID, cloudID }) => {
    if (!cloudAccountID || !cloudID) return;

    regionLoading.value = true;
    regionList.value = await cloudRegionByAccount({
      $cloudId: cloudID,
      accountID: cloudAccountID,
    }).catch(() => []);
    regionLoading.value = false;
    return regionList.value;
  };

  // 管控区域
  const nodemanCloudList = ref<Array<INodeManCloud>>([]);
  const nodemanCloudLoading = ref(false);
  const handleGetNodeManCloud = async () => {
    nodemanCloudLoading.value = true;
    nodemanCloudList.value = await nodemanCloud().catch(() => []);
    nodemanCloudLoading.value = false;
    return nodemanCloudList.value;
  };

  // vpc列表
  const vpcLoading = ref(false);
  const vpcList = ref<Array<IVpcItem>>($store.state.cloudMetadata.vpcList);
  const handleGetVPCList = async ({ region, cloudAccountID,  cloudID, resourceGroupName }) => {
    // gcpCloud 不支持vpc获取
    if (!region || !cloudAccountID || !cloudID || cloudID === 'gcpCloud') return;
    vpcLoading.value = true;
    vpcList.value = await cloudVPC({
      $cloudId: cloudID,
      accountID: cloudAccountID,
      region,
      resourceGroupName,
    }).catch(() => []);
    vpcLoading.value = false;
    return vpcList.value;
  };

  // 安全组
  const securityGroupLoading = ref(false);
  const securityGroups = ref<Array<ISecurityGroup>>($store.state.cloudMetadata.securityGroupsList);
  const handleGetSecurityGroups = async ({ region, cloudAccountID,  cloudID }) => {
    if (!region || !cloudAccountID || !cloudID) return;
    securityGroupLoading.value = true;
    securityGroups.value = await cloudSecurityGroups({
      $cloudId: cloudID,
      accountID: cloudAccountID,
      region,
    }).catch(() => []);
    securityGroupLoading.value = false;
    return securityGroups.value;
  };

  return {
    cloudAccountLoading,
    cloudAccountList,
    cloudAccounts,
    createCloudAccounts,
    deleteCloudAccounts,
    validateCloudAccounts,
    updateCloudAccounts,
    clusterConnect,
    regionLoading,
    regionList,
    handleGetRegionList,
    nodemanCloudList,
    nodemanCloudLoading,
    handleGetNodeManCloud,
    vpcLoading,
    vpcList,
    handleGetVPCList,
    securityGroupLoading,
    securityGroups,
    handleGetSecurityGroups,
    providerNameMap,
  };
}
