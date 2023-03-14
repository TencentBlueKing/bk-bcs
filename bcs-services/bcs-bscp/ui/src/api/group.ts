import http from "../request"
import { IGroupCategoriesQuery, IGroupEditArg } from '../../types/group'

/**
 * 获取分类列表
 * @param app_id 应用ID
 * @param params 查询参数
 * @returns 
 */
export const getGroupCategories = (app_id: number, params: IGroupCategoriesQuery) => {
  return http.get(`/config/apps/${app_id}/group_categories`, { params }).then(res => res.data);
}

/**
 * 新增分类
 * @param app_id 应用ID
 * @param name 分类名称
 * @returns 
 */
export const createCategory = (app_id: number, name: string) => {
  return http.post(`/config/apps/${app_id}/group_categories`, { name }).then(res => res.data)
}

/**
 * 删除分类
 * @param app_id 应用ID
 * @param group_category_id 分类ID
 * @returns 
 */
export const delCategory = (app_id: number, group_category_id: number) => {
  return http.delete(`/config/apps/${app_id}/groups/${group_category_id}`).then(res => res.data)
}

/**
 * 获取分类下分组列表
 * @param app_id 应用ID
 * @param group_category_id 分类ID
 * @param params 查询参数
 * @returns 
 */
export const getCategoryGroupList = (app_id: number, group_category_id: number, params: IGroupCategoriesQuery) => {
  return http.get(`/config/apps/${app_id}/group_categories/${group_category_id}/groups`, { params }).then(res => res.data)
}

/**
 * 新增分组
 * @param app_id 应用ID
 * @param params 分组编辑参数
 * @returns 
 */
export const createGroup = (app_id: number, params: IGroupEditArg) => {
  return http.post(`/config/apps/${app_id}/groups`, params).then(res => res.data)
}

/**
 * 删除分组
 * @param app_id 应用ID
 * @param group_id 分组ID
 * @returns 
 */
export const deleteGroup = (app_id: number, group_id: number) => {
  return http.delete(`/config/apps/${app_id}/groups/${group_id}`)
}