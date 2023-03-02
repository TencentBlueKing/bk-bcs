// 资源视图
const DashboardView = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/dashboard-view.vue');
const DashboardNamespace = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/namespace/namespace.vue');
const DashboardNamespaceCreate = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/namespace/create.vue');
const DashboardWorkloadDeployments = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/workload/deployments.vue');
const DashboardWorkloadDaemonSets = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/workload/daemonsets.vue');
const DashboardWorkloadStatefulSets = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/workload/statefulsets.vue');
const DashboardWorkloadCronJobs = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/workload/cronjobs.vue');
const DashboardWorkloadJobs = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/workload/jobs.vue');
const DashboardWorkloadPods = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/workload/pods.vue');
const DashboardWorkloadDetail = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/workload/detail/index.vue');

// 资源表单化创建
const DashboardFormResourceUpdate = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/resource-update/form-resource.vue');

// 自定义资源
const DashboardCRD = () => import(/* webpackChunkName: 'dashboard-custom' */'@/views/dashboard/custom/crd.vue');
const DashboardGameStatefulSets = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/custom/gamestatefulsets.vue');
const DashboardGameDeployments = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/custom/gamedeployments.vue');
const DashboardHookTemplates = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/custom/hookTemplates.vue');
const DashboardCustomObjects = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/custom/customobjects.vue');

// network
const DashboardNetworkIngress = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/network/ingress.vue');
const DashboardNetworkService = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/network/service.vue');
const DashboardNetworkEndpoints = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/network/endpoints.vue');

// configs
const DashboardConfigsConfigMaps = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/configuration/config-maps.vue');
const DashboardConfigsSecrets = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/configuration/secrets.vue');

// storage
const DashboardStoragePersistentVolumesClaims = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/storage/persistent-volumes-claims.vue');
const DashboardStoragePersistentVolumes = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/storage/persistent-volumes.vue');
const DashboardStorageStorageClass = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/storage/storage-class.vue');

// rbac
const DashboardRbacServiceAccounts = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/rbac/service-accounts.vue');

const DashboardResourceUpdate = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/resource-update/resource-update.vue');

// HPA
const DashboardHPA = () => import(/* webpackChunkName: 'dashboard' */'@/views/dashboard/hpa/hpa.vue');

const childRoutes = [
  {
    path: ':clusterId',
    name: 'dashboardHome',
    component: DashboardView,
    redirect: {
      name: 'dashboardNamespace',
    },
    children: [
      // dashboard 命名空间
      {
        path: 'namespaces',
        name: 'dashboardNamespace',
        component: DashboardNamespace,
      },
      {
        path: 'namespaces/create',
        name: 'dashboardNamespaceCreate',
        component: DashboardNamespaceCreate,
        meta: {
          menuId: 'NAMESPACE',
        },
      },
      // dashboard workload
      {
        path: 'workloads',
        name: 'dashboardWorkload',
        redirect: {
          name: 'dashboardWorkloadDeployments',
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
        component: DashboardGameStatefulSets,
      },
      // dashboard gamedeployments
      {
        path: 'gamedeployments',
        name: 'dashboardGameDeployments',
        component: DashboardGameDeployments,
      },
      // dashboard hookTemplates
      {
        path: 'hook-templates',
        name: 'dashboardHookTemplates',
        component: DashboardHookTemplates,
      },
      // dashboard customobjects
      {
        path: 'customobjects',
        name: 'dashboardCustomObjects',
        component: DashboardCustomObjects,
      },
      // dashboard workload detail
      {
        path: 'workloads/:category/namespaces/:namespace/:name',
        name: 'dashboardWorkloadDetail',
        props: route => ({ ...route.params, kind: route.query.kind, crd: route.query.crd }),
        component: DashboardWorkloadDetail,
        beforeEnter: (to, from, next) => {
          // 设置当前详情的父级菜单
          to.meta.menuId = String(to.query.kind).toUpperCase();
          next();
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
      // resource update
      {
        path: 'resource/namespaces/:namespace?/:name?',
        name: 'dashboardResourceUpdate',
        props: route => ({ ...route.params, ...route.query }),
        component: DashboardResourceUpdate,
        beforeEnter: (to, from, next) => {
          // 设置当前详情的父级菜单
          to.meta.menuId = String(to.query.kind).toUpperCase();
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
          to.meta.menuId = String(to.query.kind).toUpperCase();
          next();
        },
      },
      // hpa
      {
        path: 'hpa',
        name: 'dashboardHPA',
        component: DashboardHPA,
      },
    ],
  },
];

export default childRoutes;
