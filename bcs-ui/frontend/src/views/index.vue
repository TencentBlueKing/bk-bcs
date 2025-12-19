<template>
  <div class="flex flex-col h-full overflow-hidden" v-bkloading="{ isLoading: loading }">
    <!-- 集群列表和项目信息加载完后再渲染视图 -->
    <template v-if="!loading">
      <template v-if="curProject.kind && curProject.businessID && (curProject.businessID !== '0')">
        <!-- 页面公共导航 -->
        <ContentHeader
          :title="routeMeta.title"
          :hide-back="routeMeta.hideBack"
          :cluster-id="routeMeta.showClusterName ? $route.params.clusterId : ''"
          v-if="routeMeta.title" />
        <KeepAlive>
          <RouterView class="flex-1" v-if="$route.meta?.keepAlive" />
        </KeepAlive>
        <!-- key为了解决旧版模板集刷新问题, 资源视图视图管理切换集群后不能刷新界面(不能用path作为Key) -->
        <RouterView class="flex-1" :key="routerViewKey" v-if="!$route.meta?.keepAlive" />
        <!-- 终端 -->
        <Terminal />
      </template>
      <template v-else>
        <!-- 未注册容器服务 -->
        <Unregistry :cur-project="curProject" v-if="!hasNoAuthorizedProject" />
        <!-- 空项目引导 -->
        <ProjectGuide v-else-if="hasNoAuthorizedProject" />
      </template>
    </template>
  </div>
</template>
<script lang="ts">
/* eslint-disable camelcase */
import cookie from 'cookie';
import { computed, defineComponent, onErrorCaptured, onMounted, reactive, ref, toRef } from 'vue';

import useProjects from './project-manage/project/use-project';

import $bkMessage from '@/common/bkmagic';
import { setCookie } from '@/common/util';
import ContentHeader from '@/components/layout/Header.vue';
import { IProject } from '@/composables/use-app';
import $router from '@/router';
import $store from '@/store';
import ProjectGuide from '@/views/app/empty-project-guide.vue';
import Terminal from '@/views/app/terminal.vue';
import Unregistry from '@/views/app/unregistry.vue';

export default defineComponent({
  name: 'AppViews',
  components: {
    Terminal,
    Unregistry,
    ContentHeader,
    ProjectGuide,
  },
  setup() {
    const { fetchProjectInfo, getProjectList } = useProjects();
    const loading = ref(true);// 默认不加载视图，等待集群接口加载完
    const currentRoute = computed(() => toRef(reactive($router), 'currentRoute').value);
    const isDashboard = computed(() => currentRoute.value.matched.some(item => item.name === 'dashboardIndex'));
    const routeMeta = computed(() => currentRoute.value?.meta || {});
    const curProject = computed(() => $store.state.curProject);
    const hasNoAuthorizedProject = ref(false);

    const routerViewKey = computed(() => {
      if (routeMeta.value?.keepAlive) return routeMeta.value?.keepAlive;
      return isDashboard.value ? 'dashboard' : currentRoute.value.path;
    });

    // 设置项目缓存
    const handleSetProjectStorage = (data: IProject) => {
      // 缓存当前项目信息
      $store.commit('updateCurProject', data);
      // 设置路由projectId和projectCode信息（旧模块很多地方用到），后续路由切换时也会在全局导航钩子上注入这个两个参数
      currentRoute.value.params.projectId = data.projectID;
      currentRoute.value.params.projectCode = data.projectCode;
      // 获取上一次的项目Code
      const curCookieProjectCode = cookie.parse(document.cookie)?.['X-BCS-Project-Code'];
      // 设置当前项目Code
      setCookie('X-BCS-Project-Code', data.projectCode, window.BK_DOMAIN);
      // 判断cookie和当前项目code是否一致，不一致刷新当前界面
      if (
        cookie.parse(document.cookie)?.['X-BCS-Project-Code'] === data.projectCode // 判断是否设置cookie成功，防止无限刷新
        && curCookieProjectCode
        && curCookieProjectCode !== data.projectCode
      ) {
        window.location.reload();
      }
    };
    // 校验项目Code
    const validateProjectCode = async () => {
      const projectInfo = await getProjectList();
      if (!projectInfo?.data?.results?.length) {
        $store.commit('updateCurProject', {});
        hasNoAuthorizedProject.value = true;
        if (currentRoute.value.path !== '/') {
          $router.replace({
            path: '/',
          });
        }
        return false;
      }
      const projectCode = currentRoute.value.params?.projectCode;
      // 路由中不存在项目Code, 重新设置projectCode
      if (!projectCode) {
        const { data, web_annotations } = projectInfo;
        const authorizedProject = data?.results?.find(item => web_annotations?.perms[item.projectID]?.project_view);

        if (authorizedProject) {
          // 跳转第一个有权限项目
          // $router.replace({
          //   name: 'clusterMain',
          //   params: {
          //     projectCode: data?.results?.[0]?.projectCode,
          //   },
          // });
          handleSetProjectStorage(authorizedProject);
          return true;
        }
        // 无任何项目权限
        hasNoAuthorizedProject.value = true;
        return false;
      }

      // 路由中存在Code, 校验Code正确性
      const { data, code, web_annotations } = await fetchProjectInfo({
        $projectId: projectCode,
      });

      // 无权限
      if (code === 40403) {
        $router.replace({
          name: '403',
          params: {
            projectCode,
            perms: web_annotations?.perms,
          },
        });
        return false;
      }

      // 项目不存在
      if (code === 40404) {
        $router.replace({ name: '404' });
        return false;
      }

      // 未知异常
      if (code !== 0 && !data) {
        return false;
      }

      handleSetProjectStorage(data);
      return true;
    };

    onMounted(async () => {
      // 校验项目Code是否有权限和正确
      if (!await validateProjectCode()) {
        loading.value = false;
        return;
      };

      loading.value = true;
      const list = [
        $store.dispatch('cluster/getClusterList', curProject.value?.project_id).then(({ data }) => {
          // 校验集群ID是否正确
          const clusterList = data || [];
          const cluster = clusterList.find(item => item.clusterID ===  $store.getters.curClusterId);
          $store.commit('updateCurCluster', cluster || clusterList.find(item => item.status === 'RUNNING'));
        }),
      ];
      if (curProject.value?.kind && curProject.value?.businessID && (curProject.value?.businessID !== '0')) {
        list.push($store.dispatch('cluster/getBizMaintainers'));
      }
      await Promise.all(list).catch((err) => {
        console.error(err);
      });
      loading.value = false;
    });

    onErrorCaptured((err: Error, vm, info: string) => {
      process.env.NODE_ENV === 'development' && $bkMessage({
        theme: 'warning',
        message: `Something is wrong with the component ${vm.$options.name} ${info}`,
      });
      console.error(err, vm, info);
      return true;
    });

    return {
      routerViewKey,
      loading,
      isDashboard,
      currentRoute,
      routeMeta,
      curProject,
      hasNoAuthorizedProject,
    };
  },
});
</script>
