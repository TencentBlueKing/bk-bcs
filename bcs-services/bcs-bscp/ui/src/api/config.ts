import http from '../request';
import { IConfigEditParams, IConfigVersionQueryParams, ITemplateBoundByAppData } from '../../types/config';
import { IVariableEditParams } from '../../types/variable';
import { ICommonQuery } from '../../types/index';

// 配置文件版本下脚本配置接口可能会返回null，做数据兼容处理
export const getDefaultConfigScriptData = () => ({
  hook_id: 0,
  hook_name: '',
  hook_revision_id: 0,
  hook_revision_name: '',
  type: '',
  content: '',
});

/**
 * 获取未命名版本的配置文件列表
 * @param biz_id 空间ID
 * @param app_id 应用ID
 * @param query 查询参数
 * @returns
 */
export const getConfigList = (biz_id: string, app_id: number, query: ICommonQuery) => http
  .get(`/config/biz/${biz_id}/apps/${app_id}/config_items`, { params: { ...query, with_status: true } })
  .then(res => res.data);

/**
 * 获取已发布版本的非模板配置文件列表
 * @param biz_id 空间ID
 * @param app_id 应用ID
 * @param release_id 版本ID
 * @param params 查询参数
 * @returns
 */
export const getReleasedConfigList = (biz_id: string, app_id: number, release_id: number, params: ICommonQuery) => http.get(`/config/biz/${biz_id}/apps/${app_id}/releases/${release_id}/config_items`, { params }).then((res) => {
  res.data.details.forEach((item: any) => {
    // 接口返回的config_item_id为实际的配置文件id，id字段没有到，统一替换
    item.id = item.config_item_id;
  });
  return res.data;
});

/**
 * 新增配置
 * @param app_id 服务ID
 * @param biz_id 业务ID
 * @param params 配置参数内容
 * @returns
 */
export const createServiceConfigItem = (app_id: number, biz_id: string, params: IConfigEditParams) => http.post(`/config/create/config_item/config_item/app_id/${app_id}/biz_id/${biz_id}`, params);

/**
 * 更新配置
 * @param id 配置ID
 * @param app_id 服务ID
 * @param biz_id 业务ID
 * @param params 配置参数内容
 * @returns
 */
export const updateServiceConfigItem = (id: number, app_id: number, biz_id: string, params: IConfigEditParams) => http.put(`/config/update/config_item/config_item/config_item_id/${id}/app_id/${app_id}/biz_id/${biz_id}`, params);

/**
 * 删除配置
 * @param id 配置ID
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns
 */
export const deleteServiceConfigItem = (id: number, bizId: string, appId: number) => http.delete(`/config/delete/config_item/config_item/config_item_id/${id}/app_id/${appId}/biz_id/${bizId}`, {});

/**
 * 获取未命名版本配置文件详情
 * @param biz_id 空间ID
 * @param id 配置ID
 * @param appId 应用ID
 * @returns
 */
export const getConfigItemDetail = (biz_id: string, id: number, appId: number) => http.get(`/config/biz/${biz_id}/apps/${appId}/config_items/${id}`).then(resp => resp.data);

/**
 * 获取已发布版本配置文件详情
 * @param biz_id 空间ID
 * @param app_id 应用ID
 * @param release_id 版本ID
 * @param config_item_id 配置文件ID
 * @returns
 */
export const getReleasedConfigItemDetail = (
  biz_id: string,
  app_id: number,
  release_id: number,
  config_item_id: number,
) => http
  .get(`/config/biz/${biz_id}/apps/${app_id}/releases/${release_id}/config_items/${config_item_id}`)
  .then((resp) => {
    resp.data.config_item.id = resp.data.config_item_id;
    return resp.data;
  });

/**
 * 上传配置文件内容
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param data 配置内容
 * @param signature 文件内容的SHA256值
 * @returns
 */
export const updateConfigContent = (bizId: string, appId: number, data: string | File, signature: string) => http
  .put(`/biz/${bizId}/content/upload`, data, {
    headers: {
      'X-Bscp-App-Id': appId,
      'X-Bkapi-File-Content-Id': signature,
      'X-Bkapi-File-Content-Overwrite': 'false',
      'Content-Type': 'text/plain',
    },
  })
  .then(res => res.data);

/**
 * 下载配置文件内容
 * @param bizId 业务ID
 * @param appId 模板空间ID
 * @param signature sha256签名
 * @returns
 */
export const downloadConfigContent = (bizId: string, appId: number, signature: string) => http
  .get<string, string>(`/biz/${bizId}/content/download`, {
    headers: {
      'X-Bscp-Template-Space-Id': appId,
      'X-Bkapi-File-Content-Id': signature,
    },
    transitional: {
      forcedJSONParsing: false,
    },
  })
  .then(res => res);

/**
 * 创建配置版本
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param params 请求参数
 * @returns
 */
interface ICreateVersionParams {
  name: string;
  memo: string;
  variables: IVariableEditParams[];
}
export const createVersion = (bizId: string, appId: number, params: ICreateVersionParams) => http.post(`/config/create/release/release/app_id/${appId}/biz_id/${bizId}`, params);

/**
 * 获取版本列表
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param params 查询参数
 * @returns
 */
export const getConfigVersionList = (bizId: string, appId: number, params: IConfigVersionQueryParams) => http.get(`config/biz/${bizId}/apps/${appId}/releases`, { params });

