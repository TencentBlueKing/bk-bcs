import { Self_Request } from "../request"
import { IPageFilter, IRequestFilter, IServingEditParams } from '../types'

/**
 * 获取某个版本下配置列表
 * @param biz_id 业务ID
 * @param release_id 版本ID
 * @param filter 查询过滤条件
 * @param page 分页设置
 * @returns 
 */
 export const getVersionConfigList = (biz_id: number, release_id: number, filter: IRequestFilter = {}, page: IPageFilter) => {
  return Self_Request(`/config/list/release/config_item/release_id/${release_id}/biz_id/${biz_id}`, { biz_id, filter, page: { ...page, count: false } });
}

/**
 * 获取应用正在编辑版本下配置列表
 * @param biz_id 业务ID
 * @param app_id 应用ID
 * @param filter 查询过滤条件
 * @param page 分页设置
 * @returns 
 */
 export const getServingConfigList = (biz_id: number, app_id: number, filter: IRequestFilter = {} ,page: IPageFilter) => {
  return Self_Request(`/config/list/config_item/config_item/app_id/${app_id}/biz_id/${biz_id}`, { biz_id, app_id, filter, page }, 'POST');
}

/**
 * 新增配置
 * @param biz_id 业务ID
 * @param params 配置参数内容
 * @param page 分页设置
 * @returns 
 */
 export const createServingConfigItem = (params: IServingEditParams) => {
  const { biz_id, app_id } = params
  return Self_Request(`/config/create/config_item/config_item/app_id/${app_id}/biz_id/${biz_id}`, params, 'POST');
}

/**
 * 更新配置
 * @param biz_id 业务ID
 * @param params 配置参数内容
 * @param page 分页设置
 * @returns 
 */
 export const updateServingConfigItem = (params: IServingEditParams) => {
  const { id, biz_id, app_id } = params
  return Self_Request(`/config/update/config_item/config_item/config_item_id/${id}/app_id/${app_id}/biz_id/${biz_id}`, params, 'PUT');
}

/**
 * 删除配置
 * @param id 配置ID
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns 
 */
 export const deleteServingConfigItem = (id: number, bizId: number, appId: number) => {
  return Self_Request(`/config/delete/config_item/config_item/config_item_id/${id}/app_id/${appId}/biz_id/${bizId}`, {}, 'DELETE');
}