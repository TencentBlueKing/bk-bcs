import http from "../request"
import { IScriptEditingForm, IScriptCiteQuery } from '../../types/script'

/**
 * 创建脚本
 * @param biz_id 空间ID
 * @returns 
 */
export const createScript = (biz_id: string, params: IScriptEditingForm) => {
  return http.post(`/create/hook/hook/biz_id/${biz_id}`, params).then(res => res.data)
}

/**
 * 获取脚本列表
 * @param biz_id 空间ID
 * @param params 查询参数
 * @returns 
 */

interface IScriptListQuery {
  start: number;
  limit?: number;
  tag?: string;
  all?: boolean;
  name?: string;
}

export const getScriptList = (biz_id: string, params: IScriptListQuery) => {
  return http.get(`/config/list/hook/hook/biz_id/${biz_id}`, { params }).then(res => res.data);
}

/**
 * 删除脚本
 * @param biz_id 空间ID
 * @param params 查询参数
 * @returns 
 */
export const deleteScript = (biz_id: string, params: { start: number; limit: number; name?: string; }) => {
  return http.delete(`/config/delete/hook/hook/biz_id/${biz_id}`, { params }).then(res => res.data);
}

/**
 * 获取脚本标签列表
 * @param biz_id 空间ID
 * @param params 查询参数
 * @returns 
 */
export const getScriptTagList = (biz_id: string) => {
  return http.get(`/config/list/hook/hook/biz_id/${biz_id}`).then(res => res.data);
}

/**
 * 获取脚本版本列表
 * @param biz_id 空间ID
 * @param hook_id 脚本ID
 * @param params 查询参数
 * @returns 
 */
export const getScriptVersionList = (biz_id: string, hook_id: number, params: { start: number; limit?: number; searchKey?: string; }) => {
  return http.get(`/config/biz/${biz_id}/hooks/${hook_id}/hook_releases/list`, { params }).then(res => res.data);
}

/**
 * 获取脚本被引用列表
 * @param biz_id 空间ID
 * @param hook_id 脚本ID
 * @param params 查询参数
 * @returns 
 */
export const getScriptCiteList = (biz_id: string, hook_id: number, params: IScriptCiteQuery) => {
  return http.get(`/config/biz/${biz_id}/hooks/${hook_id}/hook_releases/list`, { params }).then(res => res.data);
}
