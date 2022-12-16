import {
  fetchProjectList,
  fetchAllProjectList,
  editProject,
  getProject,
} from '@/api/modules/project';
import store from '@/store';

export default function useProjects() {
  async function getProjectList() {
    const result = await fetchProjectList().catch(() => ({ results: [], total: 0 }));
    const projectList = result.results.map(project => ({
      ...project,
      cc_app_id: project.businessID,
      project_id: project.projectID,
      project_name: project.name,
      project_code: project.projectCode,
    }));
    store.commit('forceUpdateOnlineProjectList', projectList);
    return result.results;
  };

  async function getAllProjectList(params: any) {
    const result = await fetchAllProjectList(params,  { needRes: true })
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
        project_id: project.projectID,
        project_name: project.name,
        project_code: project.projectCode,
      })),
      web_annotations: result.webAnnotations,
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
    }
  }

  return {
    fetchProjectInfo,
    getProjectList,
    getAllProjectList,
    updateProject,
  };
}
