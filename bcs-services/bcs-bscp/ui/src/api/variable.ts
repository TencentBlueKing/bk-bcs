import http from '../request';
import { ICommonQuery } from '../../types/index';
import { IVariableEditParams, IVariableImportParams } from '../../types/variable';

/**
 * 查询变量列表
 * @param biz_id 业务ID
 * @param params
 * @returns
 */
export const getVariableList = (biz_id: string, params: ICommonQuery) =>
  http.get(`/config/biz/${biz_id}/template_variables`, { params }).then((res) => res.data);

/**
 * 创建变量
 * @param biz_id 业务ID
 * @param params 创建参数
 * @returns
 */
export const createVariable = (biz_id: string, params: IVariableEditParams) =>
  http.post(`/config/biz/${biz_id}/template_variables`, params);

/**
 * 编辑变量
 * @param biz_id 业务ID
 * @param template_variable_id 变量ID
 * @param params 编辑参数
 * @returns
 */
export const updateVariable = (
  biz_id: string,
  template_variable_id: number,
  params: { default_val: string; memo: string },
) => http.put(`/config/biz/${biz_id}/template_variables/${template_variable_id}`, params);

/**
 * 删除变量
 * @param biz_id 业务ID
 * @param template_variable_id 变量ID
 * @returns
 */
export const deleteVariable = (biz_id: string, template_variable_id: number) =>
  http.delete(`/config/biz/${biz_id}/template_variables/${template_variable_id}`);

/**
 * 批量删除变量
 * @param bizId 业务ID
 * @param ids 变量ID列表
 * @param exclusion_operation 是否跨页
 */
export const batchDeleteVariable = (biz_id: string, ids: number[], exclusion_operation: boolean) =>
  http.post(`/config/biz/${biz_id}/template_variables/batch_delete`, { ids, exclusion_operation });

/**
 * 获取未命名版本服务变量列表
 * @param biz_id 业务ID
 * @param app_id 应用ID
 * @returns
 */
export const getUnReleasedAppVariables = (biz_id: string, app_id: number) =>
  http.get(`/config/biz/${biz_id}/apps/${app_id}/template_variables`).then((res) => res.data);

/**
 * 更新未命名版本服务变量列表
 * @param biz_id 业务ID
 * @param app_id 应用ID
 * @returns
 */
export const updateUnReleasedAppVariables = (biz_id: string, app_id: number, variables: IVariableEditParams[]) =>
  http.put(`/config/biz/${biz_id}/apps/${app_id}/template_variables`, { variables }).then((res) => res.data);

/**
 * 获取服务某个版本的变量列表
 * @param biz_id 业务ID
 * @param app_id 应用ID
 * @param release_id 服务版本ID
 * @returns
 */
export const getReleasedAppVariables = (biz_id: string, app_id: number, release_id: number) =>
  http
    .get(`/config/biz/${biz_id}/apps/${app_id}/releases/${release_id}/template_variables`, { params: {} })
    .then((res) => res.data);

/**
 * 查询未命名版本服务中变量被配置文件引用详情
 * @param biz_id 业务ID
 * @param app_id 应用ID
 * @returns
 */
export const getUnReleasedAppVariablesCitedDetail = (biz_id: string, app_id: number) =>
  http.get(`/config/biz/${biz_id}/apps/${app_id}/template_variables_references`).then((res) => res.data);

/**
 * 查询服务某个版本的变量被配置文件引用详情
 * @param biz_id 业务ID
 * @param app_id 应用ID
 * @param release_id 服务版本ID
 * @returns
 */
export const getReleasedAppVariablesCitedDetail = (biz_id: string, app_id: number, release_id: number) =>
  http
    .get(`/config/biz/${biz_id}/apps/${app_id}/releases/${release_id}/template_variables_references`)
    .then((res) => res.data);

/**
 * 批量导入变量
 * @param biz_id 业务ID
 * @param app_id 应用ID
 * @param release_id 服务版本ID
 * @returns
 */
export const batchImportTemplateVariables = (biz_id: string, params: IVariableImportParams) =>
  http.post(`config/biz/${biz_id}/template_variables/import`, params);
// @todo 待接口支持
