<template>
  <!-- 集群详情 -->
  <div>
    <ContentHeader
      :title="curCluster.name"
      :desc="`(${curCluster.clusterID})`"
      :hide-back="isSingleCluster"
    ></ContentHeader>
    <div class="biz-content-wrapper">
      <div class="cluster-detail">
        <div class="cluster-detail-tab">
          <div
            v-for="item in tabItems"
            :key="item.id"
            :class="['item', { active: activeId === item.id }]"
            @click="handleChangeActive(item.id)"
          >
            <span class="icon"><i :class="item.icon"></i></span>
            {{item.title}}
          </div>
          <!-- 扩缩容 -->
          <div
            :class="['item', { active: activeId === 'AutoScaler' }]"
            v-if="cloudDetail.confInfo && !cloudDetail.confInfo.disableNodeGroup"
            v-authority="{
              clickable: webAnnotations.perms[clusterId]
                && webAnnotations.perms[clusterId].cluster_manage,
              actionId: 'cluster_manage',
              resourceName: clusterName,
              disablePerms: true,
              permCtx: {
                project_id: projectID,
                cluster_id: clusterId
              }
            }"
            @click="handleChangeActive('AutoScaler')"
          >
            <span class="icon"><i class="bcs-icon bcs-icon-kuosuorong"></i></span>
            {{ $t('弹性扩缩容') }}
          </div>
        </div>
        <div class="cluster-detail-content">
          <component
            :is="activeCom"
            :node-menu="false"
            :cluster-id="clusterId"
            :hide-cluster-select="true"
          ></component>
        </div>
      </div>
    </div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, ref, toRefs } from '@vue/composition-api';
import ContentHeader from '@/components/layout/Header.vue';
import node from './node.vue';
import overview from '@/views/cluster/overview.vue';
import info from '@/views/cluster/info.vue';
import useDefaultClusterId from './use-default-clusterId';
import $i18n from '@/i18n/i18n-setup';
import AutoScaler from './cluster-autoscaler-tencent/autoscaler.vue';
import InternalAutoScaler from './cluster-autoscaler/autoscaler.vue';
import { useCluster, useProject } from '@/common/use-app';

export default defineComponent({
  components: {
    info,
    node,
    overview,
    ContentHeader,
    AutoScaler,
    InternalAutoScaler,
  },
  props: {
    active: {
      type: String,
      default: 'overview',
    },
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props, ctx) {
    const { $store, $router, $INTERNAL } = ctx.root;
    const { active, clusterId } = toRefs(props);

    const { clusterList } = useCluster();
    const { projectID } = useProject();
    const clusterName = computed(() => clusterList.value.find(item => item.clusterID === clusterId.value)?.clusterName);
    const webAnnotations = computed(() => $store.state.cluster.clusterWebAnnotations);
    const activeId = ref(active.value);
    const activeCom = computed(() => {
      if (activeId.value === 'AutoScaler') {
        return $INTERNAL ? InternalAutoScaler : AutoScaler;
      }
      return tabItems.value.find(item => item.id === activeId.value)?.com;
    });
    const curCluster = computed(() => $store.state.cluster.clusterList
      ?.find(item => item.clusterID === clusterId.value) || {});
    const tabItems = ref([
      {
        icon: 'bcs-icon bcs-icon-bar-chart',
        title: $i18n.t('总览'),
        com: 'overview',
        id: 'overview',
      },
      {
        icon: 'bcs-icon bcs-icon-list',
        title: $i18n.t('节点管理'),
        com: 'node',
        id: 'node',
      },
      {
        icon: 'bcs-icon bcs-icon-machine',
        title: $i18n.t('集群信息'),
        com: 'info',
        id: 'info',
      },
    ]);
    const handleChangeActive = (activeID) => {
      if (activeId.value === activeID) return;
      activeId.value = activeID;
      $router.replace({
        name: 'clusterDetail',
        query: {
          active: activeID,
        },
      });
    };
    const { isSingleCluster } = useDefaultClusterId();
    const cloudDetail = ref<any>({});
    const isLoading = ref(false);
    const handleGetCloudDetail = async () => {
      isLoading.value = true;
      cloudDetail.value = await $store.dispatch('clustermanager/cloudDetail', {
        $cloudId: curCluster.value.provider,
      });
      isLoading.value = false;
    };
    handleGetCloudDetail();
    return {
      projectID,
      clusterName,
      webAnnotations,
      cloudDetail,
      isLoading,
      isSingleCluster,
      curCluster,
      tabItems,
      activeId,
      activeCom,
      handleChangeActive,
    };
  },
});
</script>
<style lang="postcss" scoped>
.cluster-detail {
    border: 1px solid #dfe0e5;
    &-tab {
        display: flex;
        height: 60px;
        line-height: 60px;
        border-bottom: 1px solid #dfe0e5;
        font-size: 14px;
        .item {
            display: flex;
            align-items: center;
            justify-content: center;
            min-width: 140px;
            cursor: pointer;
            &.active {
                color: #3a84ff;
                background-color: #fff;
                border-right: 1px solid #dfe0e5;
                border-left: 1px solid #dfe0e5;
                font-weight: 700;
                i {
                    font-weight: 700;
                }
            }
            &:first-child {
                border-left: none;
            }
            .icon {
                font-size: 16px;
                margin-right: 8px;
                width: 16px;
                height: 16px;
                display: flex;
                align-items: center;
            }
        }
    }
    &-content {
        background-color: #fff;
    }
}
</style>
