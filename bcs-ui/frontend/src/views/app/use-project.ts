import {
  fetchProjectList,
  fetchAllProjectList,
  editProject,
} from '@/api/modules/project';
import store from '@/store';

export default function useProjects() {
  async function getProjectList() {
    const result = await fetchProjectList().catch(() => []);
    const projectList = result.results.map(project => ({
      ...project,
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
      data: result.data.results,
      web_annotations: result.webAnnotations,
    };
  };

  async function updateProject(params: any) {
    const result = editProject(params).then(() => true)
      .catch(() => false);
    return result;
  }

  return {
    getProjectList,
    getAllProjectList,
    updateProject,
  };
}
