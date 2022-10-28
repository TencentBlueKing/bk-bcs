<template>
  <div class="project-manage" v-bkloading="{ isLoading }">
    <div class="title mb20">
      {{$t('项目管理')}}
    </div>
    <div class="operate mb15">
      <bk-button
        class="create-btn" theme="primary" icon="plus"
        v-authority="{
          actionId: 'project_create',
          permCtx: {}
        }"
        @click="handleCreateProject">{{$t('创建项目')}}</bk-button>
      <bk-input
        class="search-input"
        clearable
        :placeholder="$t('输入关键字搜索')"
        :right-icon="'bk-icon icon-search'"
        v-model="keyword">
      </bk-input>
    </div>
    <bk-table
      :data="filterList"
      :pagination="pagination"
      size="medium"
      @page-change="handlePageChange"
      @page-limit-change="handleLimitChange">
      <bk-table-column :label="$t('项目名称')" prop="project_name">
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
                {{row.project_name}}
              </bk-button>
              <span class="time">{{ row.updated_at }}</span>
            </div>
          </div>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('项目英文名')" prop="english_name"></bk-table-column>
      <bk-table-column :label="$t('项目说明')" prop="description">
        <template #default="{ row }">
          {{ row.description || '--' }}
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('创建者')" prop="creator"></bk-table-column>
      <bk-table-column :label="$t('操作')" width="120">
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
            @click="handleEditProject(row)">{{$t('编辑项目')}}</bk-button>
          <!-- <bk-button theme="primary" text>{{$t('申请监控中心')}}</bk-button> -->
        </template>
      </bk-table-column>
    </bk-table>
    <ProjectCreate
      v-model="showCreateDialog"
      :project-data="curProjectData"
      @finished="handleGetProjectList"></ProjectCreate>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref, computed, onMounted } from '@vue/composition-api';
import ProjectCreate from './project-create.vue';
import { getProjectList } from '@/api/base';

export default defineComponent({
  name: 'ProjectManagement',
  components: {
    ProjectCreate,
  },
  setup: (props, ctx) => {
    const { $router } = ctx.root;
    const pagination = ref({
      current: 1,
      count: 0,
      limit: 20,
    });
    const projectList = ref<any[]>([]);
    const filterList = computed(() => projectList.value
      .filter(item => item.project_name.indexOf(keyword.value) > -1)
      .slice(
        pagination.value.limit * (pagination.value.current - 1),
        pagination.value.limit * pagination.value.current,
      ));

    const keyword = ref('');
    const showCreateDialog = ref(false);
    const curProjectData = ref(null);
    const handleGotoProject = (row) => {
      $router.push({
        name: 'clusterMain',
        params: {
          projectCode: row.project_code,
          projectId: row.project_id,
        },
      });
    };
    const handleEditProject = (row) => {
      curProjectData.value = row;
      showCreateDialog.value = true;
    };
    const handleCreateProject = () => {
      curProjectData.value = null;
      showCreateDialog.value = true;
    };
    const handlePageChange = (page) => {
      pagination.value.current = page;
    };
    const handleLimitChange = (limit) => {
      pagination.value.limit = limit;
    };
    const isLoading = ref(false);
    const webAnnotations = ref({ perms: {} });
    const handleGetProjectList = async () => {
      isLoading.value = true;
      const { data = [], web_annotations: webPerms } = await getProjectList({}, { needRes: true })
        .catch(() => ({
          data: [],
          web_annotations: {
            perms: {},
          },
        }));
      projectList.value = data;
      webAnnotations.value = webPerms;
      pagination.value.count = projectList.value.length;
      isLoading.value = false;
    };
    onMounted(() => {
      handleGetProjectList();
    });
    return {
      isLoading,
      webAnnotations,
      pagination,
      keyword,
      filterList,
      showCreateDialog,
      curProjectData,
      handleGotoProject,
      handleEditProject,
      handleCreateProject,
      handlePageChange,
      handleLimitChange,
      handleGetProjectList,
    };
  },
});
</script>
<style lang="postcss" scoped>
.project-manage {
  padding: 20px 60px 20px 60px;
  background: #f5f7fa;
  width: 100%;
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
