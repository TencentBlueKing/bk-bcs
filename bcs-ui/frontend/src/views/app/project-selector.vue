<template>
  <div>
    <bcs-select
      class="project-select"
      :clearable="false"
      searchable
      :value="displayName"
      :popover-min-width="320"
      enable-scroll-load
      :scroll-loading="{
        isLoading: scrollLoading
      }"
      :loading="loading"
      :remote-method="remoteSearch"
      allow-create
      ref="selectRef"
      :allow-enter="false"
      @selected="handleProjectChange"
      @scroll-end="handleScrollToBottom">
      <bcs-option
        v-for="option in projectList"
        :key="option.projectCode"
        :id="option.projectCode"
        :name="option.name"
        :disabled="!(perms[option.projectID] && perms[option.projectID].project_view)"
        v-authority="{
          clickable: perms[option.projectID]
            && perms[option.projectID].project_view,
          actionId: 'project_view',
          resourceName: option.name,
          disablePerms: true,
          permCtx: {
            project_id: option.projectID
          }
        }">
        <span
          :class="[
            'flex items-center mx-[-16px] px-[16px]',
            projectCode === option.projectCode ? 'bg-[#f4f6fa] text-[#3a84ff]' : ''
          ]"
          v-bk-tooltips="{
            content: option.businessID && (option.businessID !== '0') && option.kind
              ? `${$t('projects.project.name')}:
              ${option.name}<br/>${$t('projects.project.businessID')}: ${option.businessID}`
              : `${$t('projects.project.name')}: ${option.name}<br/>${$t('bcs.registry.toEnable')}`,
            placement: 'left',
            boundary: 'window',
            delay: [300, 0]
          }">
          <span :class="['bcs-ellipsis', { 'flex-1': !option.businessID }]">{{option.name}}</span>
          <span class="text-[#C4C6CC]" v-if="option.businessID">
            {{`(${option.businessID})`}}
          </span>
          <bcs-tag size="small" v-else>{{ $t('generic.status.notEnable') }}</bcs-tag>
        </span>
      </bcs-option>
      <template #extension>
        <div class="flex items-center">
          <div
            class="text-center flex-1 cursor-pointer"
            @click="handleGotoProjectManage">
            <i class="bcs-icon bcs-icon-apps mr5"></i>
            {{$t('nav.project')}}
          </div>
          <bcs-divider direction="vertical"></bcs-divider>
          <div
            class="text-center flex-1 cursor-pointer"
            @click="handleGotoIAM">
            <i class="bcs-icon bcs-icon-apps mr5"></i>
            {{$t('iam.button.apply2')}}
          </div>
        </div>
      </template>
    </bcs-select>
    <ProjectCreate v-model="showCreateDialog" />
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onBeforeMount, onMounted, ref, watch } from 'vue';

import useProjects, { IProjectPerm } from '../project-manage/project/use-project';

import cancelRequest from '@/common/cancel-request';
import { IProject } from '@/composables/use-app';
import useDebouncedRef from '@/composables/use-debounce';
import $router from '@/router';
import $store from '@/store';
import ProjectCreate from '@/views/project-manage/project/project-create.vue';

export default defineComponent({
  name: 'ProjectSelector',
  components: { ProjectCreate },
  setup() {
    const { getProjectList } = useProjects();
    const showCreateDialog = ref(false);
    const loading = ref(false);
    const params = ref({
      offset: 0,
      limit: 20,
    });

    const projectList = ref<IProject[]>([]);
    const perms = ref<Record<string, IProjectPerm>>({});
    const projectCode = computed(() => $router.currentRoute?.params?.projectCode);
    const displayName = computed(() => $store.state.curProject?.name || $router.currentRoute?.params?.projectCode);
    const projectCodeMap = computed(() => projectList.value.reduce((pre, item) => {
      pre[item.projectCode] = item;
      return pre;
    }, {}));

    // 初始化数据
    const handleInitProjectList = async () => {
      params.value.offset = 0;
      const res = await getProjectList({
        ...params.value,
        searchKey: searchKey.value,
      }).catch(() => ({
        data: {
          results: [],
          total: 0,
        },
        web_annotations: {
          perms: {},
        },
      }));
      projectList.value = res?.data?.results || [];
      perms.value = res?.web_annotations?.perms || {};
    };

    // 滚动加载
    const finished = ref(false);
    const scrollLoading = ref(false);
    const handleScrollToBottom = async () => {
      if (finished.value || scrollLoading.value) return;

      scrollLoading.value = true;
      params.value.offset = projectList.value.length;
      const { data, web_annotations } = await getProjectList({
        ...params.value,
        searchKey: searchKey.value,
      });
      // 过滤重复数据
      const filterData = data.results.filter(item => !projectCodeMap.value[item.projectCode]);
      if (!filterData.length) {
        finished.value = true;
      } else {
        projectList.value.push(...filterData);
        perms.value = Object.assign(perms.value, web_annotations.perms);
      }
      scrollLoading.value = false;
    };

    // 远程搜索
    const selectRef = ref();
    const searchKey = useDebouncedRef('', 600);
    const remoteSearch = (key) => {
      // hack 重置组件内部的loading
      selectRef.value.searchLoading = false;
      searchKey.value = key;
    };
    watch(searchKey, async () => {
      selectRef.value.searchLoading = true;
      await handleInitProjectList();
      selectRef.value.searchLoading = false;
    });

    // 申请项目权限
    const handleGotoIAM = () => {
      window.open(`${window.BK_IAM_HOST}/apply-join-user-group?system_id=bk_bcs_app`);
    };
    // 创建项目
    const handleCreateProject = () => {
      showCreateDialog.value = true;
    };
    // 切换项目
    const handleProjectChange = async (projectCode) => {
      const { currentRoute } = $router;
      if (projectCode === currentRoute.params?.projectCode) return;

      // 更新当前项目缓存
      // const project = projectList.value.find(item => item.projectCode === projectCode);
      // $store.commit('updateCurProject', project);
      // 特殊界面切换项目时跳转到首页
      const name = ['403', '404'].includes(currentRoute.name || '')
        ? 'dashboardIndex' : $store.state.curNav?.route;
      const { href } = $router.resolve({
        name: name || 'dashboardIndex',
        params: {
          projectCode,
        },
        // repo仓库界面不清楚query参数会导致始终是上一个项目的参数
        // query: currentRoute.query,
      });
      await cancelRequest();
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

    onBeforeMount(async () => {
      loading.value = true;
      await handleInitProjectList();
      loading.value = false;
    });

    onMounted(() => {
      // hack 禁用select输入框
      selectRef.value?.$refs?.createInput?.setAttribute('readonly', true);
    });

    return {
      perms,
      projectCode,
      displayName,
      loading,
      selectRef,
      remoteSearch,
      scrollLoading,
      showCreateDialog,
      projectList,
      handleGotoIAM,
      handleCreateProject,
      handleProjectChange,
      handleGotoProjectManage,
      handleScrollToBottom,
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
  >>> .bk-select-name {
    background:#252F43;
    cursor: pointer;
  }
  >>> .bk-select-angle {
    z-index: 2;
  }
}
</style>
