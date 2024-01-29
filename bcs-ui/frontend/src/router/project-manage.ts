// 项目管理
const OperateAudit = () => import(/* webpackChunkName: 'project' */'@/views/project-manage/operate-audit/operate-audit.vue');
const EventQuery = () => import(/* webpackChunkName: 'project' */'@/views/project-manage/event-query/new-event-query.vue');
const ProjectInfo = () => import(/* webpackChunkName: 'project' */'@/views/project-manage/project/project-info.vue');

// 云凭证
const tencentCloud = () => import(/* webpackChunkName: 'project' */'@/views/project-manage/cloudtoken/tencentCloud.vue');
const tencentPublicCloud = () => import(/* webpackChunkName: 'project' */'@/views/project-manage/cloudtoken/tencentPublicCloud.vue');
const googleCloud = () => import(/* webpackChunkName: 'project' */'@/views/project-manage/cloudtoken/googleCloud.vue');

export default [
  {
    path: 'operate-audit',
    component: OperateAudit,
    name: 'operateAudit',
  },
  {
    path: 'event-query',
    name: 'eventQuery',
    component: EventQuery,
  },
  {
    path: 'project-info',
    name: 'projectInfo',
    component: ProjectInfo,
  },
  {
    path: 'tencent-cloud',
    name: 'tencentCloud',
    component: tencentCloud,
    meta: {
      title: 'Tencent Cloud',
      hideBack: true,
    },
  },
  {
    path: 'tencent-public-cloud',
    name: 'tencentPublicCloud',
    component: tencentPublicCloud,
    meta: {
      title: 'Tencent Cloud',
      hideBack: true,
    },
  },
  {
    path: 'google-cloud',
    name: 'googleCloud',
    component: googleCloud,
    meta: {
      title: 'Google Cloud',
      hideBack: true,
    },
  },
];
