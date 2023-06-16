import http from "../request"
import { IAppEditParams } from '../../types/app'
import { IConfigListQueryParams, IConfigVersionQueryParams } from '../../types/config'

/**
 * 获取配置项列表，通过params中的release_id区分是否拿某个版本下的配置项列表
 * @param biz_id 空间ID
 * @param app_id 应用ID
 * @param params 查询参数
 * @returns 
 */
export const getConfigList = (biz_id: string, app_id: number, params: IConfigListQueryParams = {}) => {
  return http.get(`/config/biz/${biz_id}/apps/${app_id}/config_items`, { params }).then(res => res.data)
}

/**
 * 新增配置
 * @param biz_id 业务ID
 * @param params 配置参数内容
 * @param page 分页设置
 * @returns 
 */
 export const createServiceConfigItem = (params: IAppEditParams) => {
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
 export const updateServiceConfigItem = (params: IAppEditParams) => {
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
 export const deleteServiceConfigItem = (id: number, bizId: number, appId: number) => {
  return http.delete(`/config/delete/config_item/config_item/config_item_id/${id}/app_id/${appId}/biz_id/${bizId}`, {});
}

/**
 * 获取配置项详情
 * @param biz_id 空间ID
 * @param id 配置ID
 * @param appId 应用ID
 * @returns 
 */
export const getConfigItemDetail = (biz_id: string, id: number, appId: number, params: { release_id?: number } = {}) => {
  return http.get(`/config/biz/${biz_id}/apps/${appId}/config_items/${id}`, { params }).then(resp => resp.data);
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
  console.log(SHA256Str)
  return http.get<string, string>(`/api/get/content/download/biz_id/${bizId}/app_id/${appId}`, {
    headers: {
      'X-Bkapi-File-Content-Id': SHA256Str
    }
  }).then(res => res)
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
 * 获取版本列表
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param params 查询参数
 * @returns 
 */
export const getConfigVersionList = (bizId: string, appId: number, params: IConfigVersionQueryParams) => {
  return http.get(`config/biz/${bizId}/apps/${appId}/releases`, { params })
}

/**
 * 发布版本
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param name 版本名称
 * @param data 参数
 * @returns 
 */
export const publishVersion = (bizId: string, appId: number, releaseId: number, data: {
  groups: Array<number>;
  all: boolean;
  memo: string;
}) => {
  return http.post(`/config/update/strategy/publish/publish/release_id/${releaseId}/app_id/${appId}/biz_id/${bizId}`, data)
}

/**
 * 获取服务下初始化脚本引用配置
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns 
 */
export const getConfigInitScript = (bizId: string, appId: number) => {
  return http.get(`/config/biz/${bizId}/apps/${appId}/config_hooks`)
}

/**
 * 更新服务下初始化脚本引用配置
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param params 配置数据
 * @returns 
 */
export const updateConfigInitScript = (bizId: string, appId: number, params: { pre_hook_id: number; post_hook_id: number; }) => {
  return http.put(`/config/biz/${bizId}/apps/${appId}/config_hooks`, params)
}
