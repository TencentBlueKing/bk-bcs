import http from "../request"
import { IScriptEditingForm } from '../../types/script'

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
export const getScriptList = (biz_id: string, params: { start: number; limit: number; name?: string; }) => {
  return http.get(`/config/list/biz/${biz_id}/hooks`, { params }).then(res => res.data);
}
