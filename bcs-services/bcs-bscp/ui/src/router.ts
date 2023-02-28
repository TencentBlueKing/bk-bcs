import { createRouter, createWebHistory } from 'vue-router';
const routes = [
  { path: '/', component: () => import('./views/home.vue') },
  {
    path: '/serving',
    name: 'serving',
    component: () => import('./views/serving/index.vue'),
    children: [
      {
        path: 'mine/',
        name: 'serving-mine',
        component: () => import('./views/serving/serving-mine.vue'),
      },
      {
        path: 'all/',
        name: 'serving-all',
        component: () => import('./views/serving/serving-all.vue'),
      },
    ]
  },
  {
    path: '/serving/:spaceId/app/:appId',
    name: 'serving-detail',
    component: () => import('./views/serving/detail/index.vue'),
    children: [
      {
        path: 'config/',
        name: 'serving-config',
        component: () => import('./views/serving/detail/config/index.vue')
      },
      {
        path: 'group/',
        name: 'serving-group',
        component: () => import('./views/serving/detail/group/index.vue')
      },
      {
        path: 'client/',
        name: 'serving-client',
        component: () => import('./views/serving/detail/client/index.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes, // `routes: routes` 的缩写
});

export default router;
