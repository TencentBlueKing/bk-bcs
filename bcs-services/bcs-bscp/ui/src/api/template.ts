import http from '../request';
import { ICommonQuery } from '../../types/index';
import { ITemplatePackageEditParams, ITemplateVersionEditingData, ITemplateCitedByPkgs } from '../../types/template';
import { IConfigEditParams } from '../../types/config';

/**
 * 获取模板空间列表
 * @param biz_id 业务ID
 * @param params 查询参数
 * @returns
 */
export const getTemplateSpaceList = (biz_id: string, params: ICommonQuery) =>
  http.get(`/config/biz/${biz_id}/template_spaces`, { params }).then((res) => res.data);

/**
 * 创建模板空间
 * @param biz_id 业务ID
 * @param params 创建模板参数
 * @returns
 */
export const createTemplateSpace = (biz_id: string, params: { name: string; memo: string }) =>
  http.post(`/config/biz/${biz_id}/template_spaces`, params).then((res) => res.data);

/**
 * 更新模板空间
 * @param biz_id 业务ID
 * @param id 模板ID
 * @param params 模板参数
 * @returns
 */
export const updateTemplateSpace = (biz_id: string, id: number, params: { memo: string }) =>
  http.put(`/config/biz/${biz_id}/template_spaces/${id}`, params).then((res) => res.data);

/**
 * 删除模板空间
 * @param biz_id 业务ID
 * @param id 模板ID
 * @returns
 */
export const deleteTemplateSpace = (biz_id: string, id: number) =>
  http.delete(`/config/biz/${biz_id}/template_spaces/${id}`);

/**
 * 创建模板套餐
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @param params 创建模板套餐参数
 * @returns
 */
export const createTemplatePackage = (biz_id: string, template_space_id: number, params: ITemplatePackageEditParams) =>
  http.post(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets`, params).then((res) => res.data);

/**
 * 编辑模板套餐
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @param template_set_id 模板套餐ID
 * @param params 编辑模板套餐参数
 * @returns
 */
export const updateTemplatePackage = (
  biz_id: string,
  template_space_id: number,
  template_set_id: number,
  params: ITemplatePackageEditParams,
) =>
  http
    .put(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/${template_set_id}`, params)
    .then((res) => res.data);

/**
 * 删除模板套餐
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @param template_set_id 模板套餐ID
 * @returns
 */
export const deleteTemplatePackage = (biz_id: string, template_space_id: number, template_set_id: number) =>
  http
    .delete(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/${template_set_id}`, {
      params: { force: true },
    })
    .then((res) => res.data);

/**
 * 获取空间下的模板套餐列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @param params 查询参数
 * @returns
 */
export const getTemplatePackageList = (biz_id: string, template_space_id: string, params: ICommonQuery) =>
  http
    .get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets`, { params })
    .then((res) => res.data);

/**
 * 获取业务下所有模板套餐（按模板空间分组）
 * @param biz_id 业务ID
 * @param params 查询参数，传app_id可筛选服务下的套餐模板
 * @returns
 */
export const getAllPackagesGroupBySpace = (biz_id: string, params: { app_id?: number }) =>
  http.get(`/config/biz/${biz_id}/template_sets/list_all_of_biz`, { params }).then((res) => res.data);

/**
 * 获取模板空间下的全部配置文件模板列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @returns
 */
export const getTemplatesBySpaceId = (biz_id: string, template_space_id: number, params: ICommonQuery) =>
  http.get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates`, { params }).then((res) => res.data);

/**
 * 获取模板空间下未指定套餐的模板列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @returns
 */
export const getTemplatesWithNoSpecifiedPackage = (biz_id: string, template_space_id: number, params: ICommonQuery) =>
  http
    .get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/list_not_bound`, { params })
    .then((res) => res.data);

/**
 * 获取当前模板套餐被未命名版本引用的服务引用列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @param template_set_id 模板套餐Id
 * @param params 查询参数
 * @returns
 */
export const getUnNamedVersionAppsBoundByPackage = (
  biz_id: string,
  template_space_id: number,
  template_set_id: number,
  params: ICommonQuery,
) =>
  http
    .get(
      `/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/${template_set_id}/bound_unnamed_app_details`,
      { params },
    )
    .then((res) => res.data);

