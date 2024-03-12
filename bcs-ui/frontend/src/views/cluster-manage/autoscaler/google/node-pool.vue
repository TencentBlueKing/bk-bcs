<template>
  <BcsContent>
    <template #header>
      <HeaderNav :list="navList">
        <bcs-steps style="max-width: 360px" :steps="steps" :cur-step="curStep"></bcs-steps>
      </HeaderNav>
    </template>
    <div v-bkloading="{ isLoading }" class="node-pool">
      <keep-alive>
        <component
          :is="stepComMap[curStep]"
          :default-values="defaultValues"
          :schema="schema"
          :cluster="curCluster"
          :save-loading="saveLoading"
          v-if="!isLoading"
          @next="handleNextStep"
          @pre="handlePreStep"
          @confirm="handleConfirm"
        ></component>
      </keep-alive>
    </div>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, toRefs } from 'vue';

import NodeConfig from './node-config.vue';
import NodePoolInfo from './node-pool-info.vue';

import { mergeDeep } from '@/common/util';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import Schema from '@/views/cluster-manage/autoscaler/resolve-schema';
import { useClusterList } from '@/views/cluster-manage/cluster/use-cluster';
import BcsContent from '@/views/cluster-manage/components/bcs-content.vue';
import HeaderNav from '@/views/cluster-manage/components/header-nav.vue';

export default defineComponent({
  components: {
    BcsContent,
    HeaderNav,
    NodePoolInfo,
    NodeConfig,
  },
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
    nodeGroupID: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const { clusterId, nodeGroupID } = toRefs(props);
    const { clusterList } = useClusterList();
    const curCluster = computed(() => ($store.state as any).cluster.clusterList
      ?.find(item => item.clusterID === clusterId.value) || {});
    const navList = computed(() => {
      const nav: any[] = [
        {
          title: clusterList.value.find(item => item.clusterID === clusterId.value)?.clusterName,
          link: {
            name: 'clusterMain',
          },
        },
        {
          title: 'Cluster Autoscaler',
          link: {
            name: 'clusterMain',
            query: {
              active: 'autoscaler',
              clusterId: props.clusterId,
            },
          },
        },
        {
          title: $i18n.t('cluster.ca.button.createNodePool'),
          link: null,
        },
      ];
      return nav;
    });
    const steps = ref([
      {
        title: $i18n.t('cluster.ca.nodePool.title.nodeConfig'),
        icon: 1,
      },
      {
        title: $i18n.t('cluster.ca.nodePool.title.initConfig'),
        icon: 2,
      },
    ]);
    const curStep = ref(1);
    const curStepItem = computed<Record<string, any>>(() => steps.value
      .find((_, index) => index + 1 === curStep.value) || {});
    const stepComMap = {
      1: 'NodeConfig',
      2: 'NodePoolInfo',
    };
    const nodePoolData = ref<Record<string, any>>({});
    const handleNextStep = (data) => {
      nodePoolData.value = mergeDeep(nodePoolData.value, data);
      if (curStep.value + 1 <= steps.value.length) {
        curStep.value = curStep.value + 1;
      }
    };
    const handlePreStep = () => {
      curStep.value = curStep.value - 1;
    };

    const isLoading = ref(true);
    // 获取默认值
    const defaultValues = ref<any>(null);
    const schema = ref({});
    const handleGetSchemaData = async () => {
      const data = await $store.dispatch('clustermanager/resourceSchema', {
        $cloudID: curCluster.value.provider,
        $name: 'nodegroup',
      });
      schema.value = data?.schema || {};
    };

    // 获取详情
    const handleGetNodeGroupDetail = async () => {
      const data = await $store.dispatch('clustermanager/nodeGroupDetail', {
        $nodeGroupID: nodeGroupID.value,
      });
      return data;
    };

    // 创建节点规格
    const user = computed(() => $store.state.user);
    const saveLoading = ref(false);
    const handleConfirm = async () => {
      saveLoading.value = true;
      await handleCreateNodePool();
      saveLoading.value = false;
    };
    const handleCreateNodePool = async () => {
      const data = {
        ...nodePoolData.value,
        clusterID: curCluster.value.clusterID,
        region: curCluster.value.region,
        creator: user.value.username,
      };
      console.log(data);
      const result = await $store.dispatch('clustermanager/createNodeGroup', data);
      if (result) {
        $router.push({
          name: 'clusterMain',
          query: {
            active: 'autoscaler',
            scrollToBottom: true,
            clusterId: props.clusterId,
          },
        });
      }
    };

    onMounted(async () => {
      isLoading.value = true;
      await handleGetSchemaData();
      if (nodeGroupID.value) {
        defaultValues.value = await handleGetNodeGroupDetail();
        defaultValues.value.name = '';
      } else {
        defaultValues.value = Schema.getSchemaDefaultValue(schema.value);
      }
      isLoading.value = false;
    });
    return {
      saveLoading,
      curCluster,
      isLoading,
      schema,
      defaultValues,
      navList,
      steps,
      curStep,
      curStepItem,
      stepComMap,
      handleNextStep,
      handlePreStep,
      handleConfirm,
    };
  },
});
</script>
<style lang="postcss" scoped>
.node-pool {
  margin: -24px;
  height: calc(100vh - 104px);
}
</style>
