<template>
  <!-- 集群详情 -->
  <BcsContent
    :title="curCluster.clusterName"
    :desc="`(${curCluster.clusterID})`"
    v-bkloading="{ isLoading, opacity: 1 }">
    <bcs-tab :label-height="42" :active.sync="activeTabName" :validate-active="false" @tab-change="handleTabChange">
      <bcs-tab-panel name="overview" :label="$t('cluster.detail.title.overview')" render-directive="if">
        <Overview :cluster-id="clusterId" />
      </bcs-tab-panel>
      <bcs-tab-panel name="info" :label="$t('generic.title.basicInfo1')" render-directive="if">
        <Info :cluster-id="clusterId" />
      </bcs-tab-panel>
      <bcs-tab-panel name="quota" :label="$t('cluster.detail.title.quota')" render-directive="if" v-if="curCluster.clusterType === 'virtual'">
        <VClusterQuota :cluster-id="clusterId" />
      </bcs-tab-panel>
      <template v-else>
        <bcs-tab-panel name="network" :label="$t('cluster.detail.title.network')" render-directive="if">
          <Network :cluster-id="clusterId" />
        </bcs-tab-panel>
        <bcs-tab-panel name="master" :label="$t('cluster.detail.title.master')" render-directive="if">
          <Master :cluster-id="clusterId" />
        </bcs-tab-panel>
        <bcs-tab-panel name="node" :label="$t('cluster.detail.title.nodeList')" render-directive="if">
          <Node class="pb-[20px]" :cluster-id="clusterId" hide-cluster-select from-cluster />
        </bcs-tab-panel>
        <bcs-tab-panel
          name="autoscaler"
          :label="$t('cluster.detail.title.autoScaler')"
          render-directive="if"
          ref="autoScalerTabRef"
          v-if="showAutoScaler">
          <template #label>
            {{ $t('cluster.detail.title.autoScaler') }}
            <bk-tag theme="danger">NEW</bk-tag>
          </template>
          <AutoScaler :cluster-id="clusterId" />
        </bcs-tab-panel>
      </template>
    </bcs-tab>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, toRefs, watch } from 'vue';
import BcsContent from '@/components/layout/Content.vue';
import Node from '../node-list/node.vue';
import Overview from '@/views/cluster-manage/cluster/overview/overview.vue';
import Info from '@/views/cluster-manage/cluster/info/basic-info.vue';
import VClusterQuota from '@/views/cluster-manage/cluster/info/vcluster-quota.vue';
import Network from '@/views/cluster-manage/cluster/info/network.vue';
import Master from '@/views/cluster-manage/cluster/info/master.vue';
import AutoScaler from './autoscaler/autoscaler.vue';
import $store from '@/store';
import $router from '@/router';
import { useCluster } from '@/composables/use-app';

export default defineComponent({
  components: {
    BcsContent,
    Info,
    Network,
    Master,
    Node,
    Overview,
    AutoScaler,
    VClusterQuota,
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
    scrollToBottom: {
      type: Boolean,
      default: false,
    },
  },
  setup(props) {
    const { active, clusterId, scrollToBottom } = toRefs(props);
    const activeTabName = ref(active.value);
    const isLoading = ref(false);

    const { clusterList } = useCluster();
    const curCluster = computed(() => clusterList.value.find(item => item.clusterID === clusterId.value) || {
      provider: '',
      clusterName: '',
      clusterID: '',
      clusterType: '',
    });
    // 云区域详情
    const cloudDetail = ref<Record<string, any>|null>(null);
    const handleGetCloudDetail = async () => {
      cloudDetail.value = await $store.dispatch('clustermanager/cloudDetail', {
        $cloudId: curCluster.value?.provider,
      });
    };
    const showAutoScaler = computed(() => cloudDetail.value && !cloudDetail.value?.confInfo?.disableNodeGroup);

    // 滚动到底部
    const autoScalerTabRef = ref<any>();
    watch(showAutoScaler, () => {
      if (!scrollToBottom.value || active.value !== 'autoscaler' || !showAutoScaler.value) return;

      setTimeout(() => {
        autoScalerTabRef.value.$el.scrollTop = autoScalerTabRef.value.$el.scrollHeight;
      }, 10);
    }, { immediate: true });
    // tab change事件
    const handleTabChange = (name) => {
      $router.replace({
        name: 'clusterDetail',
        query: {
          active: name,
        },
      });
    };

    onMounted(async () => {
      isLoading.value = true;
      await handleGetCloudDetail();
      isLoading.value = false;
    });

    return {
      activeTabName,
      showAutoScaler,
      isLoading,
      curCluster,
      handleTabChange,
      autoScalerTabRef,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .bk-tab-section {
  padding: 0;
}
>>> .bk-tab-content {
  max-height: calc(100vh - 188px);
  overflow: auto;
}
</style>
