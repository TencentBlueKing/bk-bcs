import { createRouter, createWebHashHistory } from 'vue-router';
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
]

const router = createRouter({
  // 4. 内部提供了 history 模式的实现。为了简单起见，我们在这里使用 hash 模式。
  history: createWebHashHistory(),
  routes, // `routes: routes` 的缩写
});

export default router;
