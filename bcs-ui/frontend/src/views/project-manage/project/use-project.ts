import {
  fetchProjectList,
  fetchAllProjectList,
  editProject,
  getProject,
  businessList,
  createProject as handleCreateProject,
} from '@/api/modules/project';
import { userInfo } from '@/api/modules/user-manager';
import $store from '@/store';
import { computed } from 'vue';

export default function useProjects() {
  const projectList = computed<any[]>(() => $store.state.projectList);

  // 获取当前有权限项目
  async function getProjectList() {
    const result = await fetchProjectList({}, { cancelWhenRouteChange: false })
      .catch(() => ({ results: [], total: 0 }));
    const projectList = result.results.map(project => ({
      ...project,
      cc_app_id: project.businessID,
      cc_app_name: project.businessName,
      project_id: project.projectID,
      project_name: project.name,
      project_code: project.projectCode,
    }));
    $store.commit('updateProjectList', projectList);
    return result.results;
  };

  // 获取所有项目列表
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

  async function fetchProjectInfo(params: any) {
    const result = await getProject(params).catch(() => {});
    return {
      ...result,
      cc_app_id: result.businessID,
      project_id: result.projectID,
      project_name: result.name,
      project_code: result.projectCode,
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

  async function getUserInfo() {
    const data = await userInfo().catch(() => ({}));
    $store.commit('updateUser', data);
    return data;
  }

  return {
    getUserInfo,
    projectList,
    fetchProjectInfo,
    getProjectList,
    getAllProjectList,
    updateProject,
    getBusinessList,
    createProject,
  };
}
