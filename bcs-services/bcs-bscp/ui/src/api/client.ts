import http from '../request';
import { IClinetCommonQuery, ICreateClientSearchRecordQuery } from '../../types/client';
/**
 * 获取客户端查询列表
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns
 */
export const getClientQueryList = (bizId: string, appId: number, query: IClinetCommonQuery) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/clients`, query);

/**
 * 获取客户端配置拉取记录
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const getClientPullRecord = (bizId: string, appId: number, clientId: number, query: IClinetCommonQuery) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/client_events/${clientId}`, query);

/**
 * 获取客户端搜索记录
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const getClientSearchRecord = (bizId: string, appId: number, params: IClinetCommonQuery) =>
  http.get(`/config/biz/${bizId}/apps/${appId}/client_querys`, { params });

/**
 * 新增客户端搜索记录
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const createClientSearchRecord = (bizId: string, appId: number, query: ICreateClientSearchRecordQuery) =>
  http.post(`config/biz/${bizId}/apps/${appId}/client_querys`, query);

/**
 * 编辑客户端搜索记录
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const updateClientSearchRecord = (
  bizId: string,
  appId: number,
  recordId: number,
  query: ICreateClientSearchRecordQuery,
) => http.put(`/config/biz/${bizId}/apps/${appId}/client_querys/${recordId}`, query);

/**
 * 删除客户端搜索记录
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const deleteClientSearchRecord = (bizId: string, appId: number, recordId: number) =>
  http.delete(`config/biz/${bizId}/apps/${appId}/client_querys/${recordId}`);

/**
 * 获取客户端统计配置版本数据
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const getConfigVersionData = (bizId: string, appId: number, query: any) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/clients/config_version_statistics`, query).then((resp) => resp.data);

/**
 * 获取客户端统计拉取数量趋势数据
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const getClientPullCountData = (bizId: string, appId: number, query: any) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/clients/pull_trend_statistics`, query).then((resp) => resp.data);

/**
 * 获取客户端统计拉取状态
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const getClientPullStatusData = (bizId: string, appId: number, query: any) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/clients/pull_statistics`, query).then((resp) => resp.data);

/**
 * 获取客户端标签数据
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const getClientLabelData = (bizId: string, appId: number, query: any) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/clients/label_statistics`, query).then((resp) => resp.data);

/**
 * 获取客户端附加信息数据
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const getClientAnnotationData = (bizId: string, appId: number, query: any) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/clients/annotation_statistics`, query).then((resp) => resp.data);

/**
 * 获取客户端组件信息数据
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const getClientComponentInfoData = (bizId: string, appId: number, query: any) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/clients/version_statistics`, query).then((resp) => resp.data);

/**
 * 获取客户端标签和注释列表
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const getClientLabelsAndAnnotations = (bizId: string, appId: number, query: any) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/clients/labels_and_annotations`, query);

/**
 * 获取客户端拉取失败详细原因
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns
 */
export const getClientPullFailedReason = (bizId: string, appId: number, query: any) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/clients/specific_failed_reason`, query);
