<template>
  <div>
    <div class="biz-side-title cluster-selector">
      <!-- 全部集群 -->
      <template v-if="!curCluster">
        <span class="icon">{{ clusterType }}</span>
        <span class="cluster-name-all">
          {{ $t('全部集群')}}
        </span>
      </template>
      <!-- 单集群 -->
      <template v-else-if="curCluster.cluster_id && curCluster.name">
        <span :class="['icon', { shared: curCluster.is_shared }]">{{ clusterType }}</span>
        <span>
          <span class="cluster-name" :title="curCluster.name">{{ curCluster.name }}</span>
          <br>
          <span class="cluster-id">{{ curCluster.cluster_id }}</span>
        </span>
      </template>
      <!-- 异常情况 -->
      <template v-else>
        <img src="@/images/bcs2.svg" class="all-icon">
        <span class="cluster-name-all">{{$t('容器服务')}}</span>
      </template>
      <!-- 单集群切换 -->
      <i class="biz-conf-btn bcs-icon bcs-icon-qiehuan f12" @click.stop="handleShowClusterSelector"></i>
      <img v-if="featureCluster" class="dot" src="@/images/new.svg" />
      <cluster-selector v-model="isShowClusterSelector" @change="handleChangeCluster" />
    </div>
    <!-- 视图切换 -->
    <div class="resouce-toggle" v-if="curCluster">
      <span
        v-for="item in viewList"
        :key="item.id"
        :class="['tab bcs-ellipsis', { active: viewMode === item.id }]"
        @click="handleChangeView(item)">
        {{item.name}}
      </span>
    </div>
    <!-- 菜单 -->
    <div class="side-nav">
      <SideMenu :list="menuList" :selected="selected" @change="handleMenuChange"></SideMenu>
      <div class="bcs-footer" v-if="$INTERNAL">
        <div class="mb5 link">
          <a href="wxwork://message?uin=8444252571319680">{{ $t('联系BK助手') }}</a> |
          <a :href="paasHost" target="_blank">{{ $t('蓝鲸桌面') }}</a>
        </div>
        <p>
          Copyright © 2012-{{(new Date()).getFullYear()}} Tencent BlueKing. All Rights Reserved
        </p>
      </div>
      <div class="bcs-footer" v-else>
        <div class="mb5 link">
          <a href="https://wpa1.qq.com/KziXGWJs?_type=wpa&qidian=true" target="_blank">{{ $t('技术支持') }}</a> |
          <a href="https://bk.tencent.com/s-mart/community/" target="_blank">{{ $t('社区论坛') }}</a> |
          <a href="https://bk.tencent.com/index/" target="_blank">{{ $t('产品官网') }}</a>
        </div>
        <p>Copyright © 2012-{{(new Date()).getFullYear()}} Tencent BlueKing. All Rights Reserved. V1.28.0</p>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, ref, watch } from '@vue/composition-api';
import SideMenu from '@/views/app/menu.vue';
import clusterSelector from '@/components/cluster-selector/index.vue';
import menuConfig, { IMenuItem, ISpecialMenuItem } from '@/store/menu';
import { BCS_CLUSTER } from '@/common/constant';
import useGoHome from '@/common/use-gohome';
import { useConfig } from '@/common/use-app';

