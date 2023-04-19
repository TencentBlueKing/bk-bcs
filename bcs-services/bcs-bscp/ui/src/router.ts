import { createRouter, createWebHistory } from 'vue-router';
const routes = [
  { path: '/', name:'home', component: () => import('./views/home.vue') },
  {
    path: '/space/:spaceId/service',
    name: 'service',
    component: () => import('./views/service/list/index.vue'),
    children: [
      {
        path: 'mine/',
        name: 'service-mine',
        component: () => import('./views/service/list/mine.vue'),
      },
      {
        path: 'all/',
        name: 'service-all',
        component: () => import('./views/service/list/all.vue'),
      },
    ]
  },
  {
    path: '/space/:spaceId/service/:appId',
    name: 'service-detail',
    component: () => import('./views/service/detail/index.vue'),
    children: [
      {
        path: 'config/',
        name: 'service-config',
        component: () => import('./views/service/detail/config/index.vue')
      },
      // {
      //   path: 'group/',
      //   name: 'service-group',
      //   component: () => import('./views/service/detail/group/index.vue')
      // },
      // {
      //   path: 'client/',
      //   name: 'service-client',
      //   component: () => import('./views/service/detail/client/index.vue')
      // }
    ]
  },
  {
    path: '/space/:spaceId/groups/',
    name: 'groups-management',
    component: () => import('./views/groups/index.vue')
  },
  {
    path: '/space/:spaceId/scripts/',
    name: 'scripts-management',
    component: () => import('./views/scripts/index.vue')
  },
  {
    path: '/space/:spaceId/keys/',
    name: 'keys-management',
    component: () => import('./views/keys/index.vue')
  }
]

const router = createRouter({
  history: createWebHistory((<any>window).SITE_URL),
  routes, // `routes: routes` 的缩写
});

export default router;
