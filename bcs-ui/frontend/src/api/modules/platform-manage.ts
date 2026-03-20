// 自动生成的, 请勿手动编辑!!!
import { createRequest } from '../request';

const requestConfig: IRequestConfig = {
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/platformmanager/v1',
};
const request = createRequest(requestConfig);

export const ProjectService = {
  // 获取项目列表
  ListProject: (params?: IProjectListParams, config?: IFetchConfig): Promise<IProjectListResponse extends { data: any } ? IProjectListResponse['data'] : IProjectListResponse> => request('get', '/project')(params, config),
  // 获取项目详情
  GetProject: (params?: { $projectId: string }, config?: IFetchConfig): Promise<IPlatformProject extends { data: any } ? IPlatformProject['data'] : IPlatformProject> => request('get', '/project/$projectId')(params, config),
  // 更新项目
  UpdateProject: (params?: { $projectId: string } & IProjectUpdateParams, config?: IFetchConfig): Promise<boolean extends { data: any } ? boolean['data'] : boolean> => request('put', '/project/$projectId')(params, config),
  // 更新项目管理者
  UpdateProjectManager: (params?: { $projectId: string } & IProjectManagerUpdateParams, config?: IFetchConfig): Promise<IProjectManagerUpdateParams extends { data: any } ? IProjectManagerUpdateParams['data'] : IProjectManagerUpdateParams> => request('put', '/project/$projectId/managers')(params, config),
  // 更新项目业务
  UpdateProjectBusiness: (params?: { $projectId: string } & IProjectBusinessUpdateParams, config?: IFetchConfig): Promise<boolean extends { data: any } ? boolean['data'] : boolean> => request('put', '/project/$projectId/business')(params, config),
  // 更新项目状态
  UpdateProjectStatus: (params?: { $projectId: string } & IProjectStatusUpdateParams, config?: IFetchConfig): Promise<boolean extends { data: any } ? boolean['data'] : boolean> => request('put', '/project/$projectId/isOffline')(params, config),
};

export const BusinessService = {
  // 获取业务列表
  ListBusiness: (params?: {}, config?: IFetchConfig): Promise<IBusinessListResponse extends { data: any } ? IBusinessListResponse['data'] : IBusinessListResponse> => request('post', '/cmdb/business')(params, config),
};
