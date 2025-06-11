import $store from '@/store';
// 资源视图
const DashboardView = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/dashboard-view.vue');
// const ResourceView = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/resource-view.vue');
const DashboardWorkloadDeployments = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/workload/deployments.vue');
const DashboardWorkloadDaemonSets = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/workload/daemonsets.vue');
const DashboardWorkloadStatefulSets = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/workload/statefulsets.vue');
const DashboardWorkloadCronJobs = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/workload/cronjobs.vue');
const DashboardWorkloadJobs = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/workload/jobs.vue');
const DashboardWorkloadPods = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/workload/pods.vue');
const DashboardWorkloadDetail = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/workload/detail/index.vue');

// 资源表单化创建
const DashboardFormResourceUpdate = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/resource-update/form-resource.vue');

// 自定义资源
const DashboardCRD = () => import(/* webpackChunkName: 'dashboard-custom' */'@/views/resource-view/custom/crd.vue');
const DashboardGameStatefulSets = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/custom/gamestatefulsets.vue');
const DashboardGameDeployments = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/custom/gamedeployments.vue');
const DashboardHookTemplates = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/custom/hookTemplates.vue');
const DashboardCustomObjects = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/custom/customobjects.vue');

// network
const DashboardNetworkIngress = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/network/ingress.vue');
const DashboardNetworkService = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/network/service.vue');
const DashboardNetworkEndpoints = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/network/endpoints.vue');

// configs
const DashboardConfigsConfigMaps = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/configuration/config-maps.vue');
const DashboardConfigsSecrets = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/configuration/secrets.vue');
const dashboardConfigsBscpConfigs = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/configuration/bscp-configs.vue');

// storage
const DashboardStoragePersistentVolumesClaims = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/storage/persistent-volumes-claims.vue');
const DashboardStoragePersistentVolumes = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/storage/persistent-volumes.vue');
const DashboardStorageStorageClass = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/storage/storage-class.vue');

// rbac
const DashboardRbacServiceAccounts = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/rbac/service-accounts.vue');

const DashboardResourceUpdate = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/resource-update/resource-update.vue');

// HPA
const DashboardHPA = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/hpa/hpa.vue');

const UpdateRecord = () => import(/* webpackChunkName: 'dashboard' */'@/views/resource-view/workload/update-record.vue');