/**
 * 获取多个模板套餐被未命名版本引用的服务引用列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @returns
 */
export const getUnNamedVersionAppsBoundByPackages = (
  biz_id: string,
  template_space_id: number,
  template_sets: number[],
  params: ICommonQuery,
) =>
  http
    .get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/bound_unnamed_app_details`, {
      params: { template_set_ids: template_sets.join(','), ...params },
    })
    .then((res) => res.data);

/**
 * 获取当前模板套餐被已生成版本引用的服务引用列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @param template_set_id 模板套餐Id
 * @param params 查询参数
 * @returns
 */
export const getReleasedVersionAppsBoundByPackage = (
  biz_id: string,
  template_space_id: number,
  template_set_id: number,
  params: ICommonQuery,
) =>
  http
    .get(
      `/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/${template_set_id}/bound_unnamed_app_details`,
      { params },
    )
    .then((res) => res.data);

/**
 * 获取模板套餐下的配置文件模板列表
 * @param biz_id 业务ID
 * @param template_space_id 模板空间ID
 * @returns
 */
export const getTemplatesByPackageId = (
  biz_id: string,
  template_space_id: number,
  template_set_id: number,
  params: ICommonQuery,
) =>
  http
    .get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/template_sets/${template_set_id}/templates`, {
      params,
    })
    .then((res) => res.data);

/**
 * 创建模板
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param params 配置文件参数
 * @returns
 */
export const createTemplate = (biz_id: string, template_space_id: number, params: IConfigEditParams) =>
  http.post(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates`, params);

/**
 * 上传模板配置内容
 * @param biz_id 业务ID
 * @param templateSpaceId 模板空间ID
 * @param data 配置内容
 * @param signature sha256签名
 * @returns
 */
export const updateTemplateContent = (
  biz_id: string,
  templateSpaceId: number,
  data: string | File,
  signature: string,
  progress?: Function,
) =>
  http
    .put(`/biz/${biz_id}/content/upload`, data, {
      headers: {
        'X-Bscp-Template-Space-Id': templateSpaceId,
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
 * 下载模板配置内容
 * @param biz_id 业务ID
 * @param templateSpaceId 模板空间ID
 * @param signature sha256签名
 * @param isBlob 是否需要返回二进制流，下载配置文件时需要
 * @returns
 */
export const downloadTemplateContent = (biz_id: string, templateSpaceId: number, signature: string, isBlob = false) =>
  http
    .get<string, string>(`/biz/${biz_id}/content/download`, {
      headers: {
        'X-Bscp-Template-Space-Id': templateSpaceId,
        'X-Bkapi-File-Content-Id': signature,
      },
      transitional: {
        forcedJSONParsing: false,
      },
      ...(isBlob && { responseType: 'blob' }), // 文件为二进制流，需要设置响应类型为blob才能正确解析
    })
    .then((res) => res);

/**
 * 批量删除模板
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_ids 模板ID列表
 * @returns
 */
export const deleteTemplate = (biz_id: string, template_space_id: number, template_ids: number[]) =>
  http.delete(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates`, {
    params: { template_ids: template_ids.join(',') },
  });

/**
 * 根据模板id列表查询对应模板详情
 * @param biz_id 业务ID
 * @param ids 模板ID列表
 * @returns
 */
export const getTemplatesDetailByIds = (biz_id: string, ids: number[]) =>
  http.post(`/config/biz/${biz_id}/templates/list_by_ids`, { ids }).then((res) => res.data);

/**
 * 添加模版到模版套餐(多个模板添加到多个套餐)
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param template_set_ids 模板套餐列表
 * @returns
 */
export const addTemplateToPackage = (
  biz_id: string,
  template_space_id: number,
  template_ids: number[],
  template_set_ids: number[],
  exclusion_operation: boolean,
) =>
  http.post(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/add_to_template_sets`, {
    template_ids,
    template_set_ids,
    exclusion_operation,
  });

/**
 * 将模版移出套餐(多个模板移出多个套餐)
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param template_set_ids 模板套餐列表
 * @returns
 */
export const moveOutTemplateFromPackage = (
  biz_id: string,
  template_space_id: number,
  template_ids: number[],
  template_set_ids: number[],
) =>
  http.post(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/delete_from_template_sets`, {
    template_ids,
    template_set_ids,
  });

