import { CC_Request, Self_Request } from "../request"

export const getBizList = () => {
  return CC_Request('search_business/', { fields: ['bk_biz_id', 'bk_biz_name'] });
}

/**
 * 获取应用列表
 * @param biz_id 业务ID
 * @param params 查询过滤条件
 * @returns 
 */
export type IAppListQuery = {
  operator?: string,
  name?: string,
  start?: number,
  limit?: number 
}

export const getAppList = (biz_id: number, params: IAppListQuery = {}) => {
  return Self_Request(`config/list/app/app/biz_id/${biz_id}`, params);
}

/**
 * 删除应用
 * @param id 应用ID
 * @param biz_id 业务ID
 * @returns 
 */
export const deleteApp = (id: number, biz_id: number) => {
  return Self_Request(`config/delete/app/app/app_id/${id}/biz_id/${biz_id}`, { id, biz_id }, 'DELETE');
}

/**
 * 创建应用
 * @param biz_id 业务ID
 * @param params 
 * @returns 
 */
export const createApp = (biz_id: number, params: any) => {
  return Self_Request(`config/create/app/app/biz_id/${biz_id}`, { biz_id, ...params }, 'POST');
}

/**
 * 更新应用
 * @param params { id, biz_id, name?, memo?, reload_type?, reload_file_path? }
 * @returns 
 */
export const updateApp = (params: any) => {
  const { id, biz_id, data } = params;
  return Self_Request(`config/update/app/app/app_id/${id}/biz_id/${biz_id}`, data, 'PUT');
}