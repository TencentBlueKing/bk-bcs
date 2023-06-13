import http from "../request"
import { ISpaceDetail } from '../../types/index'
import { IAppListQuery } from '../../types/app'
import { IPermissionQuery } from '../../types/index'

/**
 * 获取空间、项目列表
 * @param biz_id 业务ID
 * @param params 查询过滤条件
 * @returns 
 */

export const getSpaceList = () => {
  return http.get('auth/user/spaces').then(resp => {
    const permissioned: ISpaceDetail[] = []
    const noPermissions: ISpaceDetail[] = []
    resp.data.items.forEach((item: ISpaceDetail) => {
      const { space_id } = item
      // @ts-ignore
      item.permission = resp.web_annotations.perms[space_id].find_business_resource
      if (item.permission) {
        permissioned.push(item)
      } else {
        noPermissions.push(item)
      }
    })
    resp.data.items = [...permissioned, ...noPermissions]
    return resp.data
  });
}

/**
 * 获取所有业务下的服务列表
 * @returns 
 */
export const getAllApp = (bizId: string) => {
  return http.get(`config/biz/${bizId}/apps`).then(resp => resp.data);
}

/**
 * 获取服务列表
 * @param biz_id 业务ID
 * @param params 查询过滤条件
 * @returns 
 */
export const getAppList = (biz_id: string, params: IAppListQuery = {}) => {
  return http.get(`config/list/app/app/biz_id/${biz_id}`, { params }).then(resp => resp.data);
}

/**
 * 获取服务下配置文件数量、更新时间等信息
 */
export const getAppsConfigData = (biz_id: string, app_id: number[]) => {
  return http.post(`/config/config_item_count/biz_id/${biz_id}`, { biz_id, app_id }).then(resp => resp.data);
}

/**
 * 获取服务详情
 * @param biz_id 业务ID
 * @param app_id 服务ID
 * @returns 
 */
export const getAppDetail = (biz_id: string, app_id: number) => {
  return http.get(`config/get/app/app/app_id/${app_id}/biz_id/${biz_id}`).then(resp => resp.data);
}

/**
 * 删除服务
 * @param id 服务ID
 * @param biz_id 业务ID
 * @returns 
 */
export const deleteApp = (id: number, biz_id: number) => {
  return http.delete(`config/delete/app/app/app_id/${id}/biz_id/${biz_id}`);
}

/**
 * 创建服务
 * @param biz_id 业务ID
 * @param params 
 * @returns 
 */
export const createApp = (biz_id: string, params: any) => {
  return http.post(`config/create/app/app/biz_id/${biz_id}`, { biz_id, ...params }).then(resp => resp.data);
}

/**
 * 更新服务
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
