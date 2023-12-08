import {
  businessList,
  createProject as handleCreateProject,
  editProject,
  fetchAllProjectList,
  fetchProjectList,
  getProject,
} from '@/api/modules/project';
import { IProject } from '@/composables/use-app';
import $store from '@/store';

export interface IProjectPerm {
  project_create: boolean
  project_delete: boolean
  project_edit: boolean
  project_view: boolean
}
export interface IProjectListData {
  data: {
    results: IProject[]
    total: number
  }
  web_annotations: {
    perms: Record<string, IProjectPerm>
  }
}

export default function useProjects() {
  // 获取项目列表
  async function getProjectList(params = {}) {
    const data = await fetchProjectList({
      all: true,
      ...params,
    }, { cancelWhenRouteChange: false, needRes: true })
      .catch(() => ({ results: [], total: 0 }));
    data.data.results = data.data.results.map(project => ({
      ...project,
      cc_app_id: project.businessID,
      cc_app_name: project.businessName,
      project_id: project.projectID,
      project_name: project.name,
      project_code: project.projectCode,
    }));
    return data as IProjectListData;
  };

  // 获取所有项目列表(项目列表页面)
  async function getAllProjectList(params: any = {}, config = {}) {
    const result = await fetchAllProjectList(params,  { needRes: true, ...config })
      .catch(() => ({
        data: {},
        webAnnotations: {
          perms: {},
        },
      }));
    return {
      data: result.data.results.map(project => ({
        ...project,
        cc_app_id: project.businessID,
        cc_app_name: project.businessName,
        project_id: project.projectID,
        project_name: project.name,
        project_code: project.projectCode,
      })),
      total: result.data.total,
      web_annotations: result.web_annotations,
    };
  };

  async function updateProject(params: any) {
    const result = await editProject(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function fetchProjectInfo(params: { $projectId: string }) {
    const { data, web_annotations, code, message } = await getProject(params, {
      needRes: true,
      globalError: false,
      cancelWhenRouteChange: false,
    }).catch(() => ({}));
    if (!data) return {
      code,
      data: data as IProject,
      web_annotations,
    };
    // 兼容历史数据
    const bcsProjectData = {
      ...data,
      cc_app_id: data.businessID,
      project_id: data.projectID,
      project_name: data.name,
      project_code: data.projectCode,
    };
    $store.commit('updateCurProject', bcsProjectData);

    return {
      code,
      data: bcsProjectData as IProject,
      web_annotations,
      message,
    };
  }

  async function getBusinessList() {
    return await businessList().catch(() => []);
  }

  async function createProject(params: any) {
    const result = handleCreateProject(params).then(() => true)
      .catch(() => false);
    return result;
  }

  return {
    fetchProjectInfo,
    getProjectList,
    getAllProjectList,
    updateProject,
    getBusinessList,
    createProject,
  };
}
