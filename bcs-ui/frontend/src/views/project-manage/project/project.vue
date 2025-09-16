<template>
  <div class="project-manage" v-bkloading="{ isLoading }">
    <div class="title mb20">
      {{$t('nav.project')}}
    </div>
    <div class="operate mb15">
      <bk-button
        class="create-btn" theme="primary" icon="plus"
        v-authority="{
          actionId: 'project_create',
          permCtx: {}
        }"
        @click="handleCreateProject">{{$t('projects.project.create')}}</bk-button>
      <bk-input
        class="search-input"
        clearable
        :placeholder="$t('projects.project.search')"
        :right-icon="'bk-icon icon-search'"
        maxlength="64"
        v-model="keyword">
      </bk-input>
    </div>
    <bk-table
      :data="projectList"
      :pagination="pagination"
      size="medium"
      @page-change="handlePageChange"
      @page-limit-change="handleLimitChange">
      <bk-table-column :label="$t('projects.project.name')" prop="project_name">
        <template #default="{ row }">
          <div class="row-name">
            <span class="row-name-left">{{row.project_name[0]}}</span>
            <div class="row-name-right">
              <bk-button
                theme="primary"
                text
                v-authority="{
                  clickable: webAnnotations.perms[row.project_id]
                    && webAnnotations.perms[row.project_id].project_view,
                  actionId: 'project_view',
                  resourceName: row.project_name,
                  disablePerms: true,
                  permCtx: {
                    project_id: row.project_id
                  }
                }"
                @click="handleGotoProject(row)">
                <span class="bcs-ellipsis">{{row.project_name}}</span>
              </bk-button>
              <span class="time">{{ row.updated_at }}</span>
            </div>
          </div>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('projects.project.engName')" prop="projectCode"></bk-table-column>
      <bk-table-column :label="$t('projects.project.intro')" prop="description">
        <template #default="{ row }">
          {{ row.description || '--' }}
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('generic.label.createdBy1')" prop="creator">
        <template #default="{ row }">
          <bk-user-display-name :user-id="row.creator"></bk-user-display-name>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('generic.label.action')" width="120">
        <template #default="{ row }">
          <bk-button
            class="mr10"
            theme="primary"
            text
            v-authority="{
              clickable: webAnnotations.perms[row.project_id]
                && webAnnotations.perms[row.project_id].project_edit,
              actionId: 'project_edit',
              resourceName: row.project_name,
              disablePerms: true,
              permCtx: {
                project_id: row.project_id
              }
            }"
            @click="handleEditProject(row)">{{$t('projects.project.edit')}}</bk-button>
          <!-- <bk-button theme="primary" text>{{$t('projects.project.bkMonitor')}}</bk-button> -->
        </template>
      </bk-table-column>
    </bk-table>
    <ProjectCreate
      v-model="showCreateDialog"
      :project-data="curProjectData"
      @finished="handleUpdateProjectList" />
  </div>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref, watch } from 'vue';

import ProjectCreate from './project-create.vue';
import useProjects from './use-project';

import { bus } from '@/common/bus';
import { IProject, useAppData } from '@/composables/use-app';
import useDebouncedRef from '@/composables/use-debounce';
import $router from '@/router';

export default defineComponent({
  name: 'ProjectManagement',
  components: {
    ProjectCreate,
  },
  setup: () => {
    // 特性开关
    const { flagsMap } = useAppData();

    const { getProjectList } = useProjects();
    const pagination = ref({
      current: 1,
      count: 0,
      limit: 20,
    });
    const projectList = ref<IProject[]>([]);
    const keyword = useDebouncedRef<string>('', 300);
    const showCreateDialog = ref(false);
    const curProjectData = ref<IProject>();
    const handleGotoProject = (row) => {
      $router.push({
        name: 'clusterMain',
        params: {
          projectCode: row.project_code,
          projectId: row.project_id,
        },
      });
    };
    watch(keyword, () => {
      pagination.value.current = 1;
      handleGetProjectList();
    });

    const handleEditProject = (row) => {
      curProjectData.value = row;
      showCreateDialog.value = true;
    };
    const handleCreateProject = () => {
      curProjectData.value = undefined;
      showCreateDialog.value = true;
    };
    const handlePageChange = (page) => {
      pagination.value.current = page;
      handleGetProjectList();
    };
    const handleLimitChange = (limit) => {
      pagination.value.limit = limit;
      pagination.value.current = 1;
      handleGetProjectList();
    };
    const isLoading = ref(false);
    const webAnnotations = ref({ perms: {} });
    const handleGetProjectList = async () => {
      isLoading.value = true;
      const { data, web_annotations: webPerms } = await getProjectList({
        searchKey: keyword.value,
        offset: pagination.value.current - 1,
        limit: pagination.value.limit,
      });
      projectList.value = data.results;
      webAnnotations.value = webPerms;
      pagination.value.count = data.total;
      isLoading.value = false;
    };

    // 更新项目列表和顶部项目选择器
    function handleUpdateProjectList() {
      handleGetProjectList();
      bus.$emit('refresh-project-list');
    }

    onMounted(() => {
      handleGetProjectList();
    });
    return {
      projectList,
      isLoading,
      webAnnotations,
      pagination,
      keyword,
      showCreateDialog,
      curProjectData,
      flagsMap,
      handleGotoProject,
      handleEditProject,
      handleCreateProject,
      handlePageChange,
      handleLimitChange,
      handleGetProjectList,
      handleUpdateProjectList,
    };
  },
});
</script>
<style lang="postcss" scoped>
.project-manage {
  padding: 20px 60px 20px 60px;
  background: #f5f7fa;
  width: 100%;
  max-height: calc(100vh - 52px);
  overflow: auto;
  .title {
      font-size: 16px;
      color: #313238;
  }
  .operate {
      display: flex;
      align-content: center;
      justify-content: space-between;
      .create-btn {
          min-width: 120px;
      }
      .search-input {
          width: 500px;
      }
  }
  .row-name {
      display: flex;
      align-items: center;
      &-left {
          display: inline-block;
          position: relative;
          margin-right: 10px;
          width: 32px;
          height: 32px;
          line-height: 30px;
          border-radius: 16px;
          text-align: center;
          color: #fff;
          font-size: 16px;
          background-color: rgb(227, 213, 194);
      }
      &-right {
          display: flex;
          flex-direction: column;
          align-items: flex-start;
          flex: 1;
      }
  }
}
</style>
