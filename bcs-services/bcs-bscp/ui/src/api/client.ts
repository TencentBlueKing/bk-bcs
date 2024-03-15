import http from '../request';
import { IClinetCommonQuery, ICreateClientSearchRecordQuery } from '../../types/client';
/**
 * 获取客户端查询列表
 * @param bizId 业务ID
 * @param appId 应用ID
 * @returns
 */
export const getClientQueryList = (bizId: string, appId: number, query: IClinetCommonQuery) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/client_metrics`, query);

/**
 * 获取客户端配置拉取记录
 * @param bizId 业务ID
 * @param appId 应用ID
 * @param clientId 客户端id
 * @returns
 */
export const getClientPullRecord = (bizId: string, appId: number, clientId: number, query: IClinetCommonQuery) =>
  http.post(`/config/biz/${bizId}/apps/${appId}/client_metrics/${clientId}/events`, query);

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
