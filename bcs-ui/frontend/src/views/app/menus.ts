import $i18n from '@/i18n/i18n-setup';

export interface IMenu {
  title?: string
  id: string
  route?: string
  icon?: string
  root?: IMenu
  tag?: string
  children?: IMenu[]
  parent?: IMenu
}

const menus: IMenu[] = [
  {
    title: $i18n.t('nav.dashboard'),
    id: 'CLUSTERRESOURCE',
    children: [
      {
        title: $i18n.t('nav.namespace'),
        icon: 'bcs-icon-namespace',
        id: 'NAMESPACE',
        route: 'dashboardNamespace',
      },
      {
        title: $i18n.t('nav.workload'),
        icon: 'bcs-icon-yy-apply',
        id: 'WORKLOAD',
        children: [
          {
            title: 'Deployments',
            route: 'dashboardWorkloadDeployments',
            id: 'DEPLOYMENT',
          },
          {
            title: 'StatefulSets',
            route: 'dashboardWorkloadStatefulSets',
            id: 'STATEFULSET',
          },
          {
            title: 'DaemonSets',
            route: 'dashboardWorkloadDaemonSets',
            id: 'DAEMONSET',
          },
          {
            title: 'Jobs',
            route: 'dashboardWorkloadJobs',
            id: 'JOB',
          },
          {
            title: 'CronJobs',
            route: 'dashboardWorkloadCronJobs',
            id: 'CRONJOB',
          },
          {
            title: 'Pods',
            route: 'dashboardWorkloadPods',
            id: 'POD',
          },
        ],
      },
      {
        title: $i18n.t('nav.network'),
        icon: 'bcs-icon-wl-network',
        id: 'NETWORK',
        children: [
          {
            title: 'Ingresses',
            route: 'dashboardNetworkIngress',
            id: 'INGRESS',
          },
          {
            title: 'Services',
            route: 'dashboardNetworkService',
            id: 'SERVICE',
          },
          {
            title: 'Endpoints',
            route: 'dashboardNetworkEndpoints',
            id: 'ENDPOINTS',
          },
        ],
      },
      {
        title: $i18n.t('nav.configuration'),
        icon: 'bcs-icon-zy-resource',
        children: [
          {
            title: 'ConfigMaps',
            route: 'dashboardConfigsConfigMaps',
            id: 'CONFIGMAP',
          },
          {
            title: 'Secrets',
            route: 'dashboardConfigsSecrets',
            id: 'SECRET',
          },
        ],
        id: 'CONFIGURATION',
      },
      {
        title: $i18n.t('nav.storage'),
        icon: 'bcs-icon-data',
        children: [
          {
            title: 'PersistentVolumes',
            route: 'dashboardStoragePersistentVolumes',
            id: 'PERSISTENTVOLUME',
          },
          {
            title: 'PersistentVolumesClaims',
            route: 'dashboardStoragePersistentVolumesClaims',
            id: 'PERSISTENTVOLUMECLAIM',
          },
          {
            title: 'StorageClass',
            route: 'dashboardStorageStorageClass',
            id: 'STORAGECLASS',
          },
        ],
        id: 'STORAGE',
      },
      {
        title: 'RBAC',
        icon: 'bcs-icon-lock-line',
        children: [
          {
            title: 'ServiceAccounts',
            route: 'dashboardRbacServiceAccounts',
            id: 'SERVICEACCOUNT',
          },
        ],
        id: 'RBAC',
      },
      {
        title: 'HPA',
        icon: 'bcs-icon-hpa',
        route: 'dashboardHPA',
        id: 'HORIZONTALPODAUTOSCALER',
      },
      {
        title: $i18n.t('nav.customResource'),
        icon: 'bcs-icon-customize',
        children: [
          {
            title: 'GameDeployments',
            route: 'dashboardGameDeployments',
            id: 'GAMEDEPLOYMENT',
          },
          {
            title: 'GameStatefulSets',
            route: 'dashboardGameStatefulSets',
            id: 'GAMESTATEFULSET',
          },
          {
            title: 'HookTemplates',
            route: 'dashboardHookTemplates',
            id: 'HOOKTEMPLATE',
          },
          {
            title: 'CustomObjects',
            route: 'dashboardCustomObjects',
            id: 'CUSTOMOBJECT',
          },
          {
            title: 'CRD',
            route: 'dashboardCRD',
            id: 'CRD',
          },
        ],
        id: 'CUSTOM_RESOURCE',
      },
    ],
  },
  {
    title: $i18n.t('nav.cluster'),
    id: 'CLUSTERMANAGE',
    route: 'clusterMain',
    children: [
      {
        title: $i18n.t('nav.clusterList'),
        icon: 'bcs-icon-jq-colony',
        route: 'clusterMain',
        id: 'CLUSTER',
      },
      {
        title: $i18n.t('nav.nodeList'),
        icon: 'bcs-icon-jd-node',
        route: 'nodeMain',
        id: 'NODE',
      },
      {
        title: $i18n.t('nav.nodeTemplate'),
        icon: 'bcs-icon-mobanpeizhi',
        route: 'nodeTemplate',
        id: 'NODETEMPLATE',
      },
      {
        title: $i18n.t('nav.cloudToken'),
        icon: 'bcs-icon-yunpingzhengguanli',
        id: 'CLOUDTOKEN',
        children: [
          {
            title: 'Tencent Cloud',
            id: 'TENCENTCLOUD',
            route: 'tencentCloud',
          },
        ],
      },
    ],
  },
  {
    title: $i18n.t('nav.deploy'),
    id: 'DEPLOYMENTMANAGE',
    route: 'releaseList',
    children: [
      {
        title: 'Helm',
        icon: 'bcs-icon-helm',
        id: 'HELM',
        children: [
          {
            title: $i18n.t('nav.releaseList'),
            id: 'RELEASELIST',
            route: 'releaseList',
          },
          {
            title: $i18n.t('nav.chartList'),
            id: 'CHARTLIST',
            route: 'chartList',
          },
        ],
      },
      {
        title: $i18n.t('nav.templateSet'),
        id: 'TEMPLATESET_v1',
        icon: 'bcs-icon-templateset',
        children: [
          {
            title: $i18n.t('deploy.templateset.list'),
            route: 'templateset',
            id: 'TEMPLATESET',
          },
          {
            title: 'Deployments',
            route: 'deployments',
            id: 'TEMPLATESET_DEPLOYMENT',
          },
          {
            title: 'StatefulSets',
            route: 'statefulset',
            id: 'TEMPLATESET_STATEFULSET',
          },
          {
            title: 'DaemonSets',
            route: 'daemonset',
            id: 'TEMPLATESET_DAEMONSET',
          },
          {
            title: 'Jobs',
            route: 'job',
            id: 'TEMPLATESET_JOB',
          },
          {
            title: 'Ingresses',
            route: 'resourceIngress',
            id: 'TEMPLATESET_INGRESSE',
          },
          {
            title: 'Services',
            route: 'service',
            id: 'TEMPLATESET_SERVICE',
          },
          {
            title: 'ConfigMaps',
            route: 'resourceConfigmap',
            id: 'TEMPLATESET_CONFIGMAP',
          },
          {
            title: 'Secrets',
            route: 'resourceSecret',
            id: 'TEMPLATESET_SECRET',
          },
          {
            title: 'HPA',
            id: 'TEMPLATESET_HPA',
            route: 'hpa',
          },
          {
            title: 'GameDeployments',
            route: 'gamedeployments',
            id: 'TEMPLATESET_GAMEDEPLOYMENT',
          },
          {
            title: 'GameStatefulSets',
            route: 'gamestatefulset',
            id: 'TEMPLATESET_GAMESTATEFULSET',
          },
          {
            title: 'CustomObjects',
            route: 'customobjects',
            id: 'TEMPLATESET_CUSTOMOBJECT',
          },
        ],
      },
      {
        title: $i18n.t('nav.variable'),
        icon: 'bcs-icon-var',
        route: 'variable',
        id: 'VARIABLE',
      },
    ],
  },
  {
    title: $i18n.t('nav.project'),
    id: 'PROJECTMANAGE',
    children: [
      {
        title: $i18n.t('nav.event'),
        id: 'EVENT',
        icon: 'bcs-icon-event-query',
        route: 'eventQuery',
      },
      {
        title: $i18n.t('nav.record'),
        id: 'AUDIT',
        icon: 'bcs-icon-operate-audit',
        route: 'operateAudit',
      },
      // {
      //   title: $i18n.t('iam.actionMap.cloud_account_manage'),
      //   id: 'CLOUDTOKEN',
      //   icon: 'bcs-icon-yunpingzhengguanli',
      //   children: [
      //     {
      //       title: 'Tencent Cloud',
      //       id: 'TENCENTCLOUD',
      //       route: 'tencentCloud',
      //     },
      //   ],
      // },
      {
        title: $i18n.t('nav.projectInfo'),
        id: 'PROJECT',
        icon: 'bcs-icon bcs-icon-apps',
        route: 'projectInfo',
      },
    ],
  },
  {
    title: $i18n.t('nav.plugin'),
    id: 'PLUGINMANAGE',
    children: [
      {
        title: $i18n.t('nav.clusterTools'),
        id: 'TOOLS',
        icon: 'bcs-icon-crd1',
        route: 'dbCrdcontroller',
      },
      {
        title: $i18n.t('nav.metric'),
        id: 'METRICS',
        icon: 'bcs-icon-control-center',
        route: 'metricManage',
      },
      {
        title: 'LoadBalancers',
        id: 'LOADBALANCERS',
        icon: 'bcs-icon bcs-icon-loadbalance',
        route: 'loadBalance',
      },
      {
        title: $i18n.t('nav.log'),
        id: 'LOG',
        icon: 'bcs-icon-log-collection',
        route: 'logCrdcontroller',
      },
      // {
      //   title: $i18n.t('nav.log'),
      //   id: 'NEW_LOG',
      //   icon: 'bcs-icon-log-collection',
      //   route: 'newLogController',
      //   tag: 'NEW',
      // },
      {
        title: $i18n.t('nav.monitor'),
        id: 'MONITOR',
        icon: 'bcs-icon-monitors',
      },
    ],
  },
];
export default menus;