export default [
  {
    path: 'clusters/:clusterId',
    name: 'dashboardHome',
    component: DashboardView,
    redirect: {
      name: 'dashboardWorkload',
      params: {
        projectCode: $store.getters.curProjectCode,
      },
    },
    children: [
      // dashboard workload
      {
        path: 'workloads',
        name: 'dashboardWorkload',
        redirect: {
          name: 'dashboardWorkloadDeployments',
          params: {
            projectCode: $store.getters.curProjectCode,
            clusterId: '-',
          },
        },
      },
      // dashboard workload deployments
      {
        path: 'workloads/deployments',
        name: 'dashboardWorkloadDeployments',
        component: DashboardWorkloadDeployments,
      },
      // dashboard workload daemonsets
      {
        path: 'workloads/daemonsets',
        name: 'dashboardWorkloadDaemonSets',
        component: DashboardWorkloadDaemonSets,
      },
      // dashboard workload statefulsets
      {
        path: 'workloads/statefulsets',
        name: 'dashboardWorkloadStatefulSets',
        component: DashboardWorkloadStatefulSets,
      },
      // dashboard workload cronjobs
      {
        path: 'workloads/cronjobs',
        name: 'dashboardWorkloadCronJobs',
        component: DashboardWorkloadCronJobs,
      },
      // dashboard workload jobs
      {
        path: 'workloads/jobs',
        name: 'dashboardWorkloadJobs',
        component: DashboardWorkloadJobs,
      },
      // dashboard workload pods
      {
        path: 'workloads/pods',
        name: 'dashboardWorkloadPods',
        component: DashboardWorkloadPods,
      },
      {
        path: 'crds',
        name: 'dashboardCRD',
        component: DashboardCRD,
      },
      // dashboard gamestatefulsets
      {
        path: 'gamestatefulsets',
        name: 'dashboardGameStatefulSets',
        props: route => ({ ...route.params, crd: route.query.crd }),
        component: DashboardGameStatefulSets,
      },
      // dashboard gamedeployments
      {
        path: 'gamedeployments',
        name: 'dashboardGameDeployments',
        props: route => ({ ...route.params, crd: route.query.crd }),
        component: DashboardGameDeployments,
      },
      // dashboard hookTemplates
      {
        path: 'hook-templates',
        name: 'dashboardHookTemplates',
        props: route => ({ ...route.params, crd: route.query.crd }),
        component: DashboardHookTemplates,
      },
      // dashboard customobjects
      {
        path: 'customobjects',
        name: 'dashboardCustomObjects',
        props: route => ({ ...route.params, crd: route.query.crd, scope: route.query.scope, kind: route.query.kind }),
        component: DashboardCustomObjects,
        meta: {
          menuId: 'CLUSTERRESOURCE',
        },
      },
      // network
      {
        path: 'networks',
        name: 'dashboardNetwork',
        redirect: {
          name: 'dashboardNetworkIngress',
        },
      },
      // network ingress
      {
        path: 'networks/ingress',
        name: 'dashboardNetworkIngress',
        component: DashboardNetworkIngress,
      },
      // network service
      {
        path: 'networks/service',
        name: 'dashboardNetworkService',
        component: DashboardNetworkService,
      },
      // network endpoints
      {
        path: 'networks/endpoints',
        name: 'dashboardNetworkEndpoints',
        component: DashboardNetworkEndpoints,
      },
      // storage
      {
        path: 'storages',
        name: 'dashboardStorage',
        redirect: {
          name: 'dashboardStoragePersistentVolumes',
        },
      },
      // storage persistent-volumes
      {
        path: 'storages/persistent-volumes',
        name: 'dashboardStoragePersistentVolumes',
        component: DashboardStoragePersistentVolumes,
      },
      // storage persistent-volumes-claims
      {
        path: 'storages/persistent-volumes-claims',
        name: 'dashboardStoragePersistentVolumesClaims',
        component: DashboardStoragePersistentVolumesClaims,
      },
      // storage storage-class
      {
        path: 'storages/storage-class',
        name: 'dashboardStorageStorageClass',
        component: DashboardStorageStorageClass,
      },
      // configs
      {
        path: 'configs',
        name: 'dashboardConfigs',
        redirect: {
          name: 'dashboardConfigsConfigMaps',
        },
      },
      // configs config-maps
      {
        path: 'configs/config-maps',
        name: 'dashboardConfigsConfigMaps',
        component: DashboardConfigsConfigMaps,
      },
      // configs secrets
      {
        path: 'configs/secrets',
        name: 'dashboardConfigsSecrets',
        component: DashboardConfigsSecrets,
      },
      // configs bscp-configs
      {
        path: 'configs/bscp-configs',
        name: 'dashboardConfigsBscpConfigs',
        component: dashboardConfigsBscpConfigs,
      },
      // rbac
      {
        path: 'rbac',
        name: 'dashboardRbac',
        redirect: {
          name: 'dashboardRbacServiceAccounts',
        },
      },
      // rbac service accounts
      {
        path: 'rbac/service-accounts',
        name: 'dashboardRbacServiceAccounts',
        component: DashboardRbacServiceAccounts,
      },
      // hpa
      {
        path: 'hpa',
        name: 'dashboardHPA',
        component: DashboardHPA,
      },
    ],
  },
  {
    path: 'clusters/:clusterId',
    component: DashboardView,
    children: [
      // dashboard workload detail
      {
        path: 'workloads/:category/namespaces/:namespace/:name',
        name: 'dashboardWorkloadDetail',
        props: route => ({
          ...route.params,
          kind: route.query.kind,
          crd: route.query.crd,
          pod: route.query.pod,
          container: route.query.container,
        }),
        component: DashboardWorkloadDetail,
        beforeEnter: (to, from, next) => {
          // 设置当前详情的父级菜单
          to.meta.menuId = String(to.query.kind).toUpperCase();
          next();
        },
      },
      // resource update
      {
        path: 'resource/namespaces/:namespace?/:name?',
        name: 'dashboardResourceUpdate',
        props: route => ({ ...route.params, ...route.query }),
        component: DashboardResourceUpdate,
        beforeEnter: (to, from, next) => {
          // 设置当前详情的父级菜单
          to.meta.menuId = String(to.query.menuId || to.query.kind).toUpperCase();
          next();
        },
      },
      // form resource update
      {
        path: 'form-resource/namespaces/:namespace?/:name?',
        name: 'dashboardFormResourceUpdate',
        props: route => ({ ...route.params, ...route.query }),
        component: DashboardFormResourceUpdate,
        beforeEnter: (to, from, next) => {
          to.meta.menuId = String(to.query.menuId || to.query.kind).toUpperCase();
          next();
        },
      },
      // update record
      {
        path: 'workloads/:category/:namespace/:name/record',
        name: 'workloadRecord',
        props: route => ({ ...route.params, ...route.query }),
        component: UpdateRecord,
        beforeEnter: (to, from, next) => {
          to.meta.menuId = String(to.query.menuId || to.query.kind).toUpperCase();
          next();
        },
      },
    ],
  },
];
