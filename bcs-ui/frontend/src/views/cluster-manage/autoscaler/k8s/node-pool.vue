<template>
  <BcsContent :padding="0">
    <template #header>
      <HeaderNav :list="navList">
        <bcs-steps style="max-width: 380px" :steps="steps" :cur-step="curStep"></bcs-steps>
      </HeaderNav>
    </template>
    <div v-bkloading="{ isLoading }" class="h-full">
      <keep-alive>
        <component
          :is="stepComMap[curStep]"
          :default-values="defaultValues"
          :cluster="curCluster"
          :save-loading="saveLoading"
          v-if="!isLoading"
          @next="handleNextStep"
          @pre="handlePreStep"
          @confirm="handleConfirm"
        />
      </keep-alive>
    </div>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, toRefs } from 'vue';

import NodeConfig from './node-config.vue';
import NodePoolInfo from './node-pool-info.vue';

import { mergeDeep } from '@/common/util';
import BcsContent from '@/components/layout/Content.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import { useClusterList } from '@/views/cluster-manage/cluster/use-cluster';
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
              clusterId: clusterId.value,
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
    const defaultValues = ref<any>({
      autoScaling: {
        vpcID: '',
        maxSize: 10,
        minSize: 0,
        scalingMode: 'CLASSIC_SCALING',
        retryPolicy: 'NO_RETRY',
      },
      clusterID: '',
      creator: '',
      enableAutoscale: true,
      launchTemplate: {
        CPU: 4,
        Mem: 8,
        dataDisks: [],
        imageInfo: {},
        internetAccess: {},
        systemDisk: {},
      },
      name: '',
      nodeTemplate: {
        dockerGraphPath: '/data/bcs/service/docker',
        taints: [],
        unSchedulable: 0,
      },
      region: '',
    });

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
        provider: 'bluekingCloud',
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
            clusterId: clusterId.value,
          },
        });
      }
    };

    onMounted(async () => {
      isLoading.value = true;
      if (nodeGroupID.value) {
        defaultValues.value = await handleGetNodeGroupDetail();
        defaultValues.value.name = '';
      }
      isLoading.value = false;
    });
    return {
      saveLoading,
      curCluster,
      isLoading,
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
