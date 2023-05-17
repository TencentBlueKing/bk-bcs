<template>
  <BcsContent>
    <template #header>
      <HeaderNav :list="navList">
        <bcs-steps style="max-width: 360px" :steps="steps" :cur-step="curStep"></bcs-steps>
      </HeaderNav>
    </template>
    <div class="node-pool-content" v-bkloading="{ isLoading }">
      <div class="mb15 title">
        {{curStepItem.title}}
      </div>
      <div class="content-wrapper" v-if="!isLoading">
        <keep-alive>
          <component
            :is="stepComMap[curStep]"
            :default-values="defaultValues"
            :node-pool-info="nodePoolInfo"
            :schema="schema"
            :cluster="curCluster"
            :is-edit="isEdit"
            @next="handleNextStep"
            @pre="handlePreStep"
          ></component>
        </keep-alive>
      </div>
    </div>
  </BcsContent>
</template>
<script lang="ts">
import { defineComponent, ref, computed, onMounted, toRefs } from 'vue';
import BcsContent from '@/views/cluster-manage/components/bcs-content.vue';
import HeaderNav from '@/views/cluster-manage/components/header-nav.vue';
import { useClusterList } from '@/views/cluster-manage/cluster/use-cluster';
import $i18n from '@/i18n/i18n-setup';
import NodePoolInfo from './node-pool-info.vue';
import NodeConfig from './node-config.vue';
import $store from '@/store/index';
import Schema from '@/views/cluster-manage/cluster/autoscaler/resolve-schema';

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
    const isEdit = computed(() => !!nodeGroupID.value);
    const navList = computed(() => {
      const nav: any[] = [
        {
          title: clusterList.value.find(item => item.clusterID === clusterId.value)?.clusterName,
          link: {
            name: 'clusterDetail',
          },
        },
        {
          title: 'Cluster Autoscaler',
          link: {
            name: 'clusterDetail',
            query: {
              active: 'autoscaler',
            },
          },
        },
      ];
      if (isEdit.value) {
        nav.push(...[
          {
            title: `${defaultValues.value?.nodeGroupID} (${defaultValues.value?.name}) `,
            link: {
              name: 'nodePoolDetail',
              params: {
                clusterId: props.clusterId,
                nodeGroupID: props.nodeGroupID,
              },
            },
          },
          {
            title: isEdit.value ? $i18n.t('编辑节点池') : $i18n.t('新建节点池'),
            link: null,
          },
        ]);
      } else {
        nav.push({
          title: $i18n.t('新建节点池'),
          link: null,
        });
      }
      return nav;
    });
    const steps = ref([
      {
        title: $i18n.t('节点池信息'),
        icon: 1,
      },
      {
        title: $i18n.t('节点配置'),
        icon: 2,
      },
    ]);
    const curStep = ref(1);
    const curStepItem = computed<any>(() => steps.value.find((_, index) => index + 1 === curStep.value) || {});
    const stepComMap = {
      1: 'NodePoolInfo',
      2: 'NodeConfig',
    };
    const nodePoolInfo = ref({});
    const handleNextStep = (data) => {
      nodePoolInfo.value = data;
      curStep.value = curStep.value + 1;
    };
    const handlePreStep = () => {
      curStep.value = curStep.value - 1;
    };

    const isLoading = ref(true);
    // 获取默认值
    const defaultValues = ref<any>(null);
    const schema = ref({});
    const handleGetCloudDefaultValues = async () => {
      const data = await $store.dispatch('clustermanager/resourceSchema', {
        $cloudID: curCluster.value.provider,
        $name: 'nodegroup',
      });
      schema.value = data?.schema || {};
    };

    // 获取详情
    const handleGetNodeGroupDetail = async () => {
      if (!isEdit.value) return;

      const data = await $store.dispatch('clustermanager/nodeGroupDetail', {
        $nodeGroupID: nodeGroupID.value,
      });
      return data;
    };
    onMounted(async () => {
      isLoading.value = true;
      await handleGetCloudDefaultValues();
      if (isEdit.value) {
        defaultValues.value = await handleGetNodeGroupDetail();
      } else {
        defaultValues.value = Schema.getSchemaDefaultValue(schema.value);
      }

      isLoading.value = false;
    });
    return {
      curCluster,
      isEdit,
      isLoading,
      schema,
      defaultValues,
      navList,
      steps,
      curStep,
      curStepItem,
      stepComMap,
      nodePoolInfo,
      handleNextStep,
      handlePreStep,
    };
  },
});
</script>
<style lang="postcss" scoped>
.node-pool-content {
    background: #fff;
    border: 1px solid #DDE4EB;
    border-radius: 2px;
    padding: 16px 32px 58px 32px;
    .title {
        font-size: 14px;
        font-weight: Bold;
        line-height: 22px;
    }
}
</style>
