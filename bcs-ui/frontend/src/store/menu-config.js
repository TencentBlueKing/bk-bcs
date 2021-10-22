/*
 * @file menu 配置
 * @author ielgnaw <wuji0223@gmail.com>
 */

/**
 * 生成左侧导航菜单
 * @return {Object} 左侧导航菜单对象
 */
export default function menuConfig () {
    const cluster = window.i18n.t('集群')
    const overview = window.i18n.t('概览')
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
    const workload = window.i18n.t('工作负载')
    const dashboardNamespace = window.i18n.t('命名空间')
    const customResource = window.i18n.t('自定义资源')

    return {
        dashboardMenuList: [
            {
                name: dashboardNamespace,
                isSaveData: true,
                icon: 'bcs-icon-namespace',
                roleId: 'workload:menu',
                pathName: ['dashboardNamespace'],
                id: 'NAMESPACE'
            },
            {
                name: workload,
                isSaveData: true,
                icon: 'bcs-icon-yy-apply',
                roleId: 'workload:menu',
                id: 'WORKLOAD',
                children: [
                    {
                        name: 'Deployments',
                        pathName: ['dashboardWorkloadDeployments', 'Deployment']
                    },
                    {
                        name: 'DaemonSets',
                        pathName: ['dashboardWorkloadDaemonSets', 'DaemonSet']
                    },
                    {
                        name: 'StatefulSets',
                        pathName: ['dashboardWorkloadStatefulSets', 'StatefulSet']
                    },
                    {
                        name: 'CronJobs',
                        pathName: ['dashboardWorkloadCronJobs', 'CronJob']
                    },
                    {
                        name: 'Jobs',
                        pathName: ['dashboardWorkloadJobs', 'Job']
                    },
                    {
                        name: 'Pods',
                        pathName: ['dashboardWorkloadPods', 'Pod']
                    }
                ]
            },
            {
                name: network,
                isSaveData: true,
                icon: 'bcs-icon-wl-network',
                roleId: 'network:menu',
                id: 'NETWORK',
                children: [
                    {
                        name: 'Ingresses',
                        pathName: ['dashboardNetworkIngress', 'Ingress']
                    },
                    {
                        name: 'Services',
                        pathName: ['dashboardNetworkService', 'Service']
                    },
                    {
                        name: 'Endpoints',
                        pathName: ['dashboardNetworkEndpoints', 'Endpoints']
                    }
                ]
            },
            {
                name: resource,
                isSaveData: true,
                icon: 'bcs-icon-zy-resource',
                roleId: 'resource:menu',
                children: [
                    {
                        name: 'ConfigMaps',
                        pathName: ['dashboardConfigsConfigMaps', 'ConfigMap']
                    },
                    {
                        name: 'Secrets',
                        pathName: ['dashboardConfigsSecrets', 'Secret']
                    }
                ],
                id: 'CONFIGURATION'
            },
            {
                name: storage,
                isSaveData: true,
                icon: 'bcs-icon-data',
                roleId: 'storage:menu',
                children: [
                    {
                        name: 'PersistentVolumes',
                        pathName: ['dashboardStoragePersistentVolumes', 'PersistentVolume']
                    },
                    {
                        name: 'PersistentVolumeClaims',
                        pathName: ['dashboardStoragePersistentVolumesClaims', 'PersistentVolumeClaim']
                    },
                    {
                        name: 'StorageClasses',
                        pathName: ['dashboardStorageStorageClass', 'StorageClass']
                    }
                ],
                id: 'STORAGE'
            },
            {
                name: 'RBAC',
                isSaveData: true,
                icon: 'bcs-icon-lock-line',
                roleId: 'RBAC:menu',
                children: [
                    {
                        name: 'ServiceAccounts',
                        pathName: ['dashboardRbacServiceAccounts', 'ServiceAccount']
                    }
                ],
                id: 'RBAC'
            },
            {
                name: 'HPA',
                isSaveData: true,
                icon: 'bcs-icon-hpa',
                roleId: 'HPA:menu',
                pathName: ['dashboardHPA', 'HorizontalPodAutoscaler'],
                id: 'HPA'

            },
            {
                name: customResource,
                isSaveData: true,
                icon: 'bcs-icon-customize',
                roleId: 'custom:menu',
                children: [
                    {
                        name: 'CRD',
                        pathName: ['dashboardCRD', 'CRD']
                    },
                    {
                        name: 'GameStatefulSets',
                        pathName: ['dashboardGameStatefulSets', 'GameStatefulSet']
                    },
                    {
                        name: 'GameDeployments',
                        pathName: ['dashboardGameDeployments', 'GameDeployment']
                    },
                    {
                        name: 'CustomObjects',
                        pathName: ['dashboardCustomObjects', 'CustomObject']
                    }
                ],
                id: 'CUSTOM_RESOURCE'
            }
        ],
        clusterk8sMenuList: [
            {
                name: overview,
                icon: 'bcs-icon-jq-colony',
                pathName: ['clusterOverview', 'clusterNode', 'clusterNodeOverview', 'clusterInfo'],
                roleId: '[overview:menu]',
                id: 'OVERVIEW'
            },
            {
                name: node,
                isSaveData: true,
                icon: 'bcs-icon-jd-node',
                pathName: ['nodeMain'],
                roleId: 'node:menu',
                id: 'NODE'
            },
            {
                name: namespace,
                isSaveData: true,
                icon: 'bcs-icon-namespace',
                roleId: 'configuration:menu',
                pathName: ['namespace'],
                id: 'NAMESPACE'
            },
            { name: 'line' },
            {
                name: templateset,
                isSaveData: true,
                icon: 'bcs-icon-templateset',
                roleId: 'configuration:menu',
                pathName: [
                    'templateset',
                    'k8sTemplatesetDeployment',
                    'k8sTemplatesetDaemonset',
                    'k8sTemplatesetJob',
                    'k8sTemplatesetStatefulset',
                    'k8sTemplatesetService',
                    'k8sTemplatesetConfigmap',
                    'k8sTemplatesetSecret',
                    'k8sTemplatesetIngress',
                    'k8sTemplatesetHPA',
                    'K8sYamlTemplateset',
                    'instantiation'
                ],
                id: 'TEMPLATESET'
            },
            {
                name: variable,
                isSaveData: true,
                icon: 'bcs-icon-var',
                roleId: 'configuration:menu',
                pathName: ['var'],
                id: 'VARIABLE'
            },
            {
                name: metric,
                isSaveData: true,
                icon: 'bcs-icon-control-center',
                pathName: ['metricManage'],
                id: 'METRICS'
            },
            { name: 'line' },
            {
                name: 'Helm',
                isSaveData: true,
                icon: 'bcs-icon-helm',
                children: [
                    {
                        name: release,
                        pathName: ['helms', 'helmAppDetail']
                    },
                    {
                        name: chart,
                        pathName: [
                            'helmTplList',
                            'helmTplDetail',
                            'helmTplInstance'
                        ]
                    }
                ],
                id: 'HELM'
            },
            { name: 'line' },
            {
                name: app,
                isSaveData: true,
                icon: 'bcs-icon-yy-apply',
                roleId: 'app:menu',
                children: [
                    {
                        name: 'Deployments',
                        pathName: [
                            'deployments', 'deploymentsInstanceDetail', 'deploymentsInstanceDetail2',
                            'deploymentsContainerDetail', 'deploymentsContainerDetail2', 'deploymentsInstantiation'
                        ]
                    },
                    {
                        name: 'DaemonSets',
                        pathName: [
                            'daemonset', 'daemonsetInstanceDetail', 'daemonsetInstanceDetail2',
                            'daemonsetContainerDetail', 'daemonsetContainerDetail2', 'daemonsetInstantiation'
                        ]
                    },
                    {
                        name: 'StatefulSets',
                        pathName: [
                            'statefulset', 'statefulsetInstanceDetail', 'statefulsetInstanceDetail2',
                            'statefulsetContainerDetail', 'statefulsetContainerDetail2', 'statefulsetInstantiation'
                        ]
                    },
                    {
                        name: 'Jobs',
                        pathName: [
                            'job', 'jobInstanceDetail', 'jobInstanceDetail2',
                            'jobContainerDetail', 'jobContainerDetail2', 'jobInstantiation'
                        ]
                    },
                    {
                        name: 'GameStatefulSets',
                        pathName: [
                            'gamestatefulset'
                        ]
                    },
                    {
                        name: 'GameDeployments',
                        pathName: [
                            'gamedeployments'
                        ]
                    },
                    {
                        name: 'CustomObjects',
                        pathName: [
                            'customobjects'
                        ]
                    }
                ],
                id: 'WORKLOAD'
            },
            {
                name: network,
                isSaveData: true,
                icon: 'bcs-icon-wl-network',
                roleId: 'network:menu',
                children: [
                    {
                        name: 'Service',
                        pathName: ['service']
                    },
                    {
                        name: 'Ingress',
                        pathName: ['resourceIngress']
                    },
                    {
                        name: 'LoadBalancer',
                        pathName: ['loadBalance', 'loadBalanceDetail']
                    }
                ],
                id: 'NETWORK'
            },
            {
                name: resource,
                isSaveData: true,
                icon: 'bcs-icon-zy-resource',
                roleId: 'resource:menu',
                children: [
                    {
                        name: 'ConfigMaps',
                        pathName: ['resourceConfigmap']
                    },
                    {
                        name: 'Secrets',
                        pathName: ['resourceSecret']
                    }
                ],
                id: 'CONFIGURATION'
            },
            {
                name: 'HPA',
                isSaveData: true,
                icon: 'bcs-icon-hpa',
                roleId: 'configuration:menu',
                pathName: ['hpa'],
                id: 'HPA'
            },
            {
                name: storage,
                isSaveData: true,
                icon: 'bcs-icon-data',
                roleId: 'storage:menu',
                children: [
                    {
                        name: 'PersistentVolume',
                        pathName: ['pv']
                    },
                    {
                        name: 'PersistentVolumeClaim',
                        pathName: ['pvc']
                    },
                    {
                        name: 'StorageClass',
                        pathName: ['storageClass']
                    }
                ],
                id: 'STORAGE'
            },
            { name: 'line' },
            {
                name: logCollection,
                isSaveData: true,
                icon: 'bcs-icon-log-collection',
                roleId: 'logCollection:menu',
                pathName: ['logCrdcontroller', 'crdcontrollerLogInstances'],
                id: 'LOG'
            },
            {
                name: crdcontroller,
                isSaveData: true,
                icon: 'bcs-icon-crd1',
                roleId: 'crdcontroller:menu',
                pathName: ['dbCrdcontroller', 'crdcontrollerDBInstances', 'crdcontrollerPolarisInstances', 'crdcontrollerPolarisInstances'],
                id: 'COMPONENTS'
            },
            {
                name: eventQuery,
                isSaveData: true,
                icon: 'bcs-icon-event-query',
                pathName: ['eventQuery'],
                id: 'EVENT'
            },
            { name: 'line' },
            {
                name: monitor,
                isSaveData: true,
                icon: 'bcs-icon-monitors',
                externalLink: '/console/monitor/',
                pathName: [],
                id: 'MONITOR'
            }
        ],
        menuList: [
            {
                name: cluster,
                icon: 'bcs-icon-jq-colony',
                pathName: [
                    'clusterMain', 'clusterCreate', 'clusterOverview',
                    'clusterNode', 'clusterInfo', 'clusterNodeOverview', 'containerDetailForNode'
                ],
                roleId: 'cluster:menu',
                id: 'CLUSTER'
            },
            {
                name: node,
                isSaveData: true,
                icon: 'bcs-icon-jd-node',
                pathName: ['nodeMain'],
                roleId: 'node:menu',
                id: 'NODE'
            },
            {
                name: namespace,
                isSaveData: true,
                icon: 'bcs-icon-namespace',
                roleId: 'configuration:menu',
                pathName: ['namespace'],
                id: 'NAMESPACE'
            },
            { name: 'line' },
            {
                name: templateset,
                isSaveData: true,
                icon: 'bcs-icon-templateset',
                roleId: 'configuration:menu',
                pathName: [
                    'templateset',
                    'instantiation'
                ],
                id: 'TEMPLATESET'
            },
            {
                name: variable,
                isSaveData: true,
                icon: 'bcs-icon-var',
                roleId: 'configuration:menu',
                pathName: ['var'],
                id: 'VARIABLE'
            },
            {
                name: metric,
                isSaveData: true,
                icon: 'bcs-icon-control-center',
                pathName: ['metricManage'],
                id: 'METRICS'
            },
            { name: 'line' },
            {
                name: app,
                isSaveData: true,
                icon: 'bcs-icon-yy-apply',
                pathName: ['instanceDetail', 'instanceDetail2', 'containerDetail', 'containerDetail2'],
                roleId: 'app:menu',
                id: 'WORKLOAD'
            },
            {
                name: network,
                isSaveData: true,
                icon: 'bcs-icon-wl-network',
                roleId: 'network:menu',
                children: [
                    {
                        name: 'Service',
                        pathName: ['service']
                    },
                    {
                        name: 'Ingress',
                        pathName: ['resourceIngress']
                    },
                    {
                        name: 'LoadBalancer',
                        pathName: ['loadBalance', 'loadBalanceDetail']
                    },
                    {
                        name: 'CloudLoadBalancer',
                        pathName: ['cloudLoadBalance', 'cloudLoadBalanceDetail']
                    }
                ],
                id: 'NETWORK'
            },
            {
                name: resource,
                isSaveData: true,
                icon: 'bcs-icon-zy-resource',
                roleId: 'resource:menu',
                children: [
                    {
                        name: 'ConfigMaps',
                        pathName: ['resourceConfigmap']
                    },
                    {
                        name: 'Secrets',
                        pathName: ['resourceSecret']
                    }
                ],
                id: 'CONFIGURATION'
            },
            {
                name: 'HPA',
                isSaveData: true,
                icon: 'bcs-icon-hpa',
                roleId: 'configuration:menu',
                pathName: ['hpa'],
                id: 'HPA'
            },
            { name: 'line' },
            {
                name: imageHub,
                icon: 'bcs-icon-ck-store',
                roleId: 'repo:menu',
                isSaveData: false,
                children: [
                    {
                        name: publicImage,
                        pathName: ['imageLibrary']
                    },
                    {
                        name: projectImage,
                        pathName: ['projectImage']
                    }
                ],
                id: 'REPO'
            },
            { name: 'line' },
            {
                name: operateAudit,
                icon: 'bcs-icon-operate-audit',
                pathName: ['operateAudit'],
                isSaveData: false,
                id: 'AUDIT'
            },
            {
                name: eventQuery,
                isSaveData: true,
                icon: 'bcs-icon-event-query',
                pathName: ['eventQuery'],
                id: 'EVENT'
            },
            { name: 'line' },
            {
                name: monitor,
                isSaveData: true,
                icon: 'bcs-icon-monitors',
                externalLink: '/console/monitor/',
                pathName: [],
                id: 'MONITOR'
            }
        ],
        k8sMenuList: [
            {
                name: cluster,
                icon: 'bcs-icon-jq-colony',
                pathName: [
                    'clusterMain', 'clusterCreate', 'clusterOverview',
                    'clusterNode', 'clusterInfo', 'clusterNodeOverview', 'containerDetailForNode'
                ],
                roleId: 'cluster:menu',
                id: 'CLUSTER'
            },
            {
                name: node,
                isSaveData: true,
                icon: 'bcs-icon-jd-node',
                pathName: ['nodeMain'],
                roleId: 'node:menu',
                id: 'NODE'
            },
            {
                name: namespace,
                isSaveData: true,
                icon: 'bcs-icon-namespace',
                roleId: 'configuration:menu',
                pathName: ['namespace'],
                id: 'NAMESPACE'
            },
            { name: 'line' },
            {
                name: templateset,
                isSaveData: true,
                icon: 'bcs-icon-templateset',
                roleId: 'configuration:menu',
                pathName: [
                    'templateset',
                    'k8sTemplatesetDeployment',
                    'k8sTemplatesetDaemonset',
                    'k8sTemplatesetJob',
                    'k8sTemplatesetStatefulset',
                    'k8sTemplatesetService',
                    'k8sTemplatesetConfigmap',
                    'k8sTemplatesetSecret',
                    'k8sTemplatesetIngress',
                    'k8sTemplatesetHPA',
                    'K8sYamlTemplateset',
                    'instantiation'
                ],
                id: 'TEMPLATESET'
            },
            {
                name: variable,
                isSaveData: true,
                icon: 'bcs-icon-var',
                roleId: 'configuration:menu',
                pathName: ['var'],
                id: 'VARIABLE'
            },
            {
                name: metric,
                isSaveData: true,
                icon: 'bcs-icon-control-center',
                pathName: ['metricManage'],
                id: 'METRICS'
            },
            { name: 'line' },
            {
                name: 'Helm',
                isSaveData: true,
                icon: 'bcs-icon-helm',
                children: [
                    {
                        name: release,
                        pathName: ['helms', 'helmAppDetail']
                    },
                    {
                        name: chart,
                        pathName: [
                            'helmTplList',
                            'helmTplDetail',
                            'helmTplInstance'
                        ]
                    }
                ],
                id: 'HELM'
            },
            { name: 'line' },
            {
                name: app,
                isSaveData: true,
                icon: 'bcs-icon-yy-apply',
                roleId: 'app:menu',
                children: [
                    {
                        name: 'Deployments',
                        pathName: [
                            'deployments', 'deploymentsInstanceDetail', 'deploymentsInstanceDetail2',
                            'deploymentsContainerDetail', 'deploymentsContainerDetail2', 'deploymentsInstantiation'
                        ]
                    },
                    {
                        name: 'DaemonSets',
                        pathName: [
                            'daemonset', 'daemonsetInstanceDetail', 'daemonsetInstanceDetail2',
                            'daemonsetContainerDetail', 'daemonsetContainerDetail2', 'daemonsetInstantiation'
                        ]
                    },
                    {
                        name: 'StatefulSets',
                        pathName: [
                            'statefulset', 'statefulsetInstanceDetail', 'statefulsetInstanceDetail2',
                            'statefulsetContainerDetail', 'statefulsetContainerDetail2', 'statefulsetInstantiation'
                        ]
                    },
                    {
                        name: 'Jobs',
                        pathName: [
                            'job', 'jobInstanceDetail', 'jobInstanceDetail2',
                            'jobContainerDetail', 'jobContainerDetail2', 'jobInstantiation'
                        ]
                    },
                    {
                        name: 'GameStatefulSets',
                        pathName: [
                            'gamestatefulset'
                        ]
                    },
                    {
                        name: 'GameDeployments',
                        pathName: [
                            'gamedeployments'
                        ]
                    },
                    {
                        name: 'CustomObjects',
                        pathName: [
                            'customobjects'
                        ]
                    }
                ],
                id: 'WORKLOAD'
            },
            {
                name: network,
                isSaveData: true,
                icon: 'bcs-icon-wl-network',
                roleId: 'network:menu',
                children: [
                    {
                        name: 'Service',
                        pathName: ['service']
                    },
                    {
                        name: 'Ingress',
                        pathName: ['resourceIngress']
                    },
                    {
                        name: 'LoadBalancer',
                        pathName: ['loadBalance', 'loadBalanceDetail']
                    }
                ],
                id: 'NETWORK'
            },
            {
                name: resource,
                isSaveData: true,
                icon: 'bcs-icon-zy-resource',
                roleId: 'resource:menu',
                children: [
                    {
                        name: 'ConfigMaps',
                        pathName: ['resourceConfigmap']
                    },
                    {
                        name: 'Secrets',
                        pathName: ['resourceSecret']
                    }
                ],
                id: 'CONFIGURATION'
            },
            {
                name: 'HPA',
                isSaveData: true,
                icon: 'bcs-icon-hpa',
                roleId: 'configuration:menu',
                pathName: ['hpa'],
                id: 'HPA'
            },
            {
                name: storage,
                isSaveData: true,
                icon: 'bcs-icon-data',
                roleId: 'storage:menu',
                children: [
                    {
                        name: 'PersistentVolume',
                        pathName: ['pv']
                    },
                    {
                        name: 'PersistentVolumeClaim',
                        pathName: ['pvc']
                    },
                    {
                        name: 'StorageClass',
                        pathName: ['storageClass']
                    }
                ],
                id: 'STORAGE'
            },
            { name: 'line' },
            {
                name: logCollection,
                isSaveData: true,
                icon: 'bcs-icon-log-collection',
                roleId: 'logCollection:menu',
                pathName: ['logCrdcontroller', 'crdcontrollerLogInstances'],
                id: 'LOG'
            },
            {
                name: crdcontroller,
                isSaveData: true,
                icon: 'bcs-icon-crd1',
                roleId: 'crdcontroller:menu',
                pathName: ['dbCrdcontroller', 'crdcontrollerDBInstances', 'crdcontrollerPolarisInstances'],
                id: 'COMPONENTS'
            },
            { name: 'line' },
            {
                name: imageHub,
                icon: 'bcs-icon-ck-store',
                roleId: 'repo:menu',
                isSaveData: false,
                children: [
                    {
                        name: publicImage,
                        pathName: ['imageLibrary']
                    },
                    {
                        name: projectImage,
                        pathName: ['projectImage']
                    }
                ],
                id: 'REPO'
            },
            { name: 'line' },
            {
                name: operateAudit,
                icon: 'bcs-icon-operate-audit',
                pathName: ['operateAudit'],
                isSaveData: false,
                id: 'AUDIT'
            },
            {
                name: eventQuery,
                isSaveData: true,
                icon: 'bcs-icon-event-query',
                pathName: ['eventQuery'],
                id: 'EVENT'
            },
            { name: 'line' },
            {
                name: monitor,
                isSaveData: true,
                icon: 'bcs-icon-monitors',
                externalLink: '/console/monitor/',
                pathName: [],
                id: 'MONITOR'
            }
        ]
    }
}
