import http from '../request';
import {
  IConfigEditParams,
  IConfigVersionQueryParams,
  ITemplateBoundByAppData,
  IConfigVersion,
} from '../../types/config';
import { IVariableEditParams } from '../../types/variable';
import { ICommonQuery } from '../../types/index';
import { localT } from '../i18n';

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
export const getConfigList = (biz_id: string, app_id: number, query: ICommonQuery) =>
  http
    .post(`/config/biz/${biz_id}/apps/${app_id}/config_items`, { ...query, with_status: true })
    .then((res) => res.data);

/**
 * 获取已发布版本的非模板配置文件列表
 * @param biz_id 空间ID
 * @param app_id 应用ID
 * @param release_id 版本ID
 * @param params 查询参数
 * @returns
 */
export const getReleasedConfigList = (biz_id: string, app_id: number, release_id: number, params: ICommonQuery) =>
  http.get(`/config/biz/${biz_id}/apps/${app_id}/releases/${release_id}/config_items`, { params }).then((res) => {
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
export const createServiceConfigItem = (app_id: number, biz_id: string, params: IConfigEditParams) =>
  http.post(`/config/create/config_item/config_item/app_id/${app_id}/biz_id/${biz_id}`, params);

/**
 * 更新配置
 * @param id 配置ID
 * @param app_id 服务ID
 * @param biz_id 业务ID
 * @param params 配置参数内容
 * @returns
 */
export const updateServiceConfigItem = (id: number, app_id: number, biz_id: string, params: IConfigEditParams) =>
  http.put(`/config/update/config_item/config_item/config_item_id/${id}/app_id/${app_id}/biz_id/${biz_id}`, params);

/**
 * 删除非模板配置
 * @param id 配置ID
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns
 */
export const deleteServiceConfigItem = (id: number, bizId: string, appId: number) =>
  http.delete(`/config/delete/config_item/config_item/config_item_id/${id}/app_id/${appId}/biz_id/${bizId}`, {});

/**
 * 批量删除非模板配置
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param ids 配置项ID列表
 * @returns
 */
export const batchDeleteServiceConfigs = (bizId: string, appId: number, ids: number[]) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/config_items/batch_delete`, { ids });

/**
 * 获取未命名版本配置文件详情
 * @param biz_id 空间ID
 * @param id 配置ID
 * @param appId 应用ID
 * @returns
 */
export const getConfigItemDetail = (biz_id: string, id: number, appId: number) =>
  http.get(`/config/biz/${biz_id}/apps/${appId}/config_items/${id}`).then((resp) => resp.data);

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
) =>
  http
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
export const updateConfigContent = (
  bizId: string,
  appId: number,
  data: string | File,
  signature: string,
  progress?: Function,
) =>
  http
    .put(`/biz/${bizId}/content/upload`, data, {
      headers: {
        'X-Bscp-App-Id': appId,
        'X-Bkapi-File-Content-Id': signature,
        'X-Bkapi-File-Content-Overwrite': 'false',
        'Content-Type': 'text/plain',
      },
      onUploadProgress: (progressEvent: any) => {
        if (progress) {
          const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          progress(percentCompleted);
        }
      },
    })
    .then((res) => res.data);

/**
 * 下载配置文件内容
 * @param bizId 业务ID
 * @param appId 模板空间ID
 * @param signature sha256签名
 * @param isBlob 是否需要返回二进制流，下载配置文件时需要
 * @returns
 */
export const downloadConfigContent = (bizId: string, appId: number, signature: string, isBlob = false) =>
  http
    .get<string, Blob | string>(`/biz/${bizId}/content/download`, {
      headers: {
        'X-Bscp-Template-Space-Id': appId,
        'X-Bkapi-File-Content-Id': signature,
      },
      transitional: {
        forcedJSONParsing: false,
      },
      ...(isBlob && { responseType: 'blob' }), // 文件为二进制流，需要设置响应类型为blob才能正确解析
    })
    .then((res) => res);

/**
 * 判断上传的配置文件是否已存在
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param data 配置内容
 * @param signature 文件内容的SHA256值
 * @returns
 */
export const getConfigUploadFileIsExist = (bizId: string, appId: number, signature: string) =>
  http
    .get(`/biz/${bizId}/content/metadata`, {
      headers: {
        'X-Bscp-App-Id': appId,
        'X-Bkapi-File-Content-Id': signature,
      },
    })
    .then((res) => res.data);

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
export const createVersion = (bizId: string, appId: number, params: ICreateVersionParams) =>
  http.post(`/config/create/release/release/app_id/${appId}/biz_id/${bizId}`, params);

/**
 * 废弃版本
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param releaseId 版本ID
 * @returns
 */
export const deprecateVersion = (bizId: string, appId: number, releaseId: number) =>
  http.put(`/config/biz/${bizId}/apps/${appId}/releases/${releaseId}/deprecate`);

/**
 * 恢复版本
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param releaseId 版本ID
 * @returns
 */
export const undeprecateVersion = (bizId: string, appId: number, releaseId: number) =>
  http.put(`/config/biz/${bizId}/apps/${appId}/releases/${releaseId}/undeprecate`);

/**
 * 删除版本
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param releaseId 版本ID
 * @returns
 */
export const deleteVersion = (bizId: string, appId: number, releaseId: number) =>
  http.delete(`/config/biz/${bizId}/apps/${appId}/releases/${releaseId}`);

/**
 * 获取版本列表
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param params 查询参数
 * @returns
 */
export const getConfigVersionList = (bizId: string, appId: number, params: IConfigVersionQueryParams) =>
  http.get(`config/biz/${bizId}/apps/${appId}/releases`, { params }).then((res) => {
    res.data.details.forEach((item: IConfigVersion) => {
      const defaultGroup = item.status.released_groups.find((group) => group.id === 0);
      if (defaultGroup) {
        defaultGroup.name = localT('全部实例');
      }
    });
    return res;
  });

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
 * 发布版本(增加审批)
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param data 参数
 * @param publish_type 上线方式
 * @param publish_time 定时上线时间
 * @param is_compare 所有待上线的分组是否为首次上线
 * @returns
 */
export const publishVerSubmit = (
  bizId: string,
  appId: number,
  releaseId: number,
  data: {
    groups: Array<number>;
    all: boolean;
    memo: string;
    publish_type: 'Manually' | 'Automatically' | 'Periodically' | 'Immediately' | '';
    publish_time: Date | string;
    is_compare: boolean;
  },
) => http.post(`/config/biz_id/${bizId}/app_id/${appId}/release_id/${releaseId}/submit`, data);

/**
 * 获取服务下初始化脚本引用配置
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param releaseId 版本ID
 * @returns
 */
export const getConfigScript = (bizId: string, appId: number, releaseId: number) =>
  http.get(`/config/biz/${bizId}/apps/${appId}/releases/${releaseId}/hooks`).then((response) => {
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
) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/template_bindings/conflict_check`, params).then((res) => {
    const conflictData: { [key: number]: number[] } = {};
    res.data.details.forEach(
      (item: { template_id: number; template_name: string; template_set_id: number; template_set_name: string }) => {
        if (Array.isArray(conflictData[item.template_set_id])) {
          conflictData[item.template_set_id].push(item.template_id);
        } else {
          conflictData[item.template_set_id] = [item.template_id];
        }
      },
    );
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
export const getBoundTemplates = (bizId: string, appId: number, query: ICommonQuery) =>
  http
    .get(`/config/biz/${bizId}/apps/${appId}/template_revisions`, { params: { ...query, with_status: true } })
    .then((res) => res.data);

/**
 * 获取服务下已命名版本绑定的模板配置文件列表
 * @param bizId
 * @param appId
 * @param releaseId
 * @returns
 */
export const getBoundTemplatesByAppVersion = (bizId: string, appId: number, releaseId: number, params: ICommonQuery) =>
  http
    .get(`/config/biz/${bizId}/apps/${appId}/releases/${releaseId}/template_revisions`, { params })
    .then((res) => res.data);

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
export const deleteBoundPkg = (bizId: string, appId: number, bindingId: number, template_set_ids: number[]) =>
  http.delete(`/config/biz/${bizId}/apps/${appId}/template_bindings/${bindingId}/template_set`, {
    params: { template_set_ids: template_set_ids.join(',') },
  });

/**
 * 删除服务下绑定的模板套餐
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param id 配置ID
 * @returns
 */
export const deleteCurrBoundPkg = (bizId: string, appId: number, id: number) =>
  http.delete(`/config/biz/${bizId}/apps/${appId}/template_set/${id}`);

/**
 * 导入非模板配置文件压缩包
 * @param biz_id 业务ID
 * @param appId 应用ID
 * @param file 导入文件
 * @returns
 */
export const importNonTemplateConfigFile = (
  biz_id: string,
  appId: number,
  file: any,
  isDecompression: boolean,
  progress: Function,
) =>
  http
    .post(`/config/biz/${biz_id}/apps/${appId}/config_item/import/${encodeURIComponent(file.name)}`, file, {
      headers: {
        'X-Bscp-Unzip': isDecompression,
      },
      onUploadProgress: (progressEvent: any) => {
        const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
        progress(percentCompleted);
      },
    })
    .then((res) => res.data);

/**
 * 批量添加非模板配置列表
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param bindingId 模板和服务绑定关系ID
 * @param template_set_ids 模板套餐ID列表
 * @returns
 */
export const batchAddConfigList = (bizId: string, appId: number, query: any) =>
  http.put(`/config/biz/${bizId}/apps/${appId}/config_items`, query).then((res) => res.data);

/**
 * 创建kv
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param kv 配置键值类型
 * @returns
 */
export const createKv = (bizId: string, appId: number, kv: any) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/kvs`, kv);

/**
 * 获取kv
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns
 */
export const getKvList = (bizId: string, appId: number, query: ICommonQuery) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/kvs/list`, query).then((res) => res.data);

/**
 * 更新kv
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param key 配置键
 * @param value 配置值
 * @returns
 */
export const updateKv = (
  bizId: string,
  appId: number,
  key: string,
  editContent: { value: string; memo: string; secret_hidden?: boolean },
) => http.put(`/config/biz/${bizId}/apps/${appId}/kvs/${key}`, editContent);

/**
 * 删除kv
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param configId 配置项ID
 * @returns
 */
export const deleteKv = (bizId: string, appId: number, configId: number) =>
  http.delete(`/config/biz/${bizId}/apps/${appId}/kvs/${configId}`);

/**
 * 批量删除kv
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param ids 配置项ID列表
 * @param exclusion_operation 是否跨页
 */
export const batchDeleteKv = (bizId: string, appId: number, ids: number[], exclusion_operation: boolean) =>
  http.post(`config/biz/${bizId}/apps/${appId}/kvs/batch_delete`, { ids, exclusion_operation });

/**
 * 获取已发布kv
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param releaseId 版本id
 * @param key 配置键
 * @returns
 */
export const getReleaseKv = (bizId: string, appId: number, releaseId: number, key: string) =>
  http.get(`/config/biz/${bizId}/apps/${appId}/releases/${releaseId}/kvs/${key}`).then((res) => res.data);

/**
 * 获取已发布kv列表
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param releaseId 版本id
 * @returns
 */
export const getReleaseKvList = (bizId: string, appId: number, releaseId: number, query: ICommonQuery) =>
  http.get(`/config/biz/${bizId}/apps/${appId}/releases/${releaseId}/kvs`, { params: query }).then((res) => res.data);

/**
 * 撤销删除kv
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param kv 配置键值类型
 * @returns
 */
export const undeleteKv = (bizId: string, appId: number, key: string) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/kvs/${key}/undelete`);

/**
 * 恢复修改kv
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param kv 配置键值类型
 * @returns
 */
export const unModifyKv = (bizId: string, appId: number, key: string) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/kvs/${key}/undo`);

/**
 * 批量导入kv配置文件
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param File 配置文件
 * @returns
 */
export const batchImportKvFile = (bizId: string, appId: number, File: any) =>
  http.post(`/biz/${bizId}/apps/${appId}/kvs/import`, File);

/**
 * 导出kv配置文件
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param releaseId 发布版本ID
 * @returns
 */
export const getExportKvFile = (bizId: string, appId: number, releaseId: number, format: string) =>
  http.get(`biz/${bizId}/apps/${appId}/releases/${releaseId}/kvs/export`, { params: { format } });

/**
 * 撤销修改配置文件
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param id 配置文件ID
 * @returns
 */
export const unModifyConfigItem = (bizId: string, appId: number, id: number) =>
  http.post(`/config/undo/config_item/config_item/config_item_id/${id}/app_id/${appId}/biz_id/${bizId}`);

/**
 * 恢复删除配置文件
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param id 配置文件ID
 * @returns
 */
export const unDeleteConfigItem = (bizId: string, appId: number, id: number) =>
  http.post(`/config/undelete/config_item/config_item/config_item_id/${id}/app_id/${appId}/biz_id/${bizId}`);

/**
 * 从历史版本导入配置项
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param other_app_id 导入服务id
 * @param release_id 版本id
 * @returns
 */
export const importFromHistoryVersion = (
  bizId: string,
  appId: number,
  params: { other_app_id: number; release_id: number },
) => http.get(`/config/biz/${bizId}/apps/${appId}/config_items/compare_conflicts`, { params });

/**
 * 从历史版本导入kv配置项
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param other_app_id 导入服务id
 * @param release_id 版本id
 * @returns
 */
export const importKvFromHistoryVersion = (
  bizId: string,
  appId: number,
  params: { other_app_id: number; release_id: number },
) => http.get(`/config/biz/${bizId}/apps/${appId}/kvs/compare_conflicts`, { params });

/**
 * 简单文本导入kv配置项
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param kvs 上传kv列表
 * @returns
 */
export const importKvFormText = (bizId: string, appId: number, kvs: any, replace_all: boolean) =>
  http.put(`/config/biz/${bizId}/apps/${appId}/kvs`, { kvs, replace_all });

/**
 * json文本导入kv配置项
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param kvs 上传kv列表
 * @returns
 */
export const importKvFormJson = (bizId: string, appId: number, content: string) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/kvs/json/import`, { data: content });

/**
 * yaml文本导入kv配置项
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param kvs 上传kv列表
 * @returns
 */
export const importKvFormYaml = (bizId: string, appId: number, content: string) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/kvs/yaml/import`, { data: content });

