import http from "../request"
import { ICommonQuery } from "../../types/index"
import { IVariableEditParams } from '../../types/variable'

/**
 * 查询变量列表
 * @param biz_id 业务ID
 * @param params
 * @returns
 */
export const getVariableList = (biz_id: string, params: ICommonQuery) => {
  return http.get(`/config/biz/${biz_id}/template_variables`, { params }).then(res => res.data)
}

/**
 * 创建变量
 * @param biz_id 业务ID
 * @param params 创建参数
 * @returns
 */
export const createVariable = (biz_id: string, params: IVariableEditParams) => {
  return http.post(`/config/biz/${biz_id}/template_variables`, params)
}

/**
 * 编辑变量
 * @param biz_id 业务ID
 * @param template_variable_id 变量ID
 * @param params 编辑参数
 * @returns
 */
export const updateVariable = (biz_id: string, template_variable_id: number, params: { default_val: string; memo: string; }) => {
  return http.put(`/config/biz/${biz_id}/template_variables/${template_variable_id}`, params)
}

/**
 * 删除变量
 * @param biz_id 业务ID
 * @param template_variable_id 变量ID
 * @returns
 */
export const deleteVariable = (biz_id: string, template_variable_id: number) => {
  return http.delete(`/config/biz/${biz_id}/template_variables/${template_variable_id}`)
}
