interface INodePool {
  enableAutoscale: boolean
  nodeGroupID: string
  name: string
  autoScaling: {
    maxSize: number
    desiredSize: number
  }
}

interface ICloudTemplateDetail {
  cloudID: string
  clusterManagement: {
    availableVersion: string[]
  }
  osManagement: {
    regions: Record<string, string>
    availableVersion: string[]
  }
}

interface IRuntimeModuleParams {
  enable: boolean
  flagName: string
  defaultValue: string
  flagValueList: string[]
  networkType: string
}

type CloudID = 'tencentCloud'|'gcpCloud'|'tencentPublicCloud'|'bluekingCloud';
