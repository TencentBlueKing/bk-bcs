import { Self_Request } from "../request"
import { IPage, IRequestFilter } from '../constants'

/**
 * 获取某个版本下配置列表
 * @param biz_id 业务ID
 * @param release_id 版本ID
 * @param filter 查询过滤条件
 * @param page 分页设置
 * @returns 
 */
 export const getVersionConfigList = (biz_id: number, release_id: number, filter: IRequestFilter = {}, page: IPage) => {
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
 export const getServingConfigList = (biz_id: number, app_id: number, filter: IRequestFilter = {}, page: IPage) => {
  return Self_Request(`/config/list/config_item/config_item/app_id/${app_id}/biz_id/${biz_id}`, { biz_id, app_id, filter, page: { ...page, count: false } });
}

/**
 * 新增配置
 * @param biz_id 业务ID
 * @param params 配置参数内容
 * @param page 分页设置
 * @returns 
 */
type ICreateServingParams = {
  biz_id: number,
  app_id: number,
  name: string,
  file_type: string,
  path?: string,
  file_mode?: string,
  user?: string,
  user_group?: string,
  privilege?: string
}
 export const createServingConfigItem = (params: ICreateServingParams) => {
  const { biz_id, app_id } = params
  return Self_Request(`/config/create/config_item/config_item/app_id/${app_id}/biz_id/${biz_id}`, params);
}