import { InjectionKey, Ref } from 'vue';

export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends (infer U)[]
    ? DeepPartial<U>[]
    : T[P] extends readonly (infer U)[]
      ? DeepPartial<U>[]
      : DeepPartial<T[P]>
};

export interface IInstanceItem {
  'region': string
  'zone': string
  'vpcID': string
  'subnetID': string
  'applyNum': number // 申请数量
  'CPU': number
  'Mem': number
  'GPU': number
  'instanceType': string // 机型
  'instanceChargeType': string // 收费类型：POSTPAID_BY_HOUR 按量计费 PREPAID 按月计费
  'systemDisk': {
    'diskType': string
    'diskSize': string
  }
  'imageInfo': {
    'imageID': string
    'imageName': string
    'imageType': string
  }
  'securityGroupIDs': string[] // 安全组ID
  'isSecurityService': boolean // 默认true
  'isMonitorService': boolean // 默认true
  // 数据盘设置
  'cloudDataDisks': Array<{
    'diskType': string // 类型
    'diskSize': string // 大小
    'fileSystem': string // 文件系统
    'autoFormatAndMount': boolean // 是否格式化
    'mountTarget': string // 挂载路径
  }>
  'dockerGraphPath': string // 默认空
  'nodeRole': string // 表示是master节点还是worker节点 ：MASTER_ETCD / WORKER
  'charge': { // 包年包月计费模式时候，设置该字段；当为POSTPAID_BY_HOUR时，设置为null即可
    'period': number // 数字月数
    'renewFlag': string // 是否续费
  } | null
  'internetAccess': IInternetAccess
}

export interface IClusterData {
  projectID: string
  businessID: string
  engineType: string
  isExclusive: boolean
  clusterType: string
  creator: string
  nodes: string[]
  cloudAccountID: string
  clusterName: string
  environment: string
  provider: string
  description: string
  region: string
  labels: Record<string, string>
  extraInfo: Record<string, string>
  master: string[]
  manageType: 'INDEPENDENT_CLUSTER' | 'MANAGED_CLUSTER'
  autoGenerateMasterNodes: boolean
  clusterBasicSettings: {
    version: string
    OS: string
    area: {  // 云区域
      bkCloudID: number
    }
    clusterLevel: string
    isAutoUpgradeClusterLevel: boolean
    module: {
      masterModuleID: string
    }
  }
  clusterAdvanceSettings: {
    containerRuntime: string // 运行时
    runtimeVersion: string // 运行时版本
    clusterConnectSetting: {
      isExtranet: boolean
      subnetId: string
      securityGroup: string
    }
    networkType: string  // 网络插件 GR
  }
  nodeSettings: {
    masterLogin: {
      initLoginUsername: string
      initLoginPassword: string
      keyPair: {
        keyID: string
        keySecret: string
        keyPublic: string
      }
    }
    workerLogin: {
      initLoginUsername: string
      initLoginPassword: string
      keyPair: {
        keyID: string
        keySecret: string
        keyPublic: string
      }
    }
  }
  vpcID: string
  networkType: string // overlay underlay
  networkSettings: {
    clusterIPv4CIDR: string
    serviceIPv4CIDR: string
    maxNodePodNum: number // 单节点pod数量上限
    maxServiceNum: number
    clusterIpType: string // ipv4/ipv6/dual
    isStaticIpMode: boolean
    subnetSource: {
      new: Array<{
        zone: string
        ipCnt: number
      }>
    }
  }
  instances: Array<IInstanceItem>
}

export interface IHostNode {
  alive: number
  cloudArea: {
    id: number
    name: string
  }
  hostId: number
  hostName: string
  ip: string
  ipv6: string
  osName: string
}

export interface IKeyItem{
  'KeyID': string
  'KeyName': string
  'description': string
}
export interface IImageItem {
  imageID: string
  osName: string
  provider: string
  status: string
  alias: string
  clusters: string[]
}
export interface IImageGroup {
  name: string
  provider: string
  children: Array<IImageItem>
}

export interface ICloudProject {
  projectID: string
  projectName: string
}

export interface ICloudRegion {
  region: string
  regionName: string
  regionState: string
}

export interface INodeManCloud {
  ap_id: number
  bk_cloud_id: number
  bk_cloud_name: string
  is_visible: boolean
}

export interface IScale {
  level: string
  scale: {
    maxNodePodNum: number
    maxServiceNum: number
    cidrStep: number
  }
}

export interface IZoneItem {
  zoneID: string
  zone: string
  zoneName: string
  zoneState: string
  subnetNum?: number
}

export interface ISecurityGroup {
  securityGroupID: string
  securityGroupName: string
  description: string
}

export interface ISubnet{
  'vpcID': string
  'subnetID': string
  'subnetName': string
  'cidrRange': string
  'ipv6CidrRange': string
  'zone': string
  'availableIPAddressCount': string
  'zoneName': string
  cluster: {
    clusterID: string
    clusterName: string
  }
}

export interface IInstanceType{
  'nodeType': string
  'typeName': string
  'nodeFamily': string
  'cpu': number
  'memory': number
  'gpu': number
  'status': string
  'unitPrice': number
  'zones': string[]
  'provider': string
  'resourcePoolID': string
  'systemDisk': string
  'dataDisks': any[]
}

export const ClusterDataInjectKey: InjectionKey<Ref<Partial<IClusterData>>> = Symbol('cluster-data');

export interface IDataDisk {
  diskType: string
  diskSize: string
  autoFormatAndMount: boolean
  mountTarget: string
  fileSystem: string
}

export type InternetChargeType = 'TRAFFIC_POSTPAID_BY_HOUR' | 'BANDWIDTH_PREPAID' | 'BANDWIDTH_PACKAGE';
export interface IInternetAccess {
  publicIPAssigned: boolean
  internetMaxBandwidth: string
  internetChargeType: InternetChargeType | string
  bandwidthPackageId: string
}

export interface IVpcItem {
  vpcId: string
  name: string
  allocateIpNum: number
}
