import http from "../request"
import { IScriptEditingForm, IScriptCiteQuery, IScriptListQuery } from '../../types/script'

/**
 * 创建脚本
 * @param biz_id 空间ID
 * @returns
 */
export const createScript = (biz_id: string, params: IScriptEditingForm) => {
  return http.post(`/config/biz/${biz_id}/hooks`, params).then(res => res.data)
}

/**
 * 获取脚本列表
 * @param biz_id 空间ID
 * @param params 查询参数
 * @returns
 */

export const getScriptList = (biz_id: string, params: IScriptListQuery) => {
  return http.get(`/config/biz/${biz_id}/hooks`, { params }).then(res => res.data);
}

/**
 * 获取脚本详情
 * @param biz_id 空间ID
 * @param id 脚本ID
 * @returns
 */
export const getScriptDetail = (biz_id: string, id: number) => {
  return http.get(`/config/biz/${biz_id}/hooks/${id}`).then(res => res.data);
}

/**
 * 获取脚本某个版本下详情
 * @param biz_id 空间ID
 * @param id 脚本ID
 * @param release_id 版本ID
 * @returns
 */
export const getScriptVersionDetail = (biz_id: string, id: number, release_id: number) => {
  return http.get(`/config/biz/${biz_id}/hooks/${id}/hook_releases/${release_id}`).then(res => res.data);
}

/**
 * 删除脚本
 * @param biz_id 空间ID
 * @param id 脚本ID
 * @returns
 */
export const deleteScript = (biz_id: string, id: number) => {
  return http.delete(`/config/biz/${biz_id}/hooks/${id}`).then(res => res.data);
}

/**
 * 获取脚本标签列表
 * @param biz_id 空间ID
 * @param params 查询参数
 * @returns
 */
export const getScriptTagList = (biz_id: string) => {
  return http.get(`/config/biz/${biz_id}/hook_tags`).then(res => res.data);
}

/**
 * 获取脚本版本列表
 * @param biz_id 空间ID
 * @param hook_id 脚本ID
 * @param params 查询参数
 * @returns
 */
export const getScriptVersionList = (biz_id: string, hook_id: number, params: { start: number; limit?: number; searchKey?: string; all?: boolean }) => {
  return http.get(`/config/biz/${biz_id}/hooks/${hook_id}/hook_releases`, { params }).then(res => res.data);
}

/**
 * 创建脚本版本
 * @param biz_id 空间ID
 * @param hook_id 脚本ID
 * @param params 查询参数
 * @returns
 */
export const createScriptVersion = (biz_id: string, hook_id: number, params: { name: string; memo: string; content: string; }) => {
  return http.post(`/config/biz/${biz_id}/hooks/${hook_id}/hook_releases`, params).then(res => res.data);
}

/**
 * 更新脚本版本
 * @param biz_id 空间ID
 * @param hook_id 脚本ID
 * @param params 查询参数
 * @returns
 */
export const updateScriptVersion = (biz_id: string, hook_id: number, release_id: number, params: { name: string; memo: string; content: string; }) => {
  return http.put(`/config/biz/${biz_id}/hooks/${hook_id}/hook_releases/${release_id}`, params).then(res => res.data);
}

/**
 * 删除脚本版本
 * @param biz_id 空间ID
 * @param hook_id 脚本ID
 * @param release_id 版本ID
 * @returns
 */
export const deleteScriptVersion = (biz_id: string, hook_id: number, release_id: number) => {
  return http.delete(`/config/biz/${biz_id}/hooks/${hook_id}/hook_releases/${release_id}`).then(res => res.data);
}

/**
 * 上线版本
 * @param biz_id 空间ID
 * @param hook_id 脚本ID
 * @param release_id 版本ID
 * @returns
 */
export const publishVersion = (biz_id: string, hook_id: number, release_id: number) => {
  return http.put(`/config/biz/${biz_id}/hooks/${hook_id}/hook_releases/${release_id}/publish`).then(res => res.data);
}

/**
 * 获取脚本被引用列表
 * @param biz_id 空间ID
 * @param hook_id 脚本ID
 * @param params 查询参数
 * @returns
 */
export const getScriptCiteList = (biz_id: string, hook_id: number, params: IScriptCiteQuery) => {
  return http.get(`/config/biz/${biz_id}/hooks/${hook_id}/references`, { params }).then(res => res.data);
}

// 获取服务初始化脚本
export const getConfigInitScript = (biz_id: string, app_id: number) => {
  return http.get(`/config/biz/${biz_id}/apps/${app_id}/config_hooks`)
}

// 配置服务初始化脚本
export const createConfigInitScript = (biz_id: string, app_id: number) => {
  return http.post(`/config/biz/${biz_id}/apps/${app_id}/config_hooks`)
}

// 更新服务初始化脚本
export const updateConfigInitScript = (biz_id: string, app_id: number) => {
  return http.post(`/config/biz/${biz_id}/apps/${app_id}/config_hooks`)
}
