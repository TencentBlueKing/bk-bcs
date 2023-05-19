import http from "../request"
import { IRuleUpdateParams } from '../../types/credential'

/**
 * 创建新密钥
 * @param biz_id 空间ID
 * @returns 
 */
export const createCredential = (biz_id: string, params: { memo: string }) => {
  return http.post(`/config/biz_id/${biz_id}/credentials`, params).then(res => res.data)
}

/**
 * 获取密钥列表
 * @param biz_id 空间ID
 * @returns 
 */
export const getCredentialList = (biz_id: string, query: { limit: number, start: number, searchKey?: string }) => {
  return http.get(`/config/biz_id/${biz_id}/credentials`, { params: query }).then(res => res.data)
}

/**
 * 删除密钥
 * @param biz_id 空间ID
 * @param id 密钥ID
 * @returns 
 */
export const deleteCredential = (biz_id: string, id: number) => {
  return http.delete(`/config/biz_id/${biz_id}/credential`, { params: { id } }).then(res => res.data)
}

/**
 * 更新密钥
 * @param biz_id 空间ID
 * @returns 
 */
export const updateCredential = (biz_id: string, params: { id: number; memo?: string; enable?: boolean }) => {
  return http.put(`/config/biz_id/${biz_id}/credential`, params).then(res => res.data)
}

/**
 * 获取密钥关联的配置项规则
 * @param biz_id 空间ID
 * @param credential_id 密钥ID
 * @returns 
 */
export const getCredentialScopes = (biz_id: string, credential_id: number) => {
  return http.get(`/config/biz_id/${biz_id}/credential/${credential_id}/scopes`).then(res => res.data)
}

/**
 * 更新密钥关联的配置项规则
 * @param biz_id 空间ID
 * @param credential_id 密钥ID
 * @returns 
 */
export const updateCredentialScopes = (biz_id: string, credential_id: number, params: IRuleUpdateParams) => {
  return http.put(`/config/biz_id/${biz_id}/credential/${credential_id}/scope`, params).then(res => res.data)
}