import { has } from 'lodash';
import { computed } from 'vue';
import { Route } from 'vue-router';

import { useAppData } from '@/composables/use-app';
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
  meta?: Record<string, any>
}

export type MenuID =
  'CLUSTERRESOURCE'
  |'WORKLOAD'
  |'DEPLOYMENT'
  |'STATEFULSET'
  |'DAEMONSET'
  |'JOB'
  |'CRONJOB'
  |'POD'
  |'NETWORK'
  |'INGRESS'
  |'SERVICE'
  |'ENDPOINTS'
  |'CONFIGURATION'
  |'CONFIGMAP'
  |'SECRET'
  |'STORAGE'
  |'PERSISTENTVOLUME'
  |'PERSISTENTVOLUMECLAIM'
  |'STORAGECLASS'
  |'RBAC'
  |'SERVICEACCOUNT'
  |'HORIZONTALPODAUTOSCALER'
  |'CRD'
  |'CUSTOM_GAME_RESOURCE'
  |'GAMEDEPLOYMENT'
  |'GAMESTATEFULSET'
  |'HOOKTEMPLATE'
  |'CLUSTERMANAGE'
  |'CLUSTER'
  |'NODETEMPLATE'
  |'DEPLOYMENTMANAGE'
  |'HELM'
  |'RELEASELIST'
  |'CHARTLIST'
  |'TEMPLATESET_v1'
  |'TEMPLATESET'
  |'TEMPLATESET_DEPLOYMENT'
  |'TEMPLATESET_STATEFULSET'
  |'TEMPLATESET_DAEMONSET'
  |'TEMPLATESET_JOB'
  |'TEMPLATESET_INGRESSE'
  |'TEMPLATESET_SERVICE'
  |'TEMPLATESET_CONFIGMAP'
  |'TEMPLATESET_SECRET'
  |'TEMPLATESET_HPA'
  |'TEMPLATESET_GAMEDEPLOYMENT'
  |'TEMPLATESET_GAMESTATEFULSET'
  |'TEMPLATESET_CUSTOMOBJECT'
  |'TEMPLATE_FILE'
  |'VARIABLE'
  |'PLATFORMMANAGE'
  |'PLATFORMPROJECT'
  |'PROJECTMANAGE'
  |'EVENT'
  |'AUDIT'
  |'CLOUDTOKEN'
  |'TENCENTCLOUD'
  |'TENCENTPUBLICCLOUD'
  |'GOOGLECLOUD'
  |'AZURECLOUD'
  |'HUAWEICLOUD'
  |'AMAZONCLOUD'
  |'PROJECT'
  |'PLUGINMANAGE'
  |'TOOLS'
  |'METRICS'
  |'LOG'
  |'MONITOR'
  |'PROJECTQUOTAS'
  |'SERVICEMESH';

export interface MenuItem {
  title: string
  id: MenuID
  icon?: string
  route?: string
  meta?: Record<string, any>
  children?: MenuItem[]
}

