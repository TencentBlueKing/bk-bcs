<template>
  <div class="biz-container app-container flex-col" v-bkloading="{ isLoading: loading }">
    <!-- 集群列表和项目信息加载完后再渲染视图 -->
    <template v-if="!loading">
      <template v-if="curProject.kind">
        <!-- 页面公共导航 -->
        <ContentHeader
          :title="routeMeta.title"
          :hide-back="routeMeta.hideBack"
          v-if="routeMeta.title" />
        <RouterView :key="$route.path" />
        <!-- 终端 -->
        <Terminal />
      </template>
      <!-- 未注册容器服务 -->
      <Unregistry :cur-project="curProject" v-else />
    </template>
  </div>
</template>
<script lang="ts">
/* eslint-disable camelcase */
import { defineComponent, toRef, reactive, computed, onMounted, ref, onErrorCaptured } from 'vue';
import $router from '@/router';
import Terminal from '@/views/app/terminal.vue';
import Unregistry from '@/views/app/unregistry.vue';
import ContentHeader from '@/components/layout/Header.vue';
import useProjects from './project-manage/project/use-project';
import $store from '@/store';
import $bkMessage from '@/common/bkmagic';

export default defineComponent({
  name: 'AppViews',
  components: {
    Terminal,
    Unregistry,
    ContentHeader,
  },
  setup() {
    const { projectList, getAllProjectList } = useProjects();
    const loading = ref(true);// 默认不加载视图，等待集群接口加载完
    const currentRoute = computed(() => toRef(reactive($router), 'currentRoute').value);
    const routeMeta = computed(() => currentRoute.value?.meta || {});
    const curProject = computed(() => $store.state.curProject);

    // 校验项目
    const validateProjectCode = async () => {
      const projectCode = currentRoute.value.params?.projectCode;
      const project = projectList.value.find(item => item.projectCode === projectCode);
      if (!project) {
        // projectCode不在当前项目列表中时判断当前项目是否有权限
        const { data } = await getAllProjectList({
          projectCode,
          all: true,
        }, { cancelWhenRouteChange: false });
        loading.value = false;

        data.length
          ? $router.replace({
            name: '403', // 无权限
            query: {
              actionId: 'project_view',
              resourceName: data[0]?.project_name,
              permCtx: JSON.stringify({
                project_id: data[0]?.project_id,
              }),
              fromRoute: window.location.href,
            },
          })
          : $router.replace({ name: '404' });// 错误项目
      } else {
        // 缓存当前项目信息
        $store.commit('updateCurProject', project);
        // 设置路由projectId和projectCode信息（旧模块很多地方用到），后续路由切换时也会在全局导航钩子上注入这个两个参数
        currentRoute.value.params.projectId = project.project_id;
        currentRoute.value.params.projectCode = project.project_code;
      }
      return !!project;
    };

    onMounted(async () => {
      if (!currentRoute.value.params?.projectCode) {
        // 路由中不存在项目Code, 设置projectCode并跳转
        const route = $router.resolve({
          name: 'clusterMain',
          params: {
            projectCode: $store.getters.curProjectCode || projectList.value[0]?.projectCode,
          },
        });
        window.location.href = route.href;
        return;
      }
      // 校验项目Code是否有权限和正确
      if (!await validateProjectCode()) return;

      loading.value = true;
      const list = [
        $store.dispatch('getProject', { projectId: curProject.value?.project_id }),
        $store.dispatch('cluster/getClusterList', curProject.value?.project_id),
      ];
      if (curProject.value?.kind) {
        list.push($store.dispatch('cluster/getBizMaintainers'));
      }
      const data = await Promise.all(list).catch((err) => {
        console.error(err);
      });
      // 校验集群ID是否正确
      const clusterList = data[1]?.data || [];
      const cluster = clusterList.find(item => item.clusterID ===  $store.getters.curClusterId);
      $store.commit('updateCurCluster', cluster || clusterList[0]);
      loading.value = false;
    });

    onErrorCaptured((err: Error, vm, info: string) => {
      process.env.NODE_ENV === 'development' && $bkMessage({
        theme: 'warning',
        message: `Something is wrong with the component ${vm.$options.name} ${info}`,
      });
      return true;
    });

    return {
      loading,
      currentRoute,
      routeMeta,
      curProject,
    };
  },
});
</script>