/**
 * 判断生成版本名称是否重名
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param kvs 上传kv列表
 * @returns
 */
export const createVersionNameCheck = (bizId: string, appId: number, name: string) =>
  http.get(`/config/biz_id/${bizId}/app_id/${appId}/release/${name}/check`);

/**
 * 从配置模板导入配置文件
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param bindingId 模板和服务绑定关系ID
 * @param params 更新参数
 * @returns
 */
export const importConfigFromTemplate = (bizId: string, appId: number, query: any) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/template_bindings/import_template_set`, query);

/**
 * 上次上线方式查询
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns
 */
export const publishType = (bizId: string, appId: number) =>
  http.get(`/config/biz_id/${bizId}/app_id/${appId}/last/select`);

/**
 * 当前版本状态查询
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param releaseId 版本ID
 * @returns
 */
export const versionStatusQuery = (bizId: string, appId: number, releaseId: number) =>
  http.get(`/config/biz_id/${bizId}/app_id/${appId}/release_id/${releaseId}/status`);

/**
 * 当前服务下 所有版本上线状态检查
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param releaseId 版本ID
 * @returns
 */
export const versionStatusCheck = (bizId: string, appId: number) =>
  http.get(`/config/biz_id/${bizId}/app_id/${appId}/last/publish`);
