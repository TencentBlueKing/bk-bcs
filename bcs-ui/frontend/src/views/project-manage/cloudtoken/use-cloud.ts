import { computed } from 'vue';

import {
  cloudAccounts as cloudAccountsAPI,
  clusterConnect as clusterConnectAPI,
  createCloudAccounts as createCloudAccountsAPI,
  deleteCloudAccounts as deleteCloudAccountsAPI,
  updateCloudAccounts as updateCloudAccountsAPI,
  validateCloudAccounts as validateCloudAccountsAPI,
} from '@/api/modules/cluster-manager';
import $store from '@/store';

export type CloudID = 'tencentCloud'|'gcpCloud'|'tencentPublicCloud';

export interface IGoogleAccount {
  gkeProjectID?: string
  serviceAccountSecret: string
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

  // 云账号列表
  const cloudAccounts = async ($cloudId: CloudID) => {
    const res = await cloudAccountsAPI({
      $cloudId,
      projectID: curProject.value.projectID,
      operator: user.value.username,
    }, { needRes: true }).catch(() => []);
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
    account: IGoogleAccount
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
  }) => {
    const result = await clusterConnectAPI(params).then(() => '')
      .catch(data => data?.response?.data?.message || data);
    return result;
  };

  return {
    cloudAccounts,
    createCloudAccounts,
    deleteCloudAccounts,
    validateCloudAccounts,
    updateCloudAccounts,
    clusterConnect,
  };
}