export default defineComponent({
  name: 'SideNav',
  components: {
    SideMenu,
    clusterSelector,
  },
  setup(props, ctx) {
    const { $store, $i18n, $router } = ctx.root;
    const { $INTERNAL } = useConfig();
    const featureCluster = ref(!localStorage.getItem('FEATURE_CLUSTER'));
    const curCluster = computed(() => {
      const cluster = $store.state.cluster.curCluster;
      return cluster && Object.keys(cluster).length ? cluster : null;
    });
    const paasHost = computed(() => window.PAAS_HOST);
    const clusterType = computed(() =>
    // eslint-disable-next-line camelcase
      (curCluster.value?.is_shared ? $i18n.t('共享') : $i18n.t('专用')));

    const isShowClusterSelector = ref(false);
    const handleShowClusterSelector = () => {
      isShowClusterSelector.value = true;
    };
    const { goHome } = useGoHome();
    // 切换单集群
    const handleChangeCluster = async (cluster) => {
      localStorage.setItem('FEATURE_CLUSTER', 'done');
      localStorage.setItem(BCS_CLUSTER, cluster.cluster_id);
      sessionStorage.setItem(BCS_CLUSTER, cluster.cluster_id);
      $store.commit('cluster/forceUpdateCurCluster', cluster.cluster_id ? cluster : {});
      $store.commit('updateCurClusterId', cluster.cluster_id);
      $store.dispatch('getFeatureFlag');
      goHome(ctx.root.$route);
    };

    // 视图类型
    const viewMode = computed<'dashboard' | 'cluster'>(() => $store.state.viewMode);
    const viewList = ref([
      {
        id: 'cluster',
        name: $i18n.t('集群管理'),
      },
      {
        id: 'dashboard',
        name: $i18n.t('资源视图'),
      },
    ]);
    // 视图切换
    const handleChangeView = async (item) => {
      if (viewMode.value === item.id) return;

      $store.commit('updateViewMode', item.id);
      $store.dispatch('getFeatureFlag');
      goHome(ctx.root.$route);
    };

    // 菜单列表
    watch(viewMode, () => {
      if (viewMode.value === 'dashboard') {
        $store.commit('updateMenuList', menuConfig.dashboardMenuList);
      } else {
        $store.commit('updateMenuList', menuConfig.k8sMenuList);
      }
    }, { immediate: true });

    const featureFlag = computed(() => $store.getters.featureFlag || {});
    // 数组去重
    const removeDuplicates = (data: (IMenuItem | ISpecialMenuItem)[]) => {
      let slow = 0;
      // eslint-disable-next-line @typescript-eslint/prefer-for-of
      for (let fast = 0; fast < data.length; fast++) {
        if (data[slow].id !== data[fast].id) {
          // eslint-disable-next-line no-plusplus
          data[++slow] = data[fast];
        }
      }
      return data.slice(0, slow + 1);
    };
    const menuConfigList = computed<IMenuItem[]>(() => JSON.parse(JSON.stringify($store.state.menuList)));
    const menuList = computed(() => {
      const data = menuConfigList.value.reduce<(IMenuItem | ISpecialMenuItem)[]>((pre, item) => {
        if (item.id && featureFlag.value[item.id]) {
          // todo 特殊处理公共集群下自定义资源的二级菜单
          // eslint-disable-next-line camelcase
          if (item.id === 'CUSTOM_RESOURCE' && curCluster.value?.is_shared) {
            item.children = item.children?.filter(child => !['dashboardCRD', 'dashboardCustomObjects'].includes(child.id));
          }
          // eslint-disable-next-line camelcase
          if (item.id === 'WORKLOAD' && curCluster.value?.is_shared) {
            item.children = item.children?.filter(child => child.id !== 'dashboardWorkloadDaemonSets');
          }
          pre.push(item);
        } else if (!item.id) {
          pre.push(item);
        }
        return pre;
      }, []);
      return removeDuplicates(data);
    });

    const selected = computed(() => {
      // 当前选择菜单在全局导航守卫中设置的
      // eslint-disable-next-line camelcase
      if ($store.state.curMenuId === 'CLUSTER' && curCluster.value?.cluster_id) {
        // 特殊：单集群时预览界面是归属于概览菜单下的
        return 'OVERVIEW';
      }
      return $store.state.curMenuId;
    });
    const projectCode = computed(() => $store.state.curProjectCode);
    const projectId = computed(() => $store.state.curProjectId);
    const curProject = computed(() => $store.state.curProject);
    // 菜单切换
    const handleMenuChange = (item: IMenuItem) => {
      // 直接取$route会存在缓存，需要重新从root上获取最新路由信息
      if (ctx.root.$route.name === item.routeName) return;

      if (item.id === 'MONITOR') {
        if ($INTERNAL.value) {
          window.open(`${window.DEVOPS_HOST}/console/monitor/${projectCode.value}/?project_id=${projectId.value}`);
        } else {
          window.open(`${window.BKMONITOR_HOST}/?bizId=${curProject.value.cc_app_id}#/k8s`);
        }
      } else {
        $router.push({
          name: item.routeName,
          params: {
            // eslint-disable-next-line camelcase
            clusterId: curCluster.value?.cluster_id,
          },
        });
      }
    };

    return {
      featureCluster,
      curCluster,
      isShowClusterSelector,
      viewMode,
      viewList,
      menuList,
      selected,
      clusterType,
      paasHost,
      handleChangeCluster,
      handleShowClusterSelector,
      handleChangeView,
      handleMenuChange,
    };
  },
});
</script>

<style scoped lang="postcss">
    .biz-side-title {
        position: relative;
    }
    .cluster-selector {
        background: #fafbfd;
    }
    .resouce-toggle {
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 10px 0;
        .tab {
            display: flex;
            align-items: center;
            justify-content: center;
            background: #f7f8f9;
            border: 1px solid #dde4eb;
            margin-left: -1px;
            font-size: 12px;
            height: 24px;
            padding: 0 26px;
            cursor: pointer;
            white-space: nowrap;
            &.active {
                background: #fff;
                color: #3a84ff;
            }
            &.disabled {
                cursor: not-allowed;
            }
            &:first-child {
                border-radius: 3px 0 0 3px;
            }
            &:last-child {
                border-radius: 0 3px 3px 0;
            }
        }
    }
    .biz-conf-btn {
        position: absolute;
        right: 10px;
        top: 16px;
        font-size: 12px;
        cursor: pointer;
        width: 30px;
        height: 30px;
        text-align: center;
        line-height: 30px;
        z-index: 100;
    }
    .cluster-name {
        max-width: 150px;
        overflow: hidden;
        text-overflow: ellipsis;
        display: inline-block;
        white-space: nowrap;
        margin-top: 2px;
    }
    .cluster-name-all {
        font-size: 16px;
    }
    .dot {
        position: absolute;
        display: inline-block;
        width: 16px;
        height: 16px;
        top: 16px;
        right: 4px;
        z-index: 1;
        padding: 2px;
    }
    .side-nav {
        flex: 1;
        overflow: auto;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
    }
    .bcs-footer {
      font-size: 12px;
      color: #b7c0ca;
      width: 100%;
      text-align: center;
      line-height: 20px;
      padding: 25px 15px;
      .link a {
        color: #3a84ff;
      }
    }
</style>
