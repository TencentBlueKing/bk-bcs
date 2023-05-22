<template>
  <!-- 集群详情 -->
  <BcsContent
    :title="curCluster ? curCluster.clusterName : ''"
    :desc="curCluster ? `(${curCluster.clusterID})` : ''"
    v-bkloading="{ isLoading, opacity: 1 }">
    <bcs-tab :label-height="42" :active.sync="activeTabName" :validate-active="false" @tab-change="handleTabChange">
      <bcs-tab-panel name="overview" :label="$t('集群总览')" render-directive="if">
        <Overview :cluster-id="clusterId" />
      </bcs-tab-panel>
      <bcs-tab-panel name="info" :label="$t('基本信息')" render-directive="if">
        <Info :cluster-id="clusterId" />
      </bcs-tab-panel>
      <bcs-tab-panel name="network" :label="$t('网络配置')" render-directive="if">
        <Network :cluster-id="clusterId" />
      </bcs-tab-panel>
      <bcs-tab-panel name="master" :label="$t('Master配置')" render-directive="if">
        <Master :cluster-id="clusterId" />
      </bcs-tab-panel>
      <bcs-tab-panel name="node" :label="$t('节点列表')" render-directive="if">
        <Node class="pb-[20px]" :cluster-id="clusterId" hide-cluster-select from-cluster />
      </bcs-tab-panel>
      <bcs-tab-panel
        name="autoscaler"
        :label="$t('弹性扩缩容')"
        render-directive="if"
        v-if="showAutoScaler">
        <template #label>
          {{ $t('弹性扩缩容') }}
          <bk-tag theme="danger">NEW</bk-tag>
        </template>
        <InternalAutoScaler :cluster-id="clusterId" v-if="$INTERNAL" />
        <AutoScaler :cluster-id="clusterId" v-else />
      </bcs-tab-panel>
    </bcs-tab>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, toRefs } from 'vue';
import BcsContent from '@/components/layout/Content.vue';
import Node from '../node-list/node.vue';
import Overview from '@/views/cluster-manage/cluster/overview/overview.vue';
import Info from '@/views/cluster-manage/cluster/info/basic-info.vue';
import Network from '@/views/cluster-manage/cluster/info/network.vue';
import Master from '@/views/cluster-manage/cluster/info/master.vue';
import AutoScaler from './autoscaler/tencent/autoscaler.vue';
import InternalAutoScaler from './autoscaler/internal/autoscaler.vue';
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
  setup(props) {
    const { active, clusterId } = toRefs(props);
    const activeTabName = ref(active.value);
    const isLoading = ref(false);

    const { clusterList } = useCluster();
    const curCluster = computed(() => clusterList.value.find(item => item.clusterID === clusterId.value));
    // 云区域详情
    const cloudDetail = ref<Record<string, any>|null>(null);
    const handleGetCloudDetail = async () => {
      cloudDetail.value = await $store.dispatch('clustermanager/cloudDetail', {
        $cloudId: curCluster.value?.provider,
      });
    };
    const showAutoScaler = computed(() => cloudDetail.value && !cloudDetail.value?.confInfo?.disableNodeGroup);
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
