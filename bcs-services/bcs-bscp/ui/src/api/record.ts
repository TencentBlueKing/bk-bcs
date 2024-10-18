import http from '../request';
import { IRecordQuery } from '../../types/record';

/**
 * 获取操作记录列表
 * @param biz_id 空间ID
 * @param params 查询参数
 * @returns
 */
export const getRecordList = (biz_id: string, params: IRecordQuery) =>
  http.get(`/config/biz_id/${biz_id}/audits`, { params }).then((res) => res.data);

/**
 * 审批操作：撤销/驳回/通过/手动上线
 * @param biz_id 空间ID
 * @param app_id 服务ID
 * @param release_id 版本ID
 * @param params 参数
 * @returns
 */
export const approve = (
  biz_id: string,
  app_id: number,
  release_id: number,
  params: { publish_status: string; reason?: string },
) =>
  http
    .post(`/config/biz_id/${biz_id}/app_id/${app_id}/release_id/${release_id}/approve`, { ...params })
    .then((res) => res.data);
