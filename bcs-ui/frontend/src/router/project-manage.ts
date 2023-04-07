// 项目管理
const OperateAudit = () => import(/* webpackChunkName: 'project' */'@/views/project-manage/operate-audit/operate-audit.vue');
const EventQuery = () => import(/* webpackChunkName: 'project' */'@/views/project-manage/event-query/new-event-query.vue');
const tencentCloud = () => import(/* webpackChunkName: 'project' */'@/views/project-manage/cloudtoken/tencentCloud.vue');
const ProjectInfo = () => import(/* webpackChunkName: 'project' */'@/views/project-manage/project/project-info.vue');

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
    path: 'tencent-cloud',
    name: 'tencentCloud',
    component: tencentCloud,
    meta: {
      title: 'Tencent Cloud',
      hideBack: true,
    },
  },
  {
    path: 'project-info',
    name: 'projectInfo',
    component: ProjectInfo,
  },
];
