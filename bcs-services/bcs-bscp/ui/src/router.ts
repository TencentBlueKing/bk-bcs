import { createRouter, createWebHistory } from 'vue-router';
import useGlobalStore from './store/global';
import { ISpaceDetail } from '../types/index';

const routes = [
  {
    path: '/',
    name: 'home',
    redirect: () => {
      // 访问首页，默认跳到服务管理列表页
      // 优先取localstorage里存的上次访问的空间id
      // 不存在时取空间列表中第一个有权限的空间
      // 仍不存在时取空间列表中第一个空间
      let spaceId = localStorage.getItem('lastAccessedSpace');
      if (!spaceId) {
        const { spaceList } = useGlobalStore();
        const firstHasPermSpace = spaceList.find((item: ISpaceDetail) => item.permission);
        spaceId = firstHasPermSpace ? firstHasPermSpace.space_id : spaceList[0]?.space_id;
      }
      return { name: 'service-all', params: { spaceId } };
    },
  },
  {
    path: '/space/:spaceId',
    name: 'space',
    component: () => import('./views/space/index.vue'),
    children: [
      {
        path: 'service',
        children: [
          {
            path: 'mine',
            name: 'service-mine',
            meta: {
              navModule: 'service',
            },
            component: () => import('./views/space/service/list/index.vue'),
          },
          {
            path: 'all',
            name: 'service-all',
            meta: {
              navModule: 'service',
            },
            component: () => import('./views/space/service/list/index.vue'),
          },
          {
            path: ':appId(\\d+)',
            component: () => import('./views/space/service/detail/index.vue'),
            children: [
              {
                path: 'config/:versionId?',
                name: 'service-config',
                meta: {
                  navModule: 'service',
                },
                component: () => import('./views/space/service/detail/config/index.vue'),
              },
              {
                path: 'script/:versionId?',
                name: 'init-script',
                meta: {
                  navModule: 'service',
                },
                component: () => import('./views/space/service/detail/init-script/index.vue'),
              },
            ],
          },
        ],
      },
      {
        path: 'groups',
        name: 'groups-management',
        meta: {
          navModule: 'groups',
        },
        component: () => import('./views/space/groups/index.vue'),
      },
      {
        path: 'variables',
        name: 'variables-management',
        meta: {
          navModule: 'variables',
        },
        component: () => import('./views/space/variables/index.vue'),
      },
      {
        path: 'templates',
        meta: {
          navModule: 'templates',
        },
        children: [
          {
            path: 'list/:templateSpaceId?/:packageId?',
            name: 'templates-list',
            meta: {
              navModule: 'templates',
            },
            component: () => import('./views/space/templates/list/index.vue'),
          },
          {
            path: ':templateSpaceId/:packageId/version_manage/:templateId',
            name: 'template-version-manage',
            meta: {
              navModule: 'templates',
            },
            component: () => import('./views/space/templates/version-manage/index.vue'),
          },
        ],
      },
      {
        path: 'scripts',
        name: 'scripts-management',
        meta: {
          navModule: 'scripts',
        },
        component: () => import('./views/space/scripts/index.vue'),
        children: [
          {
            path: 'list',
            name: 'script-list',
            meta: {
              navModule: 'scripts',
            },
            component: () => import('./views/space/scripts/list/script-list.vue'),
          },
          {
            path: 'version_manage/:scriptId',
            name: 'script-version-manage',
            meta: {
              navModule: 'scripts',
            },
            component: () => import('./views/space/scripts/version-manage/index.vue'),
          },
        ],
      },
      {
        path: 'client_statistics/:appId?',
        name: 'client-statistics',
        meta: {
          navModule: 'client-statistics',
        },
        component: () => import('./views/space/client/statistics/index.vue'),
      },
      {
        path: 'client_search/:appId?',
        name: 'client-search',
        meta: {
          navModule: 'client-search',
        },
        component: () => import('./views/space/client/search/index.vue'),
      },
      {
        path: 'client_credentials',
        name: 'credentials-management',
        meta: {
          navModule: 'credentials',
        },
        component: () => import('./views/space/credentials/index.vue'),
      },
      {
        path: 'configuration_example/:appId?',
        name: 'configuration-example',
        meta: {
          navModule: 'example',
        },
        component: () => import('./views/space/client/example/index.vue'),
      },
      {
        path: 'records',
        children: [
          {
            path: 'all',
            name: 'records-all',
            component: () => import('./views/space/records/index.vue'),
            meta: {
              navModule: 'records',
            },
          },
          {
            path: ':appId(\\d+)',
            name: 'records-app',
            component: () => import('./views/space/records/index.vue'),
            meta: {
              navModule: 'records',
            },
          },
        ],
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('./views/404.vue'),
  },
];

const router = createRouter({
  history: createWebHistory((window as any).SITE_URL),
  routes,
});

// 路由切换时，取消无权限页面
router.afterEach(() => {
  const globalStore = useGlobalStore();
  globalStore.$patch((state) => {
    state.showPermApplyPage = false;
  });
});

export default router;
