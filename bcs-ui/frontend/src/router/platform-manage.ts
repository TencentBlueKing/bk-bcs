// 平台管理
const PlatformProjectList = () => import(/* webpackChunkName: 'platform' */'@/views/platform-manage/project/project-list.vue');

export default [
  {
    path: 'platform-project-list',
    name: 'platformProjectList',
    component: PlatformProjectList,
    meta: {
      resource: window.i18n.t('nav.platformProject'),
      menuId: 'PLATFORMMANAGE',
    },
  },
];
