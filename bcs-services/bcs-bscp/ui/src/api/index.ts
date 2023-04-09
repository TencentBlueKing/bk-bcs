import http from "../request"
import { ISpaceItem } from '../../types/index'
import { IAppListQuery } from '../../types/app'
import { IPermissionQuery } from '../../types/index'

/**
 * 获取空间、项目列表
 * @param biz_id 业务ID
 * @param params 查询过滤条件
 * @returns 
 */

export const getBizList = () => {
  return http.get('auth/user/spaces').then(resp => {
    resp.data.items.forEach((item: ISpaceItem) => {
      const { space_id } = item
      // @ts-ignore
      item.permission = resp.web_annotations.perms[space_id].find_business_resource
    })
    return resp.data
  });
}

/**
 * 获取所有业务下的应用列表
 * @returns 
 */
export const getAllApp = () => {
  return http.get('config/apps').then(resp => resp.data);
}

/**
 * 获取应用列表
 * @param biz_id 业务ID
 * @param params 查询过滤条件
 * @returns 
 */
export const getAppList = (biz_id: number, params: IAppListQuery = {}) => {
  return http.get(`config/list/app/app/biz_id/${biz_id}`, { params }).then(resp => resp.data);
}

/**
 * 获取应用下配置文件数量、更新时间等信息
 */
export const getAppsConfigData = (biz_id: number, app_id: number[]) => {
  return http.post(`/config/config_item_count/biz_id/${biz_id}`, { biz_id, app_id }).then(resp => resp.data);
}

/**
 * 
 * @param biz_id 业务ID
 * @param app_id 应用ID
 * @returns 
 */
export const getAppDetail = (biz_id: string, app_id: number) => {
  return http.get(`config/get/app/app/app_id/${app_id}/biz_id/${biz_id}`).then(resp => resp.data);
}

/**
 * 删除应用
 * @param id 应用ID
 * @param biz_id 业务ID
 * @returns 
 */
export const deleteApp = (id: number, biz_id: number) => {
  return http.delete(`config/delete/app/app/app_id/${id}/biz_id/${biz_id}`);
}

/**
 * 创建应用
 * @param biz_id 业务ID
 * @param params 
 * @returns 
 */
export const createApp = (biz_id: number, params: any) => {
  return http.post(`config/create/app/app/biz_id/${biz_id}`, { biz_id, ...params }).then(resp => resp.data);
}

/**
 * 更新应用
 * @param params { id, biz_id, name?, memo?, reload_type?, reload_file_path? }
 * @returns 
 */
export const updateApp = (params: any) => {
  const { id, biz_id, data } = params;
  return http.put(`config/update/app/app/app_id/${id}/biz_id/${biz_id}`, data).then(resp => resp.data);
}

/**
 * 查询资源权限以及返回权限申请链接
 * @param params IPermissionQuery 查询参数
 */
export const permissionCheck = (params: IPermissionQuery) => {
  return http.post(`/auth/iam/permission/check`, params).then(resp => resp.data);
}
