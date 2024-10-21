import http from '../request';
import { ISpaceDetail, IPermissionQueryResourceItem } from '../../types/index';
import { IAppItem, IAppListQuery } from '../../types/app';

/**
 * 获取空间、项目列表
 * @param biz_id 业务ID
 * @param params 查询过滤条件
 * @returns
 */

export const getSpaceList = () =>
  http.get('auth/user/spaces').then((resp) => {
    const permissioned: ISpaceDetail[] = [];
    const noPermissions: ISpaceDetail[] = [];
    resp.data.items.forEach((item: ISpaceDetail) => {
      const { space_id } = item;
      // @ts-ignore
      item.permission = resp.web_annotations.perms[space_id].find_business_resource;
      if (item.permission) {
        permissioned.push(item);
      } else {
        noPermissions.push(item);
      }
    });
    resp.data.items = [...permissioned, ...noPermissions];
    return resp.data;
  });

/**
 * 获取业务的特性开关配置
 * @param biz 业务ID
 * @returns
 */
export const getSpaceFeatureFlag = (biz: string) =>
  http.get('feature_flags', { params: { biz } }).then((resp) => resp.data);

/**
 * 获取服务列表
 * @param biz_id 业务ID
 * @param params 查询过滤条件
 * @returns
 */
export const getAppList = (biz_id: string, params: IAppListQuery = {}) =>
  http.get(`config/list/app/app/biz_id/${biz_id}`, { params }).then((resp) => {
    resp.data.details.forEach((item: IAppItem) => {
      // @ts-ignore
      item.permissions = resp.web_annotations.perms[item.id] || {};
    });
    return resp.data;
  });

/**
 * 获取服务下配置文件数量、更新时间等信息
 */
export const getAppsConfigData = (biz_id: string, app_id: number[]) =>
  http.post(`/config/config_item_count/biz_id/${biz_id}`, { biz_id, app_id }).then((resp) => resp.data);

/**
 * 获取服务详情
 * @param biz_id 业务ID
 * @param app_id 服务ID
 * @returns
 */
export const getAppDetail = (biz_id: string, app_id: number) =>
  http.get(`config/biz/${biz_id}/apps/${app_id}`).then((resp) => resp.data);

/**
 * 删除服务
 * @param id 服务ID
 * @param biz_id 业务ID
 * @returns
 */
export const deleteApp = (id: number, biz_id: number) =>
  http.delete(`config/delete/app/app/app_id/${id}/biz_id/${biz_id}`);

/**
 * 创建服务
 * @param biz_id 业务ID
 * @param params
 * @returns
 */
export const createApp = (biz_id: string, params: any) =>
  http.post(`config/create/app/app/biz_id/${biz_id}`, { biz_id, ...params }).then((resp) => resp.data);

/**
 * 更新服务
 * @param params { id, biz_id, name?, memo?, reload_type?, reload_file_path? }
 * @returns
 */
export const updateApp = (params: any) => {
  const { id, biz_id, data } = params;
  return http.put(`config/update/app/app/app_id/${id}/biz_id/${biz_id}`, data).then((resp) => resp.data);
};

/**
 * 查询资源权限以及返回权限申请链接
 * @param params IPermissionQueryResourceItem 查询参数
 */
export const permissionCheck = (params: { resources: IPermissionQueryResourceItem[] }) =>
  http.post('/auth/iam/permission/check', params).then((resp) => resp.data);

/**
 * 获取消息通知信息
 * @returns
 */
export const getNotice = () => http.get('/announcements').then((resp) => resp.data);

/**
 * 退出登录
 * @returns
 */
export const loginOut = () =>
  http.get('/logout').then((resp) => {
    window.location.href = `${resp.data.login_url}${encodeURIComponent(window.location.href)}&is_from_logout=1`;
  });

/**
 * 审批人员名单
 * @returns
 */
export const getApproverListApi = () =>
  `${(window as any).USER_MAN_HOST}/api/c/compapi/v2/usermanage/fs_list_users/?app_code=bk-magicbox&page_size=1000&page=1`;
// /api/c/compapi/v2/usermanage/fs_list_users/?app_code=bk-magicbox&page_size=1000&page=1"