/**
 * 发布版本
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param name 版本名称
 * @param data 参数
 * @returns
 */
export const publishVersion = (
  bizId: string,
  appId: number,
  releaseId: number,
  data: {
    groups: Array<number>;
    all: boolean;
    memo: string;
  },
) => http.post(`/config/update/strategy/publish/publish/release_id/${releaseId}/app_id/${appId}/biz_id/${bizId}`, data);

/**
 * 获取服务下初始化脚本引用配置
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param releaseId 版本ID
 * @returns
 */
export const getConfigScript = (bizId: string, appId: number, releaseId: number) => http.get(`/config/biz/${bizId}/apps/${appId}/releases/${releaseId}/hooks`).then((response) => {
  const { pre_hook, post_hook } = response.data;
  const data = {
    pre_hook: getDefaultConfigScriptData(),
    post_hook: getDefaultConfigScriptData(),
  };
  if (pre_hook) {
    data.pre_hook = pre_hook;
  }
  if (post_hook) {
    data.post_hook = post_hook;
  }
  return data;
});

/**
 * 更新服务下初始化脚本引用配置
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param params 配置数据
 * @returns
 */
export const updateConfigInitScript = (
  bizId: string,
  appId: number,
  params: { pre_hook_id: number | undefined; post_hook_id: number | undefined },
) => http.put(`/config/biz/${bizId}/apps/${appId}/config_hooks`, params);

/**
 * 检测导入模板与已存在配置文件的冲突详情
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns
 */
export const checkAppTemplateBinding = (
  bizId: string,
  appId: number,
  params: { bindings: ITemplateBoundByAppData[] },
) => http.post(`/config/biz/${bizId}/apps/${appId}/template_bindings/conflict_check`, params).then((res) => {
  const conflictData: { [key: number]: number[] } = {};
  res.data.details.forEach((item:
    { template_id: number; template_name: string; template_set_id: number; template_set_name: string }) => {
    if (Array.isArray(conflictData[item.template_set_id])) {
      conflictData[item.template_set_id].push(item.template_id);
    } else {
      conflictData[item.template_set_id] = [item.template_id];
    }
  });
  return conflictData;
});

/**
 * 新建模板配置文件和服务绑定关系
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param params 查询参数
 * @returns
 */
export const importTemplateConfigPkgs = (
  bizId: string,
  appId: number,
  params: { bindings: ITemplateBoundByAppData[] },
) => http.post(`/config/biz/${bizId}/apps/${appId}/template_bindings`, params);

/**
 * 更新模板配置文件和服务绑定关系
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param bindingId 模板和服务绑定关系ID
 * @param params 更新参数
 * @returns
 */
export const updateTemplateConfigPkgs = (
  bizId: string,
  appId: number,
  bindingId: number,
  params: { bindings: ITemplateBoundByAppData[] },
) => http.put(`/config/biz/${bizId}/apps/${appId}/template_bindings/${bindingId}`, params);

/**
 * 获取服务下未命名版本绑定的模板配置文件列表
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param query 查询参数
 * @returns
 */
export const getBoundTemplates = (bizId: string, appId: number, query: ICommonQuery) => http
  .get(`/config/biz/${bizId}/apps/${appId}/template_revisions`, { params: { ...query, with_status: true } })
  .then(res => res.data);

/**
 * 获取服务下已命名版本绑定的模板配置文件列表
 * @param bizId
 * @param appId
 * @param releaseId
 * @returns
 */
export const getBoundTemplatesByAppVersion =
  (bizId: string, appId: number, releaseId: number, params: ICommonQuery) => http
    .get(`/config/biz/${bizId}/apps/${appId}/releases/${releaseId}/template_revisions`, { params })
    .then(res => res.data);

/**
 * 更新服务下模板配置文件版本
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param bindingId 模板和服务绑定关系ID
 * @param params 更新参数
 * @returns
 */
export const updateBoundTemplateVersion = (
  bizId: string,
  appId: number,
  bindingId: number,
  params: { bindings: ITemplateBoundByAppData[] },
) => http.put(`/config/biz/${bizId}/apps/${appId}/template_bindings/${bindingId}/template_revisions`, params);

/**
 * 删除服务下绑定的模板套餐
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param bindingId 模板和服务绑定关系ID
 * @param template_set_ids 模板套餐ID列表
 * @returns
 */
export const deleteBoundPkg = (bizId: string, appId: number, bindingId: number, template_set_ids: number[]) => http.delete(`/config/biz/${bizId}/apps/${appId}/template_bindings/${bindingId}/template_sets`, {
  params: { template_set_ids: template_set_ids.join(',') },
});

/**
 * 导入非模板配置文件压缩包
 * @param biz_id 业务ID
 * @param appId 应用ID
 * @param fill 导入文件
 * @returns
 */
export const importNonTemplateConfigFile = (biz_id: string, appId: number, fill: any) => http
  .post(`/config/biz/${biz_id}/apps/${appId}/config_item/import`, fill, {
    headers: {
      'Content-Type': 'application/zip',
    },
  })
  .then(res => res.data);

/**
 * 批量添加非模板配置列表
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param bindingId 模板和服务绑定关系ID
 * @param template_set_ids 模板套餐ID列表
 * @returns
 */
export const batchAdddConfigList = (bizId: string, appId: number, list: any) => http.put(`/config/biz/${bizId}/apps/${appId}/config_items`, {
  items: list,
});
