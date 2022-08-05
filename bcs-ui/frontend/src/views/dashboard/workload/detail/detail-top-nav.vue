<template>
  <div class="header-router">
    <span class="icon-wrapper" @click="handleBack">
      <i class="bcs-icon bcs-icon-arrows-left icon-back"></i>
    </span>
    <span class="title-wrapper" v-for="(item, index) in titles" :key="index">
      <span class="title-item" @click="routeHop(item, index)">{{item.name}}</span>
      <span class="separator" v-if="index < (titles.length - 1)">/</span>
    </span>
  </div>
</template>
<script lang="ts">
import { defineComponent, toRefs, computed } from '@vue/composition-api';
import { BCS_CLUSTER } from '@/common/constant';

export default defineComponent({
  name: 'DetailTopNav',
  props: {
    titles: {
      type: Array,
      default: () => [],
    },
    from: {
      type: String,
      default: '',
    },
    clusterId: {
      type: String,
      default: '',
    },
    nodeId: {
      type: String,
      default: '',
    },
    nodeName: {
      type: String,
      default: '',
    },
  },
  setup(props, ctx) {
    const { $router, $store } = ctx.root;
    const { titles, clusterId, from, nodeId, nodeName } = toRefs(props);

    const clusterList = computed(() => $store.state.cluster.clusterList || []);
    const curCluster = computed(() => clusterList.value.find(item => item.clusterID === clusterId.value));

    const handleBack = () => {
      if (from.value === 'nodePods' && titles.value.length === 2) {
        updateViewMode();
        $router.push({
          name: 'clusterNodeOverview',
          params: {
            nodeId: nodeId.value,
            nodeName: nodeName.value,
            clusterId: clusterId.value,
          },
        });
        return false;
      }
      const index = titles.value.length - 2;
      if (!titles.value[index]) return;
      routeHop(titles.value[index], index);
    };

    const updateViewMode = () => {
      localStorage.setItem('FEATURE_CLUSTER', 'done');
      localStorage.setItem(BCS_CLUSTER, curCluster.value.cluster_id);
      sessionStorage.setItem(BCS_CLUSTER, curCluster.value.cluster_id);
      $store.commit('cluster/forceUpdateCurCluster', curCluster.value.cluster_id ? curCluster.value : {});
      $store.commit('updateCurClusterId', curCluster.value.cluster_id);
      $store.commit('updateViewMode', 'cluster');
      $store.dispatch('getFeatureFlag');
    };

    const routeHop = (item, index) => {
      if (index === (props.titles.length - 1)) return;

      ctx.emit('change', item);
    };

    return {
      handleBack,
      routeHop,
    };
  },
});
</script>
<style lang="postcss" scoped>
.header-router .title-wrapper {
    &:last-of-type {
        color: #979BA5;
    }
    &:last-of-type .title-item {
        cursor: default;
    }
}
.header-router {
    display: flex;
    height: 60px;
    align-items: center;
    font-size: 16px;
    padding-left: 12px;
    border-bottom: 1px solid #F3F3F3;
    background: #fff;
    .icon-wrapper {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 32px;
        height: 32px;
        cursor: pointer;
        .icon-back {
            font-size: 16px;
            font-weight: bold;
            color: #3A84FF;
        }
    }
    .title-wrapper {
        display: flex;
        color: #3A84FF;
        .title-item {
            cursor: pointer;
        }
        .separator {
            padding: 0 6px;
            color: #979ba5;
        }
    }
}
</style>
