import $i18n from '@/i18n/i18n-setup';

export interface IMenu {
  title?: string
  id: string
  route?: string
  icon?: string
  root?: IMenu
  tag?: string
  children?: IMenu[]
}

const menus: IMenu[] = [
  {
    title: $i18n.t('资源视图'),
    id: 'CLUSTERRESOURCE',
    children: [
      {
        title: $i18n.t('命名空间'),
        icon: 'bcs-icon-namespace',
        id: 'NAMESPACE',
        route: 'dashboardNamespace',
      },
      {
        title: $i18n.t('工作负载'),
        icon: 'bcs-icon-yy-apply',
        id: 'WORKLOAD',
        children: [
          {
            title: 'Deployments',
            route: 'dashboardWorkloadDeployments',
            id: 'DEPLOYMENT',
          },
          {
            title: 'DaemonSets',
            route: 'dashboardWorkloadDaemonSets',
            id: 'DAEMONSET',
          },
          {
            title: 'StatefulSets',
            route: 'dashboardWorkloadStatefulSets',
            id: 'STATEFULSET',
          },
          {
            title: 'CronJobs',
            route: 'dashboardWorkloadCronJobs',
            id: 'CRONJOB',
          },
          {
            title: 'Jobs',
            route: 'dashboardWorkloadJobs',
            id: 'JOB',
          },
          {
            title: 'Pods',
            route: 'dashboardWorkloadPods',
            id: 'POD',
          },
        ],
      },
      {
        title: $i18n.t('网络'),
        icon: 'bcs-icon-wl-network',
        id: 'NETWORK',
        children: [
          {
            title: 'Ingress',
            route: 'dashboardNetworkIngress',
            id: 'INGRES',
          },
          {
            title: 'Services',
            route: 'dashboardNetworkService',
            id: 'SERVICE',
          },
          {
            title: 'Endpoints',
            route: 'dashboardNetworkEndpoints',
            id: 'ENDPOINT',
          },
        ],
      },
      {
        title: $i18n.t('配置'),
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
        title: $i18n.t('存储'),
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
            id: 'PERSISTENTVOLUMESCLAIM',
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
        id: 'HPA',
      },
      {
        title: $i18n.t('自定义资源'),
        icon: 'bcs-icon-customize',
        children: [
          {
            title: 'CRD',
            route: 'dashboardCRD',
            id: 'CRD',
          },
          {
            title: 'GameStatefulSets',
            route: 'dashboardGameStatefulSets',
            id: 'GAMESTATEFULSET',
          },
          {
            title: 'GameDeployments',
            route: 'dashboardGameDeployments',
            id: 'GAMEDEPLOYMENT',
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
        ],
        id: 'CUSTOM_RESOURCE',
      },
    ],
  },
  {
    title: $i18n.t('集群管理'),
    id: 'CLUSTERMANAGE',
    route: 'clusterMain',
    children: [
      {
        title: $i18n.t('集群总览'),
        icon: 'bcs-icon-jq-colony',
        route: 'clusterMain',
        id: 'CLUSTER',
      },
      {
        title: $i18n.t('节点列表'),
        icon: 'bcs-icon-jd-node',
        route: 'nodeMain',
        id: 'NODE',
      },
      {
        title: $i18n.t('节点模板'),
        icon: 'bcs-icon-mobanpeizhi',
        route: 'nodeTemplate',
        id: 'NODETEMPLATE',
      },
    ],
  },
  {
    title: $i18n.t('部署管理'),
    id: 'DEPLOYMENTMANAGE',
    route: 'releaseList',
    children: [
      {
        title: 'Helm',
        icon: 'bcs-icon-helm',
        id: 'HELM',
        children: [
          {
            title: $i18n.t('Release列表'),
            id: 'RELEASELIST',
            route: 'releaseList',
          },
        ],
      },
      {
        title: $i18n.t('变量管理'),
        icon: 'bcs-icon-var',
        route: 'variable',
        id: 'VARIABLE',
      },
      {
        title: $i18n.t('模板集1.0'),
        id: 'TEMPLATESET_v1.0',
        icon: 'bcs-icon-templateset',
        children: [
          {
            title: $i18n.t('模板集'),
            route: 'templateset',
            id: 'TEMPLATESET',
          },
          {
            title: 'Deployments',
            route: 'deployments',
            id: 'TEMPLATESET_DEPLOYMENT',
          },
          {
            title: 'DaemonSets',
            route: 'daemonset',
            id: 'TEMPLATESET_DAEMONSET',
          },
          {
            title: 'StatefulSets',
            route: 'statefulset',
            id: 'TEMPLATESET_STATEFULSET',
          },
          {
            title: 'Jobs',
            route: 'job',
            id: 'TEMPLATESET_JOB',
          },
          {
            title: 'GameStatefulSets',
            route: 'gamestatefulset',
            id: 'TEMPLATESET_GAMESTATEFULSET',
          },
          {
            title: 'GameDeployments',
            route: 'gamedeployments',
            id: 'TEMPLATESET_GAMEDEPLOYMENT',
          },
          {
            title: 'CustomObjects',
            route: 'customobjects',
            id: 'TEMPLATESET_CUSTOMOBJECT',
          },
          {
            title: 'Services',
            route: 'service',
            id: 'TEMPLATESET_SERVICE',
          },
          {
            title: 'Ingresses',
            route: 'resourceIngress',
            id: 'TEMPLATESET_INGRESSE',
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
        ],
      },
      {
        title: $i18n.t('仓库'),
        id: 'REPO',
        icon: 'bcs-icon-ck-store',
        children: [
          {
            title: $i18n.t('镜像'),
            id: 'IMAGE',
            route: 'depotMain',
          },
          {
            title: $i18n.t('Charts包'),
            id: 'CHARTLIST',
            route: 'chartList',
          },
        ],
      },
    ],
  },
  {
    title: $i18n.t('项目管理'),
    id: 'PROJECTMANAGE',
    children: [
      {
        title: $i18n.t('事件查询'),
        id: 'EVENT',
        icon: 'bcs-icon-event-query',
        route: 'eventQuery',
      },
      {
        title: $i18n.t('操作记录'),
        id: 'AUDIT',
        icon: 'bcs-icon-operate-audit',
        route: 'operateAudit',
      },
      // {
      //   title: $i18n.t('云凭证管理'),
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
      // {
      //   title: $i18n.t('项目信息'),
      //   id: 'PROJECT',
      //   route: 'project',
      // },
    ],
  },
  {
    title: $i18n.t('插件管理'),
    id: 'PLUGINMANAGE',
    children: [
      {
        title: $i18n.t('组件库'),
        id: 'TOOLS',
        icon: 'bcs-icon-crd1',
        route: 'dbCrdcontroller',
      },
      {
        title: $i18n.t('Metric 管理'),
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
        title: $i18n.t('日志采集'),
        id: 'LOG',
        icon: 'bcs-icon-log-collection',
        route: 'logCrdcontroller',
      },
      // {
      //   title: $i18n.t('日志采集'),
      //   id: 'NEW_LOG',
      //   icon: 'bcs-icon-log-collection',
      //   route: 'newLogController',
      //   tag: 'NEW',
      // },
      {
        title: $i18n.t('容器监控'),
        id: 'MONITOR',
        icon: 'bcs-icon-monitors',
      },
    ],
  },
];
export default menus;
