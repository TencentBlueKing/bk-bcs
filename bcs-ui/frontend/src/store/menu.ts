const cluster = window.i18n.t('集群')
const node = window.i18n.t('节点')
const namespace = window.i18n.t('命名空间')
const templateset = window.i18n.t('模板集')
const variable = window.i18n.t('变量管理')
const metric = window.i18n.t('Metric管理')
const app = window.i18n.t('应用')
const network = window.i18n.t('网络')
const resource = window.i18n.t('配置')
const imageHub = window.i18n.t('仓库')
const publicImage = window.i18n.t('公共镜像')
const projectImage = window.i18n.t('项目镜像')
const operateAudit = window.i18n.t('操作审计')
const eventQuery = window.i18n.t('事件查询')
const monitor = window.i18n.t('监控中心')
const release = window.i18n.t('Release列表')
const chart = window.i18n.t('Chart仓库')
const crdcontroller = window.i18n.t('组件库')
const logCollection = window.i18n.t('日志采集')
const storage = window.i18n.t('存储')
const overview = window.i18n.t('概览')
const workload = window.i18n.t('工作负载')
const dashboardNamespace = window.i18n.t('命名空间')
const customResource = window.i18n.t('自定义资源')

export interface IMenuItem {
    name: string; // 菜单中文名称
    icon?: string; // 菜单ICON
    id: string; // 菜单ID（和feature_flags接口的菜单ID匹配，用于判断菜单显示和选中的唯一标识）,注意：ID和routeName有时是不一样的
    routeName?: string; // 菜单对应的路由名称（用于路由跳转），注意：routeName不能作为唯一ID
    disable?: boolean;
    children?: IMenuItem[]; // 子菜单
}
export interface ISpecialMenuItem {
    type: 'line'; // 特殊意义的菜单项
    id?: string;
}
export interface IMenu {
    dashboardMenuList: (IMenuItem | ISpecialMenuItem)[];
    k8sMenuList: (IMenuItem | ISpecialMenuItem)[];
}
// 左侧菜单配置
const menu: IMenu = {
    dashboardMenuList: [
        {
            name: dashboardNamespace,
            icon: 'bcs-icon-namespace',
            id: 'NAMESPACE',
            routeName: 'dashboardNamespace'
        },
        {
            name: workload,
            icon: 'bcs-icon-yy-apply',
            id: 'WORKLOAD',
            children: [
                {
                    name: 'Deployments',
                    routeName: 'dashboardWorkloadDeployments',
                    id: 'dashboardWorkloadDeployments'
                },
                {
                    name: 'DaemonSets',
                    routeName: 'dashboardWorkloadDaemonSets',
                    id: 'dashboardWorkloadDaemonSets'
                },
                {
                    name: 'StatefulSets',
                    routeName: 'dashboardWorkloadStatefulSets',
                    id: 'dashboardWorkloadStatefulSets'
                },
                {
                    name: 'CronJobs',
                    routeName: 'dashboardWorkloadCronJobs',
                    id: 'dashboardWorkloadCronJobs'
                },
                {
                    name: 'Jobs',
                    routeName: 'dashboardWorkloadJobs',
                    id: 'dashboardWorkloadJobs'
                },
                {
                    name: 'Pods',
                    routeName: 'dashboardWorkloadPods',
                    id: 'dashboardWorkloadPods'
                }
            ]
        },
        {
            name: network,
            icon: 'bcs-icon-wl-network',
            id: 'NETWORK',
            children: [
                {
                    name: 'Ingresses',
                    routeName: 'dashboardNetworkIngress',
                    id: 'dashboardNetworkIngress'
                },
                {
                    name: 'Services',
                    routeName: 'dashboardNetworkService',
                    id: 'dashboardNetworkService'
                },
                {
                    name: 'Endpoints',
                    routeName: 'dashboardNetworkEndpoints',
                    id: 'dashboardNetworkEndpoints'
                }
            ]
        },
        {
            name: resource,
            icon: 'bcs-icon-zy-resource',
            children: [
                {
                    name: 'ConfigMaps',
                    routeName: 'dashboardConfigsConfigMaps',
                    id: 'dashboardConfigsConfigMaps'
                },
                {
                    name: 'Secrets',
                    routeName: 'dashboardConfigsSecrets',
                    id: 'dashboardConfigsSecrets'
                }
            ],
            id: 'CONFIGURATION'
        },
        {
            name: storage,
            icon: 'bcs-icon-data',
            children: [
                {
                    name: 'PersistentVolumes',
                    routeName: 'dashboardStoragePersistentVolumes',
                    id: 'dashboardStoragePersistentVolumes'
                },
                {
                    name: 'PersistentVolumeClaims',
                    routeName: 'dashboardStoragePersistentVolumesClaims',
                    id: 'dashboardStoragePersistentVolumesClaims'
                },
                {
                    name: 'StorageClasses',
                    routeName: 'dashboardStorageStorageClass',
                    id: 'dashboardStorageStorageClass'
                }
            ],
            id: 'STORAGE'
        },
        {
            name: 'RBAC',
            icon: 'bcs-icon-lock-line',
            children: [
                {
                    name: 'ServiceAccounts',
                    routeName: 'dashboardRbacServiceAccounts',
                    id: 'dashboardRbacServiceAccounts'
                }
            ],
            id: 'RBAC'
        },
        {
            name: 'HPA',
            icon: 'bcs-icon-hpa',
            routeName: 'dashboardHPA',
            id: 'HPA'

        },
        {
            name: customResource,
            icon: 'bcs-icon-customize',
            children: [
                {
                    name: 'CRD',
                    routeName: 'dashboardCRD',
                    id: 'dashboardCRD'
                },
                {
                    name: 'GameStatefulSets',
                    routeName: 'dashboardGameStatefulSets',
                    id: 'dashboardGameStatefulSets'
                },
                {
                    name: 'GameDeployments',
                    routeName: 'dashboardGameDeployments',
                    id: 'dashboardGameDeployments'
                },
                {
                    name: 'CustomObjects',
                    routeName: 'dashboardCustomObjects',
                    id: 'dashboardCustomObjects'
                }
            ],
            id: 'CUSTOM_RESOURCE'
        }
    ],
    k8sMenuList: [
        {
            name: cluster,
            icon: 'bcs-icon-jq-colony',
            id: 'CLUSTER',
            routeName: 'clusterMain'
        },
        {
            name: overview,
            icon: 'bcs-icon-jq-colony',
            id: 'OVERVIEW',
            routeName: 'clusterDetail'
        },
        {
            name: node,
            icon: 'bcs-icon-jd-node',
            id: 'NODE',
            routeName: 'nodeMain'
        },
        {
            name: namespace,
            icon: 'bcs-icon-namespace',
            id: 'NAMESPACE',
            routeName: 'namespace'
        },
        { type: 'line' },
        {
            name: templateset,
            icon: 'bcs-icon-templateset',
            id: 'TEMPLATESET',
            routeName: 'templateset'
        },
        {
            name: variable,
            icon: 'bcs-icon-var',
            id: 'VARIABLE',
            routeName: 'var'
        },
        {
            name: metric,
            icon: 'bcs-icon-control-center',
            id: 'METRICS',
            routeName: 'metricManage'
        },
        { type: 'line' },
        {
            name: 'Helm',
            icon: 'bcs-icon-helm',
            children: [
                {
                    name: release,
                    routeName: 'helms',
                    id: 'helms'
                },
                {
                    name: chart,
                    routeName: 'helmTplList',
                    id: 'helmTplList'
                }
            ],
            id: 'HELM'
        },
        { type: 'line' },
        {
            name: app,
            icon: 'bcs-icon-yy-apply',
            children: [
                {
                    name: 'Deployments',
                    routeName: 'deployments',
                    id: 'deployments'
                },
                {
                    name: 'DaemonSets',
                    routeName: 'daemonset',
                    id: 'daemonset'
                },
                {
                    name: 'StatefulSets',
                    routeName: 'statefulset',
                    id: 'statefulset'
                },
                {
                    name: 'Jobs',
                    routeName: 'job',
                    id: 'job'
                },
                {
                    name: 'GameStatefulSets',
                    routeName: 'gamestatefulset',
                    id: 'gamestatefulset'
                },
                {
                    name: 'GameDeployments',
                    routeName: 'gamedeployments',
                    id: 'gamedeployments'
                },
                {
                    name: 'CustomObjects',
                    routeName: 'customobjects',
                    id: 'customobjects'
                }
            ],
            id: 'WORKLOAD'
        },
        {
            name: network,
            icon: 'bcs-icon-wl-network',
            children: [
                {
                    name: 'Service',
                    routeName: 'service',
                    id: 'service'
                },
                {
                    name: 'Ingress',
                    routeName: 'resourceIngress',
                    id: 'resourceIngress'
                },
                {
                    name: 'LoadBalancer',
                    routeName: 'loadBalance',
                    id: 'loadBalance'
                }
            ],
            id: 'NETWORK'
        },
        {
            name: resource,
            icon: 'bcs-icon-zy-resource',
            children: [
                {
                    name: 'ConfigMaps',
                    routeName: 'resourceConfigmap',
                    id: 'resourceConfigmap'
                },
                {
                    name: 'Secrets',
                    routeName: 'resourceSecret',
                    id: 'resourceSecret'
                }
            ],
            id: 'CONFIGURATION'
        },
        {
            name: 'HPA',
            icon: 'bcs-icon-hpa',
            id: 'HPA',
            routeName: 'hpa'
        },
        {
            name: storage,
            icon: 'bcs-icon-data',
            children: [
                {
                    name: 'PersistentVolume',
                    routeName: 'pv',
                    id: 'pv'
                },
                {
                    name: 'PersistentVolumeClaim',
                    routeName: 'pvc',
                    id: 'pvc'
                },
                {
                    name: 'StorageClass',
                    routeName: 'storageClass',
                    id: 'storageClass'
                }
            ],
            id: 'STORAGE'
        },
        { type: 'line' },
        {
            name: logCollection,
            icon: 'bcs-icon-log-collection',
            id: 'LOG',
            routeName: 'logCrdcontroller'
        },
        {
            name: crdcontroller,
            icon: 'bcs-icon-crd1',
            id: 'COMPONENTS',
            routeName: 'dbCrdcontroller'
        },
        { type: 'line' },
        {
            name: imageHub,
            icon: 'bcs-icon-ck-store',
            children: [
                {
                    name: publicImage,
                    routeName: 'imageLibrary',
                    id: 'imageLibrary'
                },
                {
                    name: projectImage,
                    routeName: 'projectImage',
                    id: 'projectImage'
                }
            ],
            id: 'REPO'
        },
        { type: 'line' },
        {
            name: operateAudit,
            icon: 'bcs-icon-operate-audit',
            id: 'AUDIT',
            routeName: 'operateAudit'
        },
        {
            name: eventQuery,
            icon: 'bcs-icon-event-query',
            id: 'EVENT',
            routeName: 'eventQuery'
        },
        { type: 'line' },
        {
            name: monitor,
            icon: 'bcs-icon-monitors',
            id: 'MONITOR'
        }
    ]
}
export default menu
