import { CC_Request, Self_Request } from "../request"
export type IPage = {
  count?: number,
  start: number,
  limit: number
}

export enum FilterOp {
  AND = 'and',
  OR = 'or'
}

export enum RuleOp {
  eq = 'eq',
  neq = 'neq',
  gt = 'gt',
  gte = 'gte',
  lt = 'lt',
  lte = 'lte',
  in = 'in',
  nin = 'nin'
}

export type IRequestFilter = {
  op?: FilterOp,
  rules?: IRequestFilterRule[],
}

export type IRequestFilterRule = {
  field: string,
  op: RuleOp,
  value: any
}

const def_Page: IPage = {
  count: 0,
  start: 0,
  limit: 50
};

export const getBizList = () => {
  return CC_Request('search_business', { fields: ['bk_biz_id', 'bk_biz_name'] });
}

/**
 * 获取应用列表
 * @param biz_id 业务ID
 * @param filter 查询过滤条件
 * @param page 分页设置
 * @returns 
 */
export const getAppList = (biz_id: number, filter: IRequestFilter = {}, page: IPage = def_Page) => {
  return Self_Request(`config/list/app/app/biz_id/${biz_id}`, { biz_id, filter, page: { ...page, count: true } });
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
  return Self_Request(`config/create/app/app/biz_id/${biz_id}`, { biz_id, ...params });
}

/**
 * 更新应用
 * @param params { id, biz_id, name?, memo?, reload_type?, reload_file_path? }
 * @returns 
 */
export const updateApp = (params: any) => {
  const { id, biz_id } = params;
  return Self_Request(`config/update/app/app/app_id/${id}/biz_id/${biz_id}`, {}, 'PUT');
}