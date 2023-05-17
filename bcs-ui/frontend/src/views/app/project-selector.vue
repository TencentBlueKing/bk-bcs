<template>
  <div>
    <bcs-select
      class="project-select"
      :clearable="false"
      searchable
      :value="$store.getters.curProjectCode"
      :popover-min-width="320"
      @change="handleProjectChange">
      <bcs-option
        v-for="option in projectList"
        :key="option.projectCode"
        :id="option.projectCode"
        :name="option.name">
        <span
          class="flex items-center"
          v-bk-tooltips="{
            content: option.businessID
              ? `${$t('项目名称')}: ${option.name}<br/>${$t('业务ID')}: ${option.businessID}`
              : `${$t('项目名称')}: ${option.name}<br/>${$t('未启用容器服务')}`,
            placement: 'left',
            boundary: 'window'
          }">
          <span :class="['bcs-ellipsis', { 'flex-1': !option.businessID }]">{{option.name}}</span>
          <span class="text-[#C4C6CC]" v-if="option.businessID">
            {{`(${option.businessID})`}}
          </span>
          <bcs-tag size="small" v-else>{{ $t('未启用') }}</bcs-tag>
        </span>
      </bcs-option>
      <template #extension>
        <div class="flex items-center">
          <template v-if="!$INTERNAL">
            <div
              class="text-center flex-1 cursor-pointer"
              @click="handleCreateProject">
              <i class="bk-icon icon-plus-circle mr5"></i>
              {{$t('新建项目')}}
            </div>
            <bcs-divider direction="vertical"></bcs-divider>
          </template>
          <div
            class="text-center flex-1 cursor-pointer"
            @click="handleGotoProjectManage">
            <i class="bcs-icon bcs-icon-apps mr5"></i>
            {{$t('项目管理')}}
          </div>
          <bcs-divider direction="vertical"></bcs-divider>
          <div
            class="text-center flex-1 cursor-pointer"
            @click="handleGotoIAM">
            <i class="bcs-icon bcs-icon-apps mr5"></i>
            {{$t('申请权限')}}
          </div>
        </div>
      </template>
    </bcs-select>
    <ProjectCreate v-model="showCreateDialog" />
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, ref } from 'vue';
import ProjectCreate from '@/views/project-manage/project/project-create.vue';
import $router from '@/router';
import $store from '@/store';

export default defineComponent({
  name: 'ProjectSelector',
  components: { ProjectCreate },
  setup() {
    const showCreateDialog = ref(false);
    const projectList = computed<any[]>(() => $store.state.projectList);

    // 申请项目权限
    const handleGotoIAM = () => {
      window.open(`${window.BK_IAM_APP_URL}apply-join-user-group?system_id=bk_bcs_app`);
    };
    // 创建项目
    const handleCreateProject = () => {
      showCreateDialog.value = true;
    };
    // 切换项目
    const handleProjectChange = (projectCode) => {
      const { currentRoute } = $router;
      if (projectCode === currentRoute.params?.projectCode) return;

      // 更新当前项目缓存
      // const project = projectList.value.find(item => item.projectCode === projectCode);
      // $store.commit('updateCurProject', project);
      // 特殊界面切换项目时跳转到首页
      const name = ['403', '404'].includes(currentRoute.name || '')
        ? 'dashboardIndex' : $store.state.curSideMenu?.route;
      const { href } = $router.resolve({
        name: name || 'dashboardIndex',
        params: {
          projectCode,
        },
        // repo仓库界面不清楚query参数会导致始终是上一个项目的参数
        // query: currentRoute.query,
      });
      window.location.href = href;
    };
    // 项目管理
    const handleGotoProjectManage = () => {
      if (window.REGION === 'ieod') {
        window.open(`${window.DEVOPS_HOST}/console/pm`);
      } else {
        $router.push({
          name: 'projectManage',
        });
      }
    };

    return {
      showCreateDialog,
      projectList,
      handleGotoIAM,
      handleCreateProject,
      handleProjectChange,
      handleGotoProjectManage,
    };
  },
});
</script>
<style lang="postcss" scoped>
.project-select {
  border:none;
  background:#252F43;
  color:#D3D9E4;
  box-shadow: none;
}
</style>
