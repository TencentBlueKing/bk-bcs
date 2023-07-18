import http from "../request"
import { ICommonQuery } from '../../types/index'

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
