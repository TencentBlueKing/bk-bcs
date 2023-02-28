import http from "../request"
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
  return http.get(`/config/list/release/config_item/release_id/${release_id}/biz_id/${biz_id}`, { params: { filter, page: { ...page, count: false } } }).then(resp => resp.data);
}

/**
 * 获取应用正在编辑版本下配置列表
 * @param biz_id 业务ID
 * @param app_id 应用ID
 * @param filter 查询过滤条件
 * @param page 分页设置
 * @returns 
 */
 export const getServingConfigList = (biz_id: string, app_id: number, filter: IRequestFilter = {} ,page: IPageFilter) => {
  return http.post(`/config/list/config_item/config_item/app_id/${app_id}/biz_id/${biz_id}`, { biz_id, app_id, filter, page }).then(resp => resp.data);
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
  return http.post(`/config/create/config_item/config_item/app_id/${app_id}/biz_id/${biz_id}`, params);
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
  return http.put(`/config/update/config_item/config_item/config_item_id/${id}/app_id/${app_id}/biz_id/${biz_id}`, params);
}

/**
 * 删除配置
 * @param id 配置ID
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns 
 */
 export const deleteServingConfigItem = (id: number, bizId: number, appId: number) => {
  return http.delete(`/config/delete/config_item/config_item/config_item_id/${id}/app_id/${appId}/biz_id/${bizId}`, {});
}

/**npm
 * 获取配置项详情
 * @param id 配置ID
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns 
 */
export const getConfigItemDetail = (id: number, bizId: string, appId: number) => {
  return http.get(`/config/get/config_item/config_item/config_item_id/${id}/app_id/${appId}/biz_id/${bizId}`).then(resp => resp.data);
}

/**
 * 上传配置项内容
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param data 配置内容
 * @param SHA256Str 文件内容的SHA256值
 * @returns
 */
export const updateConfigContent = (bizId: string, appId: number, data: string|File, SHA256Str: string) => {
  return http.put(`/api/create/content/upload/biz_id/${bizId}/app_id/${appId}`, data, {
    headers: {
      'X-Bkapi-File-Content-Overwrite': 'false', 
      'Content-Type': 'text/plain',
      'X-Bkapi-File-Content-Id': SHA256Str
    }
  })
}
/**
 * 获取配置项内容
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param SHA256Str 文件内容的SHA256值
 * @returns 
 */
export const getConfigContent = (bizId: string, appId: number, SHA256Str: string) => {
  return http.get(`/api/get/content/download/biz_id/${bizId}/app_id/${appId}`, {
    headers: {
      'X-Bkapi-File-Content-Id': SHA256Str
    }
  })
}

/**
 * 创建配置版本
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param name 版本名称
 * @param memo 版本描述
 * @returns 
 */
export const createVersion = (bizId: string, appId: number, name: string, memo: string) => {
  return http.post(`/config/create/release/release/app_id/${appId}/biz_id/${bizId}`, { name, memo })
}

/**
 * 获取配置的版本列表
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns 
 */
export const getConfigVersionList = (bizId: string, appId: number) => {
  // @todo 接口筛选条件需要修改，目前不能拉全量数据，最大限制是200条
  return http.post(`config/list/release/release/app_id/${appId}/biz_id/${bizId}`, { filter: { op: 'and', rules: [] }, page: { start: 0, limit: 200 } })
}

/**
 * 发布版本
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param name 版本名称
 * @param memo 版本描述
 * @returns 
 */
export const publishVersion = (bizId: string, appId: number, memo: string) => {
  return http.post(`/config/create/release/release/app_id/${appId}/biz_id/${bizId}`, { memo })
}