export default function useMenu() {
  const menusData = computed<MenuItem[]>(() => [
    {
      title: $i18n.t('nav.dashboard'),
      id: 'CLUSTERRESOURCE',
      children: [
        {
          title: $i18n.t('nav.workload'),
          icon: 'bcs-icon-yy-apply',
          id: 'WORKLOAD',
          route: 'dashboardWorkloadDeployments',
          children: [
            {
              title: 'Deployments',
              route: 'dashboardWorkloadDeployments',
              id: 'DEPLOYMENT',
              meta: {
                kind: 'Deployment',
              },
            },
            {
              title: 'StatefulSets',
              route: 'dashboardWorkloadStatefulSets',
              id: 'STATEFULSET',
              meta: {
                kind: 'StatefulSet',
              },
            },
            {
              title: 'DaemonSets',
              route: 'dashboardWorkloadDaemonSets',
              id: 'DAEMONSET',
              meta: {
                kind: 'DaemonSet',
              },
            },
            {
              title: 'Jobs',
              route: 'dashboardWorkloadJobs',
              id: 'JOB',
              meta: {
                kind: 'Job',
              },
            },
            {
              title: 'CronJobs',
              route: 'dashboardWorkloadCronJobs',
              id: 'CRONJOB',
              meta: {
                kind: 'CronJob',
              },
            },
            {
              title: 'Pods',
              route: 'dashboardWorkloadPods',
              id: 'POD',
              meta: {
                kind: 'Pod',
              },
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
              meta: {
                kind: 'Ingress',
              },
            },
            {
              title: 'Services',
              route: 'dashboardNetworkService',
              id: 'SERVICE',
              meta: {
                kind: 'Service',
              },
            },
            {
              title: 'Endpoints',
              route: 'dashboardNetworkEndpoints',
              id: 'ENDPOINTS',
              meta: {
                kind: 'Endpoints',
              },
            },
          ],
        },
        {
          title: $i18n.t('nav.configuration'),
          icon: 'bcs-icon-zy-resource',
          children: [
            {
              title: 'BscpConfigs',
              route: 'dashboardConfigsBscpConfigs',
              id: 'BSCPCONFIG',
              meta: {
                kind: 'BscpConfig',
              },
              tag: 'NEW',
            },
            {
              title: 'ConfigMaps',
              route: 'dashboardConfigsConfigMaps',
              id: 'CONFIGMAP',
              meta: {
                kind: 'ConfigMap',
              },
            },
            {
              title: 'Secrets',
              route: 'dashboardConfigsSecrets',
              id: 'SECRET',
              meta: {
                kind: 'Secret',
              },
            },
          ],
          id: 'CONFIGURATION',
          tag: 'NEW',
        },
        {
          title: $i18n.t('nav.storage'),
          icon: 'bcs-icon-data',
          children: [
            {
              title: 'PersistentVolumes',
              route: 'dashboardStoragePersistentVolumes',
              id: 'PERSISTENTVOLUME',
              meta: {
                kind: 'PersistentVolume',
              },
            },
            {
              title: 'PersistentVolumesClaims',
              route: 'dashboardStoragePersistentVolumesClaims',
              id: 'PERSISTENTVOLUMECLAIM',
              meta: {
                kind: 'PersistentVolumeClaim',
              },
            },
            {
              title: 'StorageClass',
              route: 'dashboardStorageStorageClass',
              id: 'STORAGECLASS',
              meta: {
                kind: 'StorageClass',
              },
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
              meta: {
                kind: 'ServiceAccount',
              },
            },
          ],
          id: 'RBAC',
        },
        {
          title: 'HPA',
          icon: 'bcs-icon-hpa',
          route: 'dashboardHPA',
          id: 'HORIZONTALPODAUTOSCALER',
          meta: {
            kind: 'HorizontalPodAutoscaler',
          },
        },
        {
          title: 'CRD',
          icon: 'bcs-icon-crd-3',
          route: 'dashboardCRD',
          id: 'CRD',
          meta: {
            kind: 'CustomResourceDefinition',
          },
        },
        {
          title: $i18n.t('nav.tkexCRD'),
          icon: 'bcs-icon-bcs',
          children: [
            {
              title: 'GameDeployments',
              route: 'dashboardGameDeployments',
              id: 'GAMEDEPLOYMENT',
              meta: {
                kind: 'GameDeployment',
              },
            },
            {
              title: 'GameStatefulSets',
              route: 'dashboardGameStatefulSets',
              id: 'GAMESTATEFULSET',
              meta: {
                kind: 'GameStatefulSet',
              },
            },
            {
              title: 'HookTemplates',
              route: 'dashboardHookTemplates',
              id: 'HOOKTEMPLATE',
              meta: {
                kind: 'HookTemplate',
              },
            },
          ],
          id: 'CUSTOM_GAME_RESOURCE',
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
          title: $i18n.t('nav.nodeTemplate'),
          icon: 'bcs-icon-mobanpeizhi',
          route: 'nodeTemplate',
          id: 'NODETEMPLATE',
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
          title: $i18n.t('nav.templateFile'),
          id: 'TEMPLATE_FILE',
          icon: 'bcs-icon-templete',
          route: 'templatefile',
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
            {
              title: 'Tencent Cloud',
              id: 'TENCENTPUBLICCLOUD',
              route: 'tencentPublicCloud',
            },
            {
              title: 'Google Cloud',
              id: 'GOOGLECLOUD',
              route: 'googleCloud',
            },
            {
              title: 'Azure Cloud',
              id: 'AZURECLOUD',
              route: 'azureCloud',
            },
            {
              title: 'Huawei Cloud',
              id: 'HUAWEICLOUD',
              route: 'huaweiCloud',
            },
            {
              title: 'Aws Cloud',
              id: 'AMAZONCLOUD',
              route: 'amazonCloud',
            },
          ],
        },
        {
          title: $i18n.t('nav.projectInfo'),
          id: 'PROJECT',
          icon: 'bcs-icon bcs-icon-apps',
          route: 'projectInfo',
        },
        {
          title: $i18n.t('nav.projectQuotas'),
          id: 'PROJECTQUOTAS',
          icon: 'bcs-icon bcs-icon-pie-chart-fill',
          route: 'projectQuotas',
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
        // {
        //   title: 'LoadBalancers',
        //   id: 'LOADBALANCERS',
        //   icon: 'bcs-icon bcs-icon-loadbalance',
        //   route: 'loadBalance',
        // },
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
        {
          title: $i18n.t('serviceMesh.title'),
          id: 'SERVICEMESH',
          icon: 'bcs-icon-log-collection',
          route: 'serviceMesh',
        },
      ],
    },
    {
      title: $i18n.t('nav.platformManage'),
      id: 'PLATFORMMANAGE',
      children: [
        {
          title: $i18n.t('nav.platformProject'),
          id: 'PLATFORMPROJECT',
          icon: 'bcs-icon bcs-icon-apps',
          route: 'platformProjectList',
        },
      ],
    },
  ]);
  const parseTreeMenuToMap = (menus: IMenu[], initialValue = {}, parent?: IMenu) => (
    menus.reduce<Record<string, IMenu>>((pre, item) => {
      item.parent = parent;// 父节点
      if (!item.id) {
        console.warn('menu id is null', item);
      } else if (pre[item.id]) {
        console.warn('menu id is repeat', item);
      } else {
        pre[item.id] = item;
      }
      if (item.children?.length) {
        pre = parseTreeMenuToMap(item.children, pre, item);
      }
      return pre;
    }, initialValue)
  );
  // 因为ref里面不能存有递归关闭的数据，这里缓存一份含有parent指向的map数据
  const menusDataMap = parseTreeMenuToMap(menusData.value);

  const { flagsMap, getFeatureFlags } = useAppData();
  // 过滤未开启feature_flag的菜单
  const filterMenu = (featureFlags: Record<string, boolean>, data: IMenu[]) => data.reduce<IMenu[]>((pre, menu) => {
    if (has(featureFlags, menu.id) && !featureFlags[menu.id]) return pre; // 未开启菜单项
    pre.push(menu);
    if (menu.children?.length) {
      menu.children = filterMenu(featureFlags, menu.children);
    }
    return pre;
  }, []);
  const menus = computed<IMenu[]>(() => filterMenu(flagsMap.value, menusData.value));
  // 扁平化子菜单
  const flatLeafMenus = (menus: IMenu[], root?: IMenu) => {
    const data: IMenu[] = [];
    for (const item of menus) {
      const rootMenu = root ?? item;
      if (item.children?.length) {
        data.push(...flatLeafMenus(item.children, rootMenu));
      } else {
        data.push({
          root: rootMenu,
          ...item,
        });
      }
    }
    return data;
  };
  // 所有叶子菜单项
  const leafMenus = computed(() => flatLeafMenus(menus.value));
  const allLeafMenus = computed(() => flatLeafMenus(menusData.value));
  // 所有路由父节点只是用于分组（指向子路由），真正的菜单项是子节点
  const getCurrentMenuByRouteName = (name: string) => allLeafMenus.value
    .find(item => item.route === name);
  // 根据ID获取当前一级导航菜单
  const getNavByID = (id: string) => {
    if (!id || !menusDataMap[id]) return;

    let menu = menusDataMap[id];
    while (menu.parent) {
      menu = menu.parent;
    }
    return menu;
  };
  // 校验菜单是否开启
  const validateMenuID = (menu: IMenu): boolean => {
    if (has(flagsMap.value, menu?.id || '') && !flagsMap.value[menu?.id || '']) {
      return false;
    }
    // 如果父菜单没有开启，则子菜单也不能开启
    if (menu?.parent) {
      return validateMenuID(menu.parent);
    }
    return true;
  };
  const validateRouteEnable = async (route: Route) => {
    if (!route.params.projectCode) return true; // 处理根路由
    // 首次加载时获取feature_flag数据
    if (!flagsMap.value || !Object.keys(flagsMap.value)?.length) {
      await getFeatureFlags({ projectCode: route.params.projectCode });
    }
    // 路由配置上带有menuId（父菜单ID）或 ID（当前菜单ID）, 先判断配置的ID是否开启了feature_flag
    if (route.meta?.id && has(flagsMap.value, route.meta?.id) && !flagsMap.value[route.meta.id]) {
      return false;
    }
    if (route.meta?.menuId && has(flagsMap.value, route.meta.menuId) && !flagsMap.value[route.meta.menuId]) {
      return false;
    }
    // 直接返回的菜单项不包含parent信息, 需要去menusDataMap找含有parent信息的菜单项
    const menuID = getCurrentMenuByRouteName(route.name || '')?.id || '';
    return menusDataMap[menuID] ? validateMenuID(menusDataMap[menuID]) : true;
  };
  const disabledMenuIDs = computed<string[]>(() => []);

  // 获取路由对应的导航
  const getNavByRoute = (route: Route) => {
    let menu = getCurrentMenuByRouteName(route.name || '')?.root || getNavByID(route.meta?.menuId || '');
    if (!menu && route.matched.some(item => item.name === 'dashboardIndex')) {
      menu = menusDataMap.CLUSTERRESOURCE;
    }
    return menu;
  };

  return {
    leafMenus,
    menusData,
    menus,
    disabledMenuIDs,
    flatLeafMenus,
    validateRouteEnable,
    getNavByRoute,
  };
}