/**
 * 查询单个模板被套餐引用详情
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param params 列表查询参数
 * @returns
 */
export const getPackagesByTemplateId = (
  biz_id: string,
  template_space_id: number,
  template_id: number,
  params: ICommonQuery,
) =>
  http
    .get(
      `/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/${template_id}/bound_template_set_details`,
      { params },
    )
    .then((res) => res.data);

/**
 * 查询多个模板被套餐引用详情
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_ids 模板ID
 * @param params 列表查询参数
 * @returns
 */
export const getPackagesByTemplateIds = (biz_id: string, template_space_id: number, template_ids: number[]) => {
  const params = {
    template_ids: template_ids.join(','),
    start: 0,
    all: true,
  };
  return http
    .get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/bound_template_set_details`, { params })
    .then((res) => {
      const list: ITemplateCitedByPkgs[][] = [];
      template_ids.forEach((id, index) => {
        const data = res.data.details.filter((item: ITemplateCitedByPkgs) => item.template_id === id);
        list[index] = data;
      });
      return { details: list };
    });
};

/**
 * 查询模板被引用计数
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_ids 模板ID列表
 * @returns
 */
export const getCountsByTemplateIds = (biz_id: string, template_space_id: number, template_ids: number[]) =>
  http
    .post(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/bound_counts`, { template_ids })
    .then((res) => res.data);

/**
 * 查询模板被未命名版本服务引用详情
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param params 列表查询参数
 * @returns
 */
export const getUnNamedVersionAppsBoundByTemplate = (
  biz_id: string,
  template_space_id: number,
  template_id: number,
  params: ICommonQuery,
) =>
  http
    .get(
      `/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/${template_id}/bound_unnamed_app_details`,
      { params },
    )
    .then((res) => res.data);

/**
 * 创建模板版本
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param params 模板配置参数
 * @returns
 */
export const createTemplateVersion = (
  biz_id: string,
  template_space_id: number,
  template_id: number,
  params: ITemplateVersionEditingData,
) =>
  http
    .post(
      `/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/${template_id}/template_revisions`,
      params,
    )
    .then((res) => res.data);

/**
 * 获取模板版本列表
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param params 查询参数
 * @returns
 */
