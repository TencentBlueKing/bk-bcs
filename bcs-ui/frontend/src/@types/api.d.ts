interface INodePool {
  enableAutoscale: boolean
  nodeGroupID: string
  name: string
  autoScaling: {
    maxSize: number
    desiredSize: number
  }
}
