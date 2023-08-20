import http from "../request"
import { ICommonQuery } from '../../types/index'
import { ITemplatePackageEditParams } from '../../types/template'
import { IConfigEditParams } from '../../types/config'

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
  return http.delete(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/${template_set_id}`, { params: { force: true } }).then(res => res.data);
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

/**
 * 获取模板空间下的全部配置项模板列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @returns
 */
export const getTemplatesBySpaceId = (biz_id: string, template_space_id: number, params: ICommonQuery) => {
  return http.get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates`, { params }).then(res => res.data)
}

/**
 * 获取模板空间下未指定套餐的模板列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @returns
 */
export const getTemplatesWithNoSpecifiedPackage = (biz_id: string, template_space_id: number, params: ICommonQuery) => {
  return http.get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/list_not_bound`, { params }).then(res => res.data)
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
 * 获取多个模板套餐被未命名版本引用的服务引用列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @returns
 */
export const getUnNamedVersionAppsBoundByPackages = (biz_id: string, template_space_id: number, template_sets: number[], params: ICommonQuery) => {
  return http.get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/bound_unnamed_app_details`, {
    params: { template_set_ids: template_sets.join(','), ...params }
  })
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

/**
 * 获取模板套餐下的配置项模板列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @returns
 */
export const getTemplatesByPackageId = (biz_id: string, template_space_id: number, template_set_id: number, params: ICommonQuery) => {
  return http.get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/${template_set_id}/templates`, { params }).then(res => res.data)
}

/**
 * 创建模板
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param params 配置项参数
 * @returns
 */
export const createTemplate = (biz_id: string, template_space_id: number, params: IConfigEditParams) => {
  return http.post(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates`, params)
}

/**
 * 批量删除模板
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_ids 模板ID列表
 * @returns
 */
export const deleteTemplate = (biz_id: string, template_space_id: number, template_ids: number[]) => {
  return http.delete(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates`, { params: { template_ids: template_ids.join(',') } })
}

/**
 * 根据模板id列表查询对应模板详情
 * @param biz_id 业务ID
 * @param ids 模板ID列表
 * @returns
 */
export const getTemplatesDetailByIds = (biz_id: string, ids: number[]) => {
  return http.post(`/config/biz/${biz_id}/templates/list_by_ids`, { ids }).then(res => res.data)
}


/**
 * 添加模版到模版套餐(多个模板添加到多个套餐)
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param template_set_ids 模板套餐列表
 * @returns
 */
export const addTemplateToPackage = (biz_id: string, template_space_id: number, template_ids: number[], template_set_ids: number[]) => {
  return http.post(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/add_to_template_sets`, { template_ids, template_set_ids })
}

/**
 * 将模版移出套餐(多个模板移出多个套餐)
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param template_set_ids 模板套餐列表
 * @returns
 */
export const moveOutTemplateFromPackage = (biz_id: string, template_space_id: number, template_ids: number[], template_set_ids: number[]) => {
  return http.post(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/delete_from_template_sets`, { template_ids, template_set_ids })
}

/**
 * 创建模板版本
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @returns
 */
export const createTemplateVersion = (biz_id: string, template_space_id: number, template_id: number) => {
  return http.post(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/${template_id}/template_revisions`)
}

/**
 * 获取模板版本列表
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param params 查询参数
 * @returns
 */
export const getTemplateVersionList = (biz_id: string, template_space_id: number, template_id: number, params: ICommonQuery) => {
  return http.get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/${template_id}/template_revisions`, { params }).then(res => res.data)
}

/**
 * 删除模板版本
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param template_revision_id 版本ID
 * @returns
 */
export const deleteTemplateVersion = (biz_id: string, template_space_id: number, template_id: number, template_revision_id: number) => {
  return http.delete(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/${template_id}/template_revisions/${template_revision_id}`)
}
