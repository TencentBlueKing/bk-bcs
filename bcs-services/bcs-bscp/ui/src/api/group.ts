import http from "../request"
import { IGroupCategoriesQuery } from '../../types/group'

/**
 * 获取分组标签列表
 * @param app_id 应用ID
 * @param params 查询参数
 * @returns 
 */
export const getGroupCategories = (app_id: number, params: IGroupCategoriesQuery) => {
  return http.get(`/config/apps/${app_id}/group_categories`, { params }).then(resp => resp.data);
}