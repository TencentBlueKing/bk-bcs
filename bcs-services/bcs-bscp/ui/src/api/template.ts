import http from "../request"
import { ICommonQuery } from '../../types/index'
import { ITemplatePackageEditParams } from '../../types/template'

/**
 * 获取模板空间列表
 * @param biz_id 业务ID
 * @param params 查询参数
 * @returns
 */
export const getTemplateSpaceList = (biz_id: string, params: ICommonQuery) => {
  return http.get(`/config/biz/${biz_id}/template_spaces`, { params }).then(res => res.data);
}

/**
 * 创建模板空间
 * @param biz_id 业务ID
 * @param params 创建模板参数
 * @returns
 */
export const createTemplateSpace = (biz_id: string, params: { name: string; memo: string; }) => {
  return http.post(`/config/biz/${biz_id}/template_spaces`, params).then(res => res.data);
}

/**
 * 更新模板空间
 * @param biz_id 业务ID
 * @param id 模板ID
 * @param params 模板参数
 * @returns
 */
export const updateTemplateSpace = (biz_id: string, id: number, params: { memo: string; }) => {
  return http.put(`/config/biz/${biz_id}/template_spaces/${id}`, params).then(res => res.data);
}

/**
 * 删除模板空间
 * @param biz_id 业务ID
 * @param id 模板ID
 * @returns
 */
export const deleteTemplateSpace = (biz_id: string, id: number) => {
  return http.delete(`/config/biz/${biz_id}/template_spaces/${id}`);
}

/**
 * 创建模板套餐
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @param params 创建模板套餐参数
 * @returns
 */
export const createTemplatePackage = (biz_id: string, template_space_id: number, params: ITemplatePackageEditParams) => {
  return http.post(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets`, params).then(res => res.data);
}

/**
 * 编辑模板套餐
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @param template_set_id 模板套餐ID
 * @param params 编辑模板套餐参数
 * @returns
 */
export const updateTemplatePackage = (biz_id: string, template_space_id: number, template_set_id: number, params: ITemplatePackageEditParams) => {
  return http.put(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/${template_set_id}`, params).then(res => res.data);
}

/**
 * 删除模板套餐
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @param template_set_id 模板套餐ID
 * @returns
 */
export const deleteTemplatePackage = (biz_id: string, template_space_id: number, template_set_id: number) => {
  return http.delete(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/${template_set_id}`).then(res => res.data);
}

/**
 * 获取空间下的模板套餐列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @param params 查询参数
 * @returns
 */
export const getTemplatePackageList = (biz_id: string, template_space_id: string, params: ICommonQuery) => {
  return http.get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets`, { params }).then(res => res.data);
}

/* 获取模板套餐的配置项列表
* @param biz_id 业务ID
* @param template_space_id 模板空间ID
* @param params 查询参数
* @returns
*/
export const getPackageConfigList = (biz_id: string, template_space_id: string, params: ICommonQuery) => {
  return http.get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates`, { params }).then(res => res.data);
}

/**
 * 获取当前模板套餐被未命名版本引用的服务引用列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @param template_set_id 模板套餐Id
 * @param params 查询参数
 * @returns
 */
export const getUnNamedVersionAppsBoundByPackage = (biz_id: string, template_space_id: number, template_set_id: number, params: ICommonQuery) => {
  return http.get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/${template_set_id}/bound_unnamed_app_details`, { params }).then(res => res.data);
}

/**
 * 获取当前模板套餐被已生成版本引用的服务引用列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @param template_set_id 模板套餐Id
 * @param params 查询参数
 * @returns
 */
export const getReleasedVersionAppsBoundByPackage = (biz_id: string, template_space_id: number, template_set_id: number, params: ICommonQuery) => {
  return http.get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/${template_set_id}/bound_unnamed_app_details`, { params }).then(res => res.data);
}