export const getTemplateVersionList = (
  biz_id: string,
  template_space_id: number,
  template_id: number,
  params: ICommonQuery,
) =>
  http
    .get(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/${template_id}/template_revisions`, {
      params,
    })
    .then((res) => res.data);

/**
 * 根据模板版本id列表查询对应模板版本详情
 * @param biz_id 业务ID
 * @param ids 模板版本ID列表
 * @returns
 */
export const getTemplateVersionsDetailByIds = (biz_id: string, ids: number[]) =>
  http.post(`/config/biz/${biz_id}/template_revisions/list_by_ids`, { ids }).then((res) => res.data);

/**
 *
 * @param biz_id 业务ID
 * @param app_id 应用ID
 * @param release_id 应用版本ID
 * @param template_revision_id 模板版本ID
 * @returns
 */
export const getTemplateVersionDetail = (
  biz_id: string,
  app_id: number,
  release_id: number,
  template_revision_id: number,
) =>
  http
    .get(`/config/biz/${biz_id}/apps/${app_id}/releases/${release_id}/template_revisions/${template_revision_id}`)
    .then((res) => res.data);

/**
 * 删除模板版本
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param template_revision_id 版本ID
 * @returns
 */
export const deleteTemplateVersion = (
  biz_id: string,
  template_space_id: number,
  template_id: number,
  template_revision_id: number,
) =>
  http.delete(
    `/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/${template_id}/template_revisions/${template_revision_id}`,
  );

/**
 * 查询模板版本被引用计数
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param ids 版本ID列表
 * @returns
 */
export const getCountsByTemplateVersionIds = (
  biz_id: string,
  template_space_id: number,
  template_id: number,
  ids: number[],
) =>
  http
    .post(
      `/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/${template_id}/template_revisions/bound_counts`,
      { template_revision_ids: ids },
    )
    .then((res) => res.data);

/**
 * 获取模板版本被未命名版本服务引用详情
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param template_revision_id 版本ID
 * @returns
 */
export const getUnNamedVersionAppsBoundByTemplateVersion = (
  biz_id: string,
  template_space_id: number,
  template_id: number,
  template_revision_id: number,
  params: ICommonQuery,
) =>
  http
    .get(
      `/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/${template_id}/template_revisions/${template_revision_id}/bound_unnamed_app_details`,
      { params },
    )
    .then((res) => res.data);

/**
 * 获取模板latest版本被未命名版本服务引用详情
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @returns
 */
export const getUnNamedVersionAppsBoundByLatestTemplateVersion = (
  biz_id: string,
  template_space_id: number,
  template_id: number,
  params: ICommonQuery,
) =>
  http
    .get(
      `/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/${template_id}/latest/bound_unnamed_app_details`,
      { params },
    )
    .then((res) => res.data);

/**
 * 获取模板版被已生成版本服务引用详情
 * @param biz_id 业务ID
 * @param template_space_id 空间ID
 * @param template_id 模板ID
 * @param template_revision_id 版本ID
 * @returns
 */
export const getAppsVersionBoundByTemplateVersion = (
  biz_id: string,
  template_space_id: number,
  template_id: number,
  template_revision_id: number,
  params: ICommonQuery,
) =>
  http
    .get(
      `/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/${template_id}/template_revisions/${template_revision_id}/bound_named_app_details`,
      { params },
    )
    .then((res) => res.data);

/**
 * 查询模板套餐和服务的绑定关系
 * @param biz_id 业务ID
 * @param app_id 应用ID
 * @returns
 */
export const getAppPkgBindingRelations = (biz_id: string, app_id: number) =>
  http.get(`/config/biz/${biz_id}/apps/${app_id}/template_bindings`).then((res) => res.data);

/**
 * 批量查询模板的版本名称
 * @param biz_id 业务ID
 * @param template_ids 模板ID列表
 * @returns
 */
export const getTemplateVersionsNameByIds = (biz_id: string, template_ids: number[]) =>
  http
    .post(`/config/biz/${biz_id}/template_revisions/list_names_by_template_ids`, { template_ids })
    .then((res) => res.data);

/**
 * 导入配置文件压缩包
 * @param biz_id 业务ID
 * @param template_space_id 空间id
 * @param fill 导入文件
 * @returns
 */
export const importTemplateFile = (
  biz_id: string,
  template_space_id: number,
  file: any,
  isDecompression: boolean,
  progress: Function,
) =>
  http
    .post(
      `/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/import/${encodeURIComponent(file.name)}`,
      file,
      {
        headers: {
          'X-Bscp-Unzip': isDecompression,
        },
        onUploadProgress: (progressEvent: any) => {
          const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          progress(percentCompleted);
        },
      },
    )
    .then((res) => res.data);

/**
 * 模板批量新增
 * @param biz_id 业务ID
 * @param template_space_id 空间id
 * @param configData 配置列表
 * @returns
 */
export const importTemplateBatchAdd = (biz_id: string, template_space_id: number, configData: any) =>
  http
    .post(`/config/biz/${biz_id}/template_spaces/${template_space_id}/templates/batch_upsert_templates`, {
      items: configData,
    })
    .then((res) => res.data);

/**
 * 批量修改模板权限
 * @param biz_id 业务ID
 * @param template_space_id 空间id
 * @param configData 配置列表
 * @returns
 */
export const batchEditTemplatePermission = (biz_id: string, query: any) =>
  http.post(`/config/biz/${biz_id}/templates/batch_update_templates_permissions`, query);

/**
 * 获取配置模板配置项元信息
 * @param biz_id 业务ID
 * @param template_id 模板id
 * @param revision_name 版本名称
 * @returns
 */
export const getTemplateConfigMeta = (biz_id: string, template_id: number, revision_name?: string) =>
  http.get(`/config/biz/${biz_id}/templates/${template_id}/template_revisions`, { params: { revision_name } });